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
  const args = Object.entries({
    r: opts.region,
    a: opts.role?.arn,
    n: opts.role?.sessionName,
    t: opts.timeout,
    v: opts.verbose ? "" : undefined,
  })
    .filter(([_, value]) => typeof value === "string")
    .map(([key, value]) => `-${key} ${value}`)
    .join(" ");
  lambda.addEnvironment(
    "AWS_LAMBDA_EXEC_WRAPPER",
    "/opt/aws-get-secret " + args
  );
}

export function createLayerFromNodeModule(construct: Construct) {
  return new lambda.LayerVersion(construct, "aws-get-secret-layer", {
    code: lambda.Code.fromAsset(join(__dirname, "opt")),
    compatibleArchitectures: [Architecture.X86_64],
    compatibleRuntimes: Runtime.ALL,
  });
}
