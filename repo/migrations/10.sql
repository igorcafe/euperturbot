-- enable regular users to create topics when /suba'ing
ALTER TABLE chat ADD COLUMN enable_create_topics INTEGER NOT NULL DEFAULT 0;