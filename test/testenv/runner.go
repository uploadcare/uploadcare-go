package testenv

import (
	"github.com/uploadcare/uploadcare-go/v2/conversion"
	"github.com/uploadcare/uploadcare-go/v2/file"
	"github.com/uploadcare/uploadcare-go/v2/group"
	"github.com/uploadcare/uploadcare-go/v2/metadata"
	"github.com/uploadcare/uploadcare-go/v2/project"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
	"github.com/uploadcare/uploadcare-go/v2/upload"
	"github.com/uploadcare/uploadcare-go/v2/webhook"
)

// Runner holds service instances and test artifacts
type Runner struct {
	File       file.Service
	Group      group.Service
	Upload     upload.Service
	Conversion conversion.Service
	Webhook    webhook.Service
	Project    project.Service
	Metadata   metadata.Service

	Artifacts Artifacts
}

// Artifacts are test artifacts
type Artifacts struct {
	CustomStorage  string
	Files          []*file.Info
	GroupIDs       []string
	ConversionJobs []conversion.Job
	Webhook        webhook.Info
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
		Metadata:   metadata.NewService(client),
		Artifacts: Artifacts{
			CustomStorage: customStorage,
		},
	}
}
