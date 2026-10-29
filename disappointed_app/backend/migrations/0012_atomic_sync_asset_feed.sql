-- Asset sync operations existed in the API but were missing from feed checks.
ALTER TABLE sync_changes DROP CONSTRAINT sync_changes_entity_type_check;
ALTER TABLE sync_changes ADD CONSTRAINT sync_changes_entity_type_check
 CHECK (entity_type IN ('wallet', 'category', 'transaction', 'asset'));
ALTER TABLE sync_changes DROP CONSTRAINT sync_changes_operation_check;
ALTER TABLE sync_changes ADD CONSTRAINT sync_changes_operation_check
 CHECK (operation IN ('create', 'update', 'archive', 'set_default_ai',
 'set_category_active', 'add_trade', 'update_trade', 'archive_trade', 'add_price'));
