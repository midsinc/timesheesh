ALTER TABLE time_entries ADD COLUMN billing_code_id INTEGER REFERENCES billing_codes(id);
