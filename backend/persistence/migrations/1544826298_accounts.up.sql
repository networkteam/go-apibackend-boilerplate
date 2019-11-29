BEGIN;

CREATE TABLE accounts
(
    id              uuid  NOT NULL
        CONSTRAINT accounts_pkey PRIMARY KEY,
    type            text  NOT NULL,
    role_identifier text  NOT NULL,
    secret          bytea NOT NULL,
    email_address   text,
    password_hash   bytea,
    device_label    text,
    organisation_id uuid
        CONSTRAINT organisation_id_fkey REFERENCES organisations ON DELETE CASCADE,
    first_name      text,
    last_name       text,
    device_token    text,
    device_os       text
);

CREATE UNIQUE INDEX accounts_email_address_idx ON accounts (lower(email_address));
CREATE UNIQUE INDEX accounts_organisation_id_device_label_idx ON accounts (organisation_id, lower(device_label));

COMMIT;
