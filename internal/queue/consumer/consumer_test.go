package consumer

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogConsumer(t *testing.T) {
	counts := make(map[string]int)
	c := LogConsumer{
		name:     "test",
		handlers: make(map[string]NotificationHandler),
	}
	c.HandleFunc("a", func(n entity.Notification) error {
		counts["a"] = counts["a"] + 1
		return nil
	})
	c.HandleFunc("b", func(n entity.Notification) error {
		counts["b"] = counts["b"] + 1
		return nil
	})
	err := c.Submit("a", entity.Notification{ID: "1"})
	assert.NoError(t, err)
	err = c.Submit("a", entity.Notification{ID: "2"})
	assert.NoError(t, err)
	err = c.Submit("b", entity.Notification{ID: "3"})
	assert.NoError(t, err)
	err = c.Submit("c", entity.Notification{ID: "4"})
	assert.Error(t, err)
	assert.Equal(t, 2, counts["a"])
	assert.Equal(t, 1, counts["b"])
}
