package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	middleware "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

var JWTSecret = []byte("very good secret")

var DB *sql.DB

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.Use(middleware.RecoveryHandler())
	r.Use(middleware.CORS())
	r.HandleFunc("/api/ws", join).Methods("GET")
	r.HandleFunc("/api/auth", login).Methods("POST")

	var err error
	DB, err = sql.Open("postgres", "user=postgres password=cactus dbname=game sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[server] connected to DB")

	AddDefaultRooms()
	go incomingMessages()

	srv := &http.Server{
		Addr:         *addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	log.Printf("[server] starting on %s", *addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("[server] shutting down")
	os.Exit(0)
}
