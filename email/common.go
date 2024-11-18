package email

import "fmt"

type Email struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

func (e *Email) Validate() error {
	if e.From == "" {
		return fmt.Errorf("from is required")
	}
	if len(e.To) == 0 {
		return fmt.Errorf("to is required")
	}
	if e.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if e.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}
