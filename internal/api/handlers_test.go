package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
)

func TestHandleCreateNotification_Valid(t *testing.T) {
	body := `{
    "email": "test@example.com",
    "content": "test message"
}`

	m := &producer.MockProducer{}
	m.On("Submit", mock.Anything).Return(nil)
	r := &repository.MockRepository{}
	app := API{
		Producer:     m,
		Notification: r,
	}
	rr := &httptest.ResponseRecorder{}
	req, err := http.NewRequest(http.MethodPost, "/api/notifications", bytes.NewReader([]byte(body)))
	assert.NoError(t, err)
	router := chi.NewRouter()
	router.Post("/api/notifications", app.HandleCreateNotification())
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleCreateNotification_InvalidDestination(t *testing.T) {
	body := `{
    "email": "random",
    "content": "test message"
}`

	m := &producer.MockProducer{}
	m.On("Submit", mock.Anything).Return(nil)
	r := &repository.MockRepository{}
	app := API{
		Producer:     m,
		Notification: r,
	}
	rr := &httptest.ResponseRecorder{}
	req, err := http.NewRequest(http.MethodPost, "/api/notifications", bytes.NewReader([]byte(body)))
	assert.NoError(t, err)
	router := chi.NewRouter()
	router.Post("/api/notifications", app.HandleCreateNotification())
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleGetNotification(t *testing.T) {
	m := &producer.MockProducer{}
	r := &repository.MockRepository{}
	r.On("Get", mock.Anything, "2GYpi7DgQR2Jdm6zWVEtD6c78Jl").Return(entity.Notification{
		ID: "2GYpi7DgQR2Jdm6zWVEtD6c78Jl",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:     "test message",
		DeletedAt:   nil,
		CreatedAt:   time.Time{},
		ModifiedAt:  time.Time{},
		DeliveredAt: nil,
	}, nil)
	app := API{
		Producer:     m,
		Notification: r,
	}
	rr := &httptest.ResponseRecorder{}
	req, err := http.NewRequest(http.MethodGet, "/api/notifications/2GYpi7DgQR2Jdm6zWVEtD6c78Jl", nil)
	assert.NoError(t, err)
	router := chi.NewRouter()
	router.Get("/api/notifications/{notificationId}", app.HandleGetNotification())
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
