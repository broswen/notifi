package producer

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogProducer(t *testing.T) {
	p, err := NewProducer("test", "test", "logger")
	assert.NoError(t, err)
	err = p.Submit(entity.Notification{
		ID: "1",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:  "test",
		Schedule: nil,
	})
	assert.NoError(t, err)

	assert.IsType(t, p, &LogProducer{})
}
