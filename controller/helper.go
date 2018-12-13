package controller

import (
	"context"
	"os"

	"google.golang.org/appengine"
)

func getProjectID(ctx context.Context) (projectID string) {
	if appengine.AppID(ctx) == "None" {
		projectID = os.Getenv("PROJECT_ID")
	} else {
		projectID = appengine.AppID(ctx)
	}
	return
}
