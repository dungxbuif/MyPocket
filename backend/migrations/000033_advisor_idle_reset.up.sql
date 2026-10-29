-- Advisor history retention metadata only. Expiry is enforced by the
-- application/worker; this migration never deletes existing conversation data.
ALTER TABLE advisor_conversations
    ADD COLUMN last_user_message_at timestamptz;

-- Backfill from accepted user messages only. Assistant output and reads must
-- not extend the idle window.
UPDATE advisor_conversations AS conversations
SET last_user_message_at = latest.created_at
FROM (
    SELECT conversation_id, MAX(created_at) AS created_at
    FROM advisor_messages
    WHERE role = 'user'
    GROUP BY conversation_id
) AS latest
WHERE conversations.id = latest.conversation_id;

CREATE INDEX advisor_conversations_idle_idx
    ON advisor_conversations(last_user_message_at)
    WHERE last_user_message_at IS NOT NULL;
