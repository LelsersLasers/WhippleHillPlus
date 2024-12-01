-- Step 1: Add the new semesters table
CREATE TABLE IF NOT EXISTS semesters (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    sort_order INTEGER NOT NULL,

    -- 1 user : many semesters
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
);

-- Step 2: Add the semester_id column to the classes table
ALTER TABLE classes ADD COLUMN semester_id INTEGER;

-- Step 3: Create default semester for each user
INSERT INTO semesters (name, sort_order, user_id)
SELECT '2028 Fall', 1, id
FROM users;

-- Step 4: Assign the default semester to existing classes
UPDATE classes
SET semester_id = (
    SELECT id
    FROM semesters
    WHERE semesters.user_id = classes.user_id
    AND semesters.name = '2028 Fall'
);

-- Step 5: Enforce the foreign key relationship for semester_id
PRAGMA foreign_keys=OFF;

CREATE TABLE classes_new (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    user_id     INTEGER NOT NULL,
    semester_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE,
    FOREIGN KEY (semester_id) REFERENCES semesters (id)
        ON DELETE CASCADE
);

INSERT INTO classes_new (id, name, user_id, semester_id)
SELECT id, name, user_id, semester_id
FROM classes;

DROP TABLE classes;
ALTER TABLE classes_new RENAME TO classes;

PRAGMA foreign_keys=ON;
