# FilmTube - Indie Film Streaming Platform

A self-hosted video streaming platform (YouTube-like) focused on short films and feature films for emerging creators.

## Architecture

```
filmtube/
├── backend/          # Go API Service (Gin framework)
├── worker/           # Go Transcoding Worker (FFmpeg)
├── frontend/         # Next.js App Router (React + Tailwind CSS)
└── docs/            # Design documents
```

## Tech Stack

- **Frontend**: Next.js (App Router), TypeScript, Tailwind CSS, HLS.js
- **Backend API**: Go (Gin framework), PostgreSQL, Redis
- **Storage**: Cloudflare R2 (S3-compatible)
- **Streaming**: HLS (HTTP Live Streaming) with adaptive bitrate
- **Transcoding**: FFmpeg (360p, 720p)

## Quick Start

### Prerequisites

1. **PostgreSQL** (local install)
2. **Redis** (local install)
3. **FFmpeg** installed and available in PATH
4. **Cloudflare R2** account with bucket created

### 1. Database Setup

```bash
# Create database
createdb filmtube

# Run migrations
psql filmtube < backend/migrations/001_init_schema.up.sql
```

### 2. Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your values:
# - DATABASE_URL (PostgreSQL connection string)
# - R2_* credentials (Cloudflare R2)
# - JWT_SECRET (generate a secure random string)
```

### 3. Run Backend API

```bash
cd backend

# Install dependencies
go mod download

# Run server
go run cmd/server/main.go
```

API runs on `http://localhost:8080`

### 4. Run Transcoding Worker

```bash
cd worker

# Install dependencies
go mod download

# Run worker
go run cmd/worker/main.go
```

Worker polls Redis for transcoding jobs.

### 5. Run Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run dev server
npm run dev
```

Frontend runs on `http://localhost:3000`

## API Endpoints

### Auth
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `GET /api/auth/me` - Get current user (protected)

### Films
- `GET /api/films` - List films (public)
- `GET /api/films/:id` - Get film details (public)
- `GET /api/films/:id/playback` - Get HLS playback URL (public)
- `POST /api/films` - Create film (creator)
- `POST /api/films/:id/upload-url` - Generate upload URL (creator)
- `POST /api/films/:id/confirm-upload` - Confirm upload (creator)
- `POST /api/films/:id/publish` - Publish film (creator)

## Storage Structure

R2 bucket structure:
```
original/{filmId}/source.mp4      # Original uploaded video
thumb/{filmId}/poster.jpg         # Generated thumbnail
hls/{filmId}/master.m3u8        # HLS master playlist
hls/{filmId}/360p/index.m3u8    # 360p quality
hls/{filmId}/360p/seg_*.ts       # 360p segments
hls/{filmId}/720p/index.m3u8    # 720p quality
hls/{filmId}/720p/seg_*.ts       # 720p segments
```

## Upload Flow

1. Frontend creates film via `POST /api/films`
2. Frontend requests upload URL via `POST /api/films/:id/upload-url`
3. Backend generates pre-signed R2 URL (expires in 30 min)
4. Frontend uploads video DIRECTLY to R2 (not through backend)
5. Frontend confirms upload via `POST /api/films/:id/confirm-upload`
6. Backend enqueues transcoding job in Redis
7. Worker picks up job, downloads from R2, transcodes with FFmpeg
8. Worker uploads HLS output back to R2
9. Worker updates film status to READY

## Security

- Upload URLs expire (30 minutes)
- Users can only upload to their own films
- Public read access ONLY for HLS files
- JWT-based authentication with role-based access

## License

MIT
