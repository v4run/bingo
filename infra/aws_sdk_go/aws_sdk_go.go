/*
Package aws_sns provides the library to communicate to sns service
*/
package aws_sdk_go

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func ConfigAws(regionName, accessKey, secretKey string, doIam bool) (*aws.Config, error) {

	awsConf := &aws.Config{Region: aws.String(regionName)}

	if !doIam {
		if len(accessKey) == 0 || len(secretKey) == 0 {
			return nil,fmt.Errorf("ses access key and/or secret key is/are blank")
		} else {
			token := ""
			creds := credentials.NewStaticCredentials(accessKey, secretKey, token)
			_, err := creds.Get()
			if err != nil {
				return nil,err
			}
			awsConf = &aws.Config{
				Region: aws.String(regionName),
				Credentials: creds}
		}
	}
	return awsConf,nil
}

func ConnectSns(config *aws.Config) (*sns.SNS, error) {
	awsSession := session.New(config)
	snsService := sns.New(awsSession)
	return snsService,nil
}
