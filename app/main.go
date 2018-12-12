package main

import (
	"net/http"

	"github.com/CloudMile/gae_snapshot_gce_disk/controller"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/cron/snapshot", controller.SnapashotHandle)
	http.HandleFunc("/cron/snapshoting", controller.SnapashotingHandle)
	http.HandleFunc("/cron/snapshot/delete", controller.SnapashotDeleteHandle)
	http.HandleFunc("/cron/deleteing", controller.DeleteingHandle)
	appengine.Main()
}
