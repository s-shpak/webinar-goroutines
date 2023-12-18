package middleware

import (
	"migrations/internal/appctx"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func AddRequestID(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := uuid.NewString()
		ctx := appctx.SetID(r.Context(), id)
		next(w, r.WithContext(ctx), ps)
	}
}
