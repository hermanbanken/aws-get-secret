import * as lambda from "aws-cdk-lib/aws-lambda";
import { join } from "path";

import type * as lambdaPython from "@aws-cdk/aws-lambda-python-alpha";
import type * as lambdaNodejs from "aws-cdk-lib/aws-lambda-nodejs";
import type { Construct } from "constructs";
import { Architecture, Runtime } from "aws-cdk-lib/aws-lambda";

type Lambda = Pick<
  lambda.Function | lambdaNodejs.NodejsFunction | lambdaPython.PythonFunction,
  "addLayers" | "addEnvironment" | "stack"
>;

type Opts = {
  region?: string;
  role?: { arn: string; sessionName?: string };
  timeout?: string;
  verbose?: boolean;
};

export function wrapLambdasWithSecrets(lambdas: Lambda[], opts: Opts) {
  if (lambdas.length === 0) {
    return;
  }
  const layer = createLayerFromNodeModule(lambdas[0].stack);
  lambdas.forEach((lambda) => wrapLambdaWithSecrets(lambda, opts, layer));
}

export function wrapLambdaWithSecrets(
  lambda: Lambda,
  opts: Opts,
  layer: lambda.LayerVersion
) {
  lambda.addLayers(layer);
  const env = {
    AWS_GET_SECRET_REGION: opts.region,
    AWS_GET_SECRET_ROLE: opts.role?.arn,
    AWS_GET_SECRET_SESSION_NAME: opts.role?.sessionName,
    AWS_GET_SECRET_TIMEOUT: opts.timeout,
    AWS_GET_SECRET_VERBOSE: opts.verbose ? "true" : undefined,
    AWS_LAMBDA_EXEC_WRAPPER: "/opt/aws-get-secret",
  };
  Object.entries(env)
    .filter((pair): pair is [string,string] => typeof pair[1] === "string")
    .forEach(([key, value]) => lambda.addEnvironment(key, value));
}

export function createLayerFromNodeModule(construct: Construct) {
  return new lambda.LayerVersion(construct, "aws-get-secret-layer", {
    code: lambda.Code.fromAsset(join(__dirname, "opt")),
    compatibleArchitectures: [Architecture.X86_64],
    compatibleRuntimes: [
      Runtime.NODEJS_10_X,
      Runtime.NODEJS_12_X,
      Runtime.NODEJS_14_X,
      Runtime.GO_1_X,
      Runtime.PYTHON_3_9,
      Runtime.PYTHON_3_8,
      Runtime.PYTHON_3_6,
      Runtime.PYTHON_2_7,
      Runtime.JAVA_8,
      Runtime.JAVA_11,
      Runtime.PROVIDED,
    ],
  });
}
