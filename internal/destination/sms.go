package destination

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// SMSDestination delivers sms notifications using twilio
type SMSDestination struct {
	number string
	client *twilio.RestClient
}

func NewSMSDestination(accountSid, token, number string) (Destination, error) {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Password:   token,
		AccountSid: accountSid,
	})
	return &SMSDestination{
		client: client,
		number: number,
	}, nil
}

func (d *SMSDestination) Deliver(n entity.Notification) error {

	params := &twilioApi.CreateMessageParams{
		Body: &n.Content,
		From: &d.number,
		To:   &n.Destination.SMS,
	}
	//TODO add retry with exponential backoff on failure
	_, err := d.client.Api.CreateMessage(params)
	if err != nil {
		return err
	}
	NotificationDelivered.WithLabelValues("sms").Inc()
	return nil
}
