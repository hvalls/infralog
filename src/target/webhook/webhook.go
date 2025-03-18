package webhook

import (
	"fmt"
	"infralog/tfstate"
)

type WebhookTarget struct {
	URL string
}

func New(url string) *WebhookTarget {
	return &WebhookTarget{
		URL: url,
	}
}

func (t *WebhookTarget) Write(d *tfstate.StateDiff) error {
	fmt.Println("Sending diff to", t.URL)
	return nil
}
