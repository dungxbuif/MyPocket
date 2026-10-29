ALTER TABLE push_subscriptions
    ADD COLUMN last_delivered_at timestamptz NOT NULL DEFAULT '-infinity';

CREATE INDEX push_subscriptions_user_delivery_idx
    ON push_subscriptions (user_id, last_delivered_at, next_attempt_at)
    WHERE failure_count < 5;
