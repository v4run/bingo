package jwt

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hifx/bingo/mux"
	"goji.io"
	"golang.org/x/net/context"
)

type keyPair struct {
	public     string
	private    string
	passPhrase string
}

var (
	keyP = keyPair{
		private: `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: AES-256-CBC,B787E65824CB36385F5FB3B853DDB351

rGKi3pNm0VMXP6uwY+U035ssI/JvMnOCoBUDdRdROT6sXClVcb9DRT8otLRUpqym
KyFsis6H1giubSwT/6aplxVzPau2nFCF7FBzf9m9ZUmCYPb8Rr/Vu9FAIGny4uvL
QVyrppsd8OKEhUGlHV6cUoGoGyigKvNe95pDXVIqMyeN6WZzT4sD989L1zQa0p3B
FfF8GrXE8h15iS9O+RgtoMLJqLylP95OBsbSJcrWXwi0dxxLSWT9Ns08dfejR9y3
0gF1ev3liIfl6RownwKC/iB9J21ZZjDetuXw6URUgmIUtrpTcv97fVWM6Vm4lr0W
pvL45AJmSnIWjvX+gZanMqLLN02q3BfBTTRkCGf3EalnxsdVAi2jV9At2J/Rjquq
y3GvNq6aS5a6ETgbjgtx9prnQoerzWxJhQoUj0Rl9E8qCzRKy5N/Uju7O86a7cZd
+SfWT3lQASfNhnfITwOWyr7y69vLbu4Q3BG1OQk99DPmdJemxFImzKCIqrsLXOCM
PqJKaydci46Gu6a3L1LHGcmH2P/4kzlzPCwOLIjj9ZolJ3TZ/SxokN474WbDeOkY
E/H3xLIplfed6flAzgV0xPZGJ/DD45LBGIy13uBX4JrQ2De/ecxbuKld2pOJPZRt
6xFDnv0kKStlPcCxzeTOrdjhT+t3Q3JIVIXYIbKp3vOrmOVeqQxFAQpiUAVP/7Jk
EVV1PJDx9eRe0VXoMckaRr5uNNqeFnrteyk08jPbGpG22kg1Co4CeRh5eEj95y+i
f+5to7gOtoT7YYrV+H7G14n4Wda1vg7boAHeA7WxxcF14usFsDuzdUZq0GkwBycB
bkUJPYRLS2L7cmDKKPK43kPn/f5shu/XUHDkF+P2HIak3hpIU//mw4edm2dGZmd7
n6XKsFDSK9WkAhOIN8kqBIjYoUSeUvaZVu0jYmdUqK4m43RfgxxBqko1b8DAwcPF
kXrXVf6i1lcRR1M7Pdej314aUHsn0G1UHQwA15ga1yQjeWTHCKw7l0Jvv0gT04OB
j3RUfvCt/8C782Y0uzAflpf4H0yhGN2vJOee37QQJ2Pb3hblkQjGp9h3LvjpCQZh
ek1uLZ8AETL6lNvj3cLIp+Q6mOftQ4nm4p7i2OseLJtq70jNDi+Imf3AVYGHX2q2
FprlDkUF5ptPS67oaAVD9J8rCE225EspD1lXKaBS0EtLSGwiPQ2lqPzY0rMBGKjS
IlglmJouwy2uSgmQQ6d5uus3qe3OCqgrSEdpM93/p+eub18fhINHp3GIboJM0G3v
zsrp9Hs05oIqMJmsJyn+Z0ASYZqxrjkikFOYIlyIRODG6UASgKUhnNkvg7kg28da
gg/t9Zm4lSf3yqHG4YyBFhNbqHtQgCr8ArjFD+bvGe8CNsHc0+oI1hnPw+8W6wPb
02B5ds9rHQYRRQF/OvOw0UponrxmDvkjMVEAP9rFL8U3J7Mu6N+FwSfjpSk6+9pz
BhFOHbRgqYMiSAhzmaQKkgqwG/yjW2Rf12W5UvCK8ptQpQ4Jaoq5XAn3SoHYy1qH
HnLVe7PD/7Svfb6k0obGwHVd3iOC4wmgCRZPLFdpO/6xoNcCrzUBOKbIZNDWCgwe
-----END RSA PRIVATE KEY-----
`,
		public: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4g2EMP/SgF0I68xOJU2F
CGfIrNhfTFDlCjav5aSFTNixD0SMM9xcjZ4kTL4Z8SvL8PCDtsbcAmhrGtd7eu1f
buAdC002LOsQQFHu63Y71eFFwUCS/KnDGriXWI4BihhFbp02LqTK4fPR9XOQb6LX
n6Bhbq/DEdeHPCd95SCr1sJVx1f5P/SFi6UmPwYDKmaVEN8TtH87nBC1VIySnSEx
amukSXd6byPbk0glEb9j4IK2NKhrLxvlAjYDYgumlUynSBfkDZFUMNygaoFcPzjS
uisSGv0lJZTFCVJk2gGXG9ILKmv/GTRxToENRVk0NbmH5x1rwWZjt2FKpvgPKZDq
SQIDAQAB
-----END PUBLIC KEY-----
`,
		passPhrase: `testpass`,
	}
	key = keyPair{
		private: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA4g2EMP/SgF0I68xOJU2FCGfIrNhfTFDlCjav5aSFTNixD0SM
M9xcjZ4kTL4Z8SvL8PCDtsbcAmhrGtd7eu1fbuAdC002LOsQQFHu63Y71eFFwUCS
/KnDGriXWI4BihhFbp02LqTK4fPR9XOQb6LXn6Bhbq/DEdeHPCd95SCr1sJVx1f5
P/SFi6UmPwYDKmaVEN8TtH87nBC1VIySnSExamukSXd6byPbk0glEb9j4IK2NKhr
LxvlAjYDYgumlUynSBfkDZFUMNygaoFcPzjSuisSGv0lJZTFCVJk2gGXG9ILKmv/
GTRxToENRVk0NbmH5x1rwWZjt2FKpvgPKZDqSQIDAQABAoIBAQCC8ml9KPR7v2kH
jxZFrZ4+vEAXQFAUGVhUjlFeqes+FNici4zcDe7fapiEjCri9gfxzqG+I3wXOP2y
UtkI4LDDvbeVcGjNpG2JlOzeIWOQBisuQ4XiL0UCGaQyfDCQGnc+GHvmkTelpGQf
14333VEi+vj18YMCtuN0CTx4mnBwupOGpabV68auR7mK6GkrKyrPUnznahJIKFQV
MrG03NfO5Um3eSPOItG/0GMQx8ZCmWvltHx4nOjzH8OoT3wpAwosMIhWULbVFsFX
hxAVSVn5OQ5CVai4CMhED7VFL0ivTJV1pyXwZaYoFnli0ov41f6Rx5FsqMtTZyPY
BB/ZkQMxAoGBAPxr6mWU2lnleMxsd2fXBXR88Wp3L/hIvbrH+eoCOJlKE+nnBujU
l13tkeojzvWvKsjlM6UfgqbLiq0nKaiQ54agwaSukkJ31POS6/B2o2/c9QGu/RhN
uxX6mm5VYM9lkeCTv3nEPS2wz5muxVK7wVCUs9d3VwfFkH5o4nX0yF7DAoGBAOVB
51l9m7em1NMu0XjbbbXURum4CxdKZ8zYwK6Kyd4YikVswEdTPd2SfiJSgIa424m7
dAD2Xp7NqKy7ToBLA/oBtHXdOPjsKPF/8igUQofFziQW2ORRg9qc8m5oHSqofGRL
qHpGrEoGGeF3lhLeNvWJV4NEkczg/q8QQfFeiBoDAoGAcdRwhZKUzQlQak9XoXoz
uY5GiA5rkXmsJbjcmIyb3XSsekR2tzR3diIWNRIk2GI/1wyVN5d4IaOUS/VnMd72
qZ2A9bTLvDGx1I2i3HODzIRF8JZrCDS1c3npfmv+FkjlefLm3BCEzj/3voQz89U7
ng0Q9M+abaTIPlkqFqtmWGUCgYEA33NVz+ba0KzeARw/9SFClJhbqc/Fl6Tg+UtG
upjx2vRmSPaPjrV2tjDjmgZ52VXyPROlJI79eKERR5KlF+yF6ragssS1lAFygrhn
SWM92WIV4x0Vt6wv7PNOZAg8bWidHZCUnOGnadr6fMT3VFqcjMOZtYsu5NdjxTP+
Ygj2dQsCgYBSZfc+VaIjQWSfi4bl/zBPtt1lfouQJUG5ZluNNJETEMtmWHW7O20Z
iVKXDgytuXhm5dGcK1xkHe5lMxerGRPeeOoasvWC2ppQTW2cWhYmiz6KNZuPK+BR
f6C3s2ai5sc5BRlDVl1d8YDi60BxQ457utpZIgViDfIlrtE9LBDOlw==
-----END RSA PRIVATE KEY-----
`,
		public: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4g2EMP/SgF0I68xOJU2F
CGfIrNhfTFDlCjav5aSFTNixD0SMM9xcjZ4kTL4Z8SvL8PCDtsbcAmhrGtd7eu1f
buAdC002LOsQQFHu63Y71eFFwUCS/KnDGriXWI4BihhFbp02LqTK4fPR9XOQb6LX
n6Bhbq/DEdeHPCd95SCr1sJVx1f5P/SFi6UmPwYDKmaVEN8TtH87nBC1VIySnSEx
amukSXd6byPbk0glEb9j4IK2NKhrLxvlAjYDYgumlUynSBfkDZFUMNygaoFcPzjS
uisSGv0lJZTFCVJk2gGXG9ILKmv/GTRxToENRVk0NbmH5x1rwWZjt2FKpvgPKZDq
SQIDAQAB
-----END PUBLIC KEY-----
`,
	}
)

var dummyClaim = Claims{
	Iss:           "Exp",
	Aud:           "Iat",
	IsTrusted:     "AtHash",
	Sub:           "Nonce",
	Name:          "Email",
	FirstName:     "EmailVerified",
	LastName:      "ProfileImage",
	ProfileImage:  "LastName",
	EmailVerified: "FirstName",
	Email:         "Name",
	Nonce:         "Sub",
	AtHash:        "IsTrusted",
	Iat:           12,
	Exp:           9999999999,
}

var expiredDummyClaim = Claims{
	Iss:           "Exp",
	Aud:           "Iat",
	IsTrusted:     "AtHash",
	Sub:           "Nonce",
	Name:          "Email",
	FirstName:     "EmailVerified",
	LastName:      "ProfileImage",
	ProfileImage:  "LastName",
	EmailVerified: "FirstName",
	Email:         "Name",
	Nonce:         "Sub",
	AtHash:        "IsTrusted",
	Iat:           12,
	Exp:           0,
}

type TestCase struct {
	err    interface{}
	claims interface{}
	check  func(interface{}) bool
}

var commonTestCases = map[string]TestCase{
	"Bearer abcd": {ErrInvalidToken{ErrUnrecognizedTokenFormat}, nil, func(err interface{}) bool {
		return err == ErrInvalidToken{ErrUnrecognizedTokenFormat}
	}},
	"Bearer a.b.c.d.e.f": {ErrInvalidToken{ErrUnrecognizedTokenFormat}, nil, func(err interface{}) bool {
		return err == ErrInvalidToken{ErrUnrecognizedTokenFormat}
	}},
	"Bearer a.b.c": {ErrInvalidToken{}, nil, func(err interface{}) bool {
		_, ok := err.(ErrInvalidToken)
		return ok
	}},
	"Bearer a.b.c.e.f": {ErrInvalidToken{}, nil, func(err interface{}) bool {
		_, ok := err.(ErrInvalidToken)
		return ok
	}},
}

var testCasesPassProtectedKey = map[string]TestCase{
	"Bearer eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiNGcyRU1QX1NnRjBJNjh4T0pVMkZDR2ZJck5oZlRGRGxDamF2NWFTRlROaXhEMFNNTTl4Y2paNGtUTDRaOFN2TDhQQ0R0c2JjQW1ockd0ZDdldTFmYnVBZEMwMDJMT3NRUUZIdTYzWTcxZUZGd1VDU19LbkRHcmlYV0k0QmloaEZicDAyTHFUSzRmUFI5WE9RYjZMWG42QmhicV9ERWRlSFBDZDk1U0NyMXNKVngxZjVQX1NGaTZVbVB3WURLbWFWRU44VHRIODduQkMxVkl5U25TRXhhbXVrU1hkNmJ5UGJrMGdsRWI5ajRJSzJOS2hyTHh2bEFqWURZZ3VtbFV5blNCZmtEWkZVTU55Z2FvRmNQempTdWlzU0d2MGxKWlRGQ1ZKazJnR1hHOUlMS212X0dUUnhUb0VOUlZrME5ibUg1eDFyd1daanQyRktwdmdQS1pEcVNRIiwiZSI6IkFRQUIifX0.eyJpc3MiOiJFeHAiLCJhdWQiOiJJYXQiLCJpc1RydXN0ZWQiOiJBdEhhc2giLCJzdWIiOiJOb25jZSIsIm5hbWUiOiJFbWFpbCIsImZpcnN0TmFtZSI6IkVtYWlsVmVyaWZpZWQiLCJsYXN0TmFtZSI6IlByb2ZpbGVJbWFnZSIsInByb2ZpbGVJbWFnZSI6Ikxhc3ROYW1lIiwiZW1haWxWZXJpZmllZCI6IkZpcnN0TmFtZSIsImVtYWlsIjoiTmFtZSIsIm5vbmNlIjoiU3ViIiwiYXRfaGFzaCI6IklzVHJ1c3RlZCIsImlhdCI6MTIsImV4cCI6OTk5OTk5OTk5OX0.VE6W0PzNgVTLinkcTEJksLJu3CLVhilIG1ZaoOhatBIE3-_VqCESBdWuuHupSs769oblUOCQUC4VYmUJGvpzsBLEHjp_xADPQ7Tvaqh0TPhAVtqBz7_yYbqTrkbxqtGjM_Rf0xgs7b2cNnMPPtdhgZ6F_M9aWBKT-5cBr8q52cH-9a4rQGypBWbaMPm-ZGoE4x-TgOMY1-A66JDNCvB71kbRS-R-Sypa0iLspx06dWJEBt-dI3Synfht4xbYAqlTGYVcPs2-Kg9J2lNgYfsbPg5HRSQiOcJ3sbp0eeT6nkmtgkfAslKcxz0p78pBQ0ichvbzQRgbkYv5bga9eEVkSw": {nil, dummyClaim, func(err interface{}) bool {
		return err == nil
	}},
	"Bearer eyJhbGciOiJSU0ExXzUiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0.otiTebhb8bjOtb26nicPfNgtDpDQaFE6cABCtCARnbLD5r-HIrNePTnXHMkJdkeUGmowTk_HtPzi9pGtx6K2GVjPNsgPB3j3bp1nr8oMRbRI71y1BDVH6KVX0qPN9h432ynLoIZXL9wThawx5sMYF29Zbcmyhlr8z7Dop_F_-iZgvuqZDTvhWsYA_CjnRd_vFlFvTcrJuud9pPBna1xdy6Mhpg9CiluTG95VKQ1pP_SWApiJpuRjS0lNO90lhma4XvfxwiUmDNiSCEnCPQ-W3G1vn2MoL-559rb5YGEwWCv17CBhHYQwZW7_zV7fCqxd5jZTFpy21NkA0N7xEAJ5bQ.b1UP1e7oXZNTnKZ4OSqfxQ.rFNSGRUTePtDBOaQdXMhRXNmHO8OTuzdk-CCaATNsJ_ct0_9WWtrmTmfjoJAt1jEkQj5bVU5MjzWgQT-N4khlfSzfMj2SYqJd7peek16H_yICrcJkAcwqojJQSt399HqqHlRaed-b3FXymk0H8vX7Ee81URygIDP8KHpp0_oZFZzlClKp3EgRRjjqo-x8SmXpV2Cim_NbdQME95J-KkbrUOQQv1mYtOquppDFtHtX6qv5jpAobbK18iupk1JqOcb_M7O94kWYuOCrhKf8XebLZs5O2lcxd5MM-c2AQvvSUNsQ8VuYqyt-bcKfNFhZ5nDIAk5z1VCRE0CP5w44QP6rq9r80Dyg07ca-qGUm2WilZR12M5EyVzyCqYX5Lmmzgpg2Tk0YA5l9Li2Ekuu-oMHYrYfxUuSHJ3sCfqbDJbMgT8lEpei1gNzRBbqCqE08XnR8YcesvIo4QobAt_0zbxa6LF07i2Yo5-OGZeE9E7MjouMPPQyz8nVjpgNEkA6aWGuACixskL1l3FR_NLskStSjZ-XbsnBHHgj-T_V_LP4gLNz1ojlhLPMkKTpKJ0TaurIsd1fpx5uWTO8fFs-jAvYG0LuGku6bIEqd9lf8aujdvi1uaFeT2QkpotHKinA_423MMlOaYxwjt2W7ij02i2-BzNCfzLfvucAPW4QCmi56wnsXdp6uZFxTwdIquKDYtzjd-e8ix8_9CX-MP4USsxPC-ZDqASZB0f94Yx57jI3fUbrzlgXLyFZCBoZxRZ8qCsSDA5gPNGf4j0rZn8NYxju4GoVDZQ4ThMcrxhzirp0TJbS8HbOtyREQhs0s7Gk6bkJTWIhqUetP8S49C7KvT0EXBdE-apdCPMBgLdQjGelQEvwXySSFUhLMQmMWtB2aq3fh9_R8F_J-tVq6buwysOj43bjQLv6QxyMP6jVdk6HCUKQS8-i_PsfSxuNZ1d6ljIGUtGND83qSh3mWv1WWePhLxGIkkplBV5oz4uZy6r_QciuI9yAxw-tixHMhf3TBxBh8G9Bku-Wt6j84t0oe_Jli8JdzvvZ3smgkZ4BxfDl9ARbQsY_-od3k0-SdiVhmDKs_DweWXWTkadZsel56c4jr02s3wfVCzfUaxDI-AkAjLUxlYo7XujuB75HWrYr2o4KaRY8q6AYFtZ3p9wDm_8PSPQk62TWMDN0Qvv9aqDGARZvW9WlYwYEnSFbijYKaROfM02CK1JEC6nNZwVj3WwAW25eb0KSCSnAX9BD5GDkVFn7UsZPFWy4JqdGLVoy4l5beZKyQnECrcgPrLPz12mqcKPKESi_fV4GUSB3AALFaFVhf8uDwZQMW2OL-vspQEsoA8VmTT9WucGRaEMUG3aIwkFHAi81uFqxiWgIWM4n2doUdXssuCP_f43VS5tnTvDSBXW7eDF--dsMGUbXWFigdopZGasQ_csYBXgHLYK0NxuZiCwWg3YPjLGPnA53EyDRPNPhl8c7i0ig9xEizTtRV_3kqkqPyqzIHiI_wbTsfVSUPfO9p2hPd8EF0fGMLHGmNat4uns6kHS1I_C2eL3rF-qIhGISgUSzwxvNmJr31_N7AwOrqKIiMBECshGFUGKU3AQ6ggKGq5N0cTw9f3HHUnhxEan0nMRy_9c2d2wDIE.bgB2LncONHAwrrLad2RP1A": {nil, dummyClaim, func(err interface{}) bool {
		return err == nil
	}},

	// Expired tokens
	"Bearer eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiNGcyRU1QX1NnRjBJNjh4T0pVMkZDR2ZJck5oZlRGRGxDamF2NWFTRlROaXhEMFNNTTl4Y2paNGtUTDRaOFN2TDhQQ0R0c2JjQW1ockd0ZDdldTFmYnVBZEMwMDJMT3NRUUZIdTYzWTcxZUZGd1VDU19LbkRHcmlYV0k0QmloaEZicDAyTHFUSzRmUFI5WE9RYjZMWG42QmhicV9ERWRlSFBDZDk1U0NyMXNKVngxZjVQX1NGaTZVbVB3WURLbWFWRU44VHRIODduQkMxVkl5U25TRXhhbXVrU1hkNmJ5UGJrMGdsRWI5ajRJSzJOS2hyTHh2bEFqWURZZ3VtbFV5blNCZmtEWkZVTU55Z2FvRmNQempTdWlzU0d2MGxKWlRGQ1ZKazJnR1hHOUlMS212X0dUUnhUb0VOUlZrME5ibUg1eDFyd1daanQyRktwdmdQS1pEcVNRIiwiZSI6IkFRQUIifX0.eyJpc3MiOiJFeHAiLCJhdWQiOiJJYXQiLCJpc1RydXN0ZWQiOiJBdEhhc2giLCJzdWIiOiJOb25jZSIsIm5hbWUiOiJFbWFpbCIsImZpcnN0TmFtZSI6IkVtYWlsVmVyaWZpZWQiLCJsYXN0TmFtZSI6IlByb2ZpbGVJbWFnZSIsInByb2ZpbGVJbWFnZSI6Ikxhc3ROYW1lIiwiZW1haWxWZXJpZmllZCI6IkZpcnN0TmFtZSIsImVtYWlsIjoiTmFtZSIsIm5vbmNlIjoiU3ViIiwiYXRfaGFzaCI6IklzVHJ1c3RlZCIsImlhdCI6MTJ9.Wq0RmpatrM9G7E5X7qTQa4Mca9XF1t-gwDucN_W1JVD8CX1TF8M-cQ_0syDLcO96-JI_RyW54SxYzNJA7rTaAbLBAvizwdl7-cvxD4wIwzUYXS_cY8It1YASORgyGkLYmjRHzlccDF9JbCeJisR_YgXStIY1mKH_Blh-qX0Q8CkEzZ0o9w7S6n19BZ1Fps8jjXTqocChhSV0O6r6JLtuy4gIITVCuVqukbG_lfeBFk2nedd02kyfcvRnrv5AAEJcH3KCxeQOnWnzUj2zL9oNiYxSe7Jet1NYhMFt1yC7VNYZvh25Uc2oVWIEiCfPwoBsXJQ7apjARCoW1VtlEIBCpw": {ErrTokenExpired, nil, func(err interface{}) bool {
		return err == ErrTokenExpired
	}},
	"Bearer eyJhbGciOiJSU0ExXzUiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0.MhWJqkiJ_g9nj2WA82ACBcOihNFS5RgjNnbQjCUHuXeoWnWetj_lf0Dc5JnlCNjrJmxljax5cAf5QJhSDpnDn1gUxx2r_366Ez_gVCmVw1xJ6ke9-2TlZSapW_sdeAfgoScoOhvFOuHug5U9O7lbh6bTAJY-H41BVqXqGchTtQEwiwvpf7klWQttMGsTlV3XCUgO7x-8kf7QC7U8ghITYcMB_p4_IFM5z1H0I0MfzlqmjQLPjK0U42dZ7djIQFDH9MB_RxAkCRhPp0c0oXqIR9yRWUIrewEOgkjJMC2nA_C9dgSs4NN0HWsyFhqcEta0fHUcd3oiNC-GoeYkjzXjWQ.T2IjpPjYlviv1BrH2_0I-A.s37dFf4YoFJrnOWr0vKVoYTnFzD8L79cFe3C2nATIswqF2vbfXBSZFjIdA7Xlv485zWHM4yXui3WJ-i_eBsAtVTnkH0Gnhz5lgVnoqPjNXzfved6M4N2jeJnGd6-qldK3Yaebr8OuaCYAIHSTgoqelINFwgdnqbeRhNs5EsSim58HW2s_U9_A4pOY9AoJO8UuQg3YSi-5qu9VWv4x_ic6XXFZHiH_33mIpIkGTEg2R4Kf3SWPP5zPiXyFQOWgk6BTSKSRgVqfq5lcoAY09OrIgvRNtIDj8ABWVREfviVqU2XwiS4QwVIq39o-fxjCRgoLoMvYXlN4vCqlc1JOxOhQF04YdMNeDeybmNM5ls2CYt5JFwFXBdYUgCIwPggwodS8NZauQmlDN8pOXQvQkY6z5xl5NcKd0mlriWTHDZRMNHrHJaYz5NGu08QS4qEm7T1uaWCklM0AqQ7f6-gNft3VAvMvKUQbCOwjp_3xKgbNhrmLWgMLiAlUpchomR7bRonSBOfBspDbkL2u256_ggbPCmKhitgu_tno66GWslQWYUPxtYYOCYSiZgml8XgoV21snvSdmgdNyK-6RGU3JTOKPHDExGVnMuaCa3GOR9jC3DeZ5lwZFePh8lQAx8igE2vNRKrwKzNXZsUz8DNksePbsMX89HnX9Fm7fD9dGdSIgkBw40a9VhfxsFePflpe99_1huWfMBhpzKAdug0Wfbg8yp9uzfRbKKRQ68-hVDUe5Otc192TToi58Kqzf4JW7_Sshp2R7FtmR-0WJkPTsHOB5_gDdviKlICi2D6R9M15JRqhFmSEVK7efZAWrxBZh0MKIjhrXz4qxGkR6U3z-20dvWHpg8MfGjpVhrrkbVuFt2zHJ9GbLhbOj7K8Qyih1qiX2t6aHZluWpxsVE670_FsOXytuciyNnjeO7ib-EaDoK0Te6KgPHr4GJi3lCEDTtOTYKRgYq6NePli3mD5W6CD80nXWCjXGbzr7ufcaZtfRRfmifZ7YM52tDrwl3zepQUgUp9mVJaciK-0lQHwJC1PgXQ9OTbLwsUy_ApevskMIfhVoVK7ggMWZVnS4H1stqwTAJvDTPls46UZYgLlGQwLwVUZhYovIIvSBLV1bazzMk82c3DPNKgyxLYxAF628Mf87d_M67TS-2949bJgEbRqiGET50AFn3o4mWQJSXmcHavoMjwMusSe6tBoUgsfEWaxcl5nl12hut40N00eWQCQ49zQY3PBhJ7QYHhdSPNzDSmLbyJ5VKnbH5EklYezKBUak5OJ1RWCBrR6EbN56W9-P-tmo3UVj_6xbQBaZ1072lvFi8f3KFjDrAykR4W9LL-seP5FtuCTmjl19UHIk9NwIhzv4LTLzhyqEfPrRni_lY2aX4xvjRASBMNlah2Eig0O8veOQuQse650ZPKNTJyyoEBuim_HUY7byckq_rl5ty0t0DWt5oxf8PNN8kQ_Xx67AUrgwi4RUOBb3aGJxiqpTabRM3JXJBZ8R6cIlRo0oaRuEh_BtH-O_jHFYkrS20MfbVwq0ytJ82TMAyao4yPOuWHj9RTTOKE2B25-FY85czQyCLq6GIzKYqS97ngGtqV.EnhBw__xCZZ58ZdF2mB35w": {ErrTokenExpired, nil, func(err interface{}) bool {
		return err == ErrTokenExpired
	}},
}

var testCasesNoPassKey = map[string]TestCase{
	"Bearer eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiNGcyRU1QX1NnRjBJNjh4T0pVMkZDR2ZJck5oZlRGRGxDamF2NWFTRlROaXhEMFNNTTl4Y2paNGtUTDRaOFN2TDhQQ0R0c2JjQW1ockd0ZDdldTFmYnVBZEMwMDJMT3NRUUZIdTYzWTcxZUZGd1VDU19LbkRHcmlYV0k0QmloaEZicDAyTHFUSzRmUFI5WE9RYjZMWG42QmhicV9ERWRlSFBDZDk1U0NyMXNKVngxZjVQX1NGaTZVbVB3WURLbWFWRU44VHRIODduQkMxVkl5U25TRXhhbXVrU1hkNmJ5UGJrMGdsRWI5ajRJSzJOS2hyTHh2bEFqWURZZ3VtbFV5blNCZmtEWkZVTU55Z2FvRmNQempTdWlzU0d2MGxKWlRGQ1ZKazJnR1hHOUlMS212X0dUUnhUb0VOUlZrME5ibUg1eDFyd1daanQyRktwdmdQS1pEcVNRIiwiZSI6IkFRQUIifX0.eyJpc3MiOiJFeHAiLCJhdWQiOiJJYXQiLCJpc1RydXN0ZWQiOiJBdEhhc2giLCJzdWIiOiJOb25jZSIsIm5hbWUiOiJFbWFpbCIsImZpcnN0TmFtZSI6IkVtYWlsVmVyaWZpZWQiLCJsYXN0TmFtZSI6IlByb2ZpbGVJbWFnZSIsInByb2ZpbGVJbWFnZSI6Ikxhc3ROYW1lIiwiZW1haWxWZXJpZmllZCI6IkZpcnN0TmFtZSIsImVtYWlsIjoiTmFtZSIsIm5vbmNlIjoiU3ViIiwiYXRfaGFzaCI6IklzVHJ1c3RlZCIsImlhdCI6MTIsImV4cCI6OTk5OTk5OTk5OX0.VE6W0PzNgVTLinkcTEJksLJu3CLVhilIG1ZaoOhatBIE3-_VqCESBdWuuHupSs769oblUOCQUC4VYmUJGvpzsBLEHjp_xADPQ7Tvaqh0TPhAVtqBz7_yYbqTrkbxqtGjM_Rf0xgs7b2cNnMPPtdhgZ6F_M9aWBKT-5cBr8q52cH-9a4rQGypBWbaMPm-ZGoE4x-TgOMY1-A66JDNCvB71kbRS-R-Sypa0iLspx06dWJEBt-dI3Synfht4xbYAqlTGYVcPs2-Kg9J2lNgYfsbPg5HRSQiOcJ3sbp0eeT6nkmtgkfAslKcxz0p78pBQ0ichvbzQRgbkYv5bga9eEVkSw": {nil, dummyClaim, func(err interface{}) bool {
		return err == nil
	}},
	"Bearer eyJhbGciOiJSU0ExXzUiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0.XmnA3ApGLjnex2g5Fckx-pfFJBrc3F_S86xfmbzaSy8k1gHEvmGWJ40BeGYlCOR9n9jImhZ1uXYPRmkjB9AhKj6lQ1uStwigun-1eVTeWLj4LsMlZ_M7a2eTt28aRCMOk4zMLcQ4dkNLoP-HyhO8xmmr8tosLwEKnoipu9RkWRy9lquDOkM8u3JOh3Gd1pG7d2KxFxcc0X-QBHRpFDAjHADK9GyNZWh07HQ-m1ZgIGsKhD5Nvw3WrQFn7gr8FJ_S5k2YMJFORtavyy2cPG8yjZ799IQbTT8qwbXy4xz9F8dYjPui6cYtlt43-YBPuemJjVcnRtU_aOl_px4tC4w4Ag.Hi-HTfmaoTP1O1SZEy9NYA.-wiQMMpb5gPy4Bmn4qHdkT-oQSFfAt9lxGCYDQ7IKS-Jt63TRMy9YVP1sR_NWmdaIDUcgCph2E1BuVEwTZMsQUlmch7XClao_zkJI6Dw8DyxdjMBIFdYitmO9WLh9I6jJCOtfBjiqTcb_bFHYV1RTDpJpVohogIqKOlx8_WK2HtINqX0QL_1Aff0izTqmsJOm_bJ4fWghUbBzeqcDT4LblHIwQcSJYq2gS1XicVvG4hxJBvOXtJhN0xdUdz1el-G-P2Sr829tnpqorRTa76ghWxVB9ayNzLZYNoWtt76uipA8_2XpuXCjlXTMalNPao1B7MFW8JtqbTgdhytZCnKThORIMwtLkTLD9wtzGZ1qwwLajJ6xsaz2my0xUAXMhjxeT_UxHdEmj7nKauqfAxMnUz91p9Xh9jwNP26GMuXlIwOmgODmLVt8Cvx47KPbrKK4mXwBVhOChY8OMGesTevKfAzWN9NhUyolAtW_jVz7mGWD2WVXN3pmzR8ARLjaJY6Zt1yyhaEJG0E8_1v4sm-hJtBer3TP0zYCUFN47cBrtW4N-CzKcoyZxMKF0ewG02FzF4kZkuuy7eVmTI_vX7OcCBs20i88NbttHUCfE2_t57ryhSlCGOqyiIFxyJeePNcJvdE27KW0NU9zELltYJgvPrZZRcrmayshIZZOQubvFKAQjRMCWl5KrHYk-_QAgzSKk9v7yPHkF0vzJYgtHn5zc26cXbxrYPY7wCnhsh_0IWCJVd-HSbpAQBCojLWTlczLz-UV4pLxWyjYYtL2Q4Qpui5AJhj4OOdXErmmu8s2wUyxC26UEMtL1IgHNAZ0dTOQXIiujy-qBIMHHqLju481o653-JY4QpZJ45NChntG9tbnyZxSR1AltUo9r6NS_z_yMM217dlIyWP04VKZNA0a0S3oaluWInmFSGOYvXB2Xyfx_oyApeWT7RB6hBBrUNbBRZxvkeVhRaD6u28FXSGcWvpKWjpkvmCiZboP2w4PZftGaO_oYWBXITojHL5CJ8g4joG0WykrkBX_zlrcP9Jme18Btsfp_yikeCKn038z4Msn_P35F8Cd-F_I_e6ZRiIIFj4dr2LcAQp9CPzYe_Ag94Fq3_IzKieaPsm5OkzhToXlb7imkE65IlfrWAXdr2bFQ1iweWV3KP7dbLKpZiZtqG_hnRAhrBsVWhT1F4__4UCma40Y9fwiwobrFxV8D9YGCIfK5IoDXTNP34or_UnQtUl4chJ7a0WlF0EIvLvv-kTqefRZN7-VIb258GLQ92E3RSkbBZAFo4BR1uGBi85gd81LJA5fcNL_lB-BCPRIoWDBuQJbsfRnhetkljlNe0FD8TXIzg7x_nkDpUXM_DeSPiGgaLcIwiAT0zdz33vWUhNVBDXGp3H2qPqCM_EhLLWjRPhR8p8CUZMmrIDON5RenN_oq7VPJ1YssIEAVSDeMWZhH6vHmHU2sWyhkUcmyuqIf4LMJ6CIBGRkRMAuWigQiKqH9Vu3ladN957c7o4cj0B3K-JXQwX7JBGQ8tZM_3rolCq0vftsuQ9_R6rJPdLYW-WmoyiDOn6MD-JtdgSXkdDxMTz5zPwp06uSPvzZFwbHjICVQlCw3C6X5OUQ5bD075fvkjdFHbtkctAb680mm8.y7RTO49uoR4eX6IrAF_U1A": {nil, dummyClaim, func(err interface{}) bool {
		return err == nil
	}},

	// Expired tokens
	"Bearer eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoiNGcyRU1QX1NnRjBJNjh4T0pVMkZDR2ZJck5oZlRGRGxDamF2NWFTRlROaXhEMFNNTTl4Y2paNGtUTDRaOFN2TDhQQ0R0c2JjQW1ockd0ZDdldTFmYnVBZEMwMDJMT3NRUUZIdTYzWTcxZUZGd1VDU19LbkRHcmlYV0k0QmloaEZicDAyTHFUSzRmUFI5WE9RYjZMWG42QmhicV9ERWRlSFBDZDk1U0NyMXNKVngxZjVQX1NGaTZVbVB3WURLbWFWRU44VHRIODduQkMxVkl5U25TRXhhbXVrU1hkNmJ5UGJrMGdsRWI5ajRJSzJOS2hyTHh2bEFqWURZZ3VtbFV5blNCZmtEWkZVTU55Z2FvRmNQempTdWlzU0d2MGxKWlRGQ1ZKazJnR1hHOUlMS212X0dUUnhUb0VOUlZrME5ibUg1eDFyd1daanQyRktwdmdQS1pEcVNRIiwiZSI6IkFRQUIifX0.eyJpc3MiOiJFeHAiLCJhdWQiOiJJYXQiLCJpc1RydXN0ZWQiOiJBdEhhc2giLCJzdWIiOiJOb25jZSIsIm5hbWUiOiJFbWFpbCIsImZpcnN0TmFtZSI6IkVtYWlsVmVyaWZpZWQiLCJsYXN0TmFtZSI6IlByb2ZpbGVJbWFnZSIsInByb2ZpbGVJbWFnZSI6Ikxhc3ROYW1lIiwiZW1haWxWZXJpZmllZCI6IkZpcnN0TmFtZSIsImVtYWlsIjoiTmFtZSIsIm5vbmNlIjoiU3ViIiwiYXRfaGFzaCI6IklzVHJ1c3RlZCIsImlhdCI6MTJ9.Wq0RmpatrM9G7E5X7qTQa4Mca9XF1t-gwDucN_W1JVD8CX1TF8M-cQ_0syDLcO96-JI_RyW54SxYzNJA7rTaAbLBAvizwdl7-cvxD4wIwzUYXS_cY8It1YASORgyGkLYmjRHzlccDF9JbCeJisR_YgXStIY1mKH_Blh-qX0Q8CkEzZ0o9w7S6n19BZ1Fps8jjXTqocChhSV0O6r6JLtuy4gIITVCuVqukbG_lfeBFk2nedd02kyfcvRnrv5AAEJcH3KCxeQOnWnzUj2zL9oNiYxSe7Jet1NYhMFt1yC7VNYZvh25Uc2oVWIEiCfPwoBsXJQ7apjARCoW1VtlEIBCpw": {ErrTokenExpired, nil, func(err interface{}) bool {
		return err == ErrTokenExpired
	}},
	"Bearer eyJhbGciOiJSU0ExXzUiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0.jxKKvEOIdvv6cUjPkvZ4LlwWqcwQRDv1Op8PGlqqvY9yB0Q8zROJT0WsSVgkOIe7fZObl-0p_H8DY78tYBiYwWgXGUCDkieC71VGBKqQ4dWgCDMzuQ6NcWqp_LbNPSOCDai5nO5DTXm7kVMls76YT3rWwNIBeGyr0JtyUuV05tlGd-OZQIVrQ_4-f9N7CacSAPwSSwIJ7Bi5pRshs6KAyRnAD1SoVPcOMI75WJGEhoXWePJFhunkE8i4wz4zXPfzYbtrMo9a1UStD-rWDaJ__H6vkYP3sA57GqTx4cUTJRygfscyzGRGlyjAq7yiOAh7SnfQv0rn5bJ7cJ0LvIqEqQ.KqgmJpwItUQ-Ncq98RUucw.ZTDVcNT5d4p0K2DGO4Ofssnts6Aupv58ZSXFyjaz_uP00Sc0uC7i_EoIXfJhYOyVWacnPJhdn3bmtzxECuVSxP8sKTRvYWDtCLkFOZjBQQMuPHRxE5tBrwNN_SHPpq2Y81067sBiLb3blshCxfsmsMgkcKd-Bv1HJa11wwhmvqkClRyauOwtRlctYuZEkdwdWfzkhYD4xAblKWofsdjyWp9536k8cmv8Kwv8oxhvehbDmqiMLmDmfS3nNOE1Bri_vOzd8RHivBhgppXMdxH7CX8RtIC-0WhfqoBa53k4XKrDJaQGMaxB4CRM3BHBruH70w1ef-qylEQGGfALuEttCEr8Zok2sB2shKF03nduPIMVv_L3q56QAQKqP09tHg5ZeAg7ZU2XsTXTTq6wQRDJA5HSOoiZmX8Vx5ppFRLpLvVJHVRFXhmgoLdix-WT3JJeUI75cOKr6YlapNe5B2mvKgHk_okgsIOt2RWHweDtfbL8eSQPD_KXUxOcNDjDJ9ENIX-3bQloeENHGp2XZOSwSo7I4UCVmJXW0oYn2L_NYdaQgRlnWTkOxyKvzSPS34CoFoNn16ZlqwCBm48Fu2DWfvI6FROm6AgeRxsHbHTdAVAGpWUfj_12T7uJADQWMOBa4B1uL1EopQsAOzEwB7ZOeF8FuTctew9qoZKdrUWGerFrVs78z4cD_k7BlRQDbni8bnAicS9jCmDrRprufR3qluRlpNwSzWAFAMSUQ9-TB-y2hP4po0AYHrDvM59uiwfOWX6zMmWgyJ8iyWGKXZAGIPDA44FuvM2ie7Is7NjNThp2Gj4MMcY-pdJqhsTlQjxRBr9uKxayaKPNzVBPGdKyAKHTLrcVVL-HG2sINlgfZI_KfnH-5ezAXaBBNU5bLh1xL2TQpY7btPizOHpCYGZJAdeeXhH193FUHacq6Z77g7z_SgYHwkVgZr5hcWZU_bwD4gwvbMmcQO1dq3UMJG8GltwQxkAKktfjrekn9j3UXB4xqGD8wpWmofwywo3hXMMxAZ7IjMi3ciWldm9ReEWSTqU2Co7Qaay2_PWJkE2G168HgPjUr9hKEIXUwQ4ngTSopC71_eAdYysgmVRszKoIKYbm3fJ3mWskfscQP4K50oBrxgnNLyJYYXHTpCBl-9AIGX-mzPCQxkOb0DIzrTh3HoiMfuXjOnxkMEkdw66JWnV1KQ8CBhv0Ou252N41ey2Vnd2kTA-TsoTikyf9uIlMcUaqqpIcZw7AZ0DUjfNwvzUUAw1RHQf3wgsWjpyqRPWt4E3KzZLC4OMYfchylzpEkMhb7Yktt9xk1zt_lrd-ln7ZCwrYTg6egkYFVzwLk9DpQX8PIoGY4DTQaubRQeqTf6IyFE9kT6EkZu6NlbFHBt9Zp5aC6LV-mNKm8rj2IZTfX2xbzT6oK66oTDAiT86b7nNBCw_V2zjFuvrXTuAQzwNS-XWPgbeV0JQH6w0YK97HAd9itKi0Rshfsx7p-V4JSw_EIYEp83bL2K_FPreEg0a8yMXK9CMdnpzQT8EdNxGULUsIhD-5nIW3NOogniHmR3gjEYEtfRrTlb51nOj_bZuFci1OG5qYRudYeqBf2XWB.lVB0ZtOMLbL3gtYxOllHuw": {ErrTokenExpired, nil, func(err interface{}) bool {
		return err == ErrTokenExpired
	}},
}

func TestInit(t *testing.T) {
	err := Init(nil, nil, "")
	if err == nil {
		t.Error("Error: Expected:", ErrPrivateKey, "Got:", err)
	}
	err = Init(strings.NewReader(""), nil, "")
	if err == nil {
		t.Error("Error: Expected:", ErrPublicKey, "Got:", err)
	}
	err = Init(strings.NewReader(""), strings.NewReader(""), "")
	if err == nil {
		t.Error("Error: Expected: non-nil error Got:", err)
	}
	err = Init(strings.NewReader(keyP.private), strings.NewReader(keyP.public), keyP.passPhrase)
	if err != nil {
		t.Error("Error: Expected:", nil, "Got:", err)
	}
	err = Init(strings.NewReader(keyP.private), strings.NewReader(keyP.public), "")
	if err == nil {
		t.Error("Error: Expected: non-nil error Got:", err)
	}
	err = Init(strings.NewReader(key.private), strings.NewReader(key.public), key.passPhrase)
	if err != nil {
		t.Error("Error: Expected:", nil, "Got:", err)
	}
}

func TestValidate(t *testing.T) {
	var tokenTest = func(testCases map[string]TestCase) {
		m := mux.New()
		m.UseC(Validate)
		m.Get("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if testCases[r.Header.Get("Authorization")].claims != ctx.Value(CLAIMS) {
				t.Error("Claims: Expected:", testCases[r.Header.Get("Authorization")].claims, "Got:", ctx.Value(CLAIMS))
			}
			if !testCases[r.Header.Get("Authorization")].check(ctx.Value(TOKENERROR)) {
				t.Error(r.Header.Get("Authorization"), "Error: Expected:", testCases[r.Header.Get("Authorization")].err, "Got:", ctx.Value(TOKENERROR))
			}
			return nil
		})
		for i := range testCases {
			m.ServeHTTPC(
				dr(
					map[string][]string{
						"Authorization": []string{i},
					},
				),
			)
		}
	}
	Init(strings.NewReader(keyP.private), strings.NewReader(keyP.public), keyP.passPhrase)
	tokenTest(commonTestCases)
	tokenTest(testCasesPassProtectedKey)
	Init(strings.NewReader(key.private), strings.NewReader(key.public), key.passPhrase)
	tokenTest(commonTestCases)
	tokenTest(testCasesNoPassKey)
}

func TestMustValidate(t *testing.T) {
	var tokenTest = func(testCases map[string]TestCase) {
		m := mux.New()
		m.UseC(MustValidate(goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if ctx.Value(CLAIMS) != nil {
				t.Error("Claims: Expected:", nil, "Got:", ctx.Value(CLAIMS))
			}
			if !testCases[r.Header.Get("Authorization")].check(ctx.Value(TOKENERROR)) {
				t.Error("Error: Expected:", testCases[r.Header.Get("Authorization")].err, "Got:", ctx.Value(TOKENERROR))
			}
		})))
		m.Get("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if testCases[r.Header.Get("Authorization")].claims != ctx.Value(CLAIMS) {
				t.Error("Claims: Expected:", testCases[r.Header.Get("Authorization")].claims, "Got:", ctx.Value(CLAIMS))
			}
			if ctx.Value(TOKENERROR) != nil {
				t.Error("Error: Expected:", nil, "Got:", ctx.Value(TOKENERROR))
			}
			return nil
		})
		for i := range testCases {
			m.ServeHTTPC(
				dr(
					map[string][]string{
						"Authorization": []string{i},
					},
				),
			)
		}
	}
	Init(strings.NewReader(keyP.private), strings.NewReader(keyP.public), keyP.passPhrase)
	tokenTest(commonTestCases)
	tokenTest(testCasesPassProtectedKey)
	Init(strings.NewReader(key.private), strings.NewReader(key.public), key.passPhrase)
	tokenTest(commonTestCases)
	tokenTest(testCasesNoPassKey)
}

func dr(headers http.Header) (context.Context, *httptest.ResponseRecorder, *http.Request) {
	c := context.Background()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	for k, v := range headers {
		r.Header.Set(k, v[0])
	}
	if err != nil {
		panic(err)
	}
	return c, w, r
}
