package mailer

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	url    string
	apiKey string
	sender string
}

func New(apiKey, sender, url string) Mailer {
	return Mailer{
		url:    url,
		apiKey: apiKey,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	payloadData := map[string]any{
		"from": map[string]string{
			"email": m.sender,
			"name":  "Comments Community",
		},
		"to": []map[string]string{
			{"email": recipient},
		},
		"subject":  subject.String(),
		"text":     plainBody.String(),
		"html":     htmlBody.String(),
		"category": "Integration Test",
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}

	// Retry loop
	for i := 1; i <= 3; i++ {
		req, err := http.NewRequest(http.MethodPost, m.url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return err
		}

		req.Header.Add("Authorization", "Bearer "+m.apiKey)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		res, err := client.Do(req)

		if err == nil {
			defer res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				return nil
			}
			body, _ := io.ReadAll(res.Body)
			err = fmt.Errorf("mailtrap api returned status %d: %s", res.StatusCode, string(body))
		}

		// If error or wrong status, sleep and retry
		time.Sleep(500 * time.Millisecond)
	}

	return err
}
