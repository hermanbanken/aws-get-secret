package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var Verbose = false

type Role struct {
	Session string
	ARN     string
}

// TODO rename to SecretRef
type Secret struct {
	// ARN or name of secret to retrieve
	// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html#API_GetSecretValue_RequestSyntax
	SecretID    string
	Template    *template.Template
	Default     string
	Destination Destination
}

type Destination interface {
	destination()
}

type DestinationFile struct {
	FileMode os.FileMode
	File     string
}

func (s *Secret) TransformValue(value []byte) ([]byte, error) {
	if len(value) == 0 {
		value = []byte(s.Default)
	}

	var data interface{} = string(value)

	if s.Template != nil {
		// Allow template to handle complex JSON secrets
		if value[0] == '{' {
			// Convert the secret into JSON
			data = make(map[string]interface{})
			if err := json.Unmarshal([]byte(value), &data); err != nil {
				return nil, fmt.Errorf("failed to parse secret as JSON: %w", err)
			}
		}
		wr := bytes.NewBuffer(nil)
		err := s.Template.Execute(wr, data)
		if err != nil {
			return nil, err
		}
		return wr.Bytes(), nil
	}

	// Default: just pass on the bytes
	return value, nil
}

func (f *DestinationFile) destination() {
}

func (f *DestinationFile) Write(value []byte) error {
	path, err := filepath.Abs(f.File)
	if err != nil {
		return fmt.Errorf("failed to determine absolute path for file %s, %s", f.File, err)
	}
	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create directory for file %s, %s", f.File, err)
	}
	file, err := os.Create(f.File)
	if err != nil {
		return fmt.Errorf("failed to open file %s to write to, %s", f.File, err)
	}
	err = f.Write(value)
	if err != nil {
		return fmt.Errorf("failed to write to file %s, %s", f.File, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s, %s", f.File, err)
	}
	if f.FileMode != 0 {
		err := os.Chmod(f.File, f.FileMode)
		if err != nil {
			return fmt.Errorf("failed to chmod file %s to %s, %s", f.File, f.FileMode, err)
		}
	}
	return nil
}

type DestinationEnv struct {
	Env string
}

func (f *DestinationEnv) destination() {}

func ParseEnvironment(environ []string) ([]Secret, error) {
	result := make([]Secret, 0, 10)
	for _, env := range environ {
		name, value := ToNameValue(env)
		if !strings.HasPrefix(value, "aws:") {
			continue
		}
		secret := new(Secret)
		err := secret.Parse(value, name)
		if err != nil {
			return nil, fmt.Errorf("failed to parse environment variable %s, %s", name, err)
		}
		result = append(result, *secret)
	}
	return result, nil
}

// Parse works just like binxio gcp:// secret ref parsing (https://github.com/binxio/gcp-get-secret/blob/fee75d41297a663215256d4c1fa96deed9507291/main.go#L220)
func (s *Secret) Parse(uri string, name string) (err error) {
	uri = s.expand(uri)
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if u.Host != "" {
		return fmt.Errorf("environment variable %s has an aws: uri, but specified a host. use aws:///", name)
	}

	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return fmt.Errorf("environment variable %s has an invalid query syntax, %s", name, err)
	}

	s.SecretID = u.Path
	s.SecretID = strings.TrimPrefix(s.SecretID, "/")

	// Feature: setting a default
	s.Default = values.Get("default")

	// Feature: setting a template to wrap around the value
	if values.Get("template") != "" {
		s.Template, err = template.New("secret").Parse(values.Get("template"))
		if err != nil {
			return fmt.Errorf("environment variable %s has an invalid template syntax, %s", name, err)
		}
	}

	// Feature: setting a destination file (and optionally chmod)
	destination := values.Get("destination")
	if destination != "" {
		var fileMode os.FileMode
		chmod := values.Get("chmod")
		if chmod != "" {
			if mode, err := strconv.ParseUint(chmod, 8, 32); err != nil {
				return fmt.Errorf("chmod '%s' is not valid, %s", chmod, err)
			} else {
				fileMode = os.FileMode(mode)
			}
		}
		s.Destination = &DestinationFile{fileMode, destination}
	} else {
		s.Destination = &DestinationEnv{name}
	}

	return nil
}

func (s *Secret) expand(value string) string {
	return os.Expand(value, s.getEnv)
}

func (s *Secret) getEnv(name string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	} else {
		if Verbose {
			log.Printf("WARN: environment variable %s is not set, returning empty string", name)
		}
	}
	return ""
}

// get the name and variable of a environment entry in the form of <name>=<value>
func ToNameValue(envEntry string) (string, string) {
	result := strings.SplitN(envEntry, "=", 2)
	return result[0], result[1]
}

func expandSecretName(path string) (string, error) {
	var project string
	var name string
	var version string

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if match, _ := regexp.MatchString("projects/[^/]+/secrets/[^/]+/.*", path); match {
		return path, nil
	}

	// if m.credentials != nil {
	// 	project = m.credentials.ProjectID
	// }

	parts := strings.Split(path, "/")
	switch len(parts) {
	case 1:
		name = parts[0]
		version = "latest"
	case 2:
		if match, _ := regexp.MatchString("([0-9]+|latest)", parts[1]); match {
			name = parts[0]
			version = parts[1]
		} else {
			project = parts[0]
			name = parts[1]
			version = "latest"
		}
	case 3:
		project = parts[0]
		name = parts[1]
		version = parts[2]
	default:
		return path, fmt.Errorf("invalid secret name specification: %s", path)
	}
	result := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, version)
	if Verbose {
		log.Printf("secret name = %s\n", result)
	}
	return result, nil
}
