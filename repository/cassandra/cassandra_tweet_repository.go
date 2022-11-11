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
	connString := fmt.Sprintf("cassandra://%s:%s/tweet_database", db, dbport)

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
