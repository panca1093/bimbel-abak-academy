ALTER TABLE users ADD COLUMN auth_provider TEXT NOT NULL DEFAULT 'password'
    CHECK (auth_provider IN ('password', 'google'));
