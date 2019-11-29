BEGIN;

CREATE TABLE organisations
(
    id                uuid NOT NULL
        CONSTRAINT organisations_pkey PRIMARY KEY,
    organisation_name text NOT NULL
        CONSTRAINT organisation_name_key UNIQUE
);

CREATE UNIQUE INDEX organisations_organisation_name_unique_idx ON organisations (lower(organisation_name));

COMMIT;
