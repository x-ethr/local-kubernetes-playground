--
-- User-Verification-Status
--

CREATE TYPE "user-verification-status" AS ENUM ('VERIFIED','PENDING','TIMEOUT','REVOKED');

--
-- User-Account-Type
--

CREATE TYPE "user-account-type" AS ENUM ('MEMBER','ROOT');

--
-- User
--

CREATE TABLE "User"
(
    "id"                  bigserial
        CONSTRAINT "user-id-primary-key" primary key,

    "name"                varchar(255)               default NULL::character varying,
    "display-name"        text                                                                                 null,

    "account-type"        "user-account-type"        default 'MEMBER'                                          not null,

    "email"               varchar(255)                                                                         not null
        CONSTRAINT "user-email-validation-constraint" CHECK ("User"."email" ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')
        CONSTRAINT "user-email-unique-constraint" unique,

    "username"            varchar(255)               default NULL
        CONSTRAINT "user-username-unique-constraint" unique,
    "avatar"              text                       default NULL,
    "verification-status" "user-verification-status" default 'PENDING'                                         not null,

    "marketing"           boolean                    default true                                              not null,

    "creation"            timestamp with time zone   default now(),
    "modification"        timestamp with time zone,
    "deletion"            timestamp with time zone
);

CREATE INDEX IF NOT EXISTS "user-deletion-index" on "User" (deletion);
CREATE INDEX IF NOT EXISTS "user-email-index" on "User" (email);
CREATE INDEX IF NOT EXISTS "user-username-index" on "User" (username);
