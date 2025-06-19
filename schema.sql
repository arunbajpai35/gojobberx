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
