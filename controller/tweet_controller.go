package controller

import (
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"tweet/service"
)

type TweetController struct {
	tweetService *service.TweetService
	tracer       trace.Tracer
}

func NewTweetController(tweetService *service.TweetService, tracer trace.Tracer) *TweetController {
	return &TweetController{
		tweetService,
		tracer,
	}
}

func (c *TweetController) AddTweet(w http.ResponseWriter, req *http.Request) {

}
