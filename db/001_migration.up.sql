CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(250),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
