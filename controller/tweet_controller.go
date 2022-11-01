package controller

import (
	"net/http"
	"tweet/service"
)

type TweetController struct {
	tweetService *service.TweetService
}

func NewTweetController(tweetService *service.TweetService) *TweetController {
	return &TweetController{
		tweetService,
	}
}

func (c *TweetController) AddTweet(w http.ResponseWriter, req *http.Request) {

}
