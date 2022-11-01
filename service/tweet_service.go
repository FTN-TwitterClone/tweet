package service

import "tweet/repository"

type TweetService struct {
	tweetRepository repository.TweetRepository
}

func NewTweetService(tweetRepository repository.TweetRepository) *TweetService {
	return &TweetService{
		tweetRepository,
	}
}
