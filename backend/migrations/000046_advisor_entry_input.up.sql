-- Durable worker input for accepted entry tasks. The fields are private and
-- excluded from all task/event JSON responses; they let a restarted worker
-- continue without re-reading an uncertain client request.
ALTER TABLE advisor_entry_tasks
    ADD COLUMN input_text text NOT NULL DEFAULT '',
    ADD COLUMN timezone text NOT NULL DEFAULT 'UTC';

ALTER TABLE advisor_entry_tasks
    ADD CONSTRAINT advisor_entry_tasks_input_text_check CHECK (char_length(input_text) <= 16000),
    ADD CONSTRAINT advisor_entry_tasks_timezone_check CHECK (char_length(timezone) BETWEEN 1 AND 128);

ALTER TABLE transaction_attachments
    ADD COLUMN input_order integer NOT NULL DEFAULT 0,
    DROP CONSTRAINT IF EXISTS transaction_attachments_mime_type_check;

ALTER TABLE transaction_attachments
    ADD CONSTRAINT transaction_attachments_input_order_check CHECK (input_order >= 0),
    ADD CONSTRAINT transaction_attachments_mime_type_check CHECK (mime_type IN ('image/jpeg', 'image/png', 'application/pdf', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'));
