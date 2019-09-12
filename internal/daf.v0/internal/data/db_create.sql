PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';
-- %MYAPPDATA%\daf.v0\daf.v0.sqlite

CREATE TABLE IF NOT EXISTS party
(
    party_id              INTEGER PRIMARY KEY NOT NULL,
    created_at            TIMESTAMP           NOT NULL                     DEFAULT (datetime('now')) UNIQUE,
    type                  INTEGER             NOT NULL                     DEFAULT 1 CHECK ( type > 0 ),
    component             TEXT                NOT NULL                     DEFAULT 'гексан C₆H₁₄',
    scale                 REAL                NOT NULL                     DEFAULT 1000,
    absolute_error_range  REAL                NOT NULL                     DEFAULT 200,
    absolute_error_limit  REAL                NOT NULL                     DEFAULT 50,
    relative_error_limit  REAL                NOT NULL                     DEFAULT 20,
    threshold1_production REAL                NOT NULL                     DEFAULT 200,
    threshold2_production REAL                NOT NULL                     DEFAULT 1000,
    threshold1_test       REAL                NOT NULL                     DEFAULT 200,
    threshold2_test       REAL                NOT NULL                     DEFAULT 1000,
    pgs1                  REAL                NOT NULL CHECK ( pgs1 >= 0 ) DEFAULT 0,
    pgs2                  REAL                NOT NULL CHECK ( pgs2 >= 0 ) DEFAULT 200,
    pgs3                  REAL                NOT NULL CHECK ( pgs3 >= 0 ) DEFAULT 1000,
    pgs4                  REAL                NOT NULL CHECK ( pgs4 >= 0 ) DEFAULT 2000
);

CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER PRIMARY KEY NOT NULL,
    party_id   INTEGER             NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    serial     INTEGER             NOT NULL CHECK (serial > 0 ),
    addr       SMALLINT            NOT NULL CHECK (addr > 0),

    UNIQUE (party_id, addr),
    UNIQUE (party_id, serial),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_value
(
    product_value_id INTEGER PRIMARY KEY NOT NULL,
    product_id       INTEGER             NOT NULL,
    created_at       TIMESTAMP           NOT NULL DEFAULT (datetime('now')),
    work_index       INTEGER             NOT NULL CHECK ( work_index >= 0),
    gas              INTEGER             NOT NULL CHECK ( gas IN (1, 2, 3, 4) ),
    concentration    REAL                NOT NULL,
    current          REAL                NOT NULL,
    threshold1       BOOLEAN             NOT NULL,
    threshold2       BOOLEAN             NOT NULL,
    mode             INTEGER             NOT NULL,
    failure_code     REAL                NOT NULL,
    UNIQUE (product_id, work_index),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS product_entry
(
    product_entry_id INTEGER PRIMARY KEY NOT NULL,
    product_id       INTEGER             NOT NULL,
    created_at       TIMESTAMP           NOT NULL DEFAULT (datetime('now')),
    work_name        TEXT                NOT NULL,
    ok               BOOlEAN             NOT NULL,
    message          TEXT                NOT NULL,
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS last_party AS
SELECT *
FROM party
ORDER BY created_at DESC
LIMIT 1;


CREATE VIEW IF NOT EXISTS last_party_products AS
    SELECT * FROM product WHERE party_id = (SELECT party_id FROM last_party);
