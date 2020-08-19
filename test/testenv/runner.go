package testenv

import (
	"github.com/uploadcare/uploadcare-go/conversion"
	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/group"
	"github.com/uploadcare/uploadcare-go/project"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/upload"
	"github.com/uploadcare/uploadcare-go/webhook"
)

// Runner holds service instances and test artifacts
type Runner struct {
	File       file.Service
	Group      group.Service
	Upload     upload.Service
	Conversion conversion.Service
	Webhook    webhook.Service
	Project    project.Service

	Artifacts Artifacts
}

// Artifacts are test artifacts
type Artifacts struct {
	CustomStorage  string
	Files          []*file.Info
	GroupIDs       []string
	ConversionJobs []conversion.Job
	WebhookID      int64
}

// NewRunner returns new Runner instance
func NewRunner(client ucare.Client, customStorage string) *Runner {
	return &Runner{
		File:       file.NewService(client),
		Group:      group.NewService(client),
		Upload:     upload.NewService(client),
		Conversion: conversion.NewService(client),
		Webhook:    webhook.NewService(client),
		Project:    project.NewService(client),
		Artifacts: Artifacts{
			CustomStorage: customStorage,
		},
	}
}
