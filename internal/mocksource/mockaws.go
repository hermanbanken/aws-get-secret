package mocksource

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hermanbanken/aws-get-secret/internal/source"
)

type MockSource map[string]*secretsmanager.GetSecretValueOutput

var _ source.AWS = new(MockSource)

// AttemptAssumeRole will attempt to assume the supplied role and return either an error or the assumed role
func (ms MockSource) AttemptAssumeRole() (*sts.AssumeRoleOutput, error) {
	return &sts.AssumeRoleOutput{}, nil
}

// GetSecret will return the descrypted version of the Secret from Secret Manager using the supplied
// assumed role to interact with Secret Manager.  This function will return either an error or the
// retrieved and decrypted secret.
func (ms MockSource) GetSecret(assumedRole *sts.AssumeRoleOutput, secretID string) (*secretsmanager.GetSecretValueOutput, error) {
	log.Printf("Retrieving %s", secretID)
	if v, hasValue := ms[secretID]; hasValue {
		return v, nil
	}
	return nil, errors.New("not found")
}
