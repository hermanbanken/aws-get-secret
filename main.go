package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hermanbanken/aws-secretmanager-wrapperscript/internal"
	"github.com/hermanbanken/aws-secretmanager-wrapperscript/internal/source"
)

// Constants for default values if none are supplied
const DEFAULT_TIMEOUT = 5000
const DEFAULT_REGION = "us-east-2"
const DEFAULT_SESSION = "param_session"

var (
	region      string
	roleArn     string
	timeout     int
	sessionName string
	command     []string
	verbose     bool
)

type result struct {
	ref   internal.Secret
	value *secretsmanager.GetSecretValueOutput
	err   error
}

// Does what gcp-get-secret does: https://github.com/binxio/gcp-get-secret
// - reads environment
// - find all occurrences of `aws:///`
// - inject/update environment OR files files
// - call remaining command line argument
func main() {
	// Get all of the command line data and perform the necessary validation
	getCommandParams()

	// Setup a new context to allow for limited execution time for API calls with a default of 200 milliseconds
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	source, err := source.New(ctx, region, roleArn, sessionName)
	if err != nil {
		panic("configuration error " + err.Error())
	}

	env, err := PrepareEnvAndFiles(ctx, source, os.Environ())
	if err != nil {
		panic(err.Error())
	}

	if len(command) == 1 && command[0] == "noop" {
		if verbose {
			log.Printf("INFO: noop")
		}
		return
	}

	program, err := exec.LookPath(command[0])
	if err != nil {
		log.Fatalf("could not find program %s on path, %s", command[0], err)
	}

	err = syscall.Exec(program, command, env)
	if err != nil {
		log.Fatal(err)
	}
}

func PrepareEnvAndFiles(ctx context.Context, a source.AWS, envIn []string) (env []string, err error) {
	// Get the secret refs to fill
	refs, err := internal.ParseEnvironment(envIn)
	if err != nil {
		return nil, fmt.Errorf("could not parse secret refs from environment: %w", err)
	}

	env = envIn
	if len(refs) > 0 {
		// Assume a role to retreive the parameter
		role, err := a.AttemptAssumeRole()
		if err != nil {
			return nil, fmt.Errorf("failed to assume role: %w", err)
		}

		// Retrieve all secrets
		ch := make(chan result, len(refs))
		for _, ref := range refs {
			go func(ref internal.Secret) {
				// Get the secret refs
				res, err := a.GetSecret(role, ref.SecretID)
				ch <- result{ref, res, err}
			}(ref)
		}

		// Handle all secrets
		newEnv := make(map[string]string)
		for i := 0; i < len(refs); i++ {
			res := <-ch
			if res.err != nil {
				return nil, fmt.Errorf("failed to retrieve secret %q due to error %w", res.ref.SecretID, res.err)
			}

			// AWS either puts the secret into SecretString or into SecretBinary; handle both
			var value []byte
			if v := res.value.SecretString; v != nil {
				value = []byte(*v)
			} else if len(res.value.SecretBinary) > 0 {
				value = res.value.SecretBinary
			}

			// Apply any transformations
			value, err := res.ref.TransformValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to transform secret %s value due to error %w", res.ref.SecretID, err)
			}

			// Write secret to destination
			switch dest := res.ref.Destination.(type) {
			case *internal.DestinationEnv:
				newEnv[dest.Env] = string(value)
			case *internal.DestinationFile:
				err = dest.Write(value)
				if err != nil {
					return nil, err
				}
			}
		}
		env = updateEnvironment(env, newEnv)
	}
	return env, err
}

func getCommandParams() {
	// Setup command line args
	flag.StringVar(&region, "r", DEFAULT_REGION, "The Amazon Region to use")
	flag.StringVar(&roleArn, "a", "", "The ARN for the role to assume for Secret Access")
	flag.IntVar(&timeout, "t", DEFAULT_TIMEOUT, "The amount of time to wait for any API call")
	flag.StringVar(&sessionName, "n", DEFAULT_SESSION, "The name of the session for AWS STS")
	flag.BoolVar(&verbose, "v", false, "Turn on verbose output")
	// Parse all of the command line args into the specified vars with the defaults
	flag.Parse()
	internal.Verbose = verbose

	// Verify that the user did not overwrite the defaults with empty values
	if len(region) == 0 || len(sessionName) == 0 {
		flag.PrintDefaults()
		panic("You must supply a valid region. -r REGION [-a ARN for ROLE -t TIMEOUT IN MILLISECONDS -n SESSION NAME]")
	}

	command = flag.Args()
	if len(command) == 0 {
		panic(fmt.Sprintf("You must supply a next command to run, for example \n\n\t$ %s -- echo $SECRET\n", os.Args[0]))
	}
}

func updateEnvironment(env []string, newEnv map[string]string) []string {
	result := make([]string, 0, len(env))
	for i := 0; i < len(env); i++ {
		name, _ := internal.ToNameValue(env[i])
		if newValue, ok := newEnv[name]; ok {
			result = append(result, fmt.Sprintf("%s=%s", name, newValue))
		} else {
			result = append(result, env[i])
		}
	}
	return result
}
