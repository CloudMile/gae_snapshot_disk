package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

func snapashotHandle(w http.ResponseWriter, r *http.Request) {
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

func snapashotingHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/snapshoting" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	filterParams := r.PostFormValue("filter")
	log.Infof(ctx, "filter: %v", filterParams)

	computeService := getComputeService(ctx)
	gceZoneList := getGCEZone(ctx, computeService)
	gceDiksMap := rangeDiskZone(ctx, computeService, gceZoneList, filterParams)
	rangeCreateSnapshot(ctx, computeService, gceDiksMap)
}

func snapashotDeleteHandle(w http.ResponseWriter, r *http.Request) {
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

func deleteingHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/deleteing" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	computeService := getComputeService(ctx)
	gceSnapashotsList := rangeCanDeleteSnapshotsProject(ctx, computeService)
	rangeSnapshotsDelete(ctx, computeService, gceSnapashotsList)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		fmt.Fprint(w, "404 Not Found")
	case http.StatusMethodNotAllowed:
		fmt.Fprint(w, "405 Method Not Allow")
	default:
		fmt.Fprint(w, "Bad Request")
	}
}

func getComputeService(ctx context.Context) (computeService *compute.Service) {
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(ctx, compute.ComputeScope),
			Base: &urlfetch.Transport{
				Context: ctx,
			},
		},
	}
	computeService, err := compute.New(client)
	if err != nil {
		log.Errorf(ctx, "compute error: %s", err)
	}
	return
}

func getGCEZone(ctx context.Context, computeService *compute.Service) (gceZoneList []string) {
	req := computeService.Zones.List(getProjectID(ctx))

	if err := req.Pages(ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			gceZoneList = append(gceZoneList, zone.Name)
		}
		return nil
	}); err != nil {
		log.Errorf(ctx, "compute.ZoneList error: %s", err)
	}
	return
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

func rangeDiskZone(ctx context.Context, computeService *compute.Service, gceZoneList []string, filterParams string) (gceDiksMap map[string][]*compute.Disk) {
	projectID := getProjectID(ctx)
	gceDiksMapTemp := make(map[string][]*compute.Disk)

	for _, zoneName := range gceZoneList {
		req := computeService.Disks.List(projectID, zoneName)
		if filterParams != "" {
			req = req.Filter(filterParams)
		}
		if err := req.Pages(ctx, func(page *compute.DiskList) error {
			for _, disk := range page.Items {
				if len(disk.Users) > 0 {
					gceDiksMapTemp[zoneName] = append(gceDiksMapTemp[zoneName], disk)
				}
			}
			return nil
		}); err != nil {
			log.Errorf(ctx, "compute.DiskList error: %s", err)
		}
	}
	return gceDiksMapTemp
}

func rangeCreateSnapshot(ctx context.Context, computeService *compute.Service, gceDiksMap map[string][]*compute.Disk) {
	projectID := getProjectID(ctx)
	now := time.Now()
	tString := strings.ToLower(now.Format("06010215MST"))

	for zoneName, gceDiskList := range gceDiksMap {
		for _, disk := range gceDiskList {
			diskName := disk.Name
			rb := &compute.Snapshot{
				Name:   diskName + `-` + tString,
				Labels: disk.Labels,
			}
			log.Infof(ctx, "create snapshot: %s => %s", zoneName, diskName)
			_, err := computeService.Disks.CreateSnapshot(projectID, zoneName, diskName, rb).Context(ctx).Do()
			if err != nil {
				log.Errorf(ctx, "CreateSnapshot error: %s", err)
			}
		}
	}
}

func getProjectID(ctx context.Context) (projectID string) {
	if appengine.AppID(ctx) == "None" {
		projectID = os.Getenv("PROJECT_ID")
	} else {
		projectID = appengine.AppID(ctx)
	}
	return
}
