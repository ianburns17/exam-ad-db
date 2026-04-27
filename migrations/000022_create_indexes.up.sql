CREATE INDEX idx_jobs_status_created ON jobs (status, created_at);
CREATE INDEX idx_jobs_updated_at ON jobs (updated_at DESC);
CREATE INDEX idx_jobs_payload ON jobs USING GIN (payload);
CREATE INDEX idx_jobs_result ON jobs USING GIN (result);
