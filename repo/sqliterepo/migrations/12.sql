CREATE TABLE scheduled_topic (
    chat_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL,
    topic TEXT NOT NULL,
    status TEXT NOT NULL
        CHECK (status IN ('created', 'failed', 'completed'))
        DEFAULT 'created',
    time TEXT NOT NULL,
    UNIQUE (chat_id, topic, time)
);
