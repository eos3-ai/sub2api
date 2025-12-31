-- Add remark to payment orders for user-visible notes (e.g., referral reward, promotion bonus)
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS remark VARCHAR(255) NOT NULL DEFAULT '';

