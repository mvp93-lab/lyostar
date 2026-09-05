# Lyostar 📚✨

[English](README.md) | **Tiếng Việt**

---

**Lyostar** là một máy chủ ebook tự lưu trữ (self-hosted) và trình đọc sách trực tuyến hiện đại, siêu nhẹ, được đóng gói hoàn toàn trong **một file thực thi duy nhất (single-binary)**.

Hệ thống được thiết kế theo triết lý tối giản tài nguyên: dung lượng file thực thi `< 30MB`, mức tiêu thụ RAM khi chạy nhàn rỗi `< 30MB`, không phụ thuộc các dịch vụ ngoài (Redis, Docker daemon bắt buộc, v.v.), hoạt động mượt mà trên cả máy tính thông thường lẫn các thiết bị phần cứng khiêm tốn như Raspberry Pi hoặc máy đọc sách màn hình E-Ink.

---

## ✨ Tính năng nổi bật

- 🚀 **Single Binary Deployment**: Toàn bộ giao diện người dùng (Vue 3 SPA) được nhúng trực tiếp vào file thực thi Go thông qua `embed.FS`. Bạn chỉ cần đúng 1 file duy nhất để chạy toàn bộ hệ thống (cả Backend lẫn Frontend).
- 📖 **Trình đọc sách tích hợp cao cấp**:
  - **EPUB Reader**: Tích hợp engine Foliate.js hiển thị mượt mà, hỗ trợ font chữ, dàn trang, chuyển trang nhanh, tương thích tốt với màn hình OLED và trình duyệt E-Ink có tần số quét chậm.
  - **PDF Reader**: Tích hợp Mozilla PDF.js official viewer với khả năng render vector chuẩn Hi-DPI, bôi đen/copy chữ sắc nét (TextLayer) và cuộn dọc liên tục.
- 🔄 **Tự động lưu & Phục hồi tiến độ đọc**:
  - Theo dõi vị trí đọc chi tiết cho từng tài khoản (Foliate CFI đối với EPUB, số trang đối với PDF).
  - Tự động đồng bộ tỉ lệ hoàn thành (%) và mở lại đúng trang đang đọc dang dở ở mục **"Tiếp tục đọc" (Continue Reading)**.
- ⚡ **Quét & Phân tích thư viện tự động**:
  - Quét thư mục sách cục bộ (định dạng `.epub`, `.pdf`), trích xuất metadata (tiêu đề, tác giả, mô tả) và ảnh bìa tự động.
  - Pipeline tối ưu ảnh bìa: tự động nén và chuyển đổi sang định dạng **WebP** lưu trong bộ nhớ đệm cache.
  - Thư mục sách `/books` được bảo vệ theo chế độ **STRICTLY READ-ONLY** (không bao giờ đổi tên hay ghi đè vào file gốc của bạn).
- 🔐 **Bảo mật & Quản lý người dùng (RBAC)**:
  - Hệ thống tài khoản phân quyền: **Admin** (quản trị viên) và **Reader** (người đọc).
  - Hashing mật khẩu bằng `bcrypt`, phiên đăng nhập lưu trong SQLite với token hex CSPRNG 64 ký tự truyền qua Cookie bảo mật (`HttpOnly`, `SameSite=Lax`).
  - Trình hướng dẫn thiết lập tài khoản Admin lần đầu (First-run Setup Wizard) trực quan ngay khi truy cập trang web.
- 🔍 **Tìm kiếm toàn văn (Full-Text Search)**: Sử dụng SQLite FTS5 cho tốc độ tìm kiếm sách, tác giả tức thì mà không cần Elasticsearch.
- 🎨 **Giao diện hiện đại (Dark-mode First)**: Gam màu Deep Slate (`#090a0f`) phối Glacier Blue (`#38bdf8`), giao diện đáp ứng linh hoạt trên Desktop, Tablet và Mobile.

---

## 🛠️ Công nghệ sử dụng

- **Backend**: Go (Go 1.22+), router `go-chi/chi/v5`
- **Database**: SQLite (Pure Go `modernc.org/sqlite`, CGO-free, WAL mode)
- **Truy vấn CSDL**: `sqlc` (Type-safe SQL queries, không dùng ORM nặng nề)
- **Frontend**: Vue 3 (Composition API, `<script setup>`), Vite, Tailwind CSS, Lucide Icons
- **Ebook Engines**: Foliate-js (EPUB), Mozilla PDF.js (PDF)

---

## 📋 Yêu cầu hệ thống

- **Để chạy ứng dụng**: Chỉ cần tải file binary `lyostar` (không cần cài Go, Node.js hay bất cứ phần mềm nào khác).
- **Để phát triển hoặc tự build từ mã nguồn**:
  - [Go](https://go.dev/) 1.22 trở lên
  - [Node.js](https://nodejs.org/) 18 trở lên & npm
  - `make` (tiện ích chạy lệnh tự động)

---

## 🚀 Hướng dẫn cài đặt & Khởi chạy

### Cách 1: Sử dụng Makefile (Khuyên dùng)

Dự án đã tích hợp sẵn [Makefile](file:///home/mvp/Workspace/Other/lyostar/Makefile) để bạn thao tác nhanh chỉ với 1 lệnh:

#### 1. Chạy trực tiếp từ mã nguồn (Development):
```bash
make dev
```
> Server sẽ khởi động ngay tại `http://localhost:8080`. Bạn chỉ cần mở trình duyệt và trải nghiệm!

#### 2. Build Single-Binary ra file thực thi (Production):
```bash
make build
```
Lệnh này sẽ tự động:
1. Cài đặt các gói npm và build mã nguồn Vue 3 thành các file tĩnh.
2. Compile file Go và nhúng toàn bộ frontend vào file binary `lyostar` (đã tối ưu kích thước bằng `-ldflags="-s -w"` ~19MB).

#### 3. Chạy file binary vừa build:
```bash
make run
# hoặc chạy trực tiếp file:
./lyostar
```

---

### Cách 2: Chạy và build bằng các câu lệnh thủ công

Nếu hệ thống của bạn không cài đặt `make`, bạn có thể thực hiện theo các bước sau:

#### Bước 1: Build Frontend
```bash
cd frontend
npm install
npm run build
cd ..
```

#### Bước 2: Compile Binary Go
```bash
go build -ldflags="-s -w" -o lyostar ./cmd/lyostar
```

#### Bước 3: Khởi động Server
```bash
./lyostar
```

Truy cập trên trình duyệt: `http://localhost:8080`.

---

### Cách 3: Triển khai bằng Docker & Docker Compose (Production Ready)

Lyostar cung cấp sẵn Docker image siêu nhẹ (chỉ khoảng `~31.6MB`), chạy với quyền bảo mật non-root user (`lyostar:1000`) và tích hợp sẵn healthcheck tự động:

#### 1. Khởi động với Docker Compose:
```bash
docker compose up -d
# hoặc qua Makefile
make docker-up
```

#### 2. Xem logs hoạt động:
```bash
docker compose logs -f
# hoặc qua Makefile
make docker-logs
```

#### 3. Dừng hệ thống:
```bash
docker compose down
# hoặc qua Makefile
make docker-down
```

> **Lưu ý về thư mục lưu trữ**:
> - Thư mục `./books` được mount ở chế độ **CHỈ ĐỌC (STRICTLY READ-ONLY)** (`:ro`), đảm bảo an toàn tuyệt đối cho kho sách gốc.
> - Thư mục `./data` lưu trữ SQLite database, cache ảnh bìa và sách tải lên, giữ nguyên dữ liệu khi restart container.


---

## ⚙️ Cấu hình biến môi trường

Lyostar cho phép tùy biến cổng và đường dẫn lưu trữ thông qua biến môi trường:

| Biến môi trường | Mặc định | Ý nghĩa |
| :--- | :--- | :--- |
| `PORT` | `8080` | Cổng HTTP mà server lắng nghe |
| `BOOKS_DIR` | `./books` | Thư mục chứa sách của bạn (chế độ **Read-only**) |
| `DATA_DIR` | `./data` | Thư mục lưu trạng thái ứng dụng (CSDL `app.db`, cache bìa sách) |

**Ví dụ chạy với thư mục sách riêng:**
```bash
PORT=9000 BOOKS_DIR=/home/user/my-ebooks DATA_DIR=/home/user/lyostar-data ./lyostar
```

---

## 🧑‍💻 Chế độ dành cho lập trình viên (Developer Guide)

Nếu bạn đang trực tiếp chỉnh sửa code giao diện Vue và muốn tính năng **Hot Module Replacement (HMR)** để trình duyệt tự động cập nhật ngay khi lưu file:

```bash
make dev-live
```
Lệnh trên sẽ tự động bật đồng thời cả Go Backend (port `8080`) và Vite Dev Server (port `3000`) qua cơ chế proxy.

### Các lệnh hữu ích khác trong Makefile:
- `make test`: Chạy toàn bộ unit test với thông tin chi tiết.
- `make test-short`: Chạy nhanh unit test ngắn gọn.
- `make clean`: Xóa file binary đã compile.
- `make sqlc`: Sinh lại mã nguồn Go cho database queries từ file `queries.sql`.
- `make help`: Xem danh sách tất cả các lệnh hỗ trợ.

---

## 📂 Cấu trúc thư mục dự án

```text
lyostar/
├── cmd/lyostar/          # Entrypoint của server Go (main.go)
├── internal/             # Code xử lý logic cốt lõi
│   ├── api/              # HTTP routers, middleware, API handlers
│   ├── auth/             # Session, xác thực & hashing bcrypt
│   ├── config/           # Quản lý cấu hình & biến môi trường
│   ├── database/         # Schema, models & queries sinh bởi sqlc
│   ├── epub/             # Engine giải mã EPUB & trích xuất bìa
│   ├── pdf/              # Engine trích xuất metadata & bìa PDF
│   └── scanner/          # Bộ quét thư mục nền chạy định kỳ
├── frontend/             # Ứng dụng Single Page App (Vue 3)
│   ├── src/              # Components, Views, Composables
│   ├── public/           # Assets tĩnh (PDF.js viewer pre-built)
│   ├── dist/             # Thư mục build được nhúng vào Go
│   └── embed.go          # Go embed.FS nhúng thư mục dist
├── Makefile              # Các lệnh tự động hóa build & dev
└── AGENTS.md             # Tài liệu nguyên tắc kiến trúc hệ thống
```

---

## 📜 Giấy phép

Dự án được phân phối dưới giấy phép mã nguồn mở. Mọi đóng góp và cải tiến đều được hoan nghênh!
