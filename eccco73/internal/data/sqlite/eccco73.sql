PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE gases (
  gas TEXT PRIMARY KEY,
  code INTEGER UNIQUE NOT NULL
);

CREATE TABLE units (
  units TEXT PRIMARY KEY,
  code INTEGER UNIQUE NOT NULL
);

INSERT INTO units (units, code ) VALUES  ('мг/м3', 2),  ('ppm', 3), ('об. дол. %', 7), ('млн-1',5);

INSERT INTO gases (gas, code ) VALUES
  ('CO', 0x11),   ('H2S', 0x22),    ('NH3', 0x33),    ('Cl2', 0x44),
  ('SO2', 0x55),  ('NO2', 0x66),    ('O2', 0x88),     ('NO', 0x99),   ('HCl', 0xAA);

CREATE TABLE product_types (
  product_type_id INTEGER PRIMARY KEY,
  product_type_name TEXT NOT NULL,
  gas TEXT NOT NULL,
  units TEXT NOT NULL,
  scale REAL NOT NULL,
  noble_metal_content REAL NOT NULL,
  lifetime_months  INTEGER NOT NULL,
  lc64 BOOLEAN NOT NULL
    CONSTRAINT boolean_lc64
    CHECK (lc64 = 0 OR lc64 = 1),
  points_method INTEGER NOT NULL
    CONSTRAINT points_method_2_or_3
    CHECK (points_method = 2 OR points_method = 3),
  max_fon_curr REAL,
  max_delta_fon_curr REAL,
  min_coefficient_sens REAL,
  max_coefficient_sens REAL,
  min_delta_temperature REAL,
  max_delta_temperature REAL,
  min_coefficient_sens40 REAL,
  max_coefficient_sens40 REAL,
  max_delta_not_measured REAL,
  FOREIGN KEY(gas) REFERENCES gases(gas),
  FOREIGN KEY(units) REFERENCES units(units)
);

CREATE TABLE parties (
  party_id INTEGER PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
  product_type_id INTEGER NOT NULL,
  gas1 REAL NOT NULL DEFAULT 0 CONSTRAINT positive_gas1 CHECK (gas1 >= 0),
  gas2 REAL NOT NULL DEFAULT 0 CONSTRAINT positive_gas1 CHECK (gas2 >= 0),
  gas3 REAL NOT NULL DEFAULT 0 CONSTRAINT positive_gas1 CHECK (gas3 >= 0),
  note TEXT DEFAULT NULL,
  FOREIGN KEY(product_type_id) REFERENCES product_types(product_type_id) ON DELETE CASCADE
);

CREATE TABLE import_parties (
  import_id TEXT PRIMARY KEY,
  party_id INTEGER NOT NULL UNIQUE,
  FOREIGN KEY(party_id) REFERENCES parties(party_id)
);



CREATE TABLE products (
  product_id INTEGER PRIMARY KEY,
  party_id INTEGER NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
  serial_number INTEGER NOT NULL DEFAULT 0,
  order_in_party INTEGER NOT NULL,
  product_type_id INTEGER DEFAULT NULL,
  note TEXT DEFAULT NULL,
  fon20 REAL,
  sens20 REAL,
  i13 REAL,
  i24 REAL,
  i35 REAL,
  i26 REAL,
  i17 REAL,
  not_measured REAL,
  fon_minus20 REAL,
  sens_minus20 REAL,
  fon50 REAL,
  sens50 REAL,
  flash BLOB,
  production BOOLEAN NOT NULL
    CONSTRAINT boolean_production
    CHECK (production = 0 OR production = 1),

  CONSTRAINT unique_order_in_party UNIQUE (party_id, order_in_party),
  CONSTRAINT positive_product_number CHECK (serial_number > 0),
  CONSTRAINT not_negative_order_in_party CHECK (order_in_party > 0 OR order_in_party = 0),

  FOREIGN KEY(product_type_id) REFERENCES product_types(product_type_id) ON DELETE CASCADE,
  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE
);



INSERT INTO product_types
(product_type_name, gas, units, scale, noble_metal_content, lifetime_months, lc64, points_method, max_fon_curr, max_delta_fon_curr, min_coefficient_sens, max_coefficient_sens, min_delta_temperature, max_delta_temperature, min_coefficient_sens40, max_coefficient_sens40, max_delta_not_measured)
VALUES
  ('035',     'CO', 'мг/м3',      200,  0.1626, 18, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035',     'CO', 'мг/м3',      200,  0.1456, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-59',  'CO', 'об. дол. %', 0.5,  0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-60',  'CO', 'мг/м3',      200,  0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-61',  'CO', 'ppm',        2000, 0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-80',  'CO', 'мг/м3',      200,  0.1456, 12, 0,  3, 1.01, 3,    0.08, 0.175,0,    2,    100, 135,   5),
  ('035-81',  'CO', 'мг/м3',      1500, 0.1456, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-92',  'CO', 'об. дол. %', 0.5,  0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-93',  'CO', 'млн-1',      200,  0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-94',  'CO', 'млн-1',      2000, 0.1891, 12, 0,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-105', 'CO', 'мг/м3',      200,  0.1456, 12, 0,  3, 1.51, 3,    0.08, 0.335,0,    3,    110,  145.01,NULL),
  ('100',     'CO', 'мг/м3',      200,  0.0816, 12, 1,  3, 1,    3,    0.08, 0.175,0,    3,    100,  135,  NULL),
  ('100-05',  'CO', 'мг/м3',      50,   0.0816, 12, 1,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('100-10',  'CO', 'мг/м3',      200,  0.0816, 12, 1,  3, 1,    3,    0.08, 0.175,0,    3,    100,  135,  NULL),
  ('100-15',  'CO', 'мг/м3',      50,   0.0816, 12, 1,  3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-40',  'CO', 'мг/м3',      200,  0.1456, 12, 0,  2, 1.51, 3,    0.08, 0.335,0,    3,    110,  145.1, NULL),
  ('035-21',  'CO', 'мг/м3',      200,  0.1456, 12, 0,  2, 1.51, 3,    0.08, 0.335,0,    3,    110,  145.01, NULL),
  ('130-01',  'CO', 'мг/м3',      200,  0.1626, 12, 0,  3, 1.01, 3,    0.08, 0.18, 0,    2,    100,  135, 5),
  ('035-70',  'CO', 'мг/м3',      200,  0.1626, 12, 0,  2, 1.51, 3,    0.08, 0.33, 0,    3,    110,  146, NULL),
  ('130-08',  'CO', 'ppm',        100,  0.1162, 12, 0,  3, 1, 3, 0.08, 0.18, 0, 2, 100,  135,  5),
  ('035-117', 'NO2', 'мг/м3',     200,  0.1626, 18, 1,   3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('010-18',  'O2', 'об. дол. %', 21,   0,      12, 1,   3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('010-18',  'O2', 'об. дол. %', 21,   0,      12, 1,   3, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('035-111', 'CO', 'мг/м3',      200,  0.1626, 12, 1,   3, 1,    3,    0.08, 0.175,0,    3,    100,  135,  5);

INSERT INTO parties (product_type_id) VALUES (1), (2), (2);

DELETE FROM parties;


select datetime ('2017-01-01 23:50');

INSERT  INTO products (party_id, order_in_party)
  SELECT (SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1), 1
;

SELECT product_type_id, product_type_name FROM product_types;
DELETE FROM import_parties;
DELETE FROM products;
DELETE FROM parties;
DELETE FROM product_types;



DROP TABLE products ;
DROP TABLE import_parties ;
DROP TABLE parties;
DROP TABLE product_types;

SELECT last_insert_rowid();

SELECT created_at, party_id FROM parties ORDER BY created_at DESC LIMIT 1;
SELECT * FROM product_types;
SELECT min(order_in_party), max(order_in_party) FROM products ;

SELECT * FROM parties ORDER BY created_at DESC LIMIT 1;


SELECT * FROM products
WHERE party_id IN (
  SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1
);

SELECT * FROM products where product_id = 15;


SELECT strftime('%Y', created_at) AS year FROM parties GROUP BY year;

SELECT strftime('%m', created_at) AS month FROM parties
WHERE strftime('%Y', created_at) = '2018'
GROUP BY month;

SELECT strftime('%d', created_at) AS day FROM parties
WHERE strftime('%Y.%m', created_at) = '2016.02'
GROUP BY day;


SELECT * FROM parties
WHERE strftime('%Y.%m.%d', created_at) = '2016.02.09';

UPDATE products set flash = '' WHERE  flash = 'NULL' ;

select * from products where flash = 'NULL' ;



PRAGMA foreign_keys = ON;


CREATE TABLE works(
  work_id INTEGER NOT NULL PRIMARY KEY,
  parent_id INTEGER,
  root_id INTEGER,

  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,

  ordinal INTEGER NOT NULL,
  name TEXT NOT NULL,
  CONSTRAINT parent_id_foreign_key FOREIGN KEY(parent_id) REFERENCES works(work_id) ON DELETE CASCADE,
  CONSTRAINT root_id_foreign_key FOREIGN KEY(root_id) REFERENCES works(work_id) ON DELETE CASCADE,
  CONSTRAINT unique_root_id_ordinal UNIQUE (root_id, ordinal),
  CONSTRAINT positive_ordinal CHECK ( ordinal = 0 OR ordinal > 0  ),
  CONSTRAINT valid_ordinal CHECK (
    (ordinal = 0) AND (parent_id IS NULL) AND (root_id IS NULL) OR
    (ordinal > 0) AND (parent_id IS NOT NULL ) AND (root_id IS NOT NULL)
  )

);

CREATE TABLE journal (
  created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
  party_id INTEGER,
  product_id INTEGER,
  work_id INTEGER,
  level INTEGER NOT NULL DEFAULT 0,
  message TEXT NOT NULL,
  FOREIGN KEY(work_id) REFERENCES works(work_id) ON DELETE CASCADE,
  FOREIGN KEY(party_id) REFERENCES parties(party_id) ON DELETE CASCADE,
  FOREIGN KEY(product_id) REFERENCES products(product_id) ON DELETE CASCADE
);

drop table journal;

INSERT INTO journal (party_id, work_id, level, message)
VALUES
  ( (SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1),
    (SELECT work_id FROM works WHERE ordinal = 1 ORDER BY created_at DESC LIMIT 1), 0, 'message1' );

INSERT INTO journal (party_id, work_id, level, message)
VALUES
  ( (SELECT party_id FROM parties ORDER BY created_at DESC LIMIT 1),
    (SELECT work_id FROM works WHERE ordinal = 5 ORDER BY created_at DESC LIMIT 1), 0, 'message5' );


DELETE from works;

WITH current_root(work_id)
AS (
    SELECT w.work_id FROM works w WHERE w.root_id ISNULL ORDER BY created_at DESC LIMIT 1
)
SELECT count(*) FROM works
WHERE  root_id ISNULL OR root_id IN current_root;

WITH RECURSIVE xs AS (
  SELECT w.work_id, w.name, w.ordinal
  FROM works w
  WHERE w.ordinal = 1
  UNION
  SELECT w.work_id, w.name, w.ordinal
  FROM xs
    INNER JOIN works w ON w.parent_id = xs.work_id
)
SELECT
  xs.work_id, xs.ordinal, xs.name,
  j.message, j.created_at
FROM xs
  INNER JOIN journal j ON xs.work_id = j.work_id
ORDER BY j.created_at ASC;


delete from works;

WITH
    current_root(work_id) AS (SELECT work_id FROM works WHERE root_id ISNULL ORDER BY created_at DESC LIMIT 1),
    next_ordinal AS (
      SELECT count(*) FROM works
      WHERE  root_id ISNULL OR root_id IN current_root
  )
INSERT INTO works (root_id,  ordinal, parent_id, name )
VALUES
  ( ( SELECT * FROM current_root), ( SELECT * FROM next_ordinal), 0, 'ghjled' );


WITH
    current_root(work_id) AS (
      SELECT work_id
      FROM works
      WHERE root_id ISNULL
      ORDER BY created_at DESC LIMIT 1),
    next_ordinal AS (
      SELECT count(*) FROM works
      WHERE  root_id ISNULL OR
             root_id IN current_root
  )
INSERT INTO works (root_id,  ordinal, parent_id, name )
VALUES
  (
    ( SELECT work_id FROM current_root),
    ( SELECT * FROM next_ordinal),
    ( SELECT work_id
      FROM works
      WHERE ordinal=0 AND
            root_id ISNULL OR
            root_id IN current_root
    ), 'sfd' );

INSERT INTO works (name, ordinal) VALUES  ( 'Настройка ЭХЯ', 0);

select * from works;
select * from journal;
delete from works;
delete from journal;
drop table works;

SELECT count(*) FROM works
WHERE  root_id ISNULL OR root_id IN (SELECT work_id FROM works WHERE root_id ISNULL ORDER BY created_at DESC LIMIT 1);

SELECT work_id FROM (SELECT work_id FROM works WHERE root_id ISNULL ORDER BY created_at DESC LIMIT 1);



WITH RECURSIVE xs(work_id, name) AS (
  SELECT work_id, name
  FROM works
  WHERE work_id = 11
  UNION
  SELECT w.work_id, w.name
  FROM xs
    INNER JOIN works w ON w.parent_id = xs.work_id
)
SELECT * FROM xs;

WITH RECURSIVE xs(work_id, name) AS (
  SELECT work_id, name
  FROM works
  WHERE work_id = 11
  UNION
  SELECT w.work_id, w.name
  FROM xs
    INNER JOIN works w ON w.parent_id = xs.work_id
)
SELECT j.created_at, j.message, j.level
FROM xs
  INNER JOIN journal j ON xs.work_id = j.work_id;