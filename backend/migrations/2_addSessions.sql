CREATE TABLE IF NOT EXISTS sessions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	token         TEXT    NOT NULL,
    expiration    TEXT    NOT NULL, -- Unix timestamp

    -- 1 user : many sessions
    user_id       INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
);
