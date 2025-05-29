CREATE TABLE plans(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    features JSONB NOT NULL,  
    duration_days INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO plans (name, price, features, duration_days) 
VALUES
('Basic Plan', 9.99, '["Feature A", "Feature B"]', 30),
('Pro Plan', 19.99, '["Feature A", "Feature B", "Feature C"]', 60),
('Enterprise Plan', 49.99, '["Feature A", "Feature B", "Feature C", "Priority Support"]', 90);
