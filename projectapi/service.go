package projectapi

import (
	"context"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

// Service describes all Project API operations.
// The client passed to NewService must be created with ucare.NewBearerClient.
type Service interface {
	// List returns a paginated list of accessible projects.
	List(ctx context.Context, params *ListParams) (ProjectList, error)
	// Create creates a new project.
	Create(ctx context.Context, params CreateProjectParams) (Project, error)
	// Get returns project info by public key.
	Get(ctx context.Context, pubKey string) (Project, error)
	// Update updates project settings.
	Update(ctx context.Context, pubKey string, params UpdateProjectParams) (Project, error)
	// Delete deletes a project.
	Delete(ctx context.Context, pubKey string) error

	// ListSecrets returns secret keys for a project.
	ListSecrets(ctx context.Context, pubKey string, params *ListParams) (SecretList, error)
	// CreateSecret creates a new secret key for a project.
	CreateSecret(ctx context.Context, pubKey string) (SecretRevealed, error)
	// DeleteSecret deletes a secret key.
	DeleteSecret(ctx context.Context, pubKey string, secretID string) error

	// GetUsage returns combined usage metrics for a project.
	GetUsage(ctx context.Context, pubKey string, params UsageDateRange) (UsageMetricsCombined, error)
	// GetUsageMetric returns daily usage for a specific metric.
	GetUsageMetric(ctx context.Context, pubKey string, metric string, params UsageDateRange) (UsageMetric, error)
}

type service struct {
	svc svc.Service
}

// NewService returns a new instance of the Project API Service.
// The client must be created with ucare.NewBearerClient.
func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
