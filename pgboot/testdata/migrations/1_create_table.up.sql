CREATE TABLE IF NOT EXISTS test_table (
    id SERIAL,
    name varchar(100) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);
