package cassandra

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
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

func (r *CassandraTweetRepository) SaveTweet(ctx context.Context, tweet *model.Tweet) error {
	_, span := r.tracer.Start(ctx, "CassandraTweetRepository.SaveTweet")
	defer span.End()

	err := r.session.Query("INSERT INTO tweets (id, username, text, timestamp) VALUES (?, ?, ?, ?)").
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
