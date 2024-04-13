CREATE TABLE user (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    first_name TEXT NOT NULL
);

CREATE TABLE user_topic (
    id INTEGER PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    topic TEXT NOT NULL,
    UNIQUE(chat_id, user_id, topic)
);

CREATE TABLE event (
    id INTEGER PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    time TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    msg_id INTEGER NOT NULL,
    UNIQUE(chat_id, msg_id, name)
);

CREATE TABLE poll (
    id TEXT PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    topic TEXT NOT NULL,
    result_message_id INTEGER NOT NULL
);

CREATE TABLE poll_vote (
    poll_id TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    vote INTEGER NOT NULL,
    FOREIGN KEY (poll_id) REFERENCES poll(id),
    PRIMARY KEY(poll_id, user_id)
);

CREATE TABLE voice (
    file_id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL
);