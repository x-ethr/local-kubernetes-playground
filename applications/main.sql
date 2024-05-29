-- Notes
--
-- * `BIGSERIAL` and `SERIAL` keyword identifiers call for auto-increment upon their associated column
-- * A domain is essentially a data type with optional constraints (restrictions on the allowed set of values). The user who defines a domain becomes its owner.

-- create database "Sandbox" with owner postgres;
-- comment on database "Sandbox" is 'Testing Purposes Only';

DROP TABLE IF EXISTS "User" CASCADE;

DROP TYPE IF EXISTS "User-Verification-Status";
create type "User-Verification-Status" AS ENUM IF NOT EXISTS (
    'VERIFIED',
    'PENDING',
    'TIMEOUT',
    'REVOKED'
    );

DROP TYPE IF EXISTS "User-Account-Type";
create type "User-Account-Type" as enum ('ROOT','MEMBER');

--
-- User
--
create table if not exists "User"
(
    "id"                  bigserial
        constraint "User-ID-Primary-Key" primary key,

    "first-name"          text                                                                                 not null,
    "last-name"           text                                                                                 not null,
    "middle-name"         text                                                                                 null,
    "display-name"        text                                                                                 null,

    "account-type"        "User-Account-Type"        default 'MEMBER'                                          not null,

    "email-address"       varchar(255)                                                                         not null unique
        CONSTRAINT "User-Email-Address-Validation-Constraint" CHECK (
            "User"."email-address" ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'
            )
        constraint "User-Email-Address-Unique-Constraint" unique,

    "username"            varchar(255)
        constraint "User-Username-Unique-Constraint" unique                                                    not null,
    "password"            text                                                                                 not null,
    "avatar"              text                       default '/client-assets/users/avatars/default/avatar.png' not null,
    "verification-status" "User-Verification-Status" default 'PENDING'                                         not null,

    "creation"            timestamp with time zone   default now(),
    "modification"        timestamp with time zone,
    "deletion"            timestamp with time zone
);

--- Testing
INSERT INTO "User" ("first-name", "last-name", "middle-name", "display-name", "account-type", "email-address", username, password, "verification-status")
VALUES ('Jacob', 'Sanders', 'Brian', 'Jake', 'ROOT', 'jsanders4129@gmail.com', 'segmentational', 'placeholder', 'VERIFIED');
