package controller

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CloudMile/gae_snapshot_gce_disk/model"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// SnapashotingHandle is snapshoting now
func SnapashotingHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cron/snapshoting" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ctx := appengine.NewContext(r)
	filterParams := r.PostFormValue("filter")
	log.Infof(ctx, "filter: %v", filterParams)
	cs := model.ComputeService{Ctx: ctx}
	cs.Get()
	if cs.Error != nil {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	gceZoneList := getGCEZone(ctx, cs.ComputeService)
	gceDiksMap := rangeDiskZone(ctx, cs.ComputeService, gceZoneList, filterParams)
	rangeCreateSnapshot(ctx, cs.ComputeService, gceDiksMap)
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
			if os.Getenv("STORAGE_LOCATION") != "none" {
				rb.StorageLocations = []string{os.Getenv("STORAGE_LOCATION")}
			}

			log.Infof(ctx, "create snapshot: %s => %s", zoneName, diskName)
			_, err := computeService.Disks.CreateSnapshot(projectID, zoneName, diskName, rb).Context(ctx).Do()
			if err != nil {
				log.Errorf(ctx, "CreateSnapshot error: %s", err)
			}
		}
	}
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
