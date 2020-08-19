package project

import (
	"context"
	"net/http"
)

// Info gets information about account project.
func (s service) Info(
	ctx context.Context,
) (data Info, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		infoPathFormat,
		nil,
		&data,
	)
	return
}

// Info holds project related information
type Info struct {
	// Project login name
	Name string `json:"name"`
	// Project public key
	PubKey string `json:"pub_key"`

	// AutostoreEnabled
	AutostoreEnabled bool `json:"autostore_enabled"`

	// List of project collaborators
	Collaborators []Collaborator `json:"collaborators"`
}

// Collaborator name and email
type Collaborator struct {
	// Collaborator name
	Name string `json:"name"`
	// Collaborator email
	Email string `json:"email"`
}
