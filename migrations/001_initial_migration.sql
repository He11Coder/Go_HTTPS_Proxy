CREATE SCHEMA request_data;
SET search_path TO request_data;

CREATE TABLE IF NOT EXISTS response (
    id serial PRIMARY KEY CONSTRAINT id_is_positive CHECK (id > 0),
    status_code smallint NOT NULL CONSTRAINT status_is_positive CHECK (status_code > 0),
    header json DEFAULT NULL,
    cookie bytea DEFAULT NULL,
    body bytea DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS request (
    id serial PRIMARY KEY CONSTRAINT id_is_positive CHECK (id > 0),
    response_id int REFERENCES response ON DELETE CASCADE,
    "url" TEXT NOT NULL CONSTRAINT url_is_not_empty CHECK (length("url") > 0),
    method TEXT NOT NULL CONSTRAINT method_is_not_empty CHECK (length(method) > 0),
    "path" TEXT DEFAULT NULL,
    "query_param" TEXT DEFAULT NULL,
    header json DEFAULT NULL,
    cookie bytea DEFAULT NULL,
    body bytea DEFAULT NULL
);