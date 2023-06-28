package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"migrations/internal/config"
	"migrations/internal/server"
	"migrations/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("bye-bye")
}

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func run() (err error) {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	cfg := config.GetConfig()
	db, err := store.NewDB(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize a new DB: %w", err)
	}

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer log.Print("closed DB")
		defer wg.Done()
		<-ctx.Done()

		db.Close()
	}()

	componentsErrs := make(chan error, 1)

	h := server.NewHandlers(db)
	srv := server.InitServer(h)
	go func(errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and server has failed: %w", err)
		}
	}(componentsErrs)

	wg.Add(1)
	go func() {
		defer log.Print("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		log.Print(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	}()

	return nil
}
