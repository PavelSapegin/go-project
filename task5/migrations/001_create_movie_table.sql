-- +goose Up
CREATE TABLE IF NOT EXISTS movies 
(
    id INT PRIMARY KEY,
    title TEXT NOT NULL,
    year INT NOT NULL,
    director TEXT NOT NULL
);


-- +goose Down
DROP TABLE IF EXISTS movies;