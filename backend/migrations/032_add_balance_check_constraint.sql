-- Migration: 添加用户余额非负约束，防止透支
-- Date: 2026-01-12
-- Description:
--   添加 CHECK 约束确保 users.balance >= 0
--   防止用户透支，提供数据库层面的最后一道防线

-- 检查约束是否已存在，如果存在则跳过
-- 这确保了迁移的幂等性（可重复执行）
DO $$
BEGIN
    -- 尝试添加约束
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'check_balance_non_negative'
        AND conrelid = 'users'::regclass
    ) THEN
        -- 约束不存在，添加之
        ALTER TABLE users ADD CONSTRAINT check_balance_non_negative CHECK (balance >= 0);

        -- 添加约束注释
        COMMENT ON CONSTRAINT check_balance_non_negative ON users IS '确保用户余额不能为负数，防止透支';

        RAISE NOTICE 'Successfully added check_balance_non_negative constraint';
    ELSE
        RAISE NOTICE 'Constraint check_balance_non_negative already exists, skipping';
    END IF;
END
$$;
