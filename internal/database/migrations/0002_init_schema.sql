-- +migrate Up

CREATE TABLE transactions (
    id            UUID                     NOT NULL UNIQUE PRIMARY KEY,
    name          VARCHAR                  NOT NULL,
    first_wallet  UUID                     NOT NULL REFERENCES wallets (id),
    second_wallet UUID                     REFERENCES wallets (id),
    currency      VARCHAR                  NOT NULL,
    money         NUMERIC                  NOT NULL CHECK ( money > 0 ),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- +migrate Down

DROP TABLE transactions CASCADE;