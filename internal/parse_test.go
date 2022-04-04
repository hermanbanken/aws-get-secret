package internal

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

// Full ARN; with slash in the name (=valid)
var arn = "arn:aws:secretsmanager:us-east-2:123456789012:secret:rds-db-credentials/cluster-1"
var secretName = "some-db-credentials"

func TestParse(t *testing.T) {
	var secret = new(Secret)
	err := secret.Parse("aws:///"+arn+"?template={{.}}&region=eu-west-1&default=foobar", "SECRET")
	if assert.NoError(t, err) {
		assert.Equal(t, &Secret{
			SecretID:    arn,
			Destination: &DestinationEnv{"SECRET"},
			Default:     "foobar",
			Template:    template.Must(template.New("secret").Parse("{{.}}")),
		}, secret)
	}

	secret = new(Secret)
	err = secret.Parse("aws:///"+secretName+"?region=eu-west-1", "SECRET")
	if assert.NoError(t, err) {
		assert.Equal(t, &Secret{
			SecretID:    secretName,
			Destination: &DestinationEnv{"SECRET"},
		}, secret)
	}
}

func TestParseEnv(t *testing.T) {
	refs, err := ParseEnvironment([]string{
		"FOOBAR=aws:///" + arn + "?region=eu-west-1&default=foobar",
		"FOOBAR2=aws:///" + secretName,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, []Secret{{
			SecretID:    arn,
			Destination: &DestinationEnv{"FOOBAR"},
			Default:     "foobar",
		}, {
			SecretID:    secretName,
			Destination: &DestinationEnv{"FOOBAR2"},
		}}, refs)
	}
}
