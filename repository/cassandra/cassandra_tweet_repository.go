package cassandra

import (
	"context"
	"fmt"
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

func (r *CassandraTweetRepository) SaveTweet(ctx context.Context, tweet *model.Tweet) error {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.SaveTweet")
	defer span.End()

	err := r.session.Query("INSERT INTO tweets (id, username, text, timestamp) VALUES (?, ?, ?, ?)").
		Bind(tweet.ID, tweet.Username, tweet.Text, tweet.Timestamp).
		Exec()
	//table for rendering tweets on user profile
	err = r.session.Query("INSERT INTO user_profile (tweet_id, username, text, timestamp) VALUES (?, ?, ?, ?)").
		Bind(tweet.ID, tweet.Username, tweet.Text, tweet.Timestamp).
		Exec()

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

func (r *CassandraTweetRepository) GetProfileTweets(ctx context.Context, username string) (*[]model.TweetDTO, error) {
	repoCtx, span := r.tracer.Start(ctx, "CassandraTweetRepository.GetTweetsForProfile")
	defer span.End()

	var tweets []model.TweetDTO
	var tweet model.TweetDTO

	var err error

	iter := r.session.Query("SELECT username, tweet_id, text, timestamp FROM user_profile WHERE username = ?").
		Bind(username).Iter()

	for iter.Scan(&tweet.Username, &tweet.ID, &tweet.Text, &tweet.Timestamp) {

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

	return &tweets, nil
}
