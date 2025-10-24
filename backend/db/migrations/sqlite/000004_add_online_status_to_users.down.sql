-- recreate users table without online_status
CREATE TABLE IF NOT EXISTS users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth TEXT,
    avatar TEXT,
    nickname TEXT NOT NULL UNIQUE,
    about_me TEXT,
    profile_type TEXT CHECK(profile_type IN ('public', 'private')) DEFAULT 'public'
);
INSERT INTO users_new (id, email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type)
SELECT id, email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, profile_type FROM users;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
