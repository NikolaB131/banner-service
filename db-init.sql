CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(32) NOT NULL UNIQUE CHECK (username <> ''),
  password_hash BYTEA NOT NULL,
  role VARCHAR(16) NOT NULL DEFAULT 'user',
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE features (
  id SERIAL PRIMARY KEY
);

CREATE TABLE tags (
  id SERIAL PRIMARY KEY
);

CREATE TABLE banners (
  id SERIAL PRIMARY KEY,
  feature_id INT NOT NULL REFERENCES features(id),
  content JSONB NOT NULL,
  is_active BOOLEAN NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TRIGGER banners_update_timestamp
BEFORE UPDATE ON banners
FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at();

CREATE TABLE banner_tags (
  banner_id INT NOT NULL REFERENCES banners(id) ON DELETE CASCADE,
  tag_id INT NOT NULL REFERENCES tags(id),
  PRIMARY KEY (banner_id, tag_id)
);

-- Add initial mock features and tags
INSERT INTO features (id) VALUES (10), (11), (12), (13), (14), (15), (16), (17), (18), (19);
INSERT INTO tags (id) VALUES (20), (21), (22), (23), (24), (25), (26), (27), (28), (29);
