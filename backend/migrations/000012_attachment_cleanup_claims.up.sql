ALTER TABLE transaction_attachments ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE transaction_attachments DROP CONSTRAINT IF EXISTS transaction_attachments_ocr_status_check;
ALTER TABLE transaction_attachments ADD CONSTRAINT transaction_attachments_ocr_status_check CHECK (ocr_status IN ('pending','completed','failed','deleting','delete_failed','deleted'));
DROP INDEX IF EXISTS transaction_attachments_cleanup;
CREATE INDEX transaction_attachments_cleanup ON transaction_attachments(delete_after, updated_at) WHERE ocr_status IN ('pending','completed','failed','deleting','delete_failed');
