CREATE TABLE test (
    id SERIAL PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);