package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/key-value-store/pkg/db"
	"github.com/key-value-store/pkg/handler"
)

func main() {
	db := db.NewDB()

	h, err := handler.New(db)
	if err != nil {
		log.Printf("could not create handler: %v", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	go func() {
		// graceful shutdown
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		<-interrupt
		log.Print("app is shutting down...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("could not shutdown: %v\n", err)
		}
	}()

	log.Printf("app is ready to listen and serve on port 8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
