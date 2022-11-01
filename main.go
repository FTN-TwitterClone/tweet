package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tweet/controller"
	"tweet/repository/cassandra"
	"tweet/service"
	"tweet/tracer"
)

func main() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	tracer, closer := tracer.Init("tweet_service")
	opentracing.SetGlobalTracer(tracer)

	tweetRepository, err := cassandra.NewCassandraTweetRepository()
	if err != nil {
		log.Fatal(err)
	}

	tweetService := service.NewTweetService(tweetRepository)

	tweetController := controller.NewTweetController(tweetService)

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/tweet/add/", tweetController.AddTweet).Methods("POST")

	// start server
	srv := &http.Server{Addr: "0.0.0.0:8002", Handler: router}
	go func() {
		log.Println("server starting")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	<-quit

	log.Println("service shutting down ...")

	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")

	if err := closer.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("traces saved")
}
