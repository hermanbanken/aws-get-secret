import * as lambda from "aws-cdk-lib/aws-lambda";
import type * as lambdaPython from "@aws-cdk/aws-lambda-python-alpha";
import type * as lambdaNodejs from "aws-cdk-lib/aws-lambda-nodejs";
import type { Construct } from "constructs";
declare type Lambda = Pick<lambda.Function | lambdaNodejs.NodejsFunction | lambdaPython.PythonFunction, "addLayers" | "addEnvironment" | "stack">;
declare type Opts = {
    region?: string;
    role?: {
        arn: string;
        sessionName?: string;
    };
    timeout?: string;
    verbose?: boolean;
};
export declare function wrapLambdasWithSecrets(lambdas: Lambda[], opts: Opts): void;
export declare function wrapLambdaWithSecrets(lambda: Lambda, opts: Opts, layer: lambda.LayerVersion): void;
export declare function createLayerFromNodeModule(construct: Construct): lambda.LayerVersion;
export {};
