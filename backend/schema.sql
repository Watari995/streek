-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Habits
CREATE TABLE habits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    label_color VARCHAR(7) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_habits_user_id ON habits(user_id);

-- Check-ins
CREATE TABLE check_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    checked_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (habit_id, checked_date)
);

CREATE INDEX idx_check_ins_habit_id ON check_ins(habit_id);
CREATE INDEX idx_check_ins_checked_date ON check_ins(checked_date);

-- Point Ledger (台帳パターン: 残高は履歴の合計で算出)
CREATE TABLE point_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    habit_id UUID REFERENCES habits(id) ON DELETE SET NULL,
    type VARCHAR(10) NOT NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    reason VARCHAR(100) NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_point_ledger_user_id ON point_ledger(user_id);

-- Outbox events (Outboxパターン: DB変更とイベント発行の原子性担保)
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    CHECK (status IN ('PENDING', 'SENT', 'FAILED'))
);

-- ポーリング高速化のためのpartial index（PENDINGのみ）
CREATE INDEX idx_outbox_events_pending ON outbox_events (created_at) WHERE status = 'PENDING';
