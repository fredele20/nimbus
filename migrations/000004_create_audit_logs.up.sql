CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    actor_id UUID,
    user_id UUID,

    action VARCHAR(100) NOT NULL,

    entity VARCHAR(100) NOT NULL,

    entity_id VARCHAR(255),

    metadata JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor
ON audit_logs(actor_id);

CREATE INDEX idx_audit_user
ON audit_logs(user_id);

CREATE INDEX idx_audit_action
ON audit_logs(action);
