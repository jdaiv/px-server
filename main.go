package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	middleware "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

var JWTSecret []byte

var DB *sql.DB

type Config struct {
	Addr      string
	DBConnStr string
	JWTSecret string
}

var configLocation = flag.String("config", "config.toml", "location of config file")

func main() {
	flag.Parse()

	config := Config{}
	if _, err := toml.DecodeFile(*configLocation, &config); err != nil {
		log.Printf("[server] error parsing config: %v", err)
		os.Exit(1)
	}

	JWTSecret = []byte(config.JWTSecret)

	rand.Seed(time.Now().UTC().UnixNano())

	r := mux.NewRouter()
	r.Use(middleware.RecoveryHandler())
	r.Use(middleware.CORS())
	r.HandleFunc("/api/ws", join).Methods("GET")
	r.HandleFunc("/api/auth/login", login).Methods("POST")
	r.HandleFunc("/api/auth/create", createUser).Methods("POST")

	var err error
	DB, err = sql.Open("postgres", config.DBConnStr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[server] connected to DB")

	go ClientMaintenace()
	go incomingMessages()
	go outgoingMessages()

	// t := time.Now()
	// go func() {
	// 	for {
	// 		_t := time.Now()
	// 		dt := _t.Sub(t).Seconds()
	// 		t = _t
	// 		for _, r := range rooms {
	// 			if r.Area != nil {
	// 				r.Area.Tick(dt)
	// 				r.Area.LateTick(dt)
	// 				update := r.Area.Send()
	// 				if update != nil {
	// 					r.Broadcast("room", "update", update)
	// 				}
	// 			}
	// 		}
	// 		time.Sleep(30 * time.Millisecond)
	// 	}
	// }()

	srv := &http.Server{
		Addr:         config.Addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	log.Printf("[server] starting on %s", config.Addr)

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
