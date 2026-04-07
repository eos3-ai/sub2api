-- Align existing scheduled_test_plans model_id values with updated defaults.
--
-- Scope:
-- Only update rows that still use legacy default model IDs, so custom user
-- selections are not overridden.

-- 1) Anthropic: haiku4.5 -> claude-haiku-4-5-20251001
UPDATE scheduled_test_plans stp
SET model_id = 'claude-haiku-4-5-20251001',
    updated_at = NOW()
FROM accounts a
WHERE a.id = stp.account_id
  AND a.platform = 'anthropic'
  AND stp.model_id = 'haiku4.5';

-- 2) OpenAI: gpt5.3-codex -> gpt-5.3-codex
UPDATE scheduled_test_plans stp
SET model_id = 'gpt-5.3-codex',
    updated_at = NOW()
FROM accounts a
WHERE a.id = stp.account_id
  AND a.platform = 'openai'
  AND stp.model_id = 'gpt5.3-codex';
