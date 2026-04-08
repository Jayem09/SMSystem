-- Remove points_required and is_reward columns from products table
ALTER TABLE products 
DROP COLUMN points_required,
DROP COLUMN is_reward;