CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tenant_profile (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    address     TEXT,
    postal_code TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS school (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    code         TEXT NOT NULL,
    npsn         TEXT,
    school_types TEXT[] NOT NULL DEFAULT '{}',
    alamat       TEXT,
    status       TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT,
    username        TEXT,
    phone           TEXT,
    password_hash   TEXT NOT NULL DEFAULT '',
    role            TEXT NOT NULL CHECK (role IN ('student', 'admin_school', 'admin_exam', 'admin_store', 'super_admin')),
    name            TEXT NOT NULL DEFAULT '',
    school_id       UUID REFERENCES school (id),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated', 'deleted')),
    otp_enabled     BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- student-only fields (NULL for admin roles)
    nis             TEXT,
    dob             DATE,
    gender          TEXT CHECK (gender IN ('m', 'f')),
    grade           INT,
    alamat_domisili TEXT,
    target_exam     TEXT,

    CHECK (email IS NOT NULL OR username IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active
    ON users (email)
    WHERE email IS NOT NULL AND status != 'deleted';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username
    ON users (username)
    WHERE username IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_school_id
    ON users (school_id);

CREATE INDEX IF NOT EXISTS idx_users_role_status
    ON users (role, status);
