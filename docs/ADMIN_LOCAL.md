# Trang quản trị local

Giao diện tiếng Việt tại **http://localhost:8080/admin/**, phục vụ trực tiếp bởi server Go. HTML/CSS/JavaScript được nhúng trong binary bằng `go:embed`; không cần npm, frontend server hoặc cấu hình CORS.

## 1. Chạy server

Từ thư mục dự án:

```sh
# Chỉ sao chép nếu bạn chưa có .env, không ghi đè cấu hình hiện tại.
# cp .env.example .env
# Đặt JWT_SECRET trong .env trước khi chạy.
LISTEN_ADDR=127.0.0.1:8080 go run ./cmd/server
```

Nếu server đã chạy từ trước khi thêm giao diện, hãy khởi động lại để nạp asset mới. Đổi port nếu cần và mở `/admin/` trên port tương ứng. `DATABASE_PATH` mặc định là `cmd/data/gorm.db`; ứng dụng sửa dữ liệu thông qua cùng API với backend.

## 2. Chuẩn bị tài khoản admin

Nếu đã có tài khoản admin, đăng nhập trực tiếp. Nếu có tài khoản learner, chủ server chỉ cần chạy CLI ở bước cuối dưới đây.

Để đăng ký tài khoản mới, gọi `POST /auth/register` bằng Postman với body:

```json
{
  "name": "Quản trị viên",
  "email": "admin@example.com",
  "password": "THAY_BANG_MAT_KHAU_RIENG"
}
```

Thay email/mật khẩu bằng thông tin của bạn. Sau đó cấp quyền cho tài khoản đã tồn tại:

```sh
go run ./cmd/admin -email admin@example.com -role admin
```

CLI phải sử dụng cùng cấu hình `DATABASE_PATH` với server. Không cần mở database để sửa role hoặc tạo dữ liệu thủ công. Giao diện không tự tạo tài khoản hay mật khẩu mặc định.

## 3. Các màn hình

| Màn hình | Thao tác |
|---|---|
| Tổng quan | Số câu hỏi, chủ đề, đề thi; mở nhanh đề mới tạo |
| Chủ đề | Xem, thêm, sửa, xóa; báo lỗi nếu chủ đề còn câu hỏi |
| Ngân hàng câu hỏi | Tìm kiếm, lọc hạng/chủ đề/điểm liệt, phân trang; thêm/sửa/xóa câu, đáp án, lời giải, URL ảnh, nháp/xuất bản |
| Đề thi | Lọc hạng, phân trang; tạo/sửa/xóa; chọn câu đã xuất bản, sắp xếp thứ tự; thiết lập điểm đạt/thời gian/điểm liệt; sinh đề theo quota |
| Nhập dữ liệu | Tải mẫu JSON, chọn file hoặc dán JSON để import; chạy crawler A1/B, xem kết quả |
| Danh mục API | Danh sách endpoint hiện có, method, quyền và chú thích |

Quy trình: **tạo chủ đề → tạo câu và đáp án → bổ sung lời giải → xuất bản câu → tạo đề → xuất bản đề**.

- Câu hỏi cần ít nhất 2 phương án và đúng 1 đáp án đúng. Nhãn và nội dung không trùng nhau.
- Câu xuất bản cần lời giải. Lưu câu hỏi thay toàn bộ tập đáp án; ID đáp án có thể đổi.
- Đề chỉ nhận câu đã xuất bản, cùng hạng; danh sách chọn lọc câu đã xuất bản ở backend trước khi phân trang, giữ các câu đã chọn giữa các trang. Đổi hạng đề sẽ xóa lựa chọn hiện tại để chọn lại đúng hạng.
- Mũi tên lên/xuống cạnh câu đã chọn thay đổi thứ tự câu trong đề.
- Sinh đề yêu cầu đủ câu đã xuất bản theo từng chủ đề và số câu điểm liệt; backend trả lý do nếu thiếu.
- Import tối đa 500 câu/5 MiB. Chọn một trong hai nguồn file/nội dung dán. Một dòng lỗi sẽ hoàn tác toàn lô. Import lại có thể tạo dữ liệu trùng.
- Crawler gọi nguồn thật khi bạn bấm xác nhận; câu mới được lưu nháp để duyệt. Chưa có upload ảnh hoặc tự xác nhận nội dung pháp luật.

## Xuất bản nhiều câu hỏi

1. Vào **Ngân hàng câu hỏi**, lọc/tìm những câu cần duyệt.
2. Tích checkbox từng câu nháp hoặc checkbox đầu bảng để chọn tất cả câu nháp trên trang hiện tại.
3. Có thể chuyển trang để chọn thêm (tối đa 500 câu). Số đã chọn tính trên tất cả các trang; đổi bộ lọc, chuyển màn hình hoặc đăng xuất sẽ bỏ lựa chọn.
4. Bấm **Xuất bản các câu đã chọn**, kiểm tra số lượng và xác nhận.
5. Xem kết quả: câu hợp lệ chuyển sang **Đã xuất bản**; câu thiếu lời giải, đáp án không hợp lệ hoặc không còn tồn tại được báo lỗi theo ID. Lựa chọn của câu lỗi được giữ lại để sửa và thử lại.

Câu đã xuất bản không có checkbox khả dụng. **Bỏ chọn** xóa toàn bộ lựa chọn. API kiểm tra nội dung hiện tại trong DB và chỉ cập nhật trạng thái xuất bản, không thay nội dung hoặc tạo lại đáp án. Đây là thao tác có thể thành công một phần; xem `published_ids` và `failed` trong response. Nếu gặp lỗi hạ tầng DB, transaction hoàn tác toàn bộ.

## 4. Phiên đăng nhập và quyền

Token được giữ trong `sessionStorage` của tab, tự refresh khi access token hết hạn. Đóng tab hoặc bấm Đăng xuất sẽ kết thúc lưu phiên trên trình duyệt; đăng xuất chưa thu hồi token ở backend vì chưa có API này. Không lưu mật khẩu.

Trang và asset công khai để hiển thị form đăng nhập; mọi dữ liệu vẫn cần token admin ở API. Learner không vào được khu vực quản trị. Thu hồi role bằng CLI có hiệu lực ở các request admin tiếp theo.

Thao tác xóa, import và crawl có xác nhận trước khi gửi. Các thông báo lỗi nghiệp vụ từ backend được hiển thị nguyên văn, bao gồm trường sai và dòng import lỗi. Nội dung dữ liệu được escape khi hiển thị trong HTML.

## 5. Phát triển và kiểm tra

Nguồn giao diện: `internal/adminui/dist/`. Router asset: `internal/adminui/adminui.go`. Sửa asset xong cần chạy lại Go server vì asset được nhúng khi build.

```sh
node --check internal/adminui/dist/app.js
GOCACHE=/tmp/driving-go-build-cache go test ./...
GOCACHE=/tmp/driving-go-build-cache go vet ./...
```

Xem [danh sách chức năng và REST API kèm chú thích](CHUC_NANG_VA_REST_API.md) và [API chi tiết](API.md).

Luồng tài khoản learner và cách chia dữ liệu: [LUONG_LEARNER.md](LUONG_LEARNER.md).


## Quản trị dữ liệu mobile (28/10/2026)

- **Chủ đề:** chọn `type`; dữ liệu cũ mặc định UNCATEGORIZED. Chọn SIGN để chứa học liệu biển báo.
- **Học biển báo:** CRUD theo hạng/nhóm biển, nội dung/ảnh/ý nghĩa và xuất bản.
- **Ngân hàng câu hỏi:** chọn biển liên kết thuộc cùng hạng và chủ đề; lọc theo category_type.
- **Cấu hình đề random:** tạo/sửa/xóa cấu hình với hạng, loại, thời gian, điểm đạt, quota; xuất bản để mobile dùng.
- **Đề thi:** thêm loại đề, kiểm tra mọi câu phù hợp loại khi lưu/bắt đầu.

Cấu hình random không phải một đề với danh sách câu cố định. Mỗi lần learner gọi API random, server chọn lại câu và tạo snapshot riêng. Xem [request/response và quy tắc](MOBILE_API.md).
