# aws-get-secret
This is a utility much like the awesome [gcp-get-secret](https://github.com/binxio/gcp-get-secret) from Binx.io. You can wrap it around your application that consumes environment variables containing secrets. By calling `aws-get-secret -- [your-cli] [--your-args]` it will call your cli command (server, serverless, etc.) and fill any environment variable that starts with the `aws:///` format documented below.

Note: use any libary dealing with secrets with great care. Actions you can take to prevent impact:

1. Do not trust me. Pin down your dependencies to exact SHA-512 hashes like with npm package-lock and Golang's go.sum.
1. Do not trust any other module (node, etc) you install: vet them & pin them.
1. Install upgrades only after careful review & again pin your dependencies.
1. Rotate your secrets frequently, to make it something you do with ease.
1. Seriously, pin your dependencies.

## Forks welcome
If you miss functionality, feel free to fork the repository & optionally send Pull Requests to contribute back.

## Usage
First, set some environment variables that define where to get the secret:

```bash
export FIRST_SECRET=aws:///arn:aws:secretsmanager:eu-central-1:1234567:secret:First-ABCDEF
export OTHER_SECRET=aws:///arn:aws:secretsmanager:eu-central-1:1234567:secret:Other-ABCDEF
```

You can define query parameters on the `aws:///` uri just like with binxio/gcp-get-secret:

- `default` to set a default value if there is no value
- `template` to pick values from a JSON secret or to wrap the value with other data
- `destination` and `chmod` to write to a file instead of using the environment

Then wrap your executable with this tool:

```bash
# quick example:
./aws-get-secret sh -c 'echo Something $SECRET;'

# NodeJS server:
./aws-get-secret node dist/server.js

# Python server server:
./aws-get-secret python3 server.py
```

## AWS Lambda
Call this tool as a Lambda extension script (wrapperscript) to preload secret manager secrets to environment variables.

To use this wrapper script, create a Layer including the go binary of this repository.
Then include the binary in another layer and invoke it by setting `AWS_LAMBDA_EXEC_WRAPPER=/opt/aws-get-secret` on your lambda.
Alternatively, you can use the NodeJS CDK-compatible package which does this for you.

```bash
npm i aws-get-secret-lambda
```

Then wrap your Lambda like this (if you're using CDK):

```typescript
import { wrapLambdasWithSecrets } from "aws-get-secret-lambda"

export class SomeStack extends Stack {
  constructor(scope: Construct) {
    super(scope, 'SomeStack');

    wrapLambdasWithSecrets(this.getAllFunctions());
  }
}
```

### Refs
1. https://github.com/aws-samples/aws-lambda-environmental-variables-from-aws-secrets-manager
1. https://dev.to/aws-builders/getting-the-most-of-aws-lambda-free-compute-wrapper-scripts-3h4b
1. https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper
1. https://github.com/binxio/gcp-get-secret
