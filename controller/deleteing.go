package controller

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudMile/gae_snapshot_gce_disk/model"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// DeleteingHandle is deleting snapshot
func DeleteingHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/deleteing" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	cs := model.ComputeService{Ctx: ctx}
	cs.Get()
	if cs.Error != nil {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	gceSnapashotsList := rangeCanDeleteSnapshotsProject(ctx, cs.ComputeService)
	rangeSnapshotsDelete(ctx, cs.ComputeService, gceSnapashotsList)
}

func rangeCanDeleteSnapshotsProject(ctx context.Context, computeService *compute.Service) (gceSnapashotsList []string) {
	req := computeService.Snapshots.List(getProjectID(ctx))
	req = req.Filter("labels.expired_can_delete:yes")

	if err := req.Pages(ctx, func(page *compute.SnapshotList) error {
		for _, snapahot := range page.Items {
			snapahotCreatedAt, _ := time.Parse(time.RFC3339Nano, snapahot.CreationTimestamp)
			daysAgo, _ := strconv.Atoi(os.Getenv("DAYS_AGO"))
			manyDayBefore := time.Now().AddDate(0, 0, (0 - daysAgo))

			if snapahotCreatedAt.Before(manyDayBefore) {
				log.Infof(ctx, "Found snapshot.Name: %s", snapahot.Name)
				gceSnapashotsList = append(gceSnapashotsList, snapahot.Name)
			}
		}
		return nil
	}); err != nil {
		log.Errorf(ctx, "compute.SnapshotList error: %s", err)
	}
	return
}

func rangeSnapshotsDelete(ctx context.Context, computeService *compute.Service, gceSnapashotsList []string) {
	projectID := getProjectID(ctx)

	for _, snapshotName := range gceSnapashotsList {
		_, err := computeService.Snapshots.Delete(projectID, snapshotName).Context(ctx).Do()
		if err != nil {
			log.Errorf(ctx, "Delete snapsht %s error: %s", snapshotName, err)
		}
	}
}
