ALTER TABLE orders
ADD COLUMN tintner_name VARCHAR(255) AFTER mechanic_name,
ADD COLUMN carwasher_name VARCHAR(255) AFTER tintner_name;