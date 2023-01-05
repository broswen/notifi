package api

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func (api *API) HandleGetNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := notificationId(r)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
		n, err := api.Notification.Get(r.Context(), id)
		if err != nil {
			log.Error().Err(err).Str("id", id).Msg("error getting notification")
			writeErr(w, nil, err)
			return
		}

		err = writeOK(w, http.StatusOK, n)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}

func (api *API) HandleListNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := pagination(r)
		deleted := r.URL.Query().Get("deleted")
		includeDeleted := deleted == "true"
		n, err := api.Notification.List(r.Context(), includeDeleted, page.Offset, page.Limit)
		if err != nil {
			log.Error().Err(err).Msg("error listing notifications")
			writeErr(w, nil, err)
			return
		}
		err = writeOK(w, http.StatusOK, n)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}

func (api *API) HandleDeleteNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := notificationId(r)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
		n, err := api.Notification.Delete(r.Context(), id)
		if err != nil {
			log.Error().Err(err).Str("id", id).Msg("error deleting notification")
			writeErr(w, nil, err)
			return
		}
		NotificationDeleted.Inc()
		err = writeOK(w, http.StatusOK, n)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}

func (api *API) HandleCreateNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &NotificationRequest{}
		err := readJSON(w, r, req)
		if err != nil {
			writeErr(w, nil, ErrBadRequest.WithError(err))
			return
		}
		defer r.Body.Close()

		if err = req.Validate(); err != nil {
			log.Error().Err(err).Msg("error invalid notification request")
			writeErr(w, nil, err)
			return
		}

		n := req.IntoEntity()
		err = api.Producer.Submit(n)
		if err != nil {
			log.Error().Err(err).Str("id", n.ID).Msg("error submitting notification")
			writeErr(w, nil, err)
			return
		}
		NotificationCreated.Inc()
		err = writeOK(w, http.StatusOK, n)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}
