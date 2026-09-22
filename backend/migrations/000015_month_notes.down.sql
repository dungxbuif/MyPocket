DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM month_notes) THEN
    RAISE EXCEPTION 'Refusing to drop user month notes';
  END IF;
END $$;

DROP TABLE IF EXISTS month_notes;
