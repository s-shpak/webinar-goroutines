package appctx

import (
	"context"
	"errors"
	"fmt"
)

func SetID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyReqID, id)
}

var (
	ErrObjectNotFound   = errors.New("object not found in context")
	ErrObjectHasBadType = errors.New("object has an incorrect type")
)

func ExtractID(ctx context.Context) (string, error) {
	idVal := ctx.Value(ctxKeyReqID)
	if idVal == nil {
		return "", fmt.Errorf("failed to extract the request ID from context: %w", ErrObjectNotFound)
	}
	id, ok := idVal.(string)
	if !ok {
		return "", fmt.Errorf(
			"failed to extract the request ID from context: %w (actual type: %T)", ErrObjectHasBadType, idVal,
		)
	}
	return id, nil
}
