CREATE TABLE tweets (
    id        uuid,
    username  text,
    text      text,
    timestamp timestamp,
    PRIMARY KEY (id)
);

CREATE TABLE likes (
    username text,
    tweet_id uuid,
    PRIMARY KEY ((username, tweet_id))
);

CREATE INDEX tweet_id ON likes (tweet_id);