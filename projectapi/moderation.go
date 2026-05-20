package projectapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/uploadcare/uploadcare-go/v2/internal/codec"
)

const moderationThresholdsPathFmt = "/projects/%s/moderation/thresholds/"

func (s service) GetModerationThresholds(
	ctx context.Context,
	pubKey string,
) (data ModerationThresholds, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodGet,
		moderationThresholdsPath(pubKey),
		nil,
		&data,
	)
	return
}

func (s service) SetModerationThresholds(
	ctx context.Context,
	pubKey string,
	thresholds []ModerationThresholdParams,
) (data ModerationThresholds, err error) {
	if err = validatePubKey(pubKey); err != nil {
		return
	}
	err = s.svc.ResourceOp(
		ctx,
		http.MethodPut,
		moderationThresholdsPath(pubKey),
		setModerationThresholdsParams(thresholds),
		&data,
	)
	return
}

type setModerationThresholdsParams []ModerationThresholdParams

func (p setModerationThresholdsParams) EncodeReq(req *http.Request) error {
	if p == nil {
		p = setModerationThresholdsParams{}
	}
	return codec.EncodeReqBody(p, req)
}

func moderationThresholdsPath(pubKey string) string {
	return fmt.Sprintf(moderationThresholdsPathFmt, url.PathEscape(pubKey))
}
