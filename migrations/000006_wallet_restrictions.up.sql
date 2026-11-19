ALTER TABLE wallets
    ADD COLUMN restriction_type VARCHAR(20) NULL,
    ADD COLUMN restriction_reason TEXT NULL,
    ADD COLUMN restricted_by UUID NULL,
    ADD COLUMN restricted_at TIMESTAMPTZ NULL;

ALTER TABLE wallets
    ADD CONSTRAINT fk_wallets_restricted_by
        FOREIGN KEY (restricted_by)
        REFERENCES users(id)
        ON DELETE RESTRICT;

ALTER TABLE wallets
    ADD CONSTRAINT chk_wallets_restriction_type
        CHECK (
            restriction_type IS NULL
            OR restriction_type IN ('user', 'admin')
        );

ALTER TABLE wallets
    ADD CONSTRAINT chk_wallets_restriction_consistency
        CHECK (
            (
                status = 'frozen'
                AND restriction_type IS NOT NULL
                AND restricted_by IS NOT NULL
                AND restricted_at IS NOT NULL
            )
            OR
            (
                status IN ('active', 'closed')
                AND restriction_type IS NULL
                AND restriction_reason IS NULL
                AND restricted_by IS NULL
                AND restricted_at IS NULL
            )
        );

CREATE INDEX idx_wallets_restriction_type
    ON wallets(restriction_type)
    WHERE restriction_type IS NOT NULL;
