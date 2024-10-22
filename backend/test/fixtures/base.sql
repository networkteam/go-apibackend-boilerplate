--
-- Organisations
--

INSERT INTO
    organisations (organisation_id, name)
VALUES ('6330de58-2761-411e-a243-bec6d0c53876',
           'Acme Inc.'),
       ('dba20d09-a3df-4975-9406-2fb6fd8f0940',
           'Other Corp');

--
-- Accounts
--

-- Account (d7037ad0-d4bb-4dcc-8759-d82fbb3354e8)
--   username: admin@example.com
--   password: myRandomPassword
--   role: SystemAdministrator

INSERT INTO
    accounts (account_id, role_identifier, secret, email_address, password_hash)
VALUES ('d7037ad0-d4bb-4dcc-8759-d82fbb3354e8',
        'SystemAdministrator',
        '\xf71ab8929ad747915e135b8e9a5e01403329cc6b202c8e540e74920a78394e36',
        'admin@example.com',
        '\x24326124303424664b4263675349637966474f6f4571534b5a566c6c4f6d4f347461395161623162545a65556c556e6b4962455269764a645930624f');

-- Account (3ad082c7-cbda-49e1-a707-c53e1962be65)
--   username: admin+acmeinc@example.com
--   password: myRandomPassword
--   role: OrganisationAdministrator

INSERT INTO
    accounts (account_id, role_identifier, secret, email_address, password_hash, organisation_id)
VALUES ('3ad082c7-cbda-49e1-a707-c53e1962be65',
        'OrganisationAdministrator',
        '\xf71ab8929ad747915e135b8e9a5e01403329cc6b202c8e540e74920a78394e36',
        'admin+acmeinc@example.com',
        '\x24326124303424664b4263675349637966474f6f4571534b5a566c6c4f6d4f347461395161623162545a65556c556e6b4962455269764a645930624f',
           -- Acme Inc.
        '6330de58-2761-411e-a243-bec6d0c53876');

-- Account (f045e5d1-cdad-4964-a7e2-139c8a87346c)
--   username: otheradmin+acmeinc@example.com
--   password: myRandomPassword
--   role: OrganisationAdministrator

INSERT INTO
    accounts (account_id, role_identifier, secret, email_address, password_hash, organisation_id)
VALUES ('f045e5d1-cdad-4964-a7e2-139c8a87346c',
        'OrganisationAdministrator',
        '\xf71ab8929ad747915e135b8e9a5e01403329cc6b202c8e540e74920a78394e36',
        'otheradmin+acmeinc@example.com',
        '\x24326124303424664b4263675349637966474f6f4571534b5a566c6c4f6d4f347461395161623162545a65556c556e6b4962455269764a645930624f',
           -- Acme Inc.
        '6330de58-2761-411e-a243-bec6d0c53876');

-- Account (2035f4da-f385-42c4-a609-02d9aa7290e5)
--   username: admin+othercorp@example.com
--   password: myRandomPassword
--   role: OrganisationAdministrator

INSERT INTO
    accounts (account_id, role_identifier, secret, email_address, password_hash, organisation_id)
VALUES ('2035f4da-f385-42c4-a609-02d9aa7290e5',
        'OrganisationAdministrator',
        '\xf71ab8929ad747915e135b8e9a5e01403329cc6b202c8e540e74920a78394e36',
        'admin+othercorp@example.com',
        '\x2424326124303424664b4263675349637966474f6f4571534b5a566c6c4f6d4f347461395161623162545a65556c556e6b4962455269764a645930624f',
           -- Other Corp
        'dba20d09-a3df-4975-9406-2fb6fd8f0940');
