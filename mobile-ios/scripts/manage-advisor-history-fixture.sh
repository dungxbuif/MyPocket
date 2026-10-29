#!/usr/bin/env bash
set -euo pipefail

# Disposable local history only; never accept caller-supplied DB/owner/IDs.
case "$(docker context inspect --format '{{(index .Endpoints "docker").Host}}')" in
  unix://*) ;;
  *) echo 'Refusing non-local Docker endpoint' >&2; exit 1 ;;
esac
action="${1:-}"
case "$action" in
  create)
    docker exec -i mypocket-native-feedback-pg-20261013 psql -v ON_ERROR_STOP=1 -U fixture -d mypocket_feedback <<'SQL'
BEGIN;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM public."user" WHERE id='e4aedf4a-1aa7-4918-ba09-0287c95b154c' AND email='fixture@example.com')
     OR EXISTS (SELECT 1 FROM advisor_conversations WHERE owner_id='e4aedf4a-1aa7-4918-ba09-0287c95b154c') THEN
    RAISE EXCEPTION 'Fixture owner absent or already has history; no overwrite';
  END IF;
END $$;
INSERT INTO advisor_conversations(id,owner_id,next_message_seq)
VALUES ('native-history-e2e-20261015-e8080c47','e4aedf4a-1aa7-4918-ba09-0287c95b154c',52);
INSERT INTO advisor_runs(id,owner_id,conversation_id,generation,client_request_id,payload_hash,credential_kind,credential_id,status,finished_at)
SELECT 'native-history-e2e-20261015-e8080c47-run-'||n,'e4aedf4a-1aa7-4918-ba09-0287c95b154c',
       'native-history-e2e-20261015-e8080c47',1,'native-history-e2e-20261015-e8080c47-request-'||n,
       'synthetic-history-only-no-provider','session','synthetic-history-only','completed',now()
FROM generate_series(1,26) n;
INSERT INTO advisor_messages(id,conversation_id,generation,seq,run_id,role,parts)
SELECT 'native-history-e2e-20261015-e8080c47-message-'||n,'native-history-e2e-20261015-e8080c47',1,n,
       'native-history-e2e-20261015-e8080c47-run-'||((n+1)/2),
       CASE WHEN n%2=1 THEN 'user' ELSE 'assistant' END,
       jsonb_build_array(jsonb_build_object('type','text','text','Native History Fixture '||lpad(n::text,2,'0')))
FROM generate_series(1,51) n;
COMMIT;
SELECT count(*) AS synthetic_messages FROM advisor_messages WHERE conversation_id='native-history-e2e-20261015-e8080c47';
SQL
    ;;
  cleanup)
    docker exec -i mypocket-native-feedback-pg-20261013 psql -v ON_ERROR_STOP=1 -U fixture -d mypocket_feedback <<'SQL'
BEGIN;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM advisor_conversations WHERE id='native-history-e2e-20261015-e8080c47') THEN RETURN; END IF;
  IF NOT EXISTS (SELECT 1 FROM advisor_conversations WHERE id='native-history-e2e-20261015-e8080c47' AND owner_id='e4aedf4a-1aa7-4918-ba09-0287c95b154c')
     OR (SELECT count(*) FROM advisor_messages WHERE conversation_id='native-history-e2e-20261015-e8080c47')<>51
     OR (SELECT count(*) FROM advisor_runs WHERE conversation_id='native-history-e2e-20261015-e8080c47')<>26
     OR EXISTS (SELECT 1 FROM advisor_messages WHERE conversation_id='native-history-e2e-20261015-e8080c47' AND (id<>'native-history-e2e-20261015-e8080c47-message-'||seq OR parts<>jsonb_build_array(jsonb_build_object('type','text','text','Native History Fixture '||lpad(seq::text,2,'0'))))) THEN
    RAISE EXCEPTION 'Unexpected fixture changes; refusing cleanup';
  END IF;
END $$;
DELETE FROM advisor_conversations WHERE id='native-history-e2e-20261015-e8080c47' AND owner_id='e4aedf4a-1aa7-4918-ba09-0287c95b154c';
COMMIT;
SELECT count(*) AS remaining_conversations FROM advisor_conversations WHERE id='native-history-e2e-20261015-e8080c47';
SQL
    ;;
  *) echo 'Usage: manage-advisor-history-fixture.sh create|cleanup' >&2; exit 2 ;;
esac
