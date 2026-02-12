You are a senior full-stack engineer and system architect.

Build an MVP video streaming platform (YouTube-like) focused on short films and feature films for emerging creators.

TECH STACK (MANDATORY):
- Frontend: Next.js (App Router, TypeScript)
- Backend API: Go (Gin framework)
- Database: PostgreSQL
- Cache / Queue: Redis
- Video processing: FFmpeg
- Storage: Cloudflare R2 (S3-compatible)
- CDN delivery: Cloudflare CDN
- Streaming format: HLS (HTTP Live Streaming)

GOAL:
Create a self-hosted video pipeline (no Cloudflare Stream, no paid services).

================================================
CORE REQUIREMENTS
================================================

1) USER & AUTH
- Simple email/password auth (JWT-based)
- Roles: USER, CREATOR, ADMIN

2) FILM MODEL
Each film must support:
- title
- description
- duration
- type: SHORT_FILM | FEATURE_FILM
- status: DRAFT | UPLOADED | TRANSCODING | READY | FAILED
- thumbnail_url
- hls_master_url
- created_by (user_id)

3) UPLOAD FLOW (IMPORTANT)
- Backend generates a pre-signed upload URL for Cloudflare R2
- Frontend uploads video DIRECTLY to R2 (backend must not receive video file)
- After upload, backend enqueues a transcoding job in Redis

4) VIDEO TRANSCODING WORKER
Implement a Go worker that:
- Pulls jobs from Redis
- Downloads source MP4 from R2
- Uses FFmpeg to generate HLS output:
  - 360p
  - 720p
- Generates:
  - master.m3u8
  - segmented .ts files
- Uploads HLS output back to R2
- Updates film status to READY or FAILED

5) STORAGE STRUCTURE (STRICT)
Use this exact structure in R2:

original/{filmId}/source.mp4
thumb/{filmId}/poster.jpg
hls/{filmId}/master.m3u8
hls/{filmId}/360p/index.m3u8
hls/{filmId}/360p/seg_00001.ts
hls/{filmId}/720p/index.m3u8
hls/{filmId}/720p/seg_00001.ts

6) PLAYBACK
- Use hls.js in Next.js
- Fallback to native HLS for Safari
- Player must auto-switch quality (adaptive bitrate)

7) API DESIGN
Implement REST APIs:
- POST /auth/register
- POST /auth/login
- POST /films
- POST /films/{id}/upload-url
- POST /films/{id}/publish
- GET  /films/{id}
- GET  /films/{id}/playback

8) DATABASE
Design PostgreSQL schema for:
- users
- films
- video_assets
- transcode_jobs

Include SQL migration files.

9) SECURITY (MINIMUM)
- Upload URLs must expire (10–30 minutes)
- Users can only upload to their own film
- Public read access ONLY for HLS files

================================================
DELIVERABLES
================================================

Produce:
1) Backend Go project structure
2) Redis worker code
3) FFmpeg command implementation
4) Next.js frontend pages:
   - Upload page
   - Film detail page
   - Video player component
5) PostgreSQL schema + migrations
6) Clear inline comments explaining design decisions

================================================
RULES
================================================
- Do NOT use mock services
- Do NOT skip FFmpeg implementation
- Do NOT abstract video logic away
- Code must be runnable locally
- Prefer clarity over overengineering

Start by scaffolding the backend Go service, then the worker, then frontend.


