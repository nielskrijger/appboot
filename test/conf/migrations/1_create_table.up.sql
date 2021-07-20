CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL,
    name varchar(100) NOT NULL UNIQUE,
    created_at timestamp NOT NULL,
    PRIMARY KEY (id)
);
