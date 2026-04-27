DROP TRIGGER IF EXISTS jobs_updated_at ON jobs;
DROP FUNCTION IF EXISTS set_updated_at();
ALTER TABLE jobs DROP COLUMN updated_at;
