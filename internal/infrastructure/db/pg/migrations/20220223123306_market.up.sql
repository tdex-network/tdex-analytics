CREATE TABLE market (
    market_id SERIAL,
    provider_name varchar(264) NOT NULL,
    url          varchar(264) NOT NULL,
    base_asset    varchar(264) NOT NULL,
    quote_asset   varchar(264) NOT NULL,
    PRIMARY KEY(provider_name, base_asset, quote_asset)
);
