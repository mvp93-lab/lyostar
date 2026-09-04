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
- Thumbnail Pipeline: Downscale cover images to WebP format (max width: 400px), store at `/data/cache/covers/{file_sha256}.webp`.
- Frontend: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS, Lucide Icons (`lucide-vue-next`).
  - Design Language: Clean, dark-mode first (Deep slate `#090a0f`, subtle 1px borders, Glacier Blue `#38bdf8` accents). Responsive and legible on OLED screens and slow-refresh E-Ink browsers.
  - Web Reader: Integrate `foliate-js`. ALWAYS wrap reader instance strictly inside `shallowRef()` (never `ref()` or `reactive()`). Clean up instances explicitly in `onBeforeUnmount`.
  - SPA Routing: Backend HTTP router must fallback unmatched non-API routes to `index.html` to prevent 404s on browser reload.

## 4. Scope Control
- MVP Scope: Scanning local EPUB files, extracting Dublin Core metadata/covers, SQLite indexing, shelf web UI, in-browser reader.
- Out of MVP Scope: Do NOT implement OPDS feeds, Send-to-Kindle, or multi-user authentication until explicitly instructed.