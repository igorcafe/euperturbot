CREATE TABLE message_temp AS SELECT * FROM message;

DROP TABLE message;

CREATE TABLE message (
    id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    date TIMESTAMP NOT NULL,
    text TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    user_name TEXT NOT NULL,
    PRIMARY KEY (id, chat_id)
);

INSERT INTO message SELECT * FROM message_temp;

DROP TABLE message_temp;