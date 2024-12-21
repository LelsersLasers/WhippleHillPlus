-- Step 1: Create a new table without the user_id column
CREATE TABLE classes_new (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    semester_id INTEGER NOT NULL,

    -- 1 semester : many classes
    FOREIGN KEY (semester_id) REFERENCES semesters (id)
        ON DELETE CASCADE
);

-- Step 2: Copy data from the old table to the new table
INSERT INTO classes_new (id, name, semester_id)
SELECT id, name, semester_id
FROM classes;

-- Step 3: Drop the old table
DROP TABLE classes;

-- Step 4: Rename the new table to the original name
ALTER TABLE classes_new RENAME TO classes;
