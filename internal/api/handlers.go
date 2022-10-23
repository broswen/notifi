package api

import (
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
			writeErr(w, nil, err)
			return
		}

		n := req.IntoEntity()
		err = api.Producer.Submit(n)
		if err != nil {
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
