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
	// корневой контекст приложения
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	// нештатное завершение программы по таймауту
	// происходит, если после завершения контекста
	// приложение не смогло завершиться за отведенный промежуток времени
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	// WaitGroup для отслеживания того, какие компоненты приложения
	// успешно завершились
	wg := &sync.WaitGroup{}
	defer func() {
		// при выходе из функции ожидаем завершения компонентов приложения
		wg.Wait()
	}()

	cfg := config.GetConfig()

	// открытие соединения – относительно долгая, io-bound операция
	// она использует контекст
	db, err := store.NewDB(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize a new DB: %w", err)
	}

	// отслеживаем успешное закрытие соединения с БД
	watchDB(ctx, wg, db)

	h := server.NewHandlers(db)
	srv := server.InitServer(h)

	// канал для получения критических ошибок компонентов
	componentsErrs := make(chan error, 1)
	// запускаем и отслеживаем успешное завершение работы сервера
	manageServer(ctx, wg, srv, componentsErrs)

	select {
	// завершаем выполнение при закрытии корневого контекста
	case <-ctx.Done():
	// завершаем выполнение, если в компоненте произошла критическая ошибка
	case err := <-componentsErrs:
		log.Print(err)
		cancelCtx()
	}

	return nil
}

func watchDB(ctx context.Context, wg *sync.WaitGroup, db *store.DB) {
	wg.Add(1)
	go func() {
		defer log.Print("closed DB")
		defer wg.Done()

		<-ctx.Done()

		db.Close()
	}()
}

func manageServer(ctx context.Context, wg *sync.WaitGroup, srv *http.Server, errs chan<- error) {
	// запуск сервера
	go func(errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and server has failed: %w", err)
		}
	}(errs)

	// отслеживаем успешное завершение работы сервера
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
}
