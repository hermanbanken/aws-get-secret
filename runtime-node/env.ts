import { Source } from "./source";
import { writeFileSync, mkdirSync } from "fs";
import { basename } from "path";

export function findAll(env: NodeJS.ProcessEnv) {
    return Object.entries(env)
        .filter(([key, value]) => value?.startsWith("aws:///"))
        .map(([key, value]) => new EnvSecret(key, value!));
}

export class EnvSecret {
    private secretId: string;
    private params: URLSearchParams;
    constructor(public env: string, public value: string) {
        const { pathname, searchParams } = new URL(value);
        this.secretId = pathname.replace(/^\//, "");
        this.params = searchParams;
        if (searchParams.has("template")) {
            console.error("template secret replacements are only supported in Golang mode");
            process.exit(1);
        }
    }
    public async handle(secret: Source): Promise<[string, string] | undefined> {
        const value = await secret.getSecret(this.secretId).catch((e) => {
            // Allow setting a default for if secret is not defined
            if (e.code === "ResourceNotFoundException" && this.params.has("default")) {
                return { SecretString: this.params.get("default"), SecretBinary: undefined };
            }
            throw e;
        }).then(({ SecretBinary, SecretString }) => {
            // Handle binary/string secrets
            return typeof SecretString === "string"
                ? SecretString
                : Buffer.from(SecretBinary as string, "base64").toString("utf8");
        });

        // Handle file destinations
        const dest = this.params.get("destination");
        const chmod = this.params.get("chmod");
        if (dest) {
            mkdirSync(basename(dest), { recursive: true });
            writeFileSync(dest, value, { encoding: "utf-8", mode: chmod || undefined });
            return;
        }

        return [this.env, value];
    }
}