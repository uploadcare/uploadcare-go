package projectapi

import (
	"context"
	"net/http"
)

const (
	metaMimeTypesPath            = "/meta/mime/types/"
	metaModerationCategoriesPath = "/meta/moderation/categories/"
)

func (s service) ListMimeTypes(ctx context.Context) (data MimeTypes, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		metaMimeTypesPath,
		nil,
		&data,
	)
	return
}

func (s service) ListModerationCategories(ctx context.Context) (data ModerationCategories, err error) {
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		metaModerationCategoriesPath,
		nil,
		&data,
	)
	return
}
