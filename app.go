package app

import (
	"net/http"

	"google.golang.org/appengine"
)

func init() {
	http.HandleFunc("/cron/snapshot", snapashotHandle)
	http.HandleFunc("/cron/snapshoting", snapashotingHandle)
	http.HandleFunc("/cron/snapshot/delete", snapashotDeleteHandle)
	http.HandleFunc("/cron/deleteing", deleteingHandle)
	appengine.Main()
}
