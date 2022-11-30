package cassandra

import (
	"context"
	"fmt"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/gocql/gocql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"tweet/model"
)

type CassandraTweetRepository struct {
	tracer  trace.Tracer
	session *gocql.Session
}

func NewCassandraTweetRepository(tracer trace.Tracer) (*CassandraTweetRepository, error) {
	err := initKeyspace()
	if err != nil {
		return nil, err
	}

	migrateDB()

	dbport := os.Getenv("DBPORT")
	db := os.Getenv("DB")
	host := fmt.Sprintf("%s:%s", db, dbport)

	cluster := gocql.NewCluster(host)
	cluster.ProtoVersion = 4
	cluster.Keyspace = "tweet_database"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	//defer session.Close()
	if err != nil {
		return nil, err
	}

	log.Printf("Connected OK!")

	return &CassandraTweetRepository{
		tracer:  tracer,
		session: session,
	}, nil
}

func initKeyspace() error {
	dbport := os.Getenv("DBPORT")
	db := os.Getenv("DB")
	host := fmt.Sprintf("%s:%s", db, dbport)

	cluster := gocql.NewCluster(host)
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	defer session.Close()

	if err != nil {
		return err
	}

	log.Printf("Connected OK!")

	err = session.Query("CREATE KEYSPACE IF NOT EXISTS tweet_database WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1}").Exec()
	if err != nil {
		return err
	}

	return nil
}

func migrateDB() error {
	dbport := os.Getenv("DBPORT")
	db := os.Getenv("DB")
	connString := fmt.Sprintf("cassandra://%s:%s/tweet_database?x-multi-statement=true", db, dbport)

	m, err := migrate.New("file://migrations", connString)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	return nil
}

func (r *CassandraTweetRepository) SaveTweet(ctx context.Context, tweet *model.Tweet, followers []*social_graph.SocialGraphUsername) error {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.SaveTweet")
	defer span.End()

	err := r.session.Query("INSERT INTO timeline_by_user (tweet_id, posted_by, text, retweet, original_posted_by) VALUES (?, ?, ?, ?, ?)").
		Bind(tweet.ID, tweet.PostedBy, tweet.Text, tweet.Retweet, tweet.OriginalPostedBy).
		Exec()

	// I want to see my tweet in feed
	followers = append(followers, &social_graph.SocialGraphUsername{Username: tweet.PostedBy})

	for _, follower := range followers {
		err = r.session.Query("INSERT INTO feed_by_user (tweet_id, username, posted_by, text, retweet, original_posted_by) VALUES (?, ?, ?, ?, ?, ?)").
			Bind(tweet.ID, follower.Username, tweet.PostedBy, tweet.Text, tweet.Retweet, tweet.OriginalPostedBy).
			Exec()
	}

	return err
}

func (r *CassandraTweetRepository) SaveLike(ctx context.Context, like *model.Like) error {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.SaveLike")
	defer span.End()

	err := r.session.Query("INSERT INTO likes (username, tweet_id) VALUES (?, ?)").
		Bind(like.Username, like.TweetId).
		Exec()

	return err
}

func (r *CassandraTweetRepository) DeleteLike(ctx context.Context, id string, username string) error {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.SaveLike")
	defer span.End()

	err := r.session.Query("DELETE FROM likes WHERE username = ? AND tweet_id = ?").
		Bind(username, id).
		Exec()

	return err
}

func (r *CassandraTweetRepository) CountLikes(ctx context.Context, tweetId *gocql.UUID) (int16, error) {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.CountLikes")
	defer span.End()

	var count int16
	err := r.session.Query("SELECT COUNT(*) FROM likes WHERE tweet_id = ?").
		Bind(tweetId).Consistency(gocql.One).Scan(&count)

	return count, err
}

func (r *CassandraTweetRepository) LikedByMe(ctx context.Context, tweetId *gocql.UUID) (bool, error) {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.LikedByMe")
	defer span.End()

	authUser := ctx.Value("authUser").(model.AuthUser)

	var count int16
	err := r.session.Query("SELECT COUNT(*) FROM likes WHERE tweet_id = ? and username = ?").
		Bind(tweetId, authUser.Username).Consistency(gocql.One).Scan(&count)

	return count >= 1, err
}

func (r *CassandraTweetRepository) GetTimelineTweets(ctx context.Context, username string, lastTweetId string) ([]model.TweetDTO, error) {
	repoCtx, span := r.tracer.Start(ctx, "CassandraTweetRepository.GetTimelineTweets")
	defer span.End()

	var tweets []model.TweetDTO
	var tweet model.TweetDTO

	var err error
	var iter *gocql.Iter

	if len(lastTweetId) > 0 {
		iter = r.session.Query("SELECT posted_by, tweet_id, text, retweet, original_posted_by, toTimestamp(tweet_id) FROM timeline_by_user WHERE posted_by = ? AND tweet_id < ? LIMIT 20").
			Bind(username, lastTweetId).Iter()
	} else {
		iter = r.session.Query("SELECT posted_by, tweet_id, text, retweet, original_posted_by, toTimestamp(tweet_id) FROM timeline_by_user WHERE posted_by = ? LIMIT 20").
			Bind(username).Iter()
	}

	for iter.Scan(&tweet.PostedBy, &tweet.ID, &tweet.Text, &tweet.Retweet, &tweet.OriginalPostedBy, &tweet.Timestamp) {

		tweet.LikesCount, err = r.CountLikes(repoCtx, &tweet.ID)
		if err != nil {
			tweet.LikesCount = 0
		}

		tweet.LikedByMe, err = r.LikedByMe(repoCtx, &tweet.ID)
		if err != nil {
			tweet.LikedByMe = false
		}

		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

func (r *CassandraTweetRepository) GetLikesByTweet(ctx context.Context, tweetId string) *[]model.Like {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.GetLikesByTweet")
	defer span.End()

	var likes []model.Like
	var like model.Like

	var iter *gocql.Iter

	iter = r.session.Query("SELECT username, tweet_id FROM likes WHERE tweet_id = ?").
		Bind(tweetId).Iter()

	for iter.Scan(&like.Username, &like.TweetId) {
		likes = append(likes, like)
	}

	return &likes
}

func (r *CassandraTweetRepository) GetFeedTweets(ctx context.Context, username string, lastTweetId string) ([]model.TweetDTO, error) {
	repoCtx, span := r.tracer.Start(ctx, "CassandraTweetRepository.GetFeedTweets")
	defer span.End()

	var tweets []model.TweetDTO
	var tweet model.TweetDTO

	var err error
	var iter *gocql.Iter

	if len(lastTweetId) > 0 {
		iter = r.session.Query("SELECT tweet_id, posted_by, text, retweet, original_posted_by, toTimestamp(tweet_id) FROM feed_by_user WHERE username = ? AND tweet_id < ? LIMIT 20").
			Bind(username, lastTweetId).Iter()
	} else {
		iter = r.session.Query("SELECT tweet_id, posted_by, text, retweet, original_posted_by, toTimestamp(tweet_id) FROM feed_by_user WHERE username = ? LIMIT 20").
			Bind(username).Iter()
	}

	for iter.Scan(&tweet.ID, &tweet.PostedBy, &tweet.Text, &tweet.Retweet, &tweet.OriginalPostedBy, &tweet.Timestamp) {

		tweet.LikesCount, err = r.CountLikes(repoCtx, &tweet.ID)
		if err != nil {
			tweet.LikesCount = 0
		}

		tweet.LikedByMe, err = r.LikedByMe(repoCtx, &tweet.ID)
		if err != nil {
			tweet.LikedByMe = false
		}

		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

func (r *CassandraTweetRepository) FindTweet(ctx context.Context, tweetId string) (model.Tweet, error) {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.FindTweet")
	defer span.End()

	var tweet model.Tweet
	err := r.session.Query("SELECT posted_by, tweet_id, text, retweet, original_posted_by, toTimestamp(tweet_id) FROM timeline_by_user WHERE tweet_id = ?").
		Bind(tweetId).Consistency(gocql.One).
		Scan(&tweet.PostedBy, &tweet.ID, &tweet.Text, &tweet.Retweet, &tweet.OriginalPostedBy, &tweet.TimeStamp)

	return tweet, err
}

func (r *CassandraTweetRepository) FindUserTweets(ctx context.Context, username string) []model.Tweet {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.FindUserTweets")
	defer span.End()

	var tweets []model.Tweet
	var tweet model.Tweet

	iter := r.session.Query("SELECT posted_by, tweet_id, text, retweet, original_posted_by FROM timeline_by_user WHERE posted_by = ?").
		Bind(username).Iter()

	for iter.Scan(&tweet.PostedBy, &tweet.ID, &tweet.Text, &tweet.Retweet, &tweet.OriginalPostedBy) {
		tweets = append(tweets, tweet)
	}

	return tweets
}

func (r *CassandraTweetRepository) UpdateFeed(ctx context.Context, authUsername string, username string) error {
	repoCtx, span := r.tracer.Start(ctx, "CassandraTweetRepository.UpdateFeed")
	defer span.End()

	tweets := r.FindUserTweets(repoCtx, username)

	var err error
	for _, tweet := range tweets {
		err = r.session.Query("INSERT INTO feed_by_user (tweet_id, username, posted_by, text, retweet, original_posted_by) VALUES (?, ?, ?, ?, ?, ?)").
			Bind(tweet.ID, authUsername, tweet.PostedBy, tweet.Text, tweet.Retweet, tweet.OriginalPostedBy).
			Exec()
	}

	return err
}
