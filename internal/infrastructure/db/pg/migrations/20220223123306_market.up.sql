CREATE TABLE market (
    market_id SERIAL,
    account_index int NOT NULL,
    provider_name varchar(264) NOT NULL,
    url          varchar(264) NOT NULL,
    credentials  varchar(264) NOT NULL,
    base_asset    varchar(264) NOT NULL,
    quote_asset   varchar(264) NOT NULL,
    PRIMARY KEY(account_index, provider_name)
);
