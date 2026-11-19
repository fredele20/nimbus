DROP INDEX IF EXISTS idx_wallets_restriction_type;

ALTER TABLE wallets
    DROP CONSTRAINT IF EXISTS chk_wallets_restriction_consistency,
    DROP CONSTRAINT IF EXISTS chk_wallets_restriction_type,
    DROP CONSTRAINT IF EXISTS fk_wallets_restricted_by;

ALTER TABLE wallets
    DROP COLUMN IF EXISTS restricted_at,
    DROP COLUMN IF EXISTS restricted_by,
    DROP COLUMN IF EXISTS restriction_reason,
    DROP COLUMN IF EXISTS restriction_type;
