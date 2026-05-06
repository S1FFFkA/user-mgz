package s3

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (r *Repository) PresignGetURL(ctx context.Context, objectKey, fileName string, asAttachment bool, expiresIn time.Duration) (string, error) {
	reqParams := make(url.Values)
	if fileName != "" {
		dispositionType := "inline"
		if asAttachment {
			dispositionType = "attachment"
		}
		reqParams.Set("response-content-disposition", dispositionType+`; filename="`+fileName+`"`)
	}

	u, err := r.client.PresignedGetObject(ctx, r.bucket, objectKey, expiresIn, reqParams)
	if err != nil {
		return "", fmt.Errorf("presign get url: %w", err)
	}
	return u.String(), nil
}
