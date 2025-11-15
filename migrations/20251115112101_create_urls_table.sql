-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls
(
    id            SERIAL PRIMARY KEY,
    url           TEXT         NOT NULL,
    http_method   VARCHAR(255)       NOT NULL,
    status        VARCHAR(255)  NOT NULL,
    monitor_type  VARCHAR(255)  NOT NULL,
    contact_email VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd
