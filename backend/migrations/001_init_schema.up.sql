-- Migration: Create initial schema for FilmTube platform
-- Up

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT users_role_check CHECK (role IN ('USER', 'CREATOR', 'ADMIN'))
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Films table
CREATE TABLE IF NOT EXISTS films (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    duration INTEGER DEFAULT 0, -- in seconds
    type VARCHAR(20) NOT NULL DEFAULT 'SHORT_FILM',
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    thumbnail_url TEXT,
    hls_master_url TEXT,
    created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT films_type_check CHECK (type IN ('SHORT_FILM', 'FEATURE_FILM')),
    CONSTRAINT films_status_check CHECK (status IN ('DRAFT', 'UPLOADED', 'TRANSCODING', 'READY', 'FAILED'))
);

-- Indexes for films
CREATE INDEX idx_films_created_by ON films(created_by_id);
CREATE INDEX idx_films_status ON films(status);
CREATE INDEX idx_films_type ON films(type);
CREATE INDEX idx_films_published_at ON films(published_at DESC);
CREATE INDEX idx_films_view_count ON films(view_count DESC);

-- Video assets table (for different quality levels)
CREATE TABLE IF NOT EXISTS video_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    film_id UUID NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    quality VARCHAR(10) NOT NULL, -- 360p, 720p, 1080p
    hls_index_url TEXT NOT NULL,
    size_bytes BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(film_id, quality)
);

-- Index for video assets
CREATE INDEX idx_video_assets_film_id ON video_assets(film_id);

-- Transcode jobs table
CREATE TABLE IF NOT EXISTS transcode_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    film_id UUID NOT NULL UNIQUE REFERENCES films(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'UPLOADED',
    error TEXT,
    progress INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT transcode_jobs_status_check CHECK (status IN ('UPLOADED', 'TRANSCODING', 'READY', 'FAILED'))
);

-- Index for transcode jobs (worker polls this)
CREATE INDEX idx_transcode_jobs_status ON transcode_jobs(status);
CREATE INDEX idx_transcode_jobs_created_at ON transcode_jobs(created_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to auto-update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_films_updated_at BEFORE UPDATE ON films
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
