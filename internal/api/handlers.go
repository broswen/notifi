package api

import (
	"net/http"
)

func handleCreateNotification() http.HandlerFunc {
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
		//submit to queue

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
