-- name: GetAllMarkets :many
SELECT * FROM market;

-- name: InsertMarket :one
INSERT INTO market (
    provider_name,url,base_asset,quote_asset,active) VALUES (
             $1, $2, $3, $4, $5
    )
    RETURNING *;

-- name: UpdateActive :exec
UPDATE market set active = $1 where market_id = $2;
