# aws-secretmanager-wrapperscript
Use this Lambda extension script (wrapperscript) to preload secret manager secrets to environment variables

Note: use any libary dealing with secrets with great care. Actions you can take to prevent impact:

1. Do not trust me. Pin down your dependencies to exact SHA-512 hashes like with npm package-lock and Golang's go.sum.
1. Do not trust any other module (node, etc) you install: vet them & pin them.
1. Install upgrades only after careful review & again pin your dependencies.
1. Rotate your secrets frequently, to make it something you do with ease.
1. Seriously, pin your dependencies.

### Forks welcome
If you miss functionality, feel free to fork the repository & optionally send Pull Requests to contribute back.

### Usage
To use this wrapper script, create a Layer including the go binary of this repository.
Then include the `wrap.sh` in another layer and invoke it by setting `AWS_LAMBDA_EXEC_WRAPPER=/opt/wrap.sh` on your lambda.
Alternatively, you can use the NodeJS CDK-compatible package which does this for you: TODO CREATE THIS.

Then attach environment variables to your Lambda that define where to get the secret:

```
FIRST_SECRET=aws:///arn:aws:secretsmanager:eu-central-1:1234567:secret:First-ABCDEF
OTHER_SECRET=aws:///arn:aws:secretsmanager:eu-central-1:1234567:secret:Other-ABCDEF
```

You can define query parameters on the `aws:///` uri just like with binxio/gcp-get-secret:

- `default` to set a default value if there is no value
- `template` to pick values from a JSON secret or to wrap the value with other data
- `destination` and `chmod` to write to a file instead of using the environment

### Testing
```bash
export SECRET=aws:///mysecret?default=foobar
./aws-secretmanager-wrapperscript sh -c 'echo Something $SECRET;'
```

### Refs
1. https://github.com/aws-samples/aws-lambda-environmental-variables-from-aws-secrets-manager
1. https://dev.to/aws-builders/getting-the-most-of-aws-lambda-free-compute-wrapper-scripts-3h4b
1. https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper
1. https://github.com/binxio/gcp-get-secret
