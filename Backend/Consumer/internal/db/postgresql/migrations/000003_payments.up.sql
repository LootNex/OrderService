CREATE TABLE IF NOT EXISTS Payments(
    transaction TEXT PRIMARY KEY,
    request_id TEXT,
    currency TEXT NOT NULL,
    provider TEXT NOT NULL,
    amount INT NOT NULL,
    payment_dt INT NOT NULL,
    bank TEXT NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT,
    order_id TEXT REFERENCES Orders(order_uid) ON DELETE CASCADE
)