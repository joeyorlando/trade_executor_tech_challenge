CREATE TABLE fulfilled_orders (
  id INTEGER PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  symbol TEXT NOT NULL,
  quantity REAL NOT NULL,
  price REAL NOT NULL
);

CREATE TABLE fulfilled_order_splits (
  id INTEGER PRIMARY KEY,
  order_id INTEGER NOT NULL,
  update_id INTEGER NOT NULL,
  quantity REAL NOT NULL,
  price REAL NOT NULL,

  FOREIGN KEY (order_id)
    REFERENCES fulfilled_orders (id)
);
