CREATE TABLE transaction_attachments (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 session_id text NOT NULL,
 object_key text NOT NULL UNIQUE,
 filename text NOT NULL,
 mime_type text NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','application/pdf')),
 size_bytes bigint NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 5242880),
 sha256 text NOT NULL,
 ocr_status text NOT NULL DEFAULT 'pending' CHECK (ocr_status IN ('pending','completed','failed','deleting','delete_failed','deleted')),
 ocr_text text NOT NULL DEFAULT '',
 delete_after timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY (session_id, owner_id) REFERENCES ai_entry_sessions(id, owner_id) ON DELETE CASCADE,
 UNIQUE (id, owner_id)
);
CREATE INDEX transaction_attachments_cleanup ON transaction_attachments(delete_after) WHERE ocr_status IN ('pending','completed','failed','deleting','delete_failed');

CREATE TABLE transaction_attachment_links (
 transaction_id text NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
 attachment_id text NOT NULL REFERENCES transaction_attachments(id) ON DELETE CASCADE,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(transaction_id, attachment_id)
);
