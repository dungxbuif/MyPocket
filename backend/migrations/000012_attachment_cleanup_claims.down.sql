DROP INDEX IF EXISTS transaction_attachments_cleanup;
ALTER TABLE transaction_attachments DROP CONSTRAINT IF EXISTS transaction_attachments_ocr_status_check;
ALTER TABLE transaction_attachments ADD CONSTRAINT transaction_attachments_ocr_status_check CHECK (ocr_status IN ('pending','completed','failed','delete_failed','deleted'));
ALTER TABLE transaction_attachments DROP COLUMN IF EXISTS updated_at;
CREATE INDEX transaction_attachments_cleanup ON transaction_attachments(delete_after) WHERE ocr_status IN ('pending','failed','delete_failed');
