package controller

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"tweet/controller/json"
	"tweet/model"
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
	ctx, span := c.tracer.Start(req.Context(), "TweetController.AddTweet")
	defer span.End()

	tweet, err := json.DecodeJson[model.Tweet](req.Body)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	newTweet, appErr := c.tweetService.AddTweet(ctx, tweet)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, newTweet)
}
