-- name: GetAllMarkets :many
SELECT * FROM market;

-- name: InsertMarket :one
INSERT INTO market (
    provider_name,url,base_asset,quote_asset) VALUES (
             $1, $2, $3, $4
    )
    RETURNING *;