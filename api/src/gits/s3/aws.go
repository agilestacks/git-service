package s3

import (
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsec2rolecreds "github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	awsec2metadata "github.com/aws/aws-sdk-go/aws/ec2metadata"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"

	"gits/config"
)

var S3 *awss3.S3

func Init() {
	awsConfig := aws.NewConfig()
	if config.AwsRegion != "" {
		awsConfig = awsConfig.WithRegion(config.AwsRegion)
	}
	if config.AwsProfile != "" {
		awsConfig = awsConfig.WithCredentials(awsCredentials(config.AwsProfile))
	}
	S3 = awss3.New(awssession.New(), awsConfig)
}

func awsCredentials(profile string) *awscredentials.Credentials {
	shared := awscredentials.SharedCredentialsProvider{}
	if profile != "" {
		shared.Profile = profile
	}
	providers := []awscredentials.Provider{
		&awscredentials.EnvProvider{},
		&shared,
	}
	if config.AwsUseIamRoleCredentials {
		providers = append(providers, &awsec2rolecreds.EC2RoleProvider{Client: awsec2metadata.New(awssession.New())})
	}
	return awscredentials.NewCredentials(&awscredentials.ChainProvider{Providers: providers, VerboseErrors: config.Verbose})
}

func openArchive(archive string) (io.ReadCloser, error) {
	location, err := url.Parse(archive)
	if err != nil {
		return nil, err
	}
	obj, err := S3.GetObject(
		&awss3.GetObjectInput{
			Bucket: &location.Host,
			Key:    &location.Path,
		})
	if err != nil {
		return nil, fmt.Errorf("Failed to get S3 object %q: %v", archive, err)
	}
	return obj.Body, nil
}
