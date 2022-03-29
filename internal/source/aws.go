package source

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWS interface {
	// AttemptAssumeRole will attempt to assume the supplied role and return either an error or the assumed role
	AttemptAssumeRole() (*sts.AssumeRoleOutput, error)

	// GetSecret will return the descrypted version of the Secret from Secret Manager using the supplied
	// assumed role to interact with Secret Manager.  This function will return either an error or the
	// retrieved and decrypted secret.
	GetSecret(assumedRole *sts.AssumeRoleOutput, secretID string) (*secretsmanager.GetSecretValueOutput, error)
}

type Impl struct {
	ctx         context.Context
	region      string
	roleArn     string
	sessionName string
	cfg         aws.Config
}

func New(ctx context.Context, region string, roleArn, sessionName string) (AWS, error) {
	// Load the config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithRetryer(func() aws.Retryer {
		// NopRetryer is used here in a global context to avoid retries on API calls
		return retry.AddWithMaxAttempts(aws.NopRetryer{}, 1)
	}))
	return Impl{ctx, region, roleArn, sessionName, cfg}, err
}

func (t Impl) AttemptAssumeRole() (*sts.AssumeRoleOutput, error) {
	if len(t.roleArn) <= 0 {
		return nil, nil
	}
	client := sts.NewFromConfig(t.cfg)
	return client.AssumeRole(t.ctx,
		&sts.AssumeRoleInput{
			RoleArn:         &t.roleArn,
			RoleSessionName: &t.sessionName,
		},
	)
}

func (t Impl) GetSecret(assumedRole *sts.AssumeRoleOutput, secretID string) (*secretsmanager.GetSecretValueOutput, error) {
	if assumedRole != nil {
		client := secretsmanager.NewFromConfig(t.cfg, func(o *secretsmanager.Options) {
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(*assumedRole.Credentials.AccessKeyId, *assumedRole.Credentials.SecretAccessKey, *assumedRole.Credentials.SessionToken))
		})
		return client.GetSecretValue(t.ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretID),
		})
	} else {
		client := secretsmanager.NewFromConfig(t.cfg)
		return client.GetSecretValue(t.ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretID),
		})
	}
}
