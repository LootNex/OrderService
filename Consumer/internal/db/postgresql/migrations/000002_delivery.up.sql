CREATE TABLE IF NOT EXISTS Delivery(
    delivery_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    zip TEXT,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT NOT NULL,
    email TEXT,
    order_id TEXT REFERENCES Orders(order_uid) ON DELETE CASCADE
)