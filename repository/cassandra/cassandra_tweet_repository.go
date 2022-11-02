package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
)

type CassandraTweetRepository struct {
	session *gocql.Session
}

func NewCassandraTweetRepository() (*CassandraTweetRepository, error) {
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
		session: session,
	}, nil
}
