-- Rollback tintner_name and carwasher_name
ALTER TABLE orders DROP COLUMN IF EXISTS tintner_name;
ALTER TABLE orders DROP COLUMN IF EXISTS carwasher_name;