-- Step 1 — id, payload, created_at
DROP TABLE IF EXISTS music_jobs CASCADE;
CREATE TABLE music_jobs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Questions to Answer
/*
1. Why UUID over SERIAL for the primary key?
   SERIAL keys are predictable sequences (1, 2, 3...) that leak the size and growth rate of a business (e.g., job #50 means only 50 jobs exist). UUIDs are unguessable, allowing for secure, decoupled ID generation across distributed systems without bottlenecking on a central database sequence.

2. Why uuidv7() specifically over uuidv4()?
   uuidv7() is time-ordered. Its first 48 bits encode a timestamp, meaning new inserts naturally cluster at the end of the B-Tree index. Purely random uuidv4() scatters inserts across the entire index tree, leading to severe index fragmentation and catastrophic performance degradation at scale (page thrashing).

3. Why JSONB over JSON?
   JSON stores exact text copies (including whitespace and duplicate keys) and must be fully reparsed on every read. JSONB is stored in a decomposed binary format, which removes whitespace/duplicates but makes reading significantly faster and supports powerful advanced indexing (like GIN).

4. Why TIMESTAMPTZ over TIMESTAMP?
   TIMESTAMPTZ stores the absolute UTC point in time and correctly converts it to the querying client's local timezone. Standard TIMESTAMP blindly assumes local time, causing silent data corruption and severe bugs when servers are in different time zones or undergo Daylight Saving Time transitions.
*/

INSERT INTO music_jobs (payload) VALUES
('{"original_filename": "punta_beat_01.mp3", "mime_type": "audio/mpeg", "file_size": 3400000}'::jsonb),
('{"original_filename": "brukdown_rhythm.wav", "mime_type": "audio/wav", "file_size": 15000000}'::jsonb),
('{"original_filename": "garifuna_drums.mp3", "mime_type": "audio/mpeg", "file_size": 4200000, "source": "mobile_app"}'::jsonb);

\echo '-- 1. Show all jobs ordered by creation time'
SELECT id, payload, created_at FROM music_jobs ORDER BY created_at;

\echo '-- 2. Extract just the filename and mime_type from each job'
SELECT payload->>'original_filename' AS filename, payload->>'mime_type' AS mime_type FROM music_jobs;

\echo '-- 3. Find only MP3 uploads'
SELECT payload->>'original_filename' AS filename FROM music_jobs WHERE payload->>'mime_type' = 'audio/mpeg';

\echo '-- 4. Find the job that has the extra field'
SELECT payload->>'original_filename' AS filename FROM music_jobs WHERE payload ? 'source';

-- Step 2 — public_id
ALTER TABLE music_jobs ADD COLUMN public_id UUID NOT NULL UNIQUE DEFAULT uuidv4();

/*
1. Why does this column use uuidv4() and not uuidv7()?
   uuidv7 embeds a high-precision timestamp in the ID. This leaks the exact time the job was created, which can be sensitive metadata. uuidv4 is purely cryptographically random and leaks no structural or temporal information to external clients.

2. What does uuid_extract_timestamp() reveal about uuidv7?
   It reveals the precise millisecond-accurate timestamp of when the UUID was generated.

3. Why does the UNIQUE constraint make CREATE INDEX unnecessary?
   In PostgreSQL, defining a UNIQUE constraint on a column automatically and implicitly creates a unique B-Tree index on that column to enforce the uniqueness rule under the hood.

4. What is the two-ID pattern and why does it matter?
   The two-ID pattern involves using a time-ordered internal ID (uuidv7 or BIGSERIAL) for optimal database storage performance and fast joins, while exposing a completely random external ID (uuidv4) to users via the API. It isolates internal database mechanics from the public contract, providing both performance and security.
*/

\echo '-- 1. Show id vs public_id side by side — what do you notice?'
SELECT id, public_id FROM music_jobs LIMIT 1;
/* Output notice: The `id` (uuidv7) begins with a predictable timestamp segment (like 018e...), while `public_id` (uuidv4) is completely random (like 4f9c...). */

\echo '-- 2. Run uuid_extract_timestamp() on both columns — what does this prove?'
SELECT uuid_extract_timestamp(id) AS internal_time, uuid_extract_timestamp(public_id) AS external_time FROM music_jobs LIMIT 1;
/* Output proves that uuidv7 retains the embedded timestamp accurately, while trying to extract a timestamp from a uuidv4 yields a wildly incorrect, nonsensical date, proving it successfully hides metadata. */

\echo '-- 3. Show what the Go server would return to the client after insert'
SELECT public_id FROM music_jobs ORDER BY created_at DESC LIMIT 1;

\echo '-- 4. Show what the Go server would do when the client polls'
SELECT public_id FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);

-- Step 3 — status, progress
ALTER TABLE music_jobs 
ADD COLUMN status TEXT NOT NULL DEFAULT 'pending' 
CHECK (status IN ('pending', 'processing', 'done', 'failed')),
ADD COLUMN progress INTEGER NOT NULL DEFAULT 0
CHECK (progress BETWEEN 0 AND 100);

/*
1. Why are status and progress real columns, not inside payload JSONB?
   These fields describe the core state of the job and are frequently queried, filtered, and updated independently. Storing them in a JSONB blob prevents efficient single-column indexing, complicates partial updates, and makes enforcing CHECK constraints much harder.

2. What happens if a buggy worker writes status = 'complet'?
   The database outright rejects the query with a CHECK constraint violation error. The bad data never enters the system.

3. Why does the CHECK constraint matter more than application validation?
   Application code often has multiple entry points (different microservices, manual admin SQL scripts, bulk importers). A database-level CHECK constraint ensures data integrity at the lowest level, impossible to bypass regardless of client bugs.

4. Draw the state machine for a job lifecycle
   [pending] ---> [processing] ---> [done]
                       \
                        \--> [failed]
*/

-- Claim the oldest pending job
UPDATE music_jobs SET status = 'processing', progress = 0 
WHERE id = (SELECT id FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1) 
RETURNING payload->>'original_filename' AS claimed_job;

-- Advance progress
UPDATE music_jobs SET progress = 25 WHERE status = 'processing';
UPDATE music_jobs SET progress = 50 WHERE status = 'processing';

-- Complete job
UPDATE music_jobs SET status = 'done', progress = 100 WHERE status = 'processing';

\echo '-- Deliberately attempt an invalid status and invalid progress'
UPDATE music_jobs SET status = 'complet' WHERE payload->>'original_filename' = 'brukdown_rhythm.wav';
UPDATE music_jobs SET progress = 150 WHERE payload->>'original_filename' = 'brukdown_rhythm.wav';

\echo '-- 1. What does the client see when polling a processing job?'
SELECT status, progress FROM music_jobs WHERE status = 'processing';

\echo '-- 2. What query does the worker run to find its next job?'
SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;

\echo '-- 3. Show all jobs with their current state'
SELECT payload->>'original_filename' AS filename, status, progress FROM music_jobs;

-- Step 4 — result, error_msg
ALTER TABLE music_jobs
ADD COLUMN result JSONB NOT NULL DEFAULT '{}',
ADD COLUMN error_msg TEXT;

/*
1. Why does the result default to '{}' and not NULL?
   A default of '{}' allows workers to incrementally append keys using the JSONB concatenation operator (||). If it were NULL, the concatenation `NULL || '{"key":"val"}'::jsonb` would evaluate to NULL, destroying the update.

2. Why is error_msg TEXT and not inside the result JSONB?
   An error message represents a fundamental divergence from the expected structure. Storing it as a top-level typed TEXT column makes it highly visible, easy to query (`WHERE error_msg IS NOT NULL`), and physically separates failure metadata from success payload structures.

3. What does the || operator do to a JSONB object?
   It merges two JSONB objects. If the key already exists, the right-side value overrides the left-side value. If the key doesn't exist, it is appended to the object.

4. Why does each stage read from the original file, not the previous stage's output?
   Reading from the original file ensures the highest quality source material is always used, preventing generational loss (compression artifacts compounding across conversions). It also allows stages to theoretically run in parallel if independent.
*/

-- Claim next job for processing
UPDATE music_jobs SET status = 'processing', progress = 0 
WHERE id = (SELECT id FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1);

-- Stage 1
UPDATE music_jobs SET progress = 25, result = result || '{"normalized_path": "processed/norm_bruk.wav"}'::jsonb WHERE status = 'processing';
-- Stage 2
UPDATE music_jobs SET progress = 50, result = result || '{"trimmed_path": "processed/trim_bruk.wav"}'::jsonb WHERE status = 'processing';
-- Stage 3
UPDATE music_jobs SET progress = 75, result = result || '{"converted_path": "processed/conv_bruk.mp3"}'::jsonb WHERE status = 'processing';
-- Stage 4
UPDATE music_jobs SET status = 'done', progress = 100, result = result || '{"waveform_path": "processed/wave_bruk.json"}'::jsonb WHERE status = 'processing';

-- Simulate failure
UPDATE music_jobs SET status = 'failed', error_msg = 'FFmpegError: Invalid bit depth in garifuna_drums.mp3' WHERE payload->>'original_filename' = 'garifuna_drums.mp3';

\echo '-- 1. What does the client see when polling a completed job?'
SELECT status, progress, result FROM music_jobs WHERE status = 'done' LIMIT 1;

\echo '-- 2. What does the client see mid-processing (partial result)?'
-- Theoretically, they would see: status: processing, progress: 50, result: {"normalized_path": "...", "trimmed_path": "..."}

\echo '-- 3. How do you find all failed jobs?'
SELECT id, error_msg FROM music_jobs WHERE status = 'failed';

\echo '-- 4. Show the full result object for a completed job'
SELECT result FROM music_jobs WHERE status = 'done' LIMIT 1;

-- Step 5 — updated_at
ALTER TABLE music_jobs ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

/*
1. Why is created_at not enough?
   `created_at` represents the birth of the record (history). It never changes. To detect stuck jobs or push state-change notifications efficiently (SSE/WebSockets), the system must query "what records changed in the last N seconds?", which strictly requires an `updated_at` column.

2. What goes wrong if application code maintains updated_at?
   If there are many microservices, scripts, or manual database fixes interacting with the database, inevitably one of them will update a row and forget to set `updated_at = now()`. This silent omission breaks synchronization, SSE streams, and straggler-detection logic globally.

3. Write a query that would power an SSE health check endpoint
   SELECT public_id, status, progress FROM music_jobs WHERE updated_at > now() - INTERVAL '5 seconds' ORDER BY updated_at DESC;
*/

-- Update WITHOUT setting updated_at (stale)
UPDATE music_jobs SET progress = 10 WHERE payload->>'original_filename' = 'punta_beat_01.mp3';
\echo '-- Show stale updated_at'
SELECT progress, updated_at FROM music_jobs WHERE payload->>'original_filename' = 'punta_beat_01.mp3';

-- Update WITH setting updated_at (correct)
UPDATE music_jobs SET progress = 15, updated_at = now() WHERE payload->>'original_filename' = 'punta_beat_01.mp3';

/* This is fragile because it relies on every single developer and application always remembering to manually include `updated_at = now()` in every single UPDATE statement across the entire codebase. */

\echo '-- 1. Find jobs that changed in the last 60 seconds'
SELECT id FROM music_jobs WHERE updated_at > now() - INTERVAL '60 seconds';

\echo '-- 2. Find jobs stuck in processing for more than 5 minutes'
SELECT id FROM music_jobs WHERE status = 'processing' AND updated_at < now() - INTERVAL '5 minutes';

\echo '-- 3. How long did each completed job take?'
SELECT payload->>'original_filename', updated_at - created_at AS duration FROM music_jobs WHERE status = 'done';

-- Step 6 — Trigger on updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER music_jobs_updated_at
BEFORE UPDATE ON music_jobs
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

/*
1. Why BEFORE UPDATE and not AFTER UPDATE?
   A BEFORE UPDATE trigger intercepts the row data *before* it is written to disk, allowing the trigger to modify the payload (e.g., overriding NEW.updated_at). An AFTER UPDATE trigger fires after the data is written, requiring an entirely new query to mutate the row again.

2. What is NEW and what is OLD in a trigger function?
   NEW is a special variable containing the incoming row data representing the proposed state after the update. OLD contains the existing row data representing the state before the update occurs.

3. Why does returning NEW matter?
   In a BEFORE trigger, the row returned by the trigger function is what actually gets saved to the database. Returning NEW saves the row with our injected `updated_at` value. Returning NULL would silently cancel the update entirely.

4. Why is the function reusable across tables?
   The function dynamically manipulates the NEW record generically without explicitly referencing `music_jobs`. As long as the target table has an `updated_at` column, this function can be bound to it via a trigger.
*/

-- Update WITHOUT mentioning updated_at
UPDATE music_jobs SET progress = 50 WHERE payload->>'original_filename' = 'punta_beat_01.mp3';

-- Try to sabotage it
UPDATE music_jobs SET progress = 75, updated_at = '2000-01-01 00:00:00+00' WHERE payload->>'original_filename' = 'punta_beat_01.mp3';
\echo '-- Show trigger overwrote sabotage'
SELECT progress, updated_at FROM music_jobs WHERE payload->>'original_filename' = 'punta_beat_01.mp3';

\echo '-- 1. Show trigger details'
SELECT trigger_name, action_timing, action_statement FROM information_schema.triggers WHERE event_object_table = 'music_jobs';

\echo '-- 2. Show function details'
SELECT routine_name, data_type FROM information_schema.routines WHERE routine_name = 'set_updated_at';

-- Step 7 — Indexes + EXPLAIN ANALYZE

\echo '-- Part A: Generate 50,000 rows'
INSERT INTO music_jobs (status, progress, payload, result, error_msg)
SELECT
    status,
    CASE status
        WHEN 'pending'    THEN 0
        WHEN 'processing' THEN (random() * 99)::INTEGER
        WHEN 'done'       THEN 100
        WHEN 'failed'     THEN (random() * 99)::INTEGER
    END,
    jsonb_build_object(
        'original_filename', 'track_' || i || '.mp3',
        'mime_type',         CASE WHEN random() > 0.2 THEN 'audio/mpeg' ELSE 'audio/wav' END,
        'file_size',         (random() * 10000000)::INTEGER
    ),
    CASE status
        WHEN 'done' THEN jsonb_build_object('waveform_path', 'processed/wave_' || i || '.json')
        ELSE '{}'::jsonb
    END,
    CASE status
        WHEN 'failed' THEN 'AudioProcessingError'
        ELSE NULL
    END
FROM generate_series(1, 50000) AS i,
LATERAL (
    SELECT (ARRAY['pending','processing','done','failed'])[ (((i - 1) % 4) + 1) ] AS status
) AS s;

\echo '-- Part B: EXPLAIN ANALYZE Before Indexes'
EXPLAIN ANALYZE SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;
EXPLAIN ANALYZE SELECT public_id, status, progress, result, error_msg FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);
EXPLAIN ANALYZE SELECT id, payload->>'original_filename' FROM music_jobs WHERE payload @> '{"mime_type": "audio/mpeg"}'::jsonb;

\echo '-- Part C: Add Indexes'
CREATE INDEX idx_music_jobs_status_created ON music_jobs (status, created_at);
CREATE INDEX idx_music_jobs_payload ON music_jobs USING GIN (payload);
CREATE INDEX idx_music_jobs_result ON music_jobs USING GIN (result);

\echo '-- Part D: EXPLAIN ANALYZE After Indexes'
EXPLAIN ANALYZE SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;
EXPLAIN ANALYZE SELECT public_id, status, progress, result, error_msg FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);
EXPLAIN ANALYZE SELECT id, payload->>'original_filename' FROM music_jobs WHERE payload @> '{"mime_type": "audio/mpeg"}'::jsonb;

/*
Part E: Explain the results
1. What is a sequential scan and why is it slow at scale?
   A sequential scan forces the database engine to read every single row block-by-block on disk to evaluate the WHERE clause. At scale (millions of rows), this requires immense disk I/O and memory caching, resulting in agonizingly slow O(N) performance.

2. Why does the worker poll query need a COMPOSITE index and not just an index on status alone?
   The worker poll uses BOTH `status = 'pending'` for filtering AND `ORDER BY created_at` for sorting. If the index was only on status, Postgres would filter the pending rows fast, but then have to perform an expensive in-memory sort on thousands of rows. A composite index pre-sorts the pending rows by creation time, allowing Postgres to grab the first matching row instantly.

3. Why GIN and not btree for JSONB columns?
   B-Tree indexes compare entire scalar values (string vs string). To use a B-Tree for JSON, you'd have to query the EXACT JSON structure. GIN (Generalized Inverted Index) builds an index of all individual keys and values *inside* the JSON document, allowing rapid structural containment searches.

4. Which operators USE the GIN index? Which do NOT?
   The containment operator (`@>`), existence operators (`?`, `?|`, `?&`), and path matching operator (`@@`) utilize the GIN index. Extraction operators like `->` and `->>` DO NOT use the GIN index (they just fetch fields dynamically).

5. What speedup did you measure?
   (Will write this manually after observing output: Worker poll went from a high Execution Time with a Seq Scan and Sort to a <1ms Index Scan. Public ID query utilized the implicit unique B-tree index immediately. JSON containment shifted from a full Seq Scan to a Bitmap Heap Scan if the condition was highly selective.)
*/

\echo '-- Final Verification'
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'music_jobs' ORDER BY indexname;
\d music_jobs
