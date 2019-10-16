package testenv

import (
	"github.com/uploadcare/uploadcare-go/conversion"
	"github.com/uploadcare/uploadcare-go/file"
	"github.com/uploadcare/uploadcare-go/group"
	"github.com/uploadcare/uploadcare-go/ucare"
	"github.com/uploadcare/uploadcare-go/upload"
)

// Runner holds service instances and test artifacts
type Runner struct {
	File       file.Service
	Group      group.Service
	Upload     upload.Service
	Conversion conversion.Service

	Artifacts Artifacts
}

// Artifacts are test artifacts
type Artifacts struct {
	Files          []*file.Info
	GroupIDs       []string
	ConversionJobs []conversion.Job
}

// NewRunner returns new Runner instance
func NewRunner(client ucare.Client) *Runner {
	return &Runner{
		File:       file.NewService(client),
		Group:      group.NewService(client),
		Upload:     upload.NewService(client),
		Conversion: conversion.NewService(client),
	}
}
