## Cassandra commands:
***
*List all keyspaces:*
```
DESCRIBE KEYSPACES ; 
```
*Create keyspace (database):*
```
CREATE KEYSPACE tweet_database WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'datacentar1' : 1} ; 
```
*Use the keyspace:*
```
USE tweet_database ;
```
*Create table for tweets:*
```
CREATE TABLE tweets (
    id text,
    author text,
    text text,
    timestamp timestamp,
    PRIMARY KEY (id)
);
```
*Get all from tweets table:*
```
SELECT * FROM tweets ;
```