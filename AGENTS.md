# Project Lyostar - System Directives & Engineering Rules

## 1. Role & Identity
You are an expert systems engineer building "Lyostar", an ultra-lightweight, single-binary, self-hosted ebook server and modern reader.

## 2. Invariant Rules & System Boundaries
- Distribution: The entire application MUST compile into a single binary. Embed Vue 3 static production assets directly via Go's `embed.FS`.
- Resource Caps: Standalone binary < 30MB, RAM idle consumption < 30MB. Production Docker image < 50MB.
- Storage Policy:
  - Directory `/books`: STRICTLY READ-ONLY. Never rename, modify, write to, or move files here.
  - Directory `/data`: Application state only (SQLite database `app.db`, WebP thumbnail cache).
- Concurrency & Architecture: Single-process model only. STRICTLY PROHIBITED to introduce external message brokers (Redis, RabbitMQ, Celery). Background tasks must use buffered Go channels with fixed-size worker pools.

## 3. Technology Stack Constraints
- Backend: Go 1.22+ using standard library `net/http` (or `go-chi/chi/v5`).
- Database: SQLite using `modernc.org/sqlite` (pure Go, CGO-free) or `mattn/go-sqlite3`.
  - Enforce PRAGMAs: `PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000; PRAGMA synchronous=NORMAL;`
  - Query layer: Use `sqlc` only. Do NOT use ORMs (e.g., GORM, Ent).
  - Search: Use SQLite FTS5 virtual tables over titles, authors, and series.
- EPUB Engine: Parse `META-INF/container.xml` and `.opf` manifests using standard library `archive/zip` and `encoding/xml`. NEVER extract the entire EPUB archive to disk.
- PDF Engine: Parse `/Info` dictionary and Catalog `/Metadata` using pure Go (CGO-free). Extract first embedded JPEG cover stream. Serve with `Accept-Ranges: bytes` for fast in-browser streaming.
- Authentication & Sessions: Pure Go `bcrypt` password hashing. Zero-dependency session management stored in SQLite `sessions` table via secure 64-char CSPRNG hex tokens. Transport via `HttpOnly`, `SameSite=Lax` cookies. Role-based Access Control (`admin`, `reader`) and granular user permissions matching Calibre-Web architecture (`can_read`, `can_download`, `can_upload`, `can_edit`, `can_delete`). Admin can configure capabilities for any user and self.
- Reading Progress & Resume: Per-user reading progress stored in SQLite `reading_progress` table (composite key `user_id, book_id`). Tracks location (Foliate EPUB CFI or PDF page number), progress fraction (0.0-1.0), and finished status. Readers resume automatically from saved position.
- Thumbnail Pipeline: Downscale cover images to WebP format (max width: 400px), store at `/data/cache/covers/{file_sha256}.webp`.
- Frontend: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS, Lucide Icons (`lucide-vue-next`).
  - Design Language: Clean, dark-mode first (Deep slate `#090a0f`, subtle 1px borders, Glacier Blue `#38bdf8` accents). Responsive and legible on OLED screens and slow-refresh E-Ink browsers.
  - Web Reader: Integrate `foliate-js` for EPUB (ALWAYS wrap reader instance strictly inside `shallowRef()`, never `ref()` or `reactive()`. Clean up instances explicitly in `onBeforeUnmount`). Use Mozilla PDF.js official pre-built viewer (matching Calibre-Web architecture) with Hi-DPI vector rendering, pixel-perfect TextLayer (selectable/copyable text), continuous vertical scrolling, and automated background reading progress tracking & resume.
  - SPA Routing: Backend HTTP router must fallback unmatched non-API routes to `index.html` to prevent 404s on browser reload.

## 4. Scope Control
- Core Scope: Scanning local EPUB & PDF files, extracting metadata/covers, SQLite indexing, shelf web UI, in-browser EPUB/PDF readers, multi-user authentication & RBAC (Admin & Reader), first-run setup wizard, per-user reading progress tracking & resume, "Continue Reading" shelf section.
- Out of MVP Scope: Do NOT implement OPDS feeds or Send-to-Kindle until explicitly instructed.

## 5. Specification Maintenance Directive
- Self-Updating Mandate: Whenever an architectural decision, supported format, major dependency, or feature is added, altered, or deprecated upon user instruction, the engineer/agent MUST proactively update this `AGENTS.md` file to keep the Directives, Technology Stack Constraints, and Scope Control sections fully synchronized with the actual codebase.
- Feature Tracking Mandate: The project feature matrix and implementation roadmap is tracked in `docs/features.md`. Whenever a feature is implemented, modified, or completed, the engineer/agent MUST proactively update `docs/features.md` to check off completed items (`[x]`) and record any newly introduced capabilities.