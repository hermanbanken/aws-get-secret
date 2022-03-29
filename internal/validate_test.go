package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateArn(t *testing.T) {
	err := ValidateArn("arn:aws:iam::123456789012:role/demo")
	assert.Nil(t, err)

	err = ValidateArn("arn:aws:secretsmanager:us-east-2:123456789012:secret:rds-db-credentials/cluster-1")
	assert.Nil(t, err)
}

func TestValidateSessionName(t *testing.T) {
	err := ValidateSessionName("foobar@=blabla,comma-hyphen+letters")
	assert.Nil(t, err)
}
