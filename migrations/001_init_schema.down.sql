-- Migration: Rollback initial schema for FilmTube platform
-- Down

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_films_updated_at ON films;
DROP FUNCTION IF EXISTS update_updated_at_column;

DROP TABLE IF EXISTS transcode_jobs;
DROP TABLE IF EXISTS video_assets;
DROP TABLE IF EXISTS films;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;
