-- Remove manually added columns from recharge_records table
-- These columns were added to the database but not defined in the code,
-- causing INSERT operations to fail with NOT NULL constraint violations.
-- This migration removes all fields that exist in the database but not in the code model.

-- Remove record_no (if it exists)
ALTER TABLE recharge_records
    DROP COLUMN IF EXISTS record_no;

-- Remove currency (NOT NULL constraint causing errors)
ALTER TABLE recharge_records
    DROP COLUMN IF EXISTS currency;

-- Remove source (NOT NULL constraint causing errors)
ALTER TABLE recharge_records
    DROP COLUMN IF EXISTS source;

-- Remove order_id (nullable but not in code model)
ALTER TABLE recharge_records
    DROP COLUMN IF EXISTS order_id;

-- Remove notes (nullable but not in code model)
ALTER TABLE recharge_records
    DROP COLUMN IF EXISTS notes;
