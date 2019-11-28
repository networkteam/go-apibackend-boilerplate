BEGIN;

CREATE TABLE app_account_request_tokens
(
    id              uuid                     NOT NULL
        CONSTRAINT app_account_request_tokens_pkey PRIMARY KEY,
    connect_token   text                     NOT NULL,
    role_identifier text                     NOT NULL,
    expiry          timestamp WITH TIME ZONE NOT NULL,
    organisation_id uuid                     NOT NULL
        CONSTRAINT organisation_id_fkey REFERENCES organisations ON DELETE CASCADE,
    device_label    text                     NOT NULL
);

CREATE UNIQUE INDEX app_account_request_tokens_organisation_id_device_label_idx ON app_account_request_tokens (organisation_id, lower(device_label));

COMMIT;
