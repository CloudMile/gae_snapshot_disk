package model

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// ComputeService is GCP compute setvice
type ComputeService struct {
	Ctx            context.Context
	ComputeService *compute.Service
	Error          error
}

// Get for setup GCP auth client service
func (cs *ComputeService) Get() {
	tokenSource, err := google.DefaultTokenSource(cs.Ctx, compute.ComputeScope)

	if err != nil {
		log.Errorf(cs.Ctx, "token error: %s", err)
		cs.Error = err
		return
	}

	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
			Base: &urlfetch.Transport{
				Context: cs.Ctx,
			},
		},
	}
	computeService, err := compute.New(client)

	if err != nil {
		log.Errorf(cs.Ctx, "compute service error: %s", err)
		cs.Error = err
		return
	}
	cs.ComputeService = computeService
}
