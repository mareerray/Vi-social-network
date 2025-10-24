
CREATE TABLE IF NOT EXISTS users (
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

