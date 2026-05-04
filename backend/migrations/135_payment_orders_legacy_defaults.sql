-- Continuation of 134_relax_legacy_payment_orders_not_null.sql.
-- 134 only relaxed NOT NULL on the legacy v0.1.119-deprecated columns; that
-- alone is not enough — old SQL paths (e.g. ListMyOrders) Scan() the same
-- columns into non-nullable Go strings/numerics, so a V119 INSERT that omits
-- the column writes NULL and the next legacy GET errors with
-- "converting NULL to string is unsupported".
--
-- Fix:
--   1) SET DEFAULT on each legacy column so V119 INSERTs that omit it land on
--      a typed empty value instead of NULL.
--   2) Backfill any existing NULL rows that 134 already produced.
--   3) Replace order_no's UNIQUE constraint with a partial unique index so
--      multiple V119 inserts sharing an empty default don't collide while real
--      legacy order_no values keep their uniqueness guarantee.
--
-- Same as 134, this is guarded with information_schema so fresh installs that
-- never had these columns are a no-op.

DO $$
DECLARE
    spec record;
BEGIN
    FOR spec IN
        SELECT * FROM (VALUES
            ('order_no',      ''''''),
            ('amount_cny',    '0'),
            ('amount_usd',    '0'),
            ('total_usd',     '0'),
            ('exchange_rate', '1'),
            ('provider',      ''''''),
            ('expire_at',     'NOW()')
        ) AS t(col, default_expr)
    LOOP
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'payment_orders'
              AND column_name = spec.col
        ) THEN
            EXECUTE format('ALTER TABLE payment_orders ALTER COLUMN %I SET DEFAULT %s', spec.col, spec.default_expr);
            EXECUTE format('UPDATE payment_orders SET %I = DEFAULT WHERE %I IS NULL', spec.col, spec.col);
        END IF;
    END LOOP;
END
$$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'payment_orders'
          AND column_name = 'order_no'
    ) THEN
        ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_order_no_key;
        DROP INDEX IF EXISTS payment_orders_order_no_key;
        CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_order_no_uniq
            ON payment_orders (order_no)
            WHERE order_no IS NOT NULL AND order_no <> '';
    END IF;
END
$$;
