## Cassandra commands:
***
*List all keyspaces:*
```
DESCRIBE KEYSPACES ; 
```
*Create keyspace (database):*
```
CREATE KEYSPACE tweet_database WITH REPLICATION = {'class' : 'SimpleStrategy', 'replication_factor' : 1} ; 
```
*Use the keyspace:*
```
USE tweet_database ;
```
*Create table for user profile:*
```
CREATE TABLE timeline_by_user (
    posted_by text,
    tweet_id timeuuid,
    text text,
    image_id text,
    retweet boolean,
    original_posted_by text,
    PRIMARY KEY ((posted_by), tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);
    
CREATE INDEX timeline_tweet_id ON timeline_by_user (tweet_id);
```
*Create table for user feed:*
```
CREATE TABLE feed_by_user (
    username text,
    tweet_id timeuuid,
    posted_by text,
    text text,
    image_id text,
    retweet boolean,
    original_posted_by text,
    PRIMARY KEY ((username), tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);
```
*Create table for likes:*
```
CREATE TABLE likes (
    tweet_id timeuuid,
    username text,
    PRIMARY KEY ((tweet_id), username)
);
```
*Get all from table:*
```
SELECT * FROM <table_name> ;
```