-- Drop NOT NULL on legacy payment_orders columns that the v0.1.119 Ent schema
-- no longer references. 008_payment_order.sql created these columns as NOT
-- NULL for the pre-v0.1.119 payment system; 091a_align_legacy_payment_orders_schema.sql
-- added the new columns but left these legacy ones unchanged, so legacy
-- databases still reject INSERTs that omit them (the V119 createOrderInTx path
-- does not set them).
--
-- Fresh databases never created these columns (092_payment_orders.sql defines
-- the table without them). Each branch is guarded with information_schema so
-- this migration is a no-op on those installs.
DO $$
DECLARE
    legacy_col text;
BEGIN
    FOREACH legacy_col IN ARRAY ARRAY[
        'order_no',
        'amount_cny',
        'amount_usd',
        'total_usd',
        'exchange_rate',
        'provider',
        'expire_at'
    ]
    LOOP
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'payment_orders'
              AND column_name = legacy_col
              AND is_nullable = 'NO'
        ) THEN
            EXECUTE format('ALTER TABLE payment_orders ALTER COLUMN %I DROP NOT NULL', legacy_col);
        END IF;
    END LOOP;
END
$$;
