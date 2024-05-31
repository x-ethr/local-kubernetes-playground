#!/usr/bin/env node

import Utilities from "node:util";
import type {Types} from "./src"
import {Authentication} from "./src";

async function Main() {
    const inputs = process.argv.map((value) => {
        return value
    });

    const services = {
        authentication: {
            main: false,
            register: false
        }
    };

    const help = "Available Options: --help, --authentication-service, --authentication-service-registration";

    for (const argument of inputs) {
        if (argument === "-h" || argument === "--help") {
            console.log(help);
            process.exit(0)
        }
    }

    for (const argument of inputs) {
        switch (argument) {
            case "--authentication-service":
                services.authentication.main = true;
                break;
            case "--authentication-service-registration":
                services.authentication.register = true;
                break
        }
    }

    const triggers: Map<string, boolean> = new Map();

    triggers.set("selection", false);
    triggers.set("authentication", false);

    Object.entries(services).forEach((entry) => {
        const service = entry[0];
        const flags = entry[1];

        Object.entries(flags).forEach((value) => {
            const flag = value[0];
            const trigger = value[1];

            if (trigger) {
                triggers.set("selection", true);
                triggers.set(service, true);
            }
        })
    });

    if (!(triggers.get("selection"))) {
        console.log(help);
        process.exit(0);
    }

    const apis = {
        authentication: Authentication.Run
    };

    const counts: { [key: string]: { [key: string]: { successes: 0, failures: 0 } } } = {};

    const failures: { [key: string]: any } = {};

    for await (const [service, properties] of Object.entries(services)) {
        if (triggers.get(service)) {
            const callable = Reflect.get(apis, service);
            const configuration = Reflect.get(services, service);

            console.log("Executing Service API Call(s)", {service, configuration});

            const responses = await (await callable(configuration) as Promise<Array<PromiseSettledResult<Awaited<Promise<Types.Fetcher>>>>>);

            for await (const result of responses) {
                if (result.status == "rejected") {
                    console.error("Runtime Error", result.reason);
                    continue
                }

                const {response, input} = result.value;
                if (response && !(service in counts)) {
                    counts[service] = {
                        [response.url]: {
                            successes: 0,
                            failures: 0
                        }
                    }
                }

                if (response && !(response.url in counts[service])) {
                    counts[service][response.url] = {
                        successes: 0,
                        failures: 0
                    }
                }

                if (response && !(response.ok)) {
                    counts[service][response.url].failures++;

                    const structure: { content: object | string | null } = {
                        content: null
                    };

                    try {
                        structure.content = await response.json()
                    } catch {
                        try {
                            structure.content = await response.text()
                        } catch {
                            structure.content = null
                        }
                    }

                    const {content} = structure;
                    const {status, url} = response;

                    console.error("Error", url, {status, content, result })
                } else if (response) {
                    counts[service][response.url].successes++;
                }
            }
        }
    }

    process.stdout.write("\n");

    console.log(Utilities.inspect({counts}, {colors: true, depth: 3}));
}

(async () => await Main())();
