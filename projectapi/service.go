package projectapi

import (
	"context"

	"github.com/uploadcare/uploadcare-go/v2/internal/config"
	"github.com/uploadcare/uploadcare-go/v2/internal/svc"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
)

// Service uses a client created with ucare.NewBearerClient.
type Service interface {
	List(ctx context.Context, params *ListParams) (*ProjectList, error)
	Create(ctx context.Context, params CreateProjectParams) (Project, error)
	Get(ctx context.Context, pubKey string) (Project, error)
	Update(ctx context.Context, pubKey string, params UpdateProjectParams) (Project, error)
	Delete(ctx context.Context, pubKey string) error

	ListSecrets(ctx context.Context, pubKey string, params *ListParams) (*SecretList, error)
	CreateSecret(ctx context.Context, pubKey string) (SecretRevealed, error)
	DeleteSecret(ctx context.Context, pubKey string, secretID string) error

	GetUsage(ctx context.Context, pubKey string, params UsageDateRange) (UsageMetricsCombined, error)
	GetUsageMetric(ctx context.Context, pubKey string, metric UsageMetricName, params UsageDateRange) (UsageMetric, error)
}

type service struct {
	svc svc.Service
}

func NewService(client ucare.Client) Service {
	return service{svc.New(config.RESTAPIEndpoint, client, log)}
}
