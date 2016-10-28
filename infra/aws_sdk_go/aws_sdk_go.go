/*
Package aws_sns provides the library to communicate to sns service
*/
package aws_sdk_go

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
)

func ConfigAws(regionName, accessKey, secretKey string, doIam bool) (*aws.Config, error) {
	sess:= session.New()
	awsConfig := aws.NewConfig().WithCredentials(credentials.NewChainCredentials([]credentials.Provider{
		&ec2rolecreds.EC2RoleProvider{
			Client: ec2metadata.New(sess),
		},
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
				SessionToken:    "",
			},
		},
	})).WithRegion(regionName)
	return awsConfig,nil
}

func ConnectSns(config *aws.Config) (*sns.SNS, error) {
	awsSession := session.New(config)
	snsService := sns.New(awsSession)
	return snsService,nil
}
