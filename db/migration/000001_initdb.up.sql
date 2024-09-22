CREATE TABLE "users"
(
    "id"                  bigserial PRIMARY KEY,
    "username"            varchar UNIQUE NOT NULL,
    "hashed_password"     varchar        NOT NULL,
    "first_name"          varchar        NOT NULL,
    "last_name"           varchar        NOT NULL,
    "email"               varchar UNIQUE NOT NULL,
    "password_changed_at" timestamp      NOT NULL DEFAULT '0001-01-01 00:00:00Z',
    "created_at"          timestamptz    NOT NULL DEFAULT (now())
);

CREATE TABLE "links"
(
    "id"         bigserial PRIMARY KEY,
    "user"       bigserial      NOT NULL,
    "code"       varchar UNIQUE NOT NULL,
    "link"       varchar        NOT NULL,
    "created_at" timestamptz    NOT NULL DEFAULT (now())
);

CREATE INDEX ON "links" ("code");

CREATE INDEX ON "links" ("code", "link");

ALTER TABLE "links"
    ADD FOREIGN KEY ("user") REFERENCES "users" ("id");
