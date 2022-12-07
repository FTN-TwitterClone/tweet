CREATE TABLE timeline_by_user (
    posted_by text,
    tweet_id timeuuid,
    text text,
    image_id text,
    retweet boolean,
    original_posted_by text,
    ad boolean,
    PRIMARY KEY ((posted_by), tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);

CREATE INDEX timeline_tweet_id ON timeline_by_user (tweet_id);

CREATE TABLE likes (
    tweet_id timeuuid,
    username text,
    PRIMARY KEY ((tweet_id), username)
);

CREATE TABLE feed_by_user (
    username text,
    tweet_id timeuuid,
    posted_by text,
    text text,
    image_id text,
    retweet boolean,
    original_posted_by text,
    ad boolean
    PRIMARY KEY ((username), tweet_id)
)
    WITH CLUSTERING ORDER BY (tweet_id DESC);