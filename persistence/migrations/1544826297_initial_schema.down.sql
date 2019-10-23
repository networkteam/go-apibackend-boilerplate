BEGIN;

DROP TABLE accounts;

DROP INDEX accounts_email_address_idx;
DROP INDEX accounts_organisation_id_device_label_idx;

DROP TABLE organisations;

DROP INDEX organisations_organisation_name_unique_idx;

COMMIT;
