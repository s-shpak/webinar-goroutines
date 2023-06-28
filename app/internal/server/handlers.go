package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"migrations/internal/model"
	"migrations/internal/store"
)

type Handlers struct {
	store *store.DB
}

func NewHandlers(s *store.DB) *Handlers {
	return &Handlers{
		store: s,
	}
}

func (h *Handlers) PutEmployee(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read the PutEmployee request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var emp model.Employee
	if err := json.Unmarshal(b, &emp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.store.PutEmployee(ctx, &emp); err != nil {
		log.Printf("failed to store employee in the DB: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) GetEmployeeByEmail(ctx context.Context, w http.ResponseWriter, email string) {
	emp, err := h.store.GetEmployeeByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrEmployeeNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("failed to get employee by email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(emp)
	if err != nil {
		log.Printf("failed to marshal the retrieved employee: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
