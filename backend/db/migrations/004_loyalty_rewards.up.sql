-- Add points_required and is_reward columns to products table
ALTER TABLE products 
ADD COLUMN points_required INT NOT NULL DEFAULT 0,
ADD COLUMN is_reward TINYINT(1) NOT NULL DEFAULT 0;