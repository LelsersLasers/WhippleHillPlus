pragma foreign_keys = ON;

-- - users
-- 	- id (int, pk)
-- 	- username (text)
-- 	- name (text)
-- 	- password_hash (text)
-- 	- ics_link (text, unique)
-- 	- timezone (text)
-- - sessions
-- 	- id (int, pk)
-- 	- token (text)
-- 	- expiration (text)
-- 	- user_id (int, fk)
-- - semesters
--  - id (int, pk)
--  - name (text)
--  - sort_order (int)
--  - user_id (int, fk)
-- - classes
-- 	- id (int, pk)
-- 	- name (text)
--  - semester_id (int, fk)
-- - assignments
-- 	- id (int, pk)
-- 	- name (text)
-- 	- description (text)
-- 	- due_date (date)
-- 	- due_time (time)
-- 	- assigned_date (date)
-- 	- class_id (int, fk)
-- 	- status (text) ["Not Started", "In Progress", "Completed"]
--  - type (text) ["Homework", "Quiz", "Test", "Project", "Paper", "Other"]

CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT    NOT NULL UNIQUE,
	name          TEXT    NOT NULL,
	password_hash TEXT    NOT NULL,
	ics_link	  TEXT    UNIQUE, 
	timezone 	  TEXT
);

CREATE TABLE IF NOT EXISTS sessions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	token         TEXT    NOT NULL,
    expiration    TEXT    NOT NULL, -- Unix timestamp

    -- 1 user : many sessions
    user_id       INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS semesters (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    sort_order INTEGER NOT NULL,

    -- 1 user : many semesters
    user_id    INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS classes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    semester_id INTEGER NOT NULL,

    -- 1 semester : many classes
    FOREIGN KEY (semester_id) REFERENCES semesters (id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS assignments (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT    NOT NULL,
	description   TEXT,
	due_date      DATE    NOT NULL,
	due_time      TIME,
	assigned_date DATE    NOT NULL,
	status        TEXT    CHECK (status IN ('Not Started', 'In Progress', 'Completed'))             DEFAULT 'Not Started',
	type          TEXT    CHECK (type IN ('Homework', 'Quiz', 'Test', 'Project', 'Paper', 'Other')) DEFAULT 'Homework',

	-- 1 class : many assignments
	class_id      INTEGER NOT NULL,
	FOREIGN KEY (class_id) REFERENCES classes (id)
		ON DELETE CASCADE
);