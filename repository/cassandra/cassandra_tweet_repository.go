package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
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
