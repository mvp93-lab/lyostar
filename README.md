# Lyostar 📚✨

**English** | [Tiếng Việt](README_vi.md)

---

**Lyostar** is an ultra-lightweight, single-binary, self-hosted ebook server and modern web reader.

Engineered with resource minimalism in mind: compiled binary `< 30MB`, idle RAM consumption `< 30MB`, zero external dependencies (no Redis, RabbitMQ, or mandatory Docker daemons). It runs effortlessly on everything from small cloud VPS instances and home servers (e.g., Raspberry Pi) to low-spec hardware and slow-refresh E-Ink devices.

---

## ✨ Features

- 🚀 **Single-Binary Deployment**: The entire Vue 3 Single Page Application (SPA) is embedded directly into the Go executable via `embed.FS`. A single binary serves both the backend API and frontend web application.
- 📖 **Modern In-Browser Readers**:
  - **EPUB Reader**: Powered by Foliate.js for smooth reading, customizable typography, fast pagination, and optimal contrast on OLED and E-Ink displays.
  - **PDF Reader**: Integrated with Mozilla's official PDF.js viewer featuring Hi-DPI vector rendering, pixel-perfect selectable text layer, and continuous vertical scrolling.
- 🔄 **Automatic Progress Tracking & Resume**:
  - Tracks reading locations per user (Foliate CFI for EPUB, page numbers for PDF).
  - Automatically records completion percentage and lets you resume right where you left off via the **"Continue Reading"** shelf.
- ⚡ **Automated Library Scanner**:
  - Scans local book directories (`.epub`, `.pdf`), extracts metadata (titles, authors, descriptions), and caches covers automatically.
  - Cover optimization pipeline: downscales and converts covers to high-quality **WebP** images.
  - Storage safety: The `/books` directory is treated as **STRICTLY READ-ONLY** (files are never moved, renamed, or modified).
- 🔐 **Authentication & Multi-User RBAC**:
  - Role-Based Access Control (**Admin** and **Reader** roles).
  - Secure `bcrypt` password hashing, SQLite session store with 64-character CSPRNG hex tokens delivered via secure `HttpOnly`, `SameSite=Lax` cookies.
  - First-run web setup wizard to configure the initial administrator account.
- 🔍 **Full-Text Search (FTS5)**: Built-in SQLite FTS5 virtual table for instantaneous search across titles, authors, and series without external search engines.
- 🎨 **Dark-Mode First Design**: Clean UI with deep slate (`#090a0f`) and glacier blue (`#38bdf8`) accents, designed for responsiveness across desktop, tablet, mobile, and E-Ink browsers.

---

## 🛠️ Tech Stack

- **Backend**: Go 1.22+, `go-chi/chi/v5` router
- **Database**: SQLite (Pure Go `modernc.org/sqlite`, CGO-free, WAL mode)
- **Query Layer**: `sqlc` (Type-safe SQL compilation, no heavy ORMs)
- **Frontend**: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS, Lucide Icons
- **Reading Engines**: Foliate-js (EPUB), Mozilla PDF.js (PDF)

---

## 📋 Prerequisites

- **To run Lyostar**: Simply download the compiled binary (no Go, Node.js, or database runtime required).
- **To build from source**:
  - [Go](https://go.dev/) 1.22+
  - [Node.js](https://nodejs.org/) 18+ and npm
  - `make` (optional, for automation commands)

---

## 🚀 Quick Start

### Method 1: Using Makefile (Recommended)

Lyostar includes a [Makefile](file:///home/mvp/Workspace/Other/lyostar/Makefile) for single-command workflows:

#### 1. Run directly from source (Development):
```bash
make dev
```
> The server starts on `http://localhost:8080` serving both the backend and embedded frontend.

#### 2. Build the production single-binary:
```bash
make build
```
This automatically builds the Vue 3 frontend assets and compiles the Go binary with size optimizations (`-ldflags="-s -w"` ~19MB).

#### 3. Run the compiled binary:
```bash
make run
# or run the binary directly:
./lyostar
```

---

### Method 2: Manual Build (Without Make)

If `make` is not available on your system, execute the following steps:

#### Step 1: Build the Frontend
```bash
cd frontend
npm install
npm run build
cd ..
```

#### Step 2: Compile the Go Binary
```bash
go build -ldflags="-s -w" -o lyostar ./cmd/lyostar
```

#### Step 3: Start the Server
```bash
./lyostar
```

Open your browser and navigate to: `http://localhost:8080`.

---

### Method 3: Using Docker & Docker Compose (Production Ready)

Lyostar provides an ultra-lightweight, multi-stage Docker image (`~31.6MB`) running with non-root security (`lyostar:1000`) and automated healthchecks.

#### 1. Start with Docker Compose:
```bash
docker compose up -d
# or via Makefile
make docker-up
```

#### 2. Check logs:
```bash
docker compose logs -f
# or via Makefile
make docker-logs
```

#### 3. Stop the container:
```bash
docker compose down
# or via Makefile
make docker-down
```

Your books in `./books` are mounted as **STRICTLY READ-ONLY** (`:ro`), while `./data` persists SQLite databases, uploads, and cover caches safely across container restarts.


---

## ⚙️ Configuration

Configure Lyostar using environment variables:

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | HTTP port for the web server |
| `BOOKS_DIR` | `./books` | Path to your ebook library (**Read-Only**) |
| `DATA_DIR` | `./data` | Application state directory (SQLite `app.db`, cover cache) |

**Example with custom paths:**
```bash
PORT=9000 BOOKS_DIR=/mnt/storage/books DATA_DIR=/mnt/storage/lyostar-data ./lyostar
```

---

## 🧑‍💻 Developer Guide

### Live-Reload Development (Backend + Frontend HMR)
If you are actively developing the Vue UI and want instant hot-reloading:

```bash
make dev-live
```
This runs the Go backend on port `8080` and the Vite dev server on port `3000` concurrently with API proxying.

### Useful Makefile Commands:
- `make test`: Run all unit tests with verbose output.
- `make test-short`: Run unit tests concisely.
- `make clean`: Remove compiled binaries.
- `make sqlc`: Regenerate Go SQL code from `queries.sql`.
- `make help`: Show all available commands.

---

## 📂 Project Structure

```text
lyostar/
├── cmd/lyostar/          # Go application entrypoint (main.go)
├── internal/             # Core backend logic
│   ├── api/              # HTTP router, middleware, REST endpoints
│   ├── auth/             # Session store, RBAC, bcrypt hashing
│   ├── config/           # Configuration & environment loader
│   ├── database/         # SQLite connection, schema, sqlc queries
│   ├── epub/             # EPUB metadata parser & cover extractor
│   ├── pdf/              # PDF metadata parser & cover extractor
│   └── scanner/          # Background directory scanner & workers
├── frontend/             # Vue 3 SPA frontend
│   ├── src/              # Components, Views, Composables
│   ├── public/           # Static assets (pre-built PDF.js viewer)
│   ├── dist/             # Production build embedded into Go binary
│   └── embed.go          # go:embed directive for dist/
├── Makefile              # Automation targets
├── README.md             # English documentation (Default)
├── README_vi.md          # Vietnamese documentation
└── AGENTS.md             # System directives & architecture rules
```

---

## 📜 License

Open source and community friendly. Contributions and feature suggestions are welcome!
