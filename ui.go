package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var dsn = os.Getenv("DSN")
var db *sql.DB

const htmlPage = `
<!DOCTYPE html>
<html>
<head>
	<title>Job Queue Query Planner Demo</title>
	<style>
		body { font-family: sans-serif; margin: 40px; background: #f4f4f9; }
		h1 { color: #333; }
		.btn { display: inline-block; padding: 10px 15px; margin: 5px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; border: none; cursor: pointer; }
		.btn:hover { background: #0056b3; }
		pre { background: #fff; border: 1px solid #ccc; padding: 15px; border-radius: 4px; overflow-x: auto; font-family: monospace; }
		.query { font-weight: bold; color: #d63384; }
	</style>
</head>
<body>
	<h1>Job Queue Query Planner Demo</h1>
	<p>The <b>jobs</b> table contains 100,000 generated records to demonstrate indexing and the Query Planner.</p>
	<div>
		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="SELECT count(*) FROM jobs;" />
			<button class="btn" type="submit">Count Jobs</button>
		</form>

		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="EXPLAIN ANALYZE SELECT id, payload FROM jobs WHERE status = 'pending' ORDER BY created_at LIMIT 1;" />
			<button class="btn" type="submit">Worker Query (Status/Created Index)</button>
		</form>

		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="EXPLAIN ANALYZE SELECT public_id, status, progress, updated_at FROM jobs WHERE updated_at > now() - INTERVAL '60 seconds' ORDER BY updated_at DESC;" />
			<button class="btn" type="submit">Updates Query (Updated_at Index)</button>
		</form>

		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="EXPLAIN ANALYZE SELECT id, payload->>'original_filename' AS filename FROM jobs WHERE payload @> '{&#34;mime_type&#34;: &#34;image/jpeg&#34;}'::jsonb;" />
			<button class="btn" type="submit">Payload GIN Index (jpeg)</button>
		</form>
		
		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="EXPLAIN ANALYZE SELECT id, payload->>'original_filename' AS filename FROM jobs WHERE payload @> '{&#34;mime_type&#34;: &#34;image/png&#34;, &#34;source&#34;: &#34;mobile_upload&#34;}'::jsonb;" />
			<button class="btn" type="submit">Payload GIN Index (png + mobile)</button>
		</form>

		<form method="POST" action="/query" style="display:inline;">
			<input type="hidden" name="q" value="EXPLAIN ANALYZE SELECT id, result->>'thumbnail_path' AS thumbnail FROM jobs WHERE result @> '{&#34;thumbnail_path&#34;: &#34;uploads/processed/thumb_a1b2c3d4.jpg&#34;}'::jsonb;" />
			<button class="btn" type="submit">Result GIN Index</button>
		</form>
	</div>

	{{if .Query}}
	<div style="margin-top: 30px;">
		<h2>Results</h2>
		<p class="query">Query: {{.Query}}</p>
		<pre>{{.Result}}</pre>
	</div>
	{{end}}
</body>
</html>
`

type PageData struct {
	Query  string
	Result string
}

func main() {
	if dsn == "" {
		log.Fatal("DSN environment variable must be set")
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	tmpl := template.Must(template.New("page").Parse(htmlPage))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, PageData{})
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		q := r.FormValue("q")
		if q == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		rows, err := db.Query(q)
		var output string
		if err != nil {
			output = fmt.Sprintf("Error: %v", err)
		} else {
			defer rows.Close()
			cols, _ := rows.Columns()

			for rows.Next() {
				values := make([]interface{}, len(cols))
				valuePtrs := make([]interface{}, len(cols))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					output += fmt.Sprintf("Scan error: %v\n", err)
					continue
				}

				for i, col := range cols {
					var v string
					val := values[i]
					switch b := val.(type) {
					case []byte:
						v = string(b)
					default:
						v = fmt.Sprintf("%v", b)
					}
					if len(cols) == 1 {
						output += v + "\n"
					} else {
						output += fmt.Sprintf("%s: %s\n", col, v)
					}
				}
				if len(cols) > 1 {
					output += "-------------------\n"
				}
			}
		}

		data := PageData{
			Query:  q,
			Result: output,
		}
		tmpl.Execute(w, data)
	})

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
