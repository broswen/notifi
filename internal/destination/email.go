package destination

import (
	"context"
	"fmt"
	"github.com/broswen/notifi/internal/entity"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
	"time"
)

// EmailDestination delivers notifications using sendgrid
type EmailDestination struct {
	name    string
	address string
	client  *sendgrid.Client
}

func NewEmailDestination(apiKey string, name, address string) (Destination, error) {
	client := sendgrid.NewSendClient(apiKey)
	return &EmailDestination{
		client: client,
	}, nil
}

func (d *EmailDestination) Deliver(n entity.Notification) error {
	m := mail.NewV3Mail()
	from := mail.NewEmail(d.name, d.address)
	m.SetFrom(from)
	m.Subject = "automated email"
	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail("", n.Destination.Email))
	m.AddPersonalizations(p)
	c := mail.NewContent("text/plain", n.Content)
	m.AddContent(c)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := d.client.SendWithContext(ctx, m)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendgrid send email: %d %s", resp.StatusCode, resp.Body)
	}
	return nil
}
