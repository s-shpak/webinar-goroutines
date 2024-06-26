package server

import (
	"log"
	"migrations/internal/appctx"
	"migrations/internal/server/middleware"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

func InitServer(h *Handlers) *http.Server {
	return &http.Server{
		Addr:    ":8080",
		Handler: initRouter(h),
	}
}

func initRouter(h *Handlers) http.Handler {
	r := httprouter.New()
	r.Handle("POST", "/employee", middleware.AddRequestID(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.PutEmployee(r.Context(), w, r)
	}))
	r.Handle("GET", "/employee-by-email/:email", middleware.AddRequestID(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id, err := appctx.ExtractID(r.Context())
		if err != nil {
			panic(err)
		}
		log.Println(id)
		emailSanitized := p.ByName("email")
		email, err := url.PathUnescape(emailSanitized)
		if err != nil {
			log.Printf("failed to unescape the received email: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.GetEmployeeByEmail(r.Context(), w, email)
	}))
	return r
}
