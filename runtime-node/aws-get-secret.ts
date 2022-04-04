#!/usr/bin/env node

import { Source } from "./source";
import { parse } from "./args";
import { findAll } from "./env";
import { execSync } from "child_process";

const [args, remainder] = parse(process.argv.slice(2), {
    region: {
        default: "us-east-2",
        env: "AWS_GET_SECRET_REGION"
    },
    role: {
        default: "",
        env: "AWS_GET_SECRET_ROLE"
    },
    timeout: {
        default: 5000,
        env: "AWS_GET_SECRET_TIMEOUT"
    },
    ['session-name']: {
        default: "param_session",
        env: "AWS_GET_SECRET_SESSION_NAME"
    },
    verbose: {
        default: false,
        env: "AWS_GET_SECRET_VERBOSE"
    },
});

(async function run() {
    if (args.verbose) {
        console.log("aws-get-secret", { options: args });
    }

    // Prepare client
    let source = new Source(args.region, args.role, args["session-name"]);
    source = await source.attemptAssumeRole();

    // Update environment
    const secretEnvs = findAll(process.env);
    if (args.verbose) {
        console.log("refs", secretEnvs);
    }
    const lookups = secretEnvs.map((envSecret) => envSecret.handle(source));
    const env = await Promise.all(lookups).then(list => Object.fromEntries(list.filter(isDefined)));

    // Run delegate command
    execSync(remainder, { env: { ...process.env, ...env }  });
})();

function isDefined<T>(v: T | undefined): v is T {
    return typeof v !== "undefined";
}