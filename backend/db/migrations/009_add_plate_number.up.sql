-- Add plate_number to orders table
ALTER TABLE orders ADD COLUMN plate_number VARCHAR(50) DEFAULT NULL;