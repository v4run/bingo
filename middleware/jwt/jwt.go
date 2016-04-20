package jwt

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"goji.io"
	"golang.org/x/net/context"
	"gopkg.in/square/go-jose.v1"
)

// Claims represents the claims for sso token
type Claims struct {
	Iss           string `json:"iss,omitempty"`
	Aud           string `json:"aud,omitempty"`
	IsTrusted     string `json:"isTrusted,omitempty"`
	Sub           string `json:"sub,omitempty"`
	Name          string `json:"name,omitempty"`
	FirstName     string `json:"firstName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	ProfileImage  string `json:"profileImage,omitempty"`
	EmailVerified string `json:"emailVerified,omitempty"`
	Email         string `json:"email,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	AtHash        string `json:"at_hash,omitempty"`
	Iat           int    `json:"iat,omitempty"`
	Exp           int    `json:"exp,omitempty"`
}

type tokenType int

const (
	jws tokenType = iota
	jwe
	absent
	invalid
)

const (
	// CLAIMS denotes the key to get the claims from context
	CLAIMS = "Claims"

	// TOKENERROR denotes the key to get the token error from context
	TOKENERROR = "TokenError"
)

// ErrInvalidToken indicates an invalid token
type ErrInvalidToken struct {
	cause error
}

func (e ErrInvalidToken) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("Invalid token. Cause: %s", e.cause.Error())
	}
	return fmt.Sprintf("Invalid token.")
}

// ErrTokenMissing indicates a missing token
var ErrTokenMissing = errors.New("Token missing.")

// ErrPrivateKey indicates a missing private key location
var ErrPrivateKey = errors.New("Private key location empty.")

// ErrPublicKey indicates a missing public key location
var ErrPublicKey = errors.New("Public key location empty.")

// ErrUnrecognizedTokenFormat indicates an invalid token format
var ErrUnrecognizedTokenFormat = errors.New("Unrecognized token format")

// ErrTokenExpired indicates an expired token
var ErrTokenExpired = errors.New("Token expired")

var (
	privateKey, publicKey interface{}
)

// Init initialises the jwt middlewares with privKeyFile, pubKeyFile and keyPass
func Init(privKeyFile, pubKeyFile io.Reader, keyPass string) error {
	if privKeyFile == nil {
		return ErrPrivateKey
	}

	if pubKeyFile == nil {
		return ErrPublicKey
	}

	// Load sso private key and public key
	key, err := ioutil.ReadAll(privKeyFile)
	if err != nil {
		return err
	}

	if keyPass != "" {
		block, _ := pem.Decode(key)
		key, err = x509.DecryptPEMBlock(block, []byte(keyPass))
		if err != nil {
			return err
		}
	}

	privateKey, err = jose.LoadPrivateKey(key)
	if err != nil {
		return err
	}

	key, err = ioutil.ReadAll(pubKeyFile)
	if err != nil {
		return err
	}

	publicKey, err = jose.LoadPublicKey(key)
	if err != nil {
		return err
	}
	return nil
}

// defaultErrorHandler is used to handle the request when "MustValidate" is used and
// an error occured.
type defaultErrorHandler struct {
	goji.Handler
}

func (h defaultErrorHandler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// Validate is a middleware for parsing and validating JSON Web Tokens
func Validate(h goji.Handler) goji.Handler {
	// passing h as error handler so that if an error occurs, it (h) is called
	// after setting the "TokenError" context variable.
	return generateHandler(h, h)
}

// MustValidate returns a middleware for parsing and validating JSON Web Tokens.
// When this middleware is used, the token must be present in the request and should
// be valid valid. If the token is invalid, the defaultErrorHandler or the errorHandler
// provided will be handling the request.
func MustValidate(errorHandler goji.Handler) func(goji.Handler) goji.Handler {
	return func(h goji.Handler) goji.Handler {
		if errorHandler == nil {
			return generateHandler(h, defaultErrorHandler{})
		}
		return generateHandler(h, errorHandler)
	}
}

func generateHandler(h, errorHandler goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		token, tokenType := parseFromRequest(r)
		switch tokenType {
		case jwe:
			claims, err := decryptJWEToken(token)
			if err != nil {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrInvalidToken{err}), w, r)
				return
			}
			var c Claims
			err = json.Unmarshal(claims, &c)
			if err != nil {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrInvalidToken{err}), w, r)
				return
			}
			if int64(c.Exp) < time.Now().UTC().Unix() {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrTokenExpired), w, r)
				return
			}
			h.ServeHTTPC(context.WithValue(ctx, CLAIMS, c), w, r)
		case jws:
			var c Claims
			claims, err := decodeJWSToken(token)
			if err != nil {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrInvalidToken{err}), w, r)
				return
			}
			err = json.Unmarshal(claims, &c)
			if err != nil {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrInvalidToken{err}), w, r)
				return
			}
			if int64(c.Exp) < time.Now().UTC().Unix() {
				errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrTokenExpired), w, r)
				return
			}
			h.ServeHTTPC(context.WithValue(ctx, CLAIMS, c), w, r)
		case invalid:
			errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrInvalidToken{ErrUnrecognizedTokenFormat}), w, r)
		case absent:
			errorHandler.ServeHTTPC(context.WithValue(ctx, TOKENERROR, ErrTokenMissing), w, r)
		}
	})
}

// parseFromRequest parses the request and gets the id token from header.
func parseFromRequest(req *http.Request) (string, tokenType) {
	if ah := req.Header.Get("Authorization"); ah != "" {
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			token := ah[7:]
			switch strings.Count(token, ".") {
			case 2:
				return token, jws
			case 4:
				return token, jwe
			default:
				return "", invalid
			}
		}
	}
	return "", absent
}

// decryptJWEToken parses a JWE token and returns the decrypted payload
func decryptJWEToken(token string) ([]byte, error) {
	e, err := jose.ParseEncrypted(token)
	if err != nil {
		return nil, err
	}
	payload, err := e.Decrypt(privateKey)
	if err != nil {
		return nil, err
	}
	return decodeJWSToken(string(payload))
}

// decodeJWSToken parses a JWS token and returns the decoded payload
func decodeJWSToken(token string) ([]byte, error) {
	s, err := jose.ParseSigned(token)
	if err != nil {
		return nil, err
	}
	d, err := s.Verify(publicKey)
	if err != nil {
		return nil, err
	}
	return d, nil
}
