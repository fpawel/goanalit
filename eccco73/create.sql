PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';
--E:\Program Data\Аналитприбор\elchese\elchese.sqlite

CREATE TABLE IF NOT EXISTS gas (
  gas_name TEXT PRIMARY KEY,
  code INTEGER UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS units (
  units_name TEXT PRIMARY KEY,
  code INTEGER UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS product_type (
  product_type_id INTEGER PRIMARY KEY,
  name                TEXT    NOT NULL,
  gas_name            TEXT    NOT NULL,
  units_name          TEXT    NOT NULL,
  scale               REAL    NOT NULL,
  noble_metal_content REAL    NOT NULL,
  lifetime_months     INTEGER NOT NULL CHECK (lifetime_months > 0),
  lc64                BOOLEAN NOT NULL
    CONSTRAINT boolean_lc64
    CHECK (lc64 IN (0, 1)),
  points_method       INTEGER NOT NULL
    CONSTRAINT points_method_2_or_3
    CHECK (points_method IN (2, 3)),
  max_fon             REAL,
  max_d_fon           REAL,
  min_k_sens20        REAL,
  max_k_sens20        REAL,
  min_d_temp          REAL,
  max_d_temp          REAL,
  min_k_sens50        REAL,
  max_k_sens50        REAL,
  max_d_not_measured  REAL,
  FOREIGN KEY (gas_name) REFERENCES gas (gas_name),
  FOREIGN KEY (units_name) REFERENCES units (units_name)
);

CREATE TABLE IF NOT EXISTS party (
  party_id INTEGER PRIMARY KEY,
  old_party_id          TEXT,
  created_at      TIMESTAMP NOT NULL DEFAULT current_timestamp,
  product_type_id INTEGER   NOT NULL,
  conc1           REAL      NOT NULL DEFAULT 0 CHECK (conc1 >= 0),
  conc2           REAL      NOT NULL DEFAULT 0 CHECK (conc2 >= 0),
  conc3           REAL      NOT NULL DEFAULT 0 CHECK (conc3 >= 0),
  note            TEXT,
  FOREIGN KEY (product_type_id) REFERENCES product_type (product_type_id)
);

CREATE TABLE IF NOT EXISTS product (
  product_id INTEGER PRIMARY KEY,
  party_id        INTEGER NOT NULL,
  serial          INTEGER,
  place           INTEGER NOT NULL CHECK (place >= 0),
  product_type_id INTEGER,
  note            TEXT,
  i13             REAL,
  i24             REAL,
  i35             REAL,
  i26             REAL,
  i17             REAL,
  not_measured    REAL,
  flash           BLOB,
  production      BOOLEAN CHECK (production IN (0, 1)),

  old_product_id TEXT,
  old_serial      INTEGER,

  CONSTRAINT unique_party_place UNIQUE (party_id, place),
  CONSTRAINT unique_party_serial UNIQUE (party_id, serial),

  FOREIGN KEY (product_type_id) REFERENCES product_type (product_type_id),
  FOREIGN KEY (party_id) REFERENCES party (party_id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_current (
  product_id  INTEGER NOT NULL,
  scale       TEXT CHECK (scale IN ('f', 's')),
  temperature REAL    NOT NULL,
  current     REAL    NOT NULL,
  CONSTRAINT product_current_primary_key UNIQUE (product_id, scale, temperature),
  FOREIGN KEY (product_id) REFERENCES product (product_id)
    ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS party_year AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year
  FROM party
  ORDER BY year;

CREATE VIEW IF NOT EXISTS party_year_month AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year,
                  cast(strftime('%m', created_at) AS INTEGER) AS month
  FROM party
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS party_year_month_day AS
  SELECT DISTINCT cast(strftime('%Y', created_at) AS INTEGER) AS year,
                  cast(strftime('%m', created_at) AS INTEGER) AS month,
                  cast(strftime('%d', created_at) AS INTEGER) AS day
  FROM party
  ORDER BY created_at;

CREATE VIEW IF NOT EXISTS current_party AS
  SELECT *
  FROM party
  ORDER BY created_at DESC
  LIMIT 1;

CREATE VIEW IF NOT EXISTS product_info1 AS
  SELECT p.product_id,
         (CASE (p.product_type_id ISNULL)
            WHEN 1 THEN pa.product_type_id
            WHEN 0 THEN p.product_type_id END)                           AS product_type_id,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 'f'
                                                  AND temperature = 20)  AS fon20,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 's'
                                                  AND temperature = 20)  AS sens20,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 'f'
                                                  AND temperature = -20) AS fon_minus_20,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 's'
                                                  AND temperature = -20) AS sens_minus_20,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 'f'
                                                  AND temperature = 50)  AS fon50,
         (SELECT current FROM product_current a WHERE a.product_id = p.product_id
                                                  AND scale = 's'
                                                  AND temperature = 50)  AS sens50,
         conc3,
         conc1,
         i13,
         not_measured
  FROM product p
         INNER JOIN party pa on p.party_id = pa.party_id;


CREATE VIEW IF NOT EXISTS product_info2 AS
  SELECT p.product_id,
         p.product_type_id,
         (100 * (sens50 - fon50) / (sens20 - fon20)) AS k_sens50,
         (sens20 - fon20) / (conc3 - conc1)          AS k_sens20,
         i13 - fon20                                 AS d_fon20,
         fon50 - fon20                               AS d_fon_temp,
         not_measured - fon20                        AS d_not_measured
  FROM product_info1 p
         INNER JOIN product_type a ON a.product_type_id = p.product_type_id;

CREATE VIEW IF NOT EXISTS product_info3 AS
  SELECT p.product_id,
         d_not_measured < max_d_not_measured                 AS ok_d_not_measured,
         d_fon_temp < max_d_temp                             AS ok_d_fon_temp,
         abs(d_fon20) < max_d_fon                            AS ok_d_fon20,
         k_sens20 < max_k_sens20 AND k_sens20 > min_k_sens20 AS ok_k_sens20,
         k_sens50 < max_k_sens50 AND k_sens50 > min_k_sens50 AS ok_k_sens50,
         fon20 < max_fon                                     AS ok_fon20

  FROM product_info2 p
         INNER JOIN product_type a ON a.product_type_id = p.product_type_id
         INNER JOIN product_info1 b ON b.product_id = p.product_id;

CREATE VIEW IF NOT EXISTS product_info4 AS
  SELECT product_id,
         NOT(ok_d_not_measured AND ok_d_fon_temp AND ok_d_fon20 AND ok_k_sens20 AND ok_k_sens50 AND ok_fon20) AS not_ok

  FROM product_info3;

CREATE VIEW IF NOT EXISTS product_info AS
  SELECT p.product_id,
         e.party_id   AS party_id,
         a.name AS product_type_name,
         b.old_product_id,
         old_serial,
         serial,
         place,
         b.note,
         gas_name,
         units_name,
         scale,
         noble_metal_content,
         lifetime_months,
         lc64,
         points_method,
         fon20,
         ok_fon20,
         sens20,
         k_sens20,
         ok_k_sens20,
         fon_minus_20,
         sens_minus_20,
         fon50,
         d_fon_temp,
         ok_d_fon_temp,
         sens50,
         k_sens50,
         ok_k_sens50,
         e.conc1,
         e.conc3,
         b.not_measured,
         d_not_measured,
         ok_d_not_measured,
         b.i13,
         d_fon20,
         ok_d_fon20,
         i24,
         i35,
         i26,
         i17,
         flash,
         production,
         not_ok

  FROM product_info1 p
         INNER JOIN product_type a ON a.product_type_id = p.product_type_id
         INNER JOIN product_info2 c ON c.product_id = p.product_id
         INNER JOIN product_info3 d ON d.product_id = p.product_id
         INNER JOIN product_info4 f ON f.product_id = p.product_id
         INNER JOIN product b ON b.product_id = p.product_id
         INNER JOIN party e ON e.party_id = b.party_id;


INSERT
OR REPLACE INTO units (units_name, code)
VALUES ('мг/м3', 2),
       ('ppm', 3),
       ('об. дол. %', 7),
       ('млн-1', 5);

INSERT
OR REPLACE INTO gas (gas_name, code)
VALUES ('CO', 0x11),
       ('H₂S', 0x22),
       ('NH₃', 0x33),
       ('Cl₂', 0x44),
       ('SO₂', 0x55),
       ('NO₂', 0x66),
       ('O₂', 0x88),
       ('NO', 0x99),
       ('HCl', 0xAA);

INSERT
OR REPLACE INTO product_type (name,
                              gas_name,
                              units_name,
                              scale,
                              noble_metal_content,
                              lifetime_months,
                              lc64,
                              points_method,
                              max_fon,
                              max_d_fon,
                              min_k_sens20,
                              max_k_sens20,
                              min_d_temp,
                              max_d_temp,
                              min_k_sens50,
                              max_k_sens50,
                              max_d_not_measured)
VALUES ('035', 'CO', 'мг/м3', 200, 0.1626, 18, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035', 'CO', 'мг/м3', 200, 0.1456, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-59', 'CO', 'об. дол. %', 0.5, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-60', 'CO', 'мг/м3', 200, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-61', 'CO', 'ppm', 2000, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-80', 'CO', 'мг/м3', 200, 0.1456, 12, 0, 3, 1.01, 3, 0.08, 0.175, 0, 2, 100, 135, 5),
       ('035-81', 'CO', 'мг/м3', 1500, 0.1456, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-92', 'CO', 'об. дол. %', 0.5, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-93', 'CO', 'млн-1', 200, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-94', 'CO', 'млн-1', 2000, 0.1891, 12, 0, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-105', 'CO', 'мг/м3', 200, 0.1456, 12, 0, 3, 1.51, 3, 0.08, 0.335, 0, 3, 110, 145.01, NULL),
       ('100', 'CO', 'мг/м3', 200, 0.0816, 12, 1, 3, 1, 3, 0.08, 0.175, 0, 3, 100, 135, NULL),
       ('100-05', 'CO', 'мг/м3', 50, 0.0816, 12, 1, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('100-10', 'CO', 'мг/м3', 200, 0.0816, 12, 1, 3, 1, 3, 0.08, 0.175, 0, 3, 100, 135, NULL),
       ('100-15', 'CO', 'мг/м3', 50, 0.0816, 12, 1, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-40', 'CO', 'мг/м3', 200, 0.1456, 12, 0, 2, 1.51, 3, 0.08, 0.335, 0, 3, 110, 145.1, NULL),
       ('035-21', 'CO', 'мг/м3', 200, 0.1456, 12, 0, 2, 1.51, 3, 0.08, 0.335, 0, 3, 110, 145.01, NULL),
       ('130-01', 'CO', 'мг/м3', 200, 0.1626, 12, 0, 3, 1.01, 3, 0.08, 0.18, 0, 2, 100, 135, 5),
       ('035-70', 'CO', 'мг/м3', 200, 0.1626, 12, 0, 2, 1.51, 3, 0.08, 0.33, 0, 3, 110, 146, NULL),
       ('130-08', 'CO', 'ppm', 100, 0.1162, 12, 0, 3, 1, 3, 0.08, 0.18, 0, 2, 100, 135, 5),
       ('035-117', 'NO₂', 'мг/м3', 200, 0.1626, 18, 1, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('010-18', 'O₂', 'об. дол. %', 21, 0, 12, 1, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('010-18', 'O₂', 'об. дол. %', 21, 0, 12, 1, 3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
       ('035-111', 'CO', 'мг/м3', 200, 0.1626, 12, 1, 3, 1, 3, 0.08, 0.175, 0, 3, 100, 135, 5);