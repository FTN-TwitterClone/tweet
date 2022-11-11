package controller

import (
	"github.com/go-playground/validator/v10"
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

func (c *TweetController) CreateTweet(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.CreateTweet")
	defer span.End()

	tweet, err := json.DecodeJson[model.Tweet](req.Body)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	newTweet, appErr := c.tweetService.CreateTweet(ctx, tweet)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, newTweet)
}

func (c *TweetController) CreateLike(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.CreateLike")
	defer span.End()

	like, err := json.DecodeJson[model.Like](req.Body)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if err := validator.New().Struct(like); err != nil {
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	newLike, appErr := c.tweetService.CreateLike(ctx, like)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, newLike)
}

func (c *TweetController) DeleteLike(w http.ResponseWriter, req *http.Request) {

}
