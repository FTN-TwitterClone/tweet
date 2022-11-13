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
*Create table for tweets:*
```
CREATE TABLE tweets (
    id uuid,
    username text,
    text text,
    timestamp timestamp,
    PRIMARY KEY (id)
);
```
*Create table for user profile:*
```
CREATE TABLE user_profile (
    username text,
    timestamp timestamp,
    tweet_id uuid,
    PRIMARY KEY (username, timestamp)) 
    WITH CLUSTERING ORDER BY (timestamp DESC);
```
*Create table for user feed:*
```
CREATE TABLE user_feed (
    username text,
    timestamp timestamp,
    tweet_id uuid,
    PRIMARY KEY (username, timestamp)) 
    WITH CLUSTERING ORDER BY (timestamp DESC);
```
*Create table for likes:*
```
CREATE TABLE likes (
    username text,
    tweet_id uuid,
    PRIMARY KEY ((username, tweet_id))
);
```
*Create second index on tweet_id in likes table:*
```
CREATE INDEX tweet_id ON likes(tweet_id);
```
*Get all from table:*
```
SELECT * FROM <table_name> ;
```