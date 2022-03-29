package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hermanbanken/aws-secretmanager-wrapperscript/internal/mocksource"
	"github.com/stretchr/testify/assert"
)

var arn1 = "arn:aws:secretsmanager:us-east-2:123456789012:secret:one"
var arn2 = "arn:aws:secretsmanager:us-east-2:123456789012:secret:two"

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
}

func strptr(str string) *string {
	return &str
}
