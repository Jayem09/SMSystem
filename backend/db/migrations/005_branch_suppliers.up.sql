CREATE TABLE IF NOT EXISTS branch_suppliers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    branch_id INTEGER NOT NULL,
    supplier_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, supplier_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_supplier_branch ON branch_suppliers(branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_supplier_supplier ON branch_suppliers(supplier_id);

ALTER TABLE purchase_orders ADD COLUMN branch_id INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_purchase_orders_branch ON purchase_orders(branch_id);
