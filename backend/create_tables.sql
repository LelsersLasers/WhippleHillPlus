pragma foreign_keys = ON;


-- - users
-- 	- id (int, pk)
-- 	- email (text)
-- 	- name (text)
-- 	- password_hash (text)
-- - classes
-- 	- id (int, pk)
-- 	- name (text)
-- 	- user_id (int, fk)
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
	id            INTEGER PRIMARY KEY,
	email         TEXT    NOT NULL UNIQUE,
	name          TEXT    NOT NULL,
	password_hash TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS classes (
	id      INTEGER PRIMARY KEY,
	name    TEXT    NOT NULL,

	-- 1 user : many classes
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users (id)
		ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS assignments (
	id            INTEGER PRIMARY KEY,
	name          TEXT    NOT NULL,
	description   TEXT    NOT NULL,
	due_date      DATE    NOT NULL,
	due_time      TIME,
	assigned_date DATE    NOT NULL,
	status        TEXT    CHECK (status IN ('Not Started', 'In Progress', 'Completed'))             DEFAULT 'Not Started',
	type          TEXT    CHECK (type IN ('Homework', 'Quiz', 'Test', 'Project', 'Paper', 'Other')) DEFAULT 'Homework',

	-- 1 class : many assignments
	class_id INTEGER NOT NULL,
	FOREIGN KEY (class_id) REFERENCES classes (id)
		ON DELETE CASCADE
);
