-- migrations/01_init.sql

CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(50),
    locale VARCHAR(10),
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMPTZ,
    oof_shard VARCHAR(10),
    delivery JSONB,
    payment JSONB
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT,
    track_number VARCHAR(50),
    price NUMERIC,
    rid VARCHAR(50),
    name VARCHAR(100),
    sale INT,
    size VARCHAR(20),
    total_price NUMERIC,
    nm_id INT,
    brand VARCHAR(50),
    status INT
);