CREATE TABLE user_profile (
    username text,
    tweet_id timeuuid,
    text text,
    PRIMARY KEY (username, tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);

CREATE TABLE likes (
    username text,
    tweet_id timeuuid,
    PRIMARY KEY ((username, tweet_id))
);

CREATE INDEX tweet_id ON likes (tweet_id);