package controller

import (
	"github.com/gorilla/mux"
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

	if len(tweet.Text) == 0 && len(tweet.ImageId) == 0 {
		http.Error(w, "Text and image can't be blank", 500)
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

func (c *TweetController) CreateAd(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.CreateAd")
	defer span.End()

	authUser := ctx.Value("authUser").(model.AuthUser)
	if authUser.Role != "ROLE_BUSINESS" {
		http.Error(w, "You are not a business user", 403)
		return
	}

	ad, err := json.DecodeJson[model.Ad](req.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if len(ad.Tweet.Text) == 0 && len(ad.Tweet.ImageId) == 0 {
		http.Error(w, "Text and image can't be blank", 500)
		return
	}

	newAd, appErr := c.tweetService.CreateAd(ctx, ad, authUser)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, newAd)
}

func (c *TweetController) CreateLike(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.CreateLike")
	defer span.End()

	id := mux.Vars(req)["id"]

	newLike, appErr := c.tweetService.CreateLike(ctx, id)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, newLike)
}

func (c *TweetController) DeleteLike(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.DeleteLike")
	defer span.End()

	id := mux.Vars(req)["id"]

	id, appErr := c.tweetService.DeleteLike(ctx, id)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, id)
}

func (c *TweetController) GetTimelineTweets(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.GetProfileTweets")
	defer span.End()

	username := mux.Vars(req)["username"]
	lastTweetId := req.URL.Query().Get("beforeId")

	tweets, appErr := c.tweetService.GetTimelineTweets(ctx, username, lastTweetId)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, tweets)
}

func (c *TweetController) GetLikesByTweet(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.GetLikesByTweet")
	defer span.End()

	tweetId := mux.Vars(req)["id"]

	tweets := c.tweetService.GetLikesByTweet(ctx, tweetId)

	json.EncodeJson(w, tweets)
}

func (c *TweetController) GetHomeFeed(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.GetHomeFeed")
	defer span.End()

	lastTweetId := req.URL.Query().Get("beforeId")

	tweets, appErr := c.tweetService.GetHomeFeed(ctx, lastTweetId)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, tweets)
}

func (c *TweetController) Retweet(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.Retweet")
	defer span.End()

	tweetId := mux.Vars(req)["id"]

	retweet, appErr := c.tweetService.Retweet(ctx, tweetId)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, retweet)
}

func (c *TweetController) SaveImage(w http.ResponseWriter, req *http.Request) {
	ctx, span := c.tracer.Start(req.Context(), "TweetController.SaveImage")
	defer span.End()

	imageName, appErr := c.tweetService.SaveImage(ctx, req)
	if appErr != nil {
		span.SetStatus(codes.Error, appErr.Error())
		http.Error(w, appErr.Message, appErr.Code)
		return
	}

	json.EncodeJson(w, imageName)
}
