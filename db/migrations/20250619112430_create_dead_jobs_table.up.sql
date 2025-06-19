CREATE TABLE dead_jobs (
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
