package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hermanbanken/aws-get-secret/internal/mocksource"
	"github.com/stretchr/testify/assert"
)

var arn1 = "arn:aws:secretsmanager:us-east-2:123456789012:secret:one"
var arn2 = "arn:aws:secretsmanager:us-east-2:123456789012:secret:two"
var arn3 = "arn:aws:secretsmanager:us-east-2:123456789012:secret:three"

func TestPrepare(t *testing.T) {
	source := mocksource.MockSource{
		arn1: &secretsmanager.GetSecretValueOutput{
			SecretString: strptr("baz"),
		},
		arn2: &secretsmanager.GetSecretValueOutput{
			SecretBinary: []byte(`{"key": "value", "key2": "value2"}`),
		},
	}
	outEnv, err := PrepareEnvAndFiles(context.TODO(), source, []string{
		"FOOBAR=aws:///" + arn1 + "?template={{.}}&region=eu-west-1&default=foobar",
		"FOOBAR2=aws:///" + arn2 + "?template={{.key2}}",
	})
	if !assert.NoError(t, err) || !assert.Equal(t, []string{
		"FOOBAR=baz",
		"FOOBAR2=value2",
	}, outEnv) {
		t.Fail()
	}

	// Non-existing secret
	_, err = PrepareEnvAndFiles(context.TODO(), source, []string{
		"FOOBAR3=aws:///" + arn3,
	})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("failed to retrieve secret \"%s\" due to error not found", arn3))
}

func strptr(str string) *string {
	return &str
}
