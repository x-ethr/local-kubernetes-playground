#!/usr/bin/env node

import Utilities from "node:util"

import {faker} from "@faker-js/faker";

function random(min: number, max: number) { // min and max included
    return Math.floor(Math.random() * (max - min + 1) + min);
}

export module User {
    const state = {total: 0, users: 0, ceiling: 1e6};

    function seed(entropy: number) {
        console.debug("[Debug] Seeding User Entropy ...", entropy);

        state.total = entropy;

        faker.seed(entropy);
    }

    async function create(): Promise<User.Type> {
        const first = faker.person.firstName();
        const last = faker.person.lastName();
        const length = crypto

        const user: User.Type = {
            name: Utilities.format("%s %s", first, last),
            email: faker.internet.email({firstName: first, lastName: last}).toLowerCase(),
            password: faker.internet.password({length: random(8, 72)}),
        };

        return user;
    }

    export async function generate(total: number) {
        seed(total);

        if (state.total === 0 || !state.total) {
            throw new Error("Data Must be Seeded");
        }

        console.debug("[Debug] Creating User(s) ...");

        const iterator = Array.from({length: total});

        const users: User.Type[] = [];

        for await (const index of iterator) {
            users.push(await create());
        }

        return users;
    }

    export interface Type {
        name: string;
        email: string;
        password: string;
    }
}

const test = async (signal: AbortSignal, user: User.Type) => {
    const request = new Request("http://localhost:8080/v1/users", {
        body: undefined,
        cache: "no-cache",
        credentials: undefined,
        headers: undefined,
        keepalive: true,
        method: "GET",
        signal: signal,
    });

    try {
        const response = await fetch(request);

        const structure = await response.json();

        return {response, structure, input: user, error: null};
    } catch (error) {
        return {response: null, structure: null, input: user, error: error};
    }
};

const api = async (signal: AbortSignal, user: User.Type) => {
    const body = JSON.stringify(user);
    const request = new Request("http://localhost:8080/v1/users/register", {
        body: body,
        cache: "no-cache",
        credentials: undefined,
        headers: undefined,
        keepalive: true,
        method: "POST",
        signal: signal,
    });

    try {
        const response = await fetch(request);

        const structure: { content: string | Object | null } = {content: null};

        if (response.status >= 400) {
            try {
                structure.content = await response.text()
            } catch {
                structure.content = await response.json()
            }
        } else {
            structure.content = await response.json();
        }

        return {response, structure, input: user, error: null};
    } catch (error) {
        return {response: null, structure: null, input: user, error: error};
    }
};

async function Test() {
    process.setMaxListeners(1000);

    const total = 100

    const users = await User.generate(total);

    const controller = new AbortController();
    const signal = controller.signal;

    const promises = users.map(user => {
        return test(signal, user)
    });

    const awaitables = await Promise.allSettled(promises);

    const counts = {success: 0, failure: 0};
    for (const awaitable of awaitables) {
        if (awaitable.status === "fulfilled") {
            const {response, structure, input, error} = awaitable.value;

            if (error) {
                counts.failure++
                console.error("API Error", error, {response, structure, input});
                continue
            }

            counts.success++;
            console.log("API", Utilities.inspect({response: {...structure}}, {colors: true}));
        } else {
            const {reason} = awaitable;
            counts.failure++
            console.error("Error", {reason});

        }
    }

    console.log(Utilities.inspect(counts, {colors: true}))
}

async function Register() {
    process.setMaxListeners(1000);

    const total = 100;

    const users = await User.generate(total);

    const controller = new AbortController();
    const signal = controller.signal;

    const promises = users.map(user => {
        return api(signal, user)
    });

    const awaitables = await Promise.allSettled(promises);

    const counts = {success: 0, failure: 0};
    for (const awaitable of awaitables) {
        if (awaitable.status === "fulfilled") {
            const {response, structure, input, error} = awaitable.value;

            if (error) {
                counts.failure++
                console.error("API Error", error, {response, structure, input});
                continue
            } else if (response && response.status >= 400) {
                if (response.status == 409) {
                    counts.success++
                    continue
                }
                counts.failure++
                console.error("API Status >= 400",  error, {response, structure, input});
                continue
            }

            counts.success++;
            console.log("API", Utilities.inspect({response: structure}, {colors: true}));
        } else {
            const {reason} = awaitable;
            counts.failure++
            console.error("Error", {reason});

        }
    }

    console.log(Utilities.inspect(counts, {colors: true}))
}

(async () => await Register())();
