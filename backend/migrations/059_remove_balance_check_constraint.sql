-- Migration: 删除用户余额非负约束，允许透支
-- Date: 2026-03-04
-- Description:
--   删除 CHECK 约束 check_balance_non_negative
--   允许用户余额变为负数（透支）
--   依赖预检查机制（checkBalanceEligibility）在余额 <= 0 时拦截新请求
--   这确保已执行的 API 调用能够正确扣费，避免用户免费获得服务

-- 检查约束是否存在，如果存在则删除
-- 这确保了迁移的幂等性（可重复执行）
DO $$
BEGIN
    -- 尝试删除约束
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'check_balance_non_negative'
        AND conrelid = 'users'::regclass
    ) THEN
        -- 约束存在，删除之
        ALTER TABLE users DROP CONSTRAINT check_balance_non_negative;

        RAISE NOTICE 'Successfully removed check_balance_non_negative constraint';
    ELSE
        RAISE NOTICE 'Constraint check_balance_non_negative does not exist, skipping';
    END IF;
END
$$;
