package controller

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

// SnapashotHandle is queue for snapshot
func SnapashotHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/snapshot" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "query: %v", r.URL.Query())

	t := taskqueue.NewPOSTTask("/cron/snapshoting", r.URL.Query())
	if _, err := taskqueue.Add(ctx, t, "create-snapshot"); err != nil {
		errorHandler(w, r, http.StatusInternalServerError)
		return
	}
}
