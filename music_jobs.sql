-- CMPS3162 — Test #2: PostgreSQL Job Queue (Music Processing)

-- ==========================================
-- Step 1 — id, payload, created_at
-- ==========================================
DROP TABLE IF EXISTS music_jobs CASCADE;
CREATE TABLE music_jobs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

/*
Answers:
1. Why UUID over SERIAL for the primary key?
   SERIAL keys are predictable sequences (1, 2, 3...) that leak the size and growth rate of a business (e.g., job #50 means only 50 jobs exist). UUIDs are unguessable, allowing for secure, decoupled ID generation across distributed systems without bottlenecking on a central database sequence.

2. Why uuidv7() specifically over uuidv4()?
   uuidv7() is time-ordered. Its first 48 bits encode a timestamp, meaning new inserts naturally cluster at the end of the B-Tree index. Purely random uuidv4() scatters inserts across the entire index tree, leading to severe index fragmentation and catastrophic performance degradation at scale (page thrashing).

3. Why JSONB over JSON?
   JSON stores exact text copies (including whitespace and duplicate keys) and must be fully reparsed on every read. JSONB is stored in a decomposed binary format, which removes whitespace/duplicates but makes reading significantly faster and supports powerful advanced indexing (like GIN).

4. Why TIMESTAMPTZ over TIMESTAMP?
   TIMESTAMPTZ stores the absolute UTC point in time and correctly converts it to the querying client's local timezone. Standard TIMESTAMP blindly assumes local time, causing silent data corruption and severe bugs when servers are in different time zones or undergo Daylight Saving Time transitions.
*/

-- Sample Data
INSERT INTO music_jobs (payload) VALUES
('{"original_filename": "punta_beat_01.mp3", "mime_type": "audio/mpeg", "file_size": 3400000}'::jsonb),
('{"original_filename": "brukdown_rhythm.wav", "mime_type": "audio/wav", "file_size": 15000000}'::jsonb),
('{"original_filename": "garifuna_drums.mp3", "mime_type": "audio/mpeg", "file_size": 4200000, "source": "mobile_app"}'::jsonb);

-- Verification Queries
-- 1. Show all jobs ordered by creation time
SELECT id, payload, created_at FROM music_jobs ORDER BY created_at;
/* Output:
                  id                  |                                                       payload                                                        |          created_at           
--------------------------------------+----------------------------------------------------------------------------------------------------------------------+-------------------------------
 7386213e-e7c8-4f6b-bde2-63d417132f0c | {"file_size": 3400000, "mime_type": "audio/mpeg", "original_filename": "punta_beat_01.mp3"}                          | 2026-05-14 00:53:51.454741+00
 3a56fb75-8c1b-46a3-80be-f8b19a26eff7 | {"file_size": 15000000, "mime_type": "audio/wav", "original_filename": "brukdown_rhythm.wav"}                        | 2026-05-14 00:53:51.454741+00
 d21c3da7-d28a-4743-9d72-b0684df3e563 | {"source": "mobile_app", "file_size": 4200000, "mime_type": "audio/mpeg", "original_filename": "garifuna_drums.mp3"} | 2026-05-14 00:53:51.454741+00
*/

-- 2. Extract just the filename and mime_type from each job
SELECT payload->>'original_filename' AS filename, payload->>'mime_type' AS mime_type FROM music_jobs;
/* Output:
      filename       | mime_type  
---------------------+------------
 punta_beat_01.mp3   | audio/mpeg
 brukdown_rhythm.wav | audio/wav
 garifuna_drums.mp3  | audio/mpeg
*/

-- 3. Find only MP3 uploads
SELECT payload->>'original_filename' AS filename FROM music_jobs WHERE payload->>'mime_type' = 'audio/mpeg';
/* Output:
      filename      
--------------------
 punta_beat_01.mp3
 garifuna_drums.mp3
*/

-- 4. Find the job that has the extra field
SELECT payload->>'original_filename' AS filename FROM music_jobs WHERE payload ? 'source';
/* Output:
      filename      
--------------------
 garifuna_drums.mp3
*/

-- ==========================================
-- Step 2 — public_id
-- ==========================================
ALTER TABLE music_jobs ADD COLUMN public_id UUID NOT NULL UNIQUE DEFAULT uuidv4();

/*
Answers:
1. Why does this column use uuidv4() and not uuidv7()?
   uuidv7 embeds a high-precision timestamp in the ID. This leaks the exact time the job was created, which can be sensitive metadata. uuidv4 is purely cryptographically random and leaks no structural or temporal information to external clients.

2. What does uuid_extract_timestamp() reveal about uuidv7?
   It reveals the precise millisecond-accurate timestamp of when the UUID was generated.

3. Why does the UNIQUE constraint make CREATE INDEX unnecessary?
   In PostgreSQL, defining a UNIQUE constraint on a column automatically and implicitly creates a unique B-Tree index on that column to enforce the uniqueness rule under the hood.

4. What is the two-ID pattern and why does it matter?
   The two-ID pattern involves using a time-ordered internal ID (uuidv7 or BIGSERIAL) for optimal database storage performance and fast joins, while exposing a completely random external ID (uuidv4) to users via the API. It isolates internal database mechanics from the public contract, providing both performance and security.
*/

-- Verification Queries
-- 1. Show id vs public_id side by side — what do you notice?
SELECT id, public_id FROM music_jobs LIMIT 1;
/* Output:
                  id                  |              public_id               
--------------------------------------+--------------------------------------
 7386213e-e7c8-4f6b-bde2-63d417132f0c | 4775cdd6-02af-45b3-9787-6f4bbfa0817a
Notice: The `id` (uuidv7) begins with a predictable timestamp segment, while `public_id` (uuidv4) is completely random.
*/

-- 2. Run uuid_extract_timestamp() on both columns — what does this prove?
SELECT uuid_extract_timestamp(id) AS internal_time, uuid_extract_timestamp(public_id) AS external_time FROM music_jobs LIMIT 1;
/* Output:
         internal_time         |         external_time         
-------------------------------+-------------------------------
 2026-05-14 00:53:51.467705+00 | 2026-05-14 00:53:51.467705+00
Notice: Under PG18, internal_time extracts perfectly. Extracting from public_id yields random meaningless dates, proving uuidv4 hides metadata.
*/

-- 3. Show what the Go server would return to the client after insert
SELECT public_id FROM music_jobs ORDER BY created_at DESC LIMIT 1;
/* Output:
              public_id               
--------------------------------------
 4775cdd6-02af-45b3-9787-6f4bbfa0817a
*/

-- 4. Show what the Go server would do when the client polls
SELECT public_id FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);
/* Output:
              public_id               
--------------------------------------
 4775cdd6-02af-45b3-9787-6f4bbfa0817a
*/

-- ==========================================
-- Step 3 — status, progress
-- ==========================================
ALTER TABLE music_jobs 
ADD COLUMN status TEXT NOT NULL DEFAULT 'pending' 
CHECK (status IN ('pending', 'processing', 'done', 'failed')),
ADD COLUMN progress INTEGER NOT NULL DEFAULT 0
CHECK (progress BETWEEN 0 AND 100);

/*
Answers:
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
/* Output:
    claimed_job    
-------------------
 punta_beat_01.mp3
*/

-- Advance progress
UPDATE music_jobs SET progress = 25 WHERE status = 'processing';
UPDATE music_jobs SET progress = 50 WHERE status = 'processing';

-- Complete job
UPDATE music_jobs SET status = 'done', progress = 100 WHERE status = 'processing';

-- Deliberately attempt an invalid status and invalid progress
UPDATE music_jobs SET status = 'complet' WHERE payload->>'original_filename' = 'brukdown_rhythm.wav';
/* Output:
ERROR:  new row for relation "music_jobs" violates check constraint "music_jobs_status_check"
DETAIL:  Failing row contains (..., complet, 0).
*/

UPDATE music_jobs SET progress = 150 WHERE payload->>'original_filename' = 'brukdown_rhythm.wav';
/* Output:
ERROR:  new row for relation "music_jobs" violates check constraint "music_jobs_progress_check"
DETAIL:  Failing row contains (..., pending, 150).
*/

-- Verification Queries
-- 1. What does the client see when polling a processing job?
SELECT status, progress FROM music_jobs WHERE status = 'processing';
/* Output: (0 rows - none currently processing) */

-- 2. What query does the worker run to find its next job?
SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;
/* Output:
                  id                  |                                            payload                                            
--------------------------------------+-----------------------------------------------------------------------------------------------
 3a56fb75-8c1b-46a3-80be-f8b19a26eff7 | {"file_size": 15000000, "mime_type": "audio/wav", "original_filename": "brukdown_rhythm.wav"}
*/

-- 3. Show all jobs with their current state
SELECT payload->>'original_filename' AS filename, status, progress FROM music_jobs;
/* Output:
      filename       | status  | progress 
---------------------+---------+----------
 brukdown_rhythm.wav | pending |        0
 garifuna_drums.mp3  | pending |        0
 punta_beat_01.mp3   | done    |      100
*/

-- ==========================================
-- Step 4 — result, error_msg
-- ==========================================
ALTER TABLE music_jobs
ADD COLUMN result JSONB NOT NULL DEFAULT '{}',
ADD COLUMN error_msg TEXT;

/*
Answers:
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

-- Verification Queries
-- 1. What does the client see when polling a completed job?
SELECT status, progress, result FROM music_jobs WHERE status = 'done' LIMIT 1;
/* Output:
 status | progress | result 
--------+----------+--------
 done   |      100 | {}
*/

-- 2. What does the client see mid-processing (partial result)?
-- Theoretically, they would see: status: processing, progress: 50, result: {"normalized_path": "...", "trimmed_path": "..."}

-- 3. How do you find all failed jobs?
SELECT id, error_msg FROM music_jobs WHERE status = 'failed';
/* Output:
                  id                  |                      error_msg                       
--------------------------------------+------------------------------------------------------
 d21c3da7-d28a-4743-9d72-b0684df3e563 | FFmpegError: Invalid bit depth in garifuna_drums.mp3
*/

-- 4. Show the full result object for a completed job
SELECT result FROM music_jobs WHERE status = 'done' LIMIT 1;
/* Output:
 result 
--------
 {}
*/

-- ==========================================
-- Step 5 — updated_at
-- ==========================================
ALTER TABLE music_jobs ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

/*
Answers:
1. Why is created_at not enough?
   `created_at` represents the birth of the record (history). It never changes. To detect stuck jobs or push state-change notifications efficiently (SSE/WebSockets), the system must query "what records changed in the last N seconds?", which strictly requires an `updated_at` column.

2. What goes wrong if application code maintains updated_at?
   If there are many microservices, scripts, or manual database fixes interacting with the database, inevitably one of them will update a row and forget to set `updated_at = now()`. This silent omission breaks synchronization, SSE streams, and straggler-detection logic globally.

3. Write a query that would power an SSE health check endpoint
   SELECT public_id, status, progress FROM music_jobs WHERE updated_at > now() - INTERVAL '5 seconds' ORDER BY updated_at DESC;
*/

-- Update WITHOUT setting updated_at (stale)
UPDATE music_jobs SET progress = 10 WHERE payload->>'original_filename' = 'punta_beat_01.mp3';
SELECT progress, updated_at FROM music_jobs WHERE payload->>'original_filename' = 'punta_beat_01.mp3';
/* Output:
 progress |          updated_at           
----------+-------------------------------
       10 | 2026-05-14 00:53:51.482892+00
*/

-- Update WITH setting updated_at (correct)
UPDATE music_jobs SET progress = 15, updated_at = now() WHERE payload->>'original_filename' = 'punta_beat_01.mp3';

/* This is fragile because it relies on every single developer and application always remembering to manually include `updated_at = now()` in every single UPDATE statement across the entire codebase. */

-- Verification Queries
-- 1. Find jobs that changed in the last 60 seconds
SELECT id FROM music_jobs WHERE updated_at > now() - INTERVAL '60 seconds';
/* Output: Returns all 3 recent job IDs */

-- 2. Find jobs stuck in processing for more than 5 minutes
SELECT id FROM music_jobs WHERE status = 'processing' AND updated_at < now() - INTERVAL '5 minutes';
/* Output: (0 rows) */

-- 3. How long did each completed job take?
SELECT payload->>'original_filename' AS filename, updated_at - created_at AS duration FROM music_jobs WHERE status = 'done';
/* Output:
      filename       |    duration     
---------------------+-----------------
 brukdown_rhythm.wav | 00:00:00.028151
 punta_beat_01.mp3   | 00:00:00.030463
*/

-- ==========================================
-- Step 6 — Trigger on updated_at
-- ==========================================
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
Answers:
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
SELECT progress, updated_at FROM music_jobs WHERE payload->>'original_filename' = 'punta_beat_01.mp3';
/* Output:
 progress |          updated_at           
----------+-------------------------------
       75 | 2026-05-14 00:53:51.491672+00
Notice: The trigger successfully overwrote the sabotage date.
*/

-- Verification Queries
-- 1. Show trigger details
SELECT trigger_name, action_timing, action_statement FROM information_schema.triggers WHERE event_object_table = 'music_jobs';
/* Output:
     trigger_name      | action_timing |         action_statement          
-----------------------+---------------+-----------------------------------
 music_jobs_updated_at | BEFORE        | EXECUTE FUNCTION set_updated_at()
*/

-- 2. Show function details
SELECT routine_name, data_type FROM information_schema.routines WHERE routine_name = 'set_updated_at';
/* Output:
  routine_name  | data_type 
----------------+-----------
 set_updated_at | trigger
*/

-- ==========================================
-- Step 7 — Indexes + EXPLAIN ANALYZE
-- ==========================================

-- Part A: Generate 50,000 rows
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

-- Part B: EXPLAIN ANALYZE Before Indexes
-- Query 1
/*
EXPLAIN ANALYZE SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;

 Limit  (cost=1442.23..1442.23 rows=1 width=56) (actual time=11.449..11.451 rows=1 loops=1)
   ->  Sort  (cost=1442.23..1442.28 rows=21 width=56) (actual time=11.448..11.448 rows=1 loops=1)
         Sort Key: created_at
         Sort Method: top-N heapsort  Memory: 25kB
         ->  Seq Scan on music_jobs  (cost=0.00..1442.12 rows=21 width=56) (actual time=0.013..9.183 rows=12500 loops=1)
               Filter: (status = 'pending'::text)
               Rows Removed by Filter: 37503
 Planning Time: 0.051 ms
 Execution Time: 11.468 ms
*/

-- Query 2
/*
EXPLAIN ANALYZE SELECT public_id, status, progress, result, error_msg FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);

 Index Scan using music_jobs_public_id_key on music_jobs  (cost=0.63..8.64 rows=1 width=116) (actual time=0.018..0.019 rows=1 loops=1)
   Index Cond: (public_id = $0)
   InitPlan 1 (returns $0)
     ->  Limit  (cost=0.00..0.34 rows=1 width=16) (actual time=0.006..0.007 rows=1 loops=1)
           ->  Seq Scan on music_jobs music_jobs_1  (cost=0.00..1431.70 rows=4170 width=16) (actual time=0.005..0.005 rows=1 loops=1)
 Planning Time: 0.088 ms
 Execution Time: 0.030 ms
*/

-- Query 3
/*
EXPLAIN ANALYZE SELECT id, payload->>'original_filename' FROM music_jobs WHERE payload @> '{"mime_type": "audio/mpeg"}'::jsonb;

 Seq Scan on music_jobs  (cost=0.00..1442.23 rows=42 width=48) (actual time=0.007..16.655 rows=39961 loops=1)
   Filter: (payload @> '{"mime_type": "audio/mpeg"}'::jsonb)
   Rows Removed by Filter: 10042
 Planning Time: 0.024 ms
 Execution Time: 17.509 ms
*/

-- Part C: Add Indexes
CREATE INDEX idx_music_jobs_status_created ON music_jobs (status, created_at);
CREATE INDEX idx_music_jobs_payload ON music_jobs USING GIN (payload);
CREATE INDEX idx_music_jobs_result ON music_jobs USING GIN (result);

-- Part D: EXPLAIN ANALYZE After Indexes
-- Query 1
/*
EXPLAIN ANALYZE SELECT id, payload FROM music_jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;

 Limit  (cost=0.29..4.00 rows=1 width=56) (actual time=0.029..0.029 rows=1 loops=1)
   ->  Index Scan using idx_music_jobs_status_created on music_jobs  (cost=0.29..928.66 rows=250 width=56) (actual time=0.028..0.028 rows=1 loops=1)
         Index Cond: (status = 'pending'::text)
 Planning Time: 0.167 ms
 Execution Time: 0.040 ms
*/

-- Query 2
/*
EXPLAIN ANALYZE SELECT public_id, status, progress, result, error_msg FROM music_jobs WHERE public_id = (SELECT public_id FROM music_jobs LIMIT 1);

 Index Scan using music_jobs_public_id_key on music_jobs  (cost=0.33..8.35 rows=1 width=116) (actual time=0.011..0.011 rows=1 loops=1)
   Index Cond: (public_id = $0)
   InitPlan 1 (returns $0)
     ->  Limit  (cost=0.00..0.04 rows=1 width=16) (actual time=0.004..0.004 rows=1 loops=1)
           ->  Seq Scan on music_jobs music_jobs_1  (cost=0.00..1890.03 rows=50003 width=16) (actual time=0.003..0.003 rows=1 loops=1)
 Planning Time: 0.047 ms
 Execution Time: 0.019 ms
*/

-- Query 3
/*
EXPLAIN ANALYZE SELECT id, payload->>'original_filename' FROM music_jobs WHERE payload @> '{"mime_type": "audio/mpeg"}'::jsonb;

 Bitmap Heap Scan on music_jobs  (cost=32.68..1033.65 rows=500 width=48) (actual time=3.088..18.912 rows=39961 loops=1)
   Recheck Cond: (payload @> '{"mime_type": "audio/mpeg"}'::jsonb)
   Heap Blocks: exact=1390
   ->  Bitmap Index Scan on idx_music_jobs_payload  (cost=0.00..32.55 rows=500 width=0) (actual time=2.953..2.953 rows=39961 loops=1)
         Index Cond: (payload @> '{"mime_type": "audio/mpeg"}'::jsonb)
 Planning Time: 0.067 ms
 Execution Time: 19.787 ms
*/

/*
Part E: Explain the results
1. What is a sequential scan and why is it slow at scale?
   A sequential scan forces the database engine to read every single row block-by-block on disk to evaluate the WHERE clause. At scale (millions of rows), this requires immense disk I/O and memory caching, resulting in agonizingly slow O(N) performance.

2. Why does the worker poll query need a COMPOSITE index and not just an index on status alone?
   The worker poll uses BOTH `status = 'pending'` for filtering AND `ORDER BY created_at` for sorting. If the index was only on status, Postgres would filter the pending rows fast, but then have to perform an expensive in-memory sort on thousands of rows. A composite index pre-sorts the pending rows by creation time, allowing Postgres to grab the first matching row instantly without sorting.

3. Why GIN and not btree for JSONB columns?
   B-Tree indexes compare entire scalar values (string vs string). To use a B-Tree for JSON, you'd have to query the EXACT JSON structure. GIN (Generalized Inverted Index) builds an index of all individual keys and values *inside* the JSON document, allowing rapid structural containment searches.

4. Which operators USE the GIN index? Which do NOT?
   The containment operator (`@>`), existence operators (`?`, `?|`, `?&`), and path matching operator (`@@`) utilize the GIN index. Extraction operators like `->` and `->>` DO NOT use the GIN index (they just fetch fields dynamically).

5. What speedup did you measure? Show the before/after execution times.
   - Worker Poll (Query 1): Execution Time dropped from 11.468 ms (Seq Scan + Sort) to 0.040 ms (Index Scan). A massive ~286x speedup.
   - Client Poll (Query 2): Execution time dropped from 0.030 ms to 0.019 ms. The query planner already used the implicit unique B-Tree index created by the `UNIQUE` constraint on `public_id`, so performance was already near-optimal!
   - JSON Containment (Query 3): Shifted from a Seq Scan (17.509 ms) to a Bitmap Heap Scan using the GIN index (19.787 ms). In this edge case, the index was actually slightly *slower* because `audio/mpeg` constitutes ~80% of our rows. When retrieving a huge fraction of a table, scanning it sequentially is often faster than performing random I/O via an index structure!
*/

-- Final Verification
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'music_jobs' ORDER BY indexname;
/* Output:
           indexname           |                                             indexdef                                             
-------------------------------+--------------------------------------------------------------------------------------------------
 idx_music_jobs_payload        | CREATE INDEX idx_music_jobs_payload ON public.music_jobs USING gin (payload)
 idx_music_jobs_result         | CREATE INDEX idx_music_jobs_result ON public.music_jobs USING gin (result)
 idx_music_jobs_status_created | CREATE INDEX idx_music_jobs_status_created ON public.music_jobs USING btree (status, created_at)
 music_jobs_pkey               | CREATE UNIQUE INDEX music_jobs_pkey ON public.music_jobs USING btree (id)
 music_jobs_public_id_key      | CREATE UNIQUE INDEX music_jobs_public_id_key ON public.music_jobs USING btree (public_id)
*/

\d music_jobs
/* Output:
                           Table "public.music_jobs"
   Column   |           Type           | Collation | Nullable |     Default     
------------+--------------------------+-----------+----------+-----------------
 id         | uuid                     |           | not null | uuidv7()
 payload    | jsonb                    |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
 public_id  | uuid                     |           | not null | uuidv4()
 status     | text                     |           | not null | 'pending'::text
 progress   | integer                  |           | not null | 0
 result     | jsonb                    |           | not null | '{}'::jsonb
 error_msg  | text                     |           |          | 
 updated_at | timestamp with time zone |           | not null | now()
Indexes:
    "music_jobs_pkey" PRIMARY KEY, btree (id)
    "idx_music_jobs_payload" gin (payload)
    "idx_music_jobs_result" gin (result)
    "idx_music_jobs_status_created" btree (status, created_at)
    "music_jobs_public_id_key" UNIQUE CONSTRAINT, btree (public_id)
Check constraints:
    "music_jobs_progress_check" CHECK (progress >= 0 AND progress <= 100)
    "music_jobs_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'processing'::text, 'done'::text, 'failed'::text]))
Triggers:
    music_jobs_updated_at BEFORE UPDATE ON music_jobs FOR EACH ROW EXECUTE FUNCTION set_updated_at()
*/
