-- Migration: Initial schema for WhatsApp Finance Bot
-- Version: 001
-- Created: 2026-02-02

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    msisdn VARCHAR(20) UNIQUE NOT NULL,
    plan VARCHAR(20) NOT NULL DEFAULT 'FREE', -- FREE, PREMIUM, PENDING_PREMIUM
    free_tx_count INT NOT NULL DEFAULT 0,
    premium_until TIMESTAMP,
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_msisdn ON users(msisdn);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    tx_id VARCHAR(50) UNIQUE NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('INCOME', 'EXPENSE')),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    category VARCHAR(100),
    description TEXT,
    transaction_date TIMESTAMP NOT NULL,
    wa_message_id VARCHAR(100),
    ai_confidence DECIMAL(3,2),
    ai_version VARCHAR(20),
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tx_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX idx_tx_wa_msg ON transactions(wa_message_id);
CREATE INDEX idx_tx_id ON transactions(tx_id);

-- Conversation states table
CREATE TABLE IF NOT EXISTS conversation_states (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    state VARCHAR(50) NOT NULL,
    context JSONB,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_conv_user ON conversation_states(user_id);
CREATE INDEX idx_conv_expires ON conversation_states(expires_at);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id BIGINT,
    old_value JSONB,
    new_value JSONB,
    performed_by VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);

-- Admin actions table
CREATE TABLE IF NOT EXISTS admin_actions (
    id BIGSERIAL PRIMARY KEY,
    admin_msisdn VARCHAR(20) NOT NULL,
    action VARCHAR(50) NOT NULL,
    target_msisdn VARCHAR(20),
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_created ON admin_actions(created_at DESC);

-- Message deduplication table
CREATE TABLE IF NOT EXISTS message_dedup (
    wa_message_id VARCHAR(100) PRIMARY KEY,
    processed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_dedup_time ON message_dedup(processed_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
