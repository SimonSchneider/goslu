package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

type Mailgun struct {
	Client     *http.Client
	BaseURL    string
	APIKey     string
	DomainName string
}

func NewMailgun(client *http.Client, baseURL, apiKey, domainName string) *Mailgun {
	return &Mailgun{
		Client:     client,
		BaseURL:    baseURL,
		APIKey:     apiKey,
		DomainName: domainName,
	}
}

func addRecepient(w *multipart.Writer, key string, recipients []string) error {
	for i, recipient := range recipients {
		if err := w.WriteField(fmt.Sprintf("%s[%d]", key, i), recipient); err != nil {
			return fmt.Errorf("writing %s: %w", key, err)
		}
	}
	return nil
}

func (m *Mailgun) SendEmail(ctx context.Context, email *Email) (string, error) {
	if err := email.Validate(); err != nil {
		return "", fmt.Errorf("validating email: %w", err)
	}
	data := &bytes.Buffer{}
	w := multipart.NewWriter(data)
	defer w.Close()
	if err := w.WriteField("from", email.From); err != nil {
		return "", fmt.Errorf("writing from: %w", err)
	}
	if err := addRecepient(w, "to", email.To); err != nil {
		return "", err
	}
	if err := addRecepient(w, "cc", email.Cc); err != nil {
		return "", err
	}
	if err := addRecepient(w, "bcc", email.Bcc); err != nil {
		return "", err
	}
	if err := w.WriteField("subject", email.Subject); err != nil {
		return "", fmt.Errorf("writing subject: %w", err)
	}
	if err := w.WriteField("text", email.Body); err != nil {
		return "", fmt.Errorf("writing text: %w", err)
	}
	w.Close()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v3/%s/messages", m.BaseURL, m.DomainName), bytes.NewReader(data.Bytes()))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	r.SetBasicAuth("api", m.APIKey)
	r.Header.Set("Content-Type", w.FormDataContentType())
	res, err := m.Client.Do(r)
	if err != nil {
		return "", fmt.Errorf("sending email: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return result.ID, nil
}
