-- Add tintner_name and carwasher_name to orders table
ALTER TABLE orders ADD COLUMN tintner_name VARCHAR(100) DEFAULT '';
ALTER TABLE orders ADD COLUMN carwasher_name VARCHAR(100) DEFAULT '';