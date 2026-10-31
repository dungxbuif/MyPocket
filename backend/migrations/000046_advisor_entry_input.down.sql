-- Disposable fixtures only. Never roll this back on a production database.
ALTER TABLE advisor_entry_tasks
    DROP CONSTRAINT IF EXISTS advisor_entry_tasks_input_text_check,
    DROP CONSTRAINT IF EXISTS advisor_entry_tasks_timezone_check,
    DROP COLUMN IF EXISTS input_text,
    DROP COLUMN IF EXISTS timezone;

ALTER TABLE transaction_attachments
    DROP CONSTRAINT IF EXISTS transaction_attachments_input_order_check,
    DROP CONSTRAINT IF EXISTS transaction_attachments_mime_type_check,
    DROP COLUMN IF EXISTS input_order;

ALTER TABLE transaction_attachments
    ADD CONSTRAINT transaction_attachments_mime_type_check CHECK (mime_type IN ('image/jpeg', 'image/png', 'application/pdf'));
