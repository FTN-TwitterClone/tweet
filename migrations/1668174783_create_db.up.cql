CREATE TABLE timeline_by_user (
    posted_by text,
    tweet_id timeuuid,
    text text,
    retweet boolean,
    original_posted_by text,
    PRIMARY KEY (posted_by, tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);

CREATE INDEX timeline_tweet_id ON timeline_by_user (tweet_id);

CREATE TABLE likes (
    username text,
    tweet_id timeuuid,
    PRIMARY KEY ((username, tweet_id))
);

CREATE INDEX likes_tweet_id ON likes (tweet_id);

CREATE TABLE feed_by_user (
    tweet_id timeuuid,
    posted_by text,
    username text,
    text text,
    retweet boolean,
    original_posted_by text,
    PRIMARY KEY (username, tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);