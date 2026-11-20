CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT fk_wallets_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT uq_wallets_user_currency
        UNIQUE (user_id, currency),

    CONSTRAINT chk_wallets_currency
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT chk_wallets_status
        CHECK (status IN ('active', 'frozen', 'closed'))
);

CREATE INDEX idx_wallets_status
    ON wallets(status);

CREATE INDEX idx_wallets_deleted_at
    ON wallets(deleted_at);
