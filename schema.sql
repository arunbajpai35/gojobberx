CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    payload TEXT,
    type TEXT,
    duration INT,
    status TEXT,
    retries INT,
    max_retries INT,
    priority TEXT DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dead_jobs (
    id UUID PRIMARY KEY,
    payload TEXT NOT NULL,
    type TEXT NOT NULL,
    duration INTEGER NOT NULL,
    retries INTEGER,
    max_retries INTEGER,
    priority TEXT,
    failure_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    failed_at TIMESTAMPTZ DEFAULT now()
);
