package email

import (
	"context"
	"github.com/SimonSchneider/goslu/sid"
)

type Fake struct{}

func (f Fake) SendEmail(ctx context.Context, email *Email) (string, error) {
	return sid.NewString(15)
}
