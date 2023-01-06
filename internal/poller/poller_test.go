package poller

import (
	"context"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var now = time.Now()

func TestSubmit(t *testing.T) {
	r := &repository.MockRepository{}
	pr := &producer.MockProducer{}
	n := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:  "test message",
		Schedule: &now,
	}
	p := NewScheduledNotificationPoller(r, pr, time.Minute, time.Minute*5, int64(10), 0, 99)
	pr.On("Submit", n).Return(nil)
	r.On("MarkSubmitted", mock.Anything, n.ID).Return(n, nil)
	err := p.Submit(context.Background(), n)
	assert.NoError(t, err)
	pr.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestPoll(t *testing.T) {
	r := &repository.MockRepository{}
	pr := &producer.MockProducer{}
	n := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:  "test message",
		Schedule: &now,
	}
	//TODO this doesn't really align with how the concrete implementation works
	p := NewScheduledNotificationPoller(r, pr, time.Second, time.Minute*5, int64(10), 0, 99)
	r.On("ListScheduled", mock.Anything, time.Minute*5, int64(0), int64(99), int64(0), int64(10)).Return([]entity.Notification{n, n}, nil)
	pr.On("Submit", n).Return(nil)
	r.On("MarkSubmitted", mock.Anything, mock.Anything).Return(n, nil)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	p.poll(ctx)
	r.AssertExpectations(t)
	pr.AssertExpectations(t)
	r.AssertNumberOfCalls(t, "MarkSubmitted", 2)
	pr.AssertNumberOfCalls(t, "Submit", 2)
}
