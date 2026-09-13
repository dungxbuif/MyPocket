ALTER TABLE categories ADD COLUMN IF NOT EXISTS icon_key text NOT NULL DEFAULT 'tag';

UPDATE categories
SET is_system = COALESCE(system_key IN (
  'expense_general','expense_transfer_out','expense_interest_paid','expense_uncategorized','expense_withdrawal',
  'income_other','income_transfer_in','income_interest','income_uncategorized','income_gift',
  'debt_lend','debt_repay','debt_loan','debt_collect'
), false);
