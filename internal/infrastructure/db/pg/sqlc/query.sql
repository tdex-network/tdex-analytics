-- name: GetAllMarkets :many
SELECT * FROM market;

-- name: InsertMarket :one
INSERT INTO market (
    account_index, provider_name,url,credentials,base_asset,quote_asset) VALUES (
             $1, $2, $3, $4, $5, $6
    )
    RETURNING *;