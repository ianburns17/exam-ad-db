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

	tmpl := template.Must(template.New("page").Parse(`
		<!DOCTYPE html>
		<html>
		<head><title>DB Query</title></head>
		<body>
			<h1>DB Query</h1>
			<form action="/query" method="post">
				<textarea name="q" rows="4" cols="50">{{.Query}}</textarea><br/>
				<button type="submit">Run</button>
			</form>
			<hr/>
			<pre>{{.Result}}</pre>
		</body>
		</html>
	`))

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
