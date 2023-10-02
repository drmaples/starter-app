
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(250),
    first_name VARCHAR(100),
    last_name VARCHAR(100)
);

INSERT INTO users (email, first_name, last_name) VALUES
    ('darrell@example.com', 'duh', 'rell'),
    ('foo@bar.com', 'foo', 'bar'),
    ('blip@blap.com', 'blip', 'blap');
