CREATE TABLE advisor_reset_requests (
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    client_request_id text NOT NULL,
    payload_hash text NOT NULL,
    conversation_id text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_id, client_request_id),
    FOREIGN KEY (conversation_id, owner_id)
        REFERENCES advisor_conversations(id, owner_id) ON DELETE CASCADE
);

CREATE INDEX advisor_reset_requests_created_idx
    ON advisor_reset_requests(owner_id, created_at DESC);
