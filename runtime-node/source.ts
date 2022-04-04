import SecretsManager, { } from "aws-sdk/clients/secretsmanager";
import STS, { } from "aws-sdk/clients/sts";

export class Source {
    private client: SecretsManager;
    constructor(private region: string, private roleArn?: string, private sessionName?: string, client?: SecretsManager) {
        this.client = client || new SecretsManager({
            region,
            maxRetries: 1,

        });
    }
    public async attemptAssumeRole() {
        if (this.roleArn && this.roleArn.length > 0) {
            const { Credentials } = await new STS({
                region: this.region,
                maxRetries: 1,
            }).assumeRole().promise()
            if (Credentials) {
                const opts: SecretsManager.ClientConfiguration = {
                    ...this.client.config,
                    credentials: {
                        sessionToken: Credentials.SessionToken,
                        secretAccessKey: Credentials.SecretAccessKey,
                        accessKeyId: Credentials.AccessKeyId,
                    }
                };
                return new Source(this.region, this.roleArn, this.sessionName, new SecretsManager(opts));
            }
        }
        return this;
    }

    public getSecret(secretId: string) {
        return this.client.getSecretValue({SecretId: secretId}).promise();
    }
}
