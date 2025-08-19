CREATE TABLE IF NOT EXISTS Items(
    chrt_id INT PRIMARY KEY,
    track_number TEXT, 
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id INT,
    brand TEXT,
    status INT,
    order_id TEXT REFERENCES Orders(order_uid) ON DELETE CASCADE
)