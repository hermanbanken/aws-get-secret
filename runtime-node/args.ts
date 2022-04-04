export function parse<C extends Record<string, {default: any, env: string}>>(argv: string[], conf: C): [Record<keyof C, any>, string] {
    const opts = <Record<keyof C,any>>{};

    // Set environment settings & defaults
    Object.entries(conf).forEach(([name, opt]: [keyof C, {default: any, env: string}]) => {
        if (process.env[opt.env]) {
            opts[name] = replace(opt.default, process.env[opt.env]);
        } else {
            opts[name] = opt.default;
        }
    });

    // Parse CLI args (they override environment settings)
    for (let i = 0; i < argv.length; i++) {
        // Only parse --[key] or -[key]; anything else could be the delegate command
        if (argv[i].startsWith("-") && argv[i] !== "--") {
            let key: keyof C = argv[i].replace(/^-{1,2}/, "");
            if (!(key in conf)) {
                throw new Error(`Unknown option ${argv[i]}`);
            }
            let value: any = true;
            if (key.indexOf("=") >= 0) {
                value = key.substring(key.indexOf("=")+1);
                key = key.substring(0, key.indexOf("="));
            } else {
                opts[key] = replace(opts[key], argv[i+1] || true);
                i += 1;
                continue;
            }
        } else {
            return [opts, argv.slice(i+1).join(" ")];
        }

    }
    return [opts, ""];
}

function replace<T>(current: T, replacer: any) {
    if (typeof current === "boolean") {
        return Boolean(replacer);
    } else if (typeof current === "number") {
        return parseFloat(replacer);
    } else {
        return replacer;
    }
}
