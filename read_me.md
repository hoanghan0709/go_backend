# Backend ôn thi bằng lái xe

Go + Gin + GORM + SQLite. Backend có quản trị ngân hàng câu hỏi, đề cố định/đề ngẫu nhiên, thi thử, ôn tập, lịch sử và thống kê tiến bộ.

- [Hướng dẫn API từng bước cho người mới, có ví dụ xuyên suốt](docs/HUONG_DAN_API_TUNG_BUOC.md)
- [API cào dữ liệu không tham số: nguồn, cách gọi và chống trùng](docs/CRAWL_DU_LIEU.md)
- [Theo dõi công việc và quyết định thiết kế](work_follow.md)
- [Danh sách API, request mẫu và quy tắc nghiệp vụ](docs/API.md)
- [Luồng HTTP mẫu để thử bằng REST Client](docs/example.http)

## Chạy dự án

Yêu cầu Go theo `go.mod` (hiện tại `1.26.3`) và trình biên dịch C cho SQLite driver.

```sh
cp .env.example .env
# Sửa JWT_SECRET trong .env thành chuỗi bí mật của môi trường của bạn.
go run ./cmd/server
```

Biến môi trường:

| Biến | Mặc định | Ý nghĩa |
|---|---|---|
| `JWT_SECRET` | Không có, bắt buộc | Khóa ký access token |
| `DATABASE_PATH` | `cmd/data/gorm.db` | Đường dẫn SQLite; thư mục cha phải tồn tại |
| `LISTEN_ADDR` | `:8080` | Địa chỉ HTTP |

`GET /health` kiểm tra HTTP server. Các API học/thi cần header `Authorization: Bearer <access_token>`.

## Tạo quản trị viên

1. Đăng ký tài khoản qua `POST /auth/register`.
2. Chủ server cấp quyền bằng CLI cục bộ:

```sh
go run ./cmd/admin -email admin@example.com -role admin
```

Thu hồi quyền bằng cùng lệnh với `-role learner`. CLI chỉ đổi quyền tài khoản đã tồn tại; không tạo tài khoản hay mật khẩu. Đăng ký luôn tạo `learner`, không chấp nhận quyền do client gửi. Quyền admin được kiểm tra từ DB trên mỗi request, nên thu hồi có hiệu lực ngay với token cũ.

## Chuẩn bị nội dung và dùng thử

1. Admin tạo chủ đề tại `/api/v1/admin/categories`.
2. Tạo câu hỏi với tối thiểu hai lựa chọn, đúng một đáp án đúng. Câu hỏi có `published: false` là bản nháp.
3. Thêm lời giải và đặt `published: true` để người học sử dụng được.
4. Tạo đề với danh sách `question_ids`, hoặc sinh đề bằng `quotas` theo chủ đề.
5. Người học lấy danh sách `/api/v1/tests`, bắt đầu phiên, lưu lựa chọn rồi nộp bài.
6. Xem lại bài, ôn câu sai và theo dõi tiến bộ.

Các con số trong ví dụ là **dữ liệu minh họa**, không phải xác nhận cấu trúc đề thi chính thức. Admin cần nhập thời gian, điểm đạt, hạng bằng, phiên bản và quy tắc điểm liệt theo bộ tài liệu đã được kiểm chứng.

## Nâng cấp từ phiên bản cũ

- Sao lưu SQLite trước khi chạy phiên bản mới. Server vẫn dùng `AutoMigrate`; chưa có hệ thống rollback migration theo phiên bản.
- Migration thêm quyền tài khoản, trạng thái xuất bản/lời giải, bảng liên kết đề, snapshot phiên thi và lịch sử luyện tập.
- Tài khoản cũ mặc định là `learner`; câu hỏi cũ mặc định là bản nháp. Không tự nâng quyền hoặc tự xuất bản dữ liệu cũ.
- API `/question/*`, `/questions/:question_id/answers*` và `/answer/list_answer` giữ lại cho **admin**. Frontend người học cần chuyển sang `/api/v1/practice/*` và `/api/v1/exams/*`.
- Response model câu hỏi/đáp án/đề dùng khóa JSON `snake_case` và có `id`. Frontend cũ dùng `ID`, `QuestionID`, `IsCorrect` cần cập nhật mapping.
- `PUT /api/v1/admin/questions/:id` thay cả tập đáp án; ID đáp án có thể đổi. Frontend quản trị phải tải lại response. Phiên thi đang làm vẫn dùng snapshot cũ.
- Dữ liệu kiểm tra ngày 23/10/2026 có 3 câu hỏi nhưng bảng chủ đề rỗng. Cần tạo/gán chủ đề, rà soát đáp án và thêm lời giải trước khi xuất bản. Không tự tạo dữ liệu pháp luật hoặc tự sửa nội dung chưa xác minh.

## Kiểm thử

```sh
go test ./...
go test -race ./internal/server
go vet ./...
```

Nếu môi trường giới hạn ghi cache mặc định:

```sh
GOCACHE=/tmp/driving-go-build-cache go test ./...
```

Integration test tạo database bằng `t.TempDir()`, không ghi vào database dự án. Bộ test kiểm tra phân quyền, đăng ký/refresh, snapshot, thời hạn, điểm liệt, nộp lại, nộp đồng thời, dữ liệu import, sinh đề và tiến bộ học tập.

## Cấu trúc và lý do thiết kế

- `internal/server`: khởi tạo router chung cho server và integration test.
- `internal/learning`: nghiệp vụ chủ đề, ngân hàng câu hỏi, đề thi, luyện tập và tiến bộ; `delivery` nhận/trả HTTP.
- `internal/exam`: delivery → usecase → repository cho vòng đời phiên thi; DTO định nghĩa dữ liệu API.
- `internal/question/validation`: quy tắc đáp án dùng chung cho tạo/sửa/import và API đáp án cũ.
- `internal/middleware/admin.go`: phân quyền quản trị.
- `cmd/admin`: công cụ cấp/thu hồi quyền dành cho chủ server.

Comment trong code tập trung vào **vì sao** dùng snapshot, transaction, kiểm tra quyền, giới hạn thời gian và upsert; tránh diễn giải lại những dòng code hiển nhiên.

SQLite dùng một connection để tuần tự hóa transaction ghi, cùng `foreign_keys=ON` và `busy_timeout`. Cách này phù hợp backend quy mô nhỏ; khi cần nhiều server hoặc tải ghi lớn, cần chuyển DB và thiết kế khóa hàng/optimistic locking trước khi tăng connection pool.

Bài hết giờ được chấm khi truy cập phiên thi/lịch sử/thống kê; không cần client giữ kết nối. Chưa có worker tự cập nhật hàng loạt khi không có request. Thống kê hiện tổng hợp từ lịch sử đầy đủ của người dùng; với dữ liệu lớn nên bổ sung bảng tổng hợp. Repository này chưa có giao diện người học/quản trị, gửi email hoặc lưu trữ file ảnh. Crawler có thể nhập bộ câu hỏi từ nguồn bên thứ ba vào bản nháp để rà soát; không tự xác nhận đây là dữ liệu chính thức mới nhất. Trường `image` nhận đường dẫn từ admin hoặc nguồn crawler.
