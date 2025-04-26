-- Migration: Add ics_link column to users table

BEGIN TRANSACTION;

-- 1. Rename old table
ALTER TABLE users RENAME TO users_old;

-- 2. Create new table with the updated schema
CREATE TABLE users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT    NOT NULL UNIQUE,
	name          TEXT    NOT NULL,
	password_hash TEXT    NOT NULL,
	ics_link      TEXT    UNIQUE -- UUID for the ICS link
);

-- 3. Copy data from old table to new one, leaving ics_link as NULL
INSERT INTO users (id, username, name, password_hash)
SELECT id, username, name, password_hash FROM users_old;

-- 4. Drop old table
DROP TABLE users_old;

COMMIT;
