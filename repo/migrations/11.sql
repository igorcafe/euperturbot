-- enable regular users to create topics when /suba'ing
ALTER TABLE chat ADD COLUMN enable_ask INTEGER NOT NULL DEFAULT 0;