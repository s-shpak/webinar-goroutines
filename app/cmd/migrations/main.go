package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"migrations/internal/config"
	"migrations/internal/server"
	"migrations/internal/store"

	"golang.org/x/sync/errgroup"
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
	// корневой контекст приложения
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	// нештатное завершение программы по таймауту
	// происходит, если после завершения контекста
	// приложение не смогло завершиться за отведенный промежуток времени
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	cfg := config.GetConfig()

	// открытие соединения – относительно долгая, io-bound операция
	// она использует контекст
	db, err := store.NewDB(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize a new DB: %w", err)
	}

	// отслеживаем успешное закрытие соединения с БД
	g.Go(func() error {
		defer log.Print("closed DB")

		<-ctx.Done()

		db.Close()
		return nil
	})

	h := server.NewHandlers(db)
	srv := server.InitServer(h)

	// запуск сервера
	g.Go(func() (err error) {
		defer func() {
			errRec := recover()
			if errRec != nil {
				err = fmt.Errorf("a panic occurred: %v", errRec)
			}
		}()
		if err = srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			return fmt.Errorf("listen and server has failed: %w", err)
		}
		return nil
	})

	// отслеживаем успешное завершение работы сервера
	g.Go(func() error {
		defer log.Print("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Print(err)
	}

	return nil
}
