package controller

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

// SnapashotDeleteHandle is queue for delete snapshot
func SnapashotDeleteHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/snapshot/delete" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)

	t := taskqueue.NewPOSTTask("/cron/deleteing", make(map[string][]string))
	if _, err := taskqueue.Add(ctx, t, "delete-snapshot"); err != nil {
		log.Errorf(ctx, "Error is %s", err)
		errorHandler(w, r, http.StatusInternalServerError)
		return
	}
}
