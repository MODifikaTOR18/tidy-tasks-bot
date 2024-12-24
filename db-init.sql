CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    name TEXT,
    is_premium BOOLEAN DEFAULT FALSE,
    premium_until DATE,
    registration_date DATE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(telegram_id),
    description TEXT NOT NULL,
    scheduled_time TIMESTAMP NOT NULL,
    is_recurring BOOLEAN DEFAULT FALSE,
    interval TEXT DEFAULT 'none',
    created_at TIMESTAMP DEFAULT NOW()
);
