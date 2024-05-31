import {faker} from "@faker-js/faker";

import Utilities from "node:util";

import { Types } from "..";

export module Authentication {
    const state = {total: 0, users: 0, ceiling: 1e6};

    function random(min: number, max: number) { // min and max included
        return Math.floor(Math.random() * (max - min + 1) + min);
    };

    function seed(entropy: number) {
        console.debug("[Debug] Seeding User Entropy ...", entropy);

        state.total = entropy;

        faker.seed(entropy);
    };

    function create(): Type {
        const first = faker.person.firstName();
        const last = faker.person.lastName();

        const user: Type = {
            email: faker.internet.email({firstName: first, lastName: last}).toLowerCase(),
            password: faker.internet.password({length: random(8, 72)}),
        };

        return user;
    };

    async function generate(total: number): Promise<Type[]> {
        seed(total);

        if (state.total === 0 || !state.total) {
            throw new Error("Data Must be Seeded");
        }

        console.debug("[Debug] Creating User(s) ...");

        const iterator = Array.from({length: total});

        const users: Type[] = [];

        for await (const index of iterator) {
            users.push(create());
        }

        return users;
    };

    interface Type {
        email: string;
        password: string;
    };

    export const API = {
        "/": async (): Promise<Types.Fetcher> => {
            const request = new Request("http://localhost:8080/v1/authentication", {
                body: undefined,
                cache: "no-cache",
                credentials: undefined,
                headers: undefined,
                keepalive: true,
                method: "GET",
            });

            const response = await fetch(request);

            return {input: null, response};
        },
        "/register": async (user: Type): Promise<Types.Fetcher> => {
            const body = JSON.stringify(user);
            const request = new Request("http://localhost:8080/v1/authentication/register", {
                body: body,
                cache: "no-cache",
                credentials: undefined,
                headers: undefined,
                keepalive: true,
                method: "POST",
            });

            const response = await fetch(request);

            return {input: user, response};
        }
    };

    export interface Configuration {
        /**
         * main represents a call to /v1/users; calls to main will generate 100 requests.
         */
        main: boolean;

        /**
         * register represents a call to /v1/users/register; calls to register will generate 100 users at a time, and 100 requests.
         */
        register: boolean;
    }

    export async function Run(options: Configuration): Promise<Array<PromiseSettledResult<Awaited<Promise<Types.Fetcher>>>>> {
        process.setMaxListeners(100);

        const responses: Promise<Types.Fetcher>[] = [];

        const total = 100;
        const controller = new AbortController();
        const signal = controller.signal;

        if (options.main) {
            for (let i = 0; i < total; ++i) {
                responses.push(API["/"]());
            }
        }

        if (options.register) {
            const users = await generate(total);

            for (const user of users) {
                responses.push(API["/register"](user))
            }
        }

        const promises = await Promise.allSettled(responses);

        return promises;
    }
}
