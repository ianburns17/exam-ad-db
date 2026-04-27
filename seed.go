package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN environment variable must be set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	log.Println("Inserting 100k rows of fake data...")

	query := `
INSERT INTO jobs (status, progress, payload, result, error_msg)
SELECT
    status,
    CASE status
        WHEN 'pending'    THEN 0
        WHEN 'processing' THEN (random() * 99)::INTEGER
        WHEN 'done'       THEN 100
        WHEN 'failed'     THEN (random() * 99)::INTEGER
    END,
    jsonb_build_object(
        'original_filename', 'file_' || i || '.jpg',
        'stored_path',       'uploads/' || md5(i::text) || '.jpg',
        'mime_type',         CASE WHEN random() > 0.2
                                  THEN 'image/jpeg'
                                  ELSE 'image/png'
                             END,
        'file_size', (random() * 10000000)::INTEGER
    ),
    CASE status
        WHEN 'done' THEN jsonb_build_object(
            'grayscale_path',  'uploads/processed/gray_'   || md5(i::text) || '.jpg',
            'blurred_path',    'uploads/processed/blur_'   || md5(i::text) || '.jpg',
            'scaled_path',     'uploads/processed/scaled_' || md5(i::text) || '.jpg',
            'thumbnail_path',  'uploads/processed/thumb_'  || md5(i::text) || '.jpg'
        )
        ELSE '{}'::jsonb
    END,
    CASE status
        WHEN 'failed' THEN 'ImageProcessingError: stage ' || (random()*4)::INTEGER
        ELSE NULL
    END
FROM generate_series(1, 100000) AS i,
LATERAL (
    SELECT (ARRAY['pending','processing','done','failed'])[
        (((i - 1) % 4) + 1)
    ] AS status
) AS s;`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	log.Println("Successfully inserted 100k rows.")
}
