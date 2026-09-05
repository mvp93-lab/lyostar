# Lyostar - Feature Matrix & Roadmap 🗺️

Tài liệu này theo dõi toàn bộ tính năng của Lyostar (so chiếu với hệ thống Calibre-Web và các tiêu chuẩn máy chủ ebook hiện đại). Mỗi khi hoàn thành hoặc bổ sung một tính năng mới, checklist tại file này sẽ được cập nhật tương ứng.

---

## 📊 Tổng quan tiến độ

- **Core MVP / Hạ tầng cốt lõi**: `100% Hoàn thành`
- **Quét thư viện & Quản lý Metadata**: `90% Hoàn thành`
- **Trình đọc sách (Web Readers)**: `90% Hoàn thành`
- **Duyệt & Tổ chức thư viện**: `95% Hoàn thành`
- **Quản lý người dùng & Bảo mật**: `95% Hoàn thành`
- **Kết nối thiết bị ngoài (OPDS / E-Reader)**: `0% Hoàn thành`

---

## 1. 🏗️ Kiến trúc & Hệ thống (Architecture & System)

- [x] **Single-Binary Deployment**: Nhúng toàn bộ Vue 3 SPA vào bên trong 1 file thực thi Go duy nhất (`embed.FS`).
- [x] **Tối ưu tài nguyên cực hạn**: Standalone binary `< 30MB`, mức tiêu thụ RAM nhàn rỗi `< 30MB`.
- [x] **Cơ sở dữ liệu SQLite tối ưu**: CGO-free (`modernc.org/sqlite`), kích hoạt chế độ `WAL`, `foreign_keys=ON`, `busy_timeout=5000`.
- [x] **Type-safe Query Layer**: Sử dụng `sqlc` để biên dịch các câu truy vấn SQL an toàn, không sử dụng ORM nặng nề.
- [x] **Single-process concurrency**: Xử lý tác vụ nền bằng Go channel có bộ đệm và fixed worker pool, không phụ thuộc broker ngoài (Redis, RabbitMQ).
- [x] **Chế độ an toàn thư mục sách**: Thư mục `/books` luôn ở chế độ **STRICTLY READ-ONLY** (không bao giờ sửa, xóa, đổi tên file gốc).
- [ ] **Docker Image tối ưu**: Dockerfile multi-stage build cho dung lượng image cuối cùng `< 50MB`.

---

## 2. 📖 Trình đọc sách trực tuyến (In-Browser Readers)

- [x] **Trình đọc EPUB (Foliate.js)**:
  - [x] Tích hợp engine Foliate.js (bọc an toàn trong `shallowRef` để tránh rò rỉ bộ nhớ).
  - [x] Tùy chỉnh kích thước chữ, font chữ, giãn dòng và theme (Dark, Light, Sepia).
  - [x] Chuyển chương mượt mà qua Mục lục (Table of Contents).
  - [x] Tối ưu hiển thị cho màn hình OLED và trình duyệt E-Ink có tần số quét chậm.
- [x] **Trình đọc PDF (Mozilla PDF.js)**:
  - [x] Tích hợp phiên bản PDF.js official pre-built viewer chuẩn vector.
  - [x] Lớp TextLayer hiển thị sắc nét, hỗ trợ bôi đen và sao chép văn bản.
  - [x] Chế độ cuộn dọc liên tục (Continuous vertical scrolling).
  - [x] Phục vụ file với header `Accept-Ranges: bytes` hỗ trợ tải và streaming nhanh.
- [x] **Lưu & Khôi phục tiến độ đọc (Reading Progress & Resume)**:
  - [x] Lưu tiến độ đọc riêng cho từng user trong bảng `reading_progress`.
  - [x] Lưu chính xác vị trí đọc (Foliate CFI đối với EPUB, số trang đối với PDF).
  - [x] Lưu tỉ lệ đọc xong (`0.0` đến `1.0`) và trạng thái hoàn thành.
  - [x] Tự động cuộn/mở đúng trang đang đọc dang dở khi người dùng quay lại.
- [x] **Đánh dấu trang (Bookmarks) & Ghi chú (Highlights/Notes)**:
  - [x] Cho phép user lưu các điểm đánh dấu trang yêu thích (1-click bookmark toggle trên thanh công cụ Reader).
  - [x] Tô màu đoạn văn bản và ghi chú trích dẫn cá nhân (4 mã màu: Vàng, Xanh lá, Xanh dương, Hồng) kèm xem / sửa / xóa.
  - [x] Ngăn kéo tiện ích bên cạnh (Side Drawer) trong trình đọc với 3 tab: Bookmarks, Notes, và Mục lục (TOC) hỗ trợ nhảy tới vị trí đọc tức thì (1-click jump).
- [ ] **Trình đọc Comic / Manga (CBZ / CBR)**:
  - [ ] Parser đọc danh sách ảnh từ file nén `.cbz` (zip) và `.cbr` (rar).
  - [ ] Web viewer đọc truyện tranh (chế độ cuộn dọc Webtoon hoặc lật trang đôi).

---

## 3. 🔍 Duyệt, Tìm kiếm & Tổ chức sách (Browsing & Organization)

- [x] **Kệ "Tất cả sách"**: Hiển thị dạng lưới thẻ bìa sách trực quan kèm phân trang / tải lướt.
- [x] **Kệ "Tiếp tục đọc" (Continue Reading)**: Hiển thị các sách user đang đọc dở kèm thanh tiến độ %.
- [x] **Tìm kiếm toàn văn (FTS5)**: Tìm kiếm siêu tốc theo Tiêu đề, Tác giả, Series thông qua SQLite FTS5.
- [x] **Sidebar Navigation Drawer (Kiểu Calibre-Web)**:
  - [x] Chuyển đổi thanh điều hướng từ Topbar quá tải sang Sidebar Drawer hiện đại, khoa học.
  - [x] Nhóm BROWSE: All Books, Continue Reading, Categories / Tags, Series, Authors.
  - [x] Nhóm SHELVES: Danh sách kệ cá nhân/công khai của user kèm số lượng sách, nút `+ Create a Shelf` và quản lý kệ.
  - [x] Nhóm MANAGEMENT: Upload sách (`can_upload`), Quản lý Users (Admin), Quét lại thư viện (Admin), Đăng xuất.
  - [x] Thiết kế Responsive: Cố định trang nhã bên trái trên Desktop (`md:w-64`), tự động chuyển thành Drawer trượt có backdrop-blur mờ trên Mobile.
- [x] **Lọc & Sắp xếp cơ bản**: Sắp xếp theo Tên sách, Tác giả, Thời gian thêm mới.
- [x] **Xem chi tiết sách (Book Detail Modal)**: Hiển thị bìa, mô tả, nhà xuất bản, ngày xuất bản, định dạng file, kích thước.
- [x] **Tải sách về máy (Direct Download)**: Nút tải file EPUB/PDF gốc về máy cá nhân.
- [x] **Kệ sách tùy chỉnh của User (Custom Shelves / Collections)**:
  - [x] Cho phép người dùng tự tạo, chỉnh sửa, xóa kệ sách cá nhân (ví dụ: *"Sách yêu thích"*, *"Muốn đọc"*, *"Học kỹ thuật"*).
  - [x] Thêm / bớt sách vào kệ sách nhanh chóng qua hộp thoại 1-click checkbox modal (`ShelfSelectModal`) tương tự YouTube playlist & Calibre-Web.
  - [x] Tùy chọn kệ sách công khai (Public) hoặc riêng tư (Private).
  - [x] Modal quản lý kệ sách (`ShelvesManageModal`) xem danh sách, số lượng sách (`book_count`), huy hiệu sở hữu (`is_owner`).
  - [x] Lọc thư viện theo kệ sách đang chọn trên giao diện chính, nút huy hiệu xóa bộ lọc nhanh để quay lại toàn bộ thư viện.
- [x] **Phân loại theo Thể loại (Tags / Categories / Genres)**:
  - [x] Trích xuất trường `<dc:subject>` từ EPUB và `/Keywords` / XMP `Subject` từ PDF.
  - [x] Lưu trữ cấu trúc quan hệ `tags` & `book_tags` trong SQLite, truy vấn không trùng lặp qua subquery.
  - [x] Endpoint `GET /api/tags` trả về danh sách thể loại kèm `book_count`.
  - [x] Lọc thư viện theo thể loại qua `GET /api/books?tag=...`.
  - [x] Thanh duyệt thể loại trực quan (Horizontal Genre Bar) trên Web UI kèm số lượng sách.
  - [x] Hiển thị tag pills trên thẻ sách và modal chi tiết, bấm vào để lọc nhanh thư viện theo tag.
  - [x] Chỉnh sửa danh sách Tags / Categories trực tiếp từ Metadata Editor (`can_edit`).
- [ ] **Đánh giá sao (Rating System)**:
  - [ ] Chấm điểm sách từ 1 đến 5 sao theo từng tài khoản.
  - [ ] Lọc sách theo số sao đánh giá.
- [ ] **Bộ lọc mở rộng**: Lọc theo Ngôn ngữ (Language), Nhà xuất bản (Publisher).

---

## 4. ⚡ Quét thư viện & Quản lý Metadata (Library & Metadata)

- [x] **Bộ quét thư mục tự động (Background Scanner)**:
  - [x] Quét thư mục `/books` đệ quy tìm các file `.epub`, `.pdf`.
  - [x] Tính hash SHA-256 để chống trùng lặp và phát hiện file đổi tên/di chuyển.
  - [x] Worker pool cố định không làm nghẽn CPU/RAM hệ thống.
- [x] **Trích xuất Metadata & Ảnh bìa**:
  - [x] EPUB: Parse `META-INF/container.xml` và `.opf` bằng `archive/zip` & `encoding/xml` (không giải nén toàn bộ ra đĩa).
  - [x] PDF: Parse `/Info` dictionary và trích xuất stream ảnh bìa JPEG đầu tiên thuần Go.
- [x] **Cover Thumbnail Pipeline**:
  - [x] Resize bìa sách và chuyển mã sang chuẩn **WebP** (chiều rộng tối đa 400px).
  - [x] Lưu cache bìa sách tại `/data/cache/covers/{file_sha256}.webp`.
- [x] **Chỉnh sửa Metadata trên Web (Metadata Editor)**:
  - [x] Cho phép người dùng có quyền `can_edit` (hoặc Admin) sửa tiêu đề, tác giả, mô tả, series, tập số (#), NXB, ngày phát hành, ngôn ngữ trực tiếp từ giao diện web.
  - [x] Dữ liệu sửa được lưu trong SQLite và tự động đồng bộ tìm kiếm toàn văn FTS5 (không ghi đè file sách gốc).
- [x] **Xóa sách (Delete Book)**:
  - [x] Cho phép người dùng có quyền `can_delete` (hoặc Admin) xóa sách khỏi thư viện với hộp thoại xác nhận an toàn.
  - [x] Tự động cascade dọn dẹp tiến độ đọc, quan hệ tác giả, thumbnail WebP và chỉ mục FTS5.
  - [x] Tuân thủ nghiêm ngặt quy tắc lưu trữ: File thuộc `/data/uploads/` được xóa khỏi đĩa; file thuộc `/books/` luôn được bảo toàn an toàn (STRICTLY READ-ONLY).
- [x] **Upload sách mới qua Web UI**:
  - [x] Cho phép tải file `.epub`, `.pdf` lên trực tiếp từ trình duyệt qua kéo thả (Drag & Drop) hoặc duyệt file.
  - [x] Lưu trữ an toàn trong thư mục `/data/uploads/` (đảm bảo thư mục `/books` luôn STRICTLY READ-ONLY).
  - [x] Kiểm soát phân quyền `can_upload` (Calibre-Web style), ngăn chặn tải trùng lặp qua SHA-256 (409 Conflict).
  - [x] Tức thì trích xuất metadata và tạo WebP thumbnail (`IndexFile`), tự động cập nhật kệ sách trên UI.
- [ ] **Tự động tải Metadata & Ảnh bìa từ Internet**:
  - [ ] Tìm kiếm thông tin sách qua ISBN / Tiêu đề từ Google Books, Open Library, Douban.

---

## 5. 👥 Quản lý người dùng & Bảo mật (Auth & RBAC)

- [x] **Xác thực & Mã hóa bảo mật**:
  - [x] Mã hóa mật khẩu một chiều bằng thuật toán `bcrypt`.
  - [x] Quản lý phiên đăng nhập (sessions) lưu trong SQLite với token CSPRNG 64-character hex.
  - [x] Vận chuyển token qua cookie `HttpOnly`, `SameSite=Lax`.
- [x] **Phân quyền người dùng (Role-Based Access Control)**:
  - [x] Role `admin`: Quản trị toàn quyền hệ thống, quản lý tài khoản người dùng, kích hoạt quét sách.
  - [x] Role `reader`: Xem sách, đọc sách, tải sách và lưu tiến độ cá nhân.
- [x] **Trình hướng dẫn thiết lập lần đầu (First-run Setup Wizard)**:
  - [x] Tự động phát hiện khi hệ thống chưa có user và dẫn hướng tạo tài khoản Admin ban đầu.
- [x] **Quản lý tài khoản (Admin User Management)**:
  - [x] Thêm mới tài khoản, đổi mật khẩu, đổi role, xóa tài khoản.
- [x] **Hệ thống phân quyền chi tiết (Granular User Permissions - Calibre-Web model)**:
  - [x] Các cờ quyền độc lập: Đọc online (`can_read`), Tải sách (`can_download`), Tải lên (`can_upload`), Chỉnh sửa (`can_edit`), Xóa sách (`can_delete`).
  - [x] Admin có thể tùy chỉnh quyền cho bất kỳ người dùng nào hoặc chính tài khoản của mình qua `PUT /api/users/{id}`.
  - [x] Bảo vệ endpoint backend: `/api/books/{id}/file` (`can_read`), `/api/books/{id}/download` (`can_download`), `/api/books/upload` (`can_upload`), `/api/books/{id}` PUT (`can_edit`), `/api/books/{id}` DELETE (`can_delete`).
  - [x] Giao diện người dùng: Huy hiệu quyền hạn trực quan, modal chỉnh sửa tài khoản và ẩn/hiện nút Đọc / Tải / Sửa / Xóa linh hoạt theo quyền tài khoản.
- [ ] **Giới hạn nội dung theo User (Content Restrictions)**:
  - [ ] Giới hạn sách hiển thị cho từng user theo Tag/Thể loại hoặc độ tuổi.
- [ ] **Đăng ký mở (Public Registration Toggle)**:
  - [ ] Cho phép bật/tắt tính năng tự tạo tài khoản mới từ trang đăng nhập.

---

## 6. 📱 Kết nối thiết bị ngoài & Tích hợp (E-Reader Sync & Integrations)

- [ ] **OPDS Feed Catalog (Chuẩn kết nối Open Publication Distribution System)**:
  - [ ] Cung cấp feed XML/Atom chuẩn OPDS 1.2 tại endpoint `/opds`.
  - [ ] Hỗ trợ phân trang, danh mục sách mới nhất, tác giả, thể loại, tìm kiếm OpenSearch.
  - [ ] Tương thích với các ứng dụng đọc sách: **KOReader** (trên máy Kobo, Kindle jailbreak), **Moon+ Reader**, **FBReader**, **Chunky**.
  - [ ] Xác thực HTTP Basic Auth cho OPDS client.
- [ ] **Gửi sách qua Kindle (Send-to-Kindle via SMTP)**:
  - [ ] Cấu hình máy chủ gửi mail SMTP (Gmail, Brevo, SendGrid, custom SMTP).
  - [ ] Lưu địa chỉ email nhận sách Kindle (`@kindle.com`) vào hồ sơ cá nhân của từng user.
  - [ ] Nút "Gửi đến Kindle" gửi trực tiếp file sách qua email chỉ với 1 click.
- [ ] **Kobo Sync API**:
  - [ ] Cung cấp API giả lập Kobo Store để đồng bộ trực tiếp sách và tiến độ đọc qua Wi-Fi với thiết bị máy đọc sách Kobo nguyên bản.
