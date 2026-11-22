# Luồng người học sau đăng nhập

## Phạm vi hiện tại

Backend đã có API learner cho ôn tập, thi thử, lịch sử và tiến bộ. Web local `/admin/` là giao diện quản trị; không phải trang learner và sẽ từ chối tài khoản không có quyền admin. Chưa có giao diện learner tự chuyển trang sau đăng nhập.

## Nội dung chung và dữ liệu cá nhân

| Dữ liệu | Phạm vi hiện tại |
|---|---|
| Chủ đề | Có type để làm tab; learner nhận chủ đề có học liệu đã xuất bản phù hợp hạng |
| Câu hỏi ôn tập | Câu đã xuất bản, lọc theo hạng/chủ đề/điểm liệt/từ khóa; danh sách không có đáp án đúng/lời giải |
| Đề thi | Đề đã xuất bản, lọc theo hạng; tất cả learner đều có thể xem, không có gán đề riêng |
| Phiên thi và lựa chọn | Tạo riêng theo user_id lấy từ JWT; giữ snapshot nội dung khi bắt đầu |
| Lượt luyện, câu sai, tiến bộ | Riêng từng tài khoản; không chia sẻ lịch sử với người khác |

Không sao chép ngân hàng câu hỏi vào từng user khi đăng nhập. Xuất bản câu hỏi đưa câu vào ôn tập; để xuất hiện trong đề thi cần thêm câu vào đề và xuất bản đề. Đề đã xuất bản vẫn được kiểm tra lại câu hỏi khi bắt đầu thi; câu bị thu hồi/xóa/sai hạng có thể làm đề không thể bắt đầu cho tới khi admin sửa lại.

Đã có `PUT /api/v1/me/license` lưu hạng đang học và `POST /api/v1/exams/random` tạo phiên riêng từ template admin. `license_type` explicit ưu tiên hơn hạng lưu; chưa lưu hạng thì danh sách vẫn đọc mọi hạng để tương thích, còn random yêu cầu hạng. Chưa có lớp học, quyền theo gói hoặc bảng gán user–đề. Xem [API mobile đầy đủ](MOBILE_API.md).

## Luồng API

Base URL: `http://localhost:8080`. Sau đăng nhập, gửi `Authorization: Bearer <access_token>` cho các API cần xác thực.

| Bước / màn hình | Request | Ý nghĩa |
|---|---|---|
| Đăng nhập | `POST /auth/login` với `{email,password}` | Lấy `data.access_token`, `data.refresh_token`, `data.role`; frontend dựa vào role để mở giao diện tương ứng |
| Hồ sơ | `GET /auth/getProfile` | Xác định tài khoản và role hiện tại |
| Chọn hạng A1 | `PUT /api/v1/me/license` với `{license_type:"A1"}` | Lưu vào hồ sơ; các danh sách learner mặc định dùng hạng này |
| Trang chủ / chủ đề | `GET /api/v1/categories` | Nhóm nội dung để ôn |
| Trang chủ / đề | `GET /api/v1/tests?license_type=A1&page=1&limit=20` | Đề A1 đã xuất bản |
| Trang chủ / tiến bộ | `GET /api/v1/progress?license_type=A1` | Tiến bộ của chính user; tài khoản mới chưa có lượt học sẽ có thống kê 0/danh sách rỗng |
| Ôn tập | `GET /api/v1/practice/questions?license_type=A1&category_id=8&page=1&limit=20` | Lấy câu đã xuất bản thuộc chủ đề; thay category_id bằng ID đang có |
| Trả lời câu ôn | `POST /api/v1/practice/questions/:question_id/answer` với `{answer_id}` | Lưu lượt làm của user; trả đúng/sai và lời giải |
| Chọn đề và bắt đầu | `POST /api/v1/exams` với `{test_id}` | Tạo phiên riêng; lấy `id` làm exam_id, đọc câu/phương án trong `exam.questions` |
| Lưu lựa chọn | `PUT /api/v1/exams/:exam_id/answers/:question_id` với `{answer_id}` | Lưu theo phiên của user; gửi 0 để bỏ chọn |
| Tiếp tục bài | `GET /api/v1/exams/:exam_id` | Khôi phục bài và lựa chọn đã lưu; thời hạn vẫn tính từ lúc bắt đầu, không được tạm dừng |
| Nộp bài | `POST /api/v1/exams/:exam_id/submit` | Server chấm và trả kết quả/lời giải; client không tự gửi điểm |
| Lịch sử | `GET /api/v1/exams?page=1&limit=20` | Danh sách phiên riêng của user, dùng ID để xem/tiếp tục |
| Ôn câu sai | `GET /api/v1/practice/wrong-questions?license_type=A1` | Câu sai/bỏ trống theo lần làm gần nhất của user |
| Thi lại | `POST /api/v1/exams` với cùng `{test_id}` | Tạo exam_id mới, giữ nguyên lịch sử lần trước |
| Token hết hạn | `POST /auth/refresh` với `{refresh_token}` | Nhận cặp token mới, refresh token cũ không dùng lại |

User A và B có thể cùng thấy đề `test_id=2`, nhưng mỗi lần bắt đầu tạo một exam_id khác, gắn với user_id tương ứng. A không được xem/lưu/nộp phiên của B. Admin cũng không có API quản trị lịch sử của mọi learner ở thời điểm này; các API lịch sử/tiến bộ luôn lấy người hiện tại từ token.

## API mới cho random và biển báo

Xem [MOBILE_API.md](MOBILE_API.md): cấu hình đề random, `POST /api/v1/exams/random`, CRUD học liệu biển báo, câu liên kết và thi thử SIGN.

## Giao diện learner cần xây tiếp

Đăng nhập → chọn hạng → trang chủ (ôn tập, thi thử, tiến bộ) → ôn/làm bài → kết quả → xem lại/ôn câu sai/thi lại. Đây là giao diện đề xuất để nối với API đã có, chưa được triển khai trong `/admin/`.

Nếu muốn admin phân bổ đề cho từng người/lớp, cần bổ sung model và API giao bài, đăng ký người học vào lớp/hạng, cùng kiểm tra quyền ở API danh sách và bắt đầu thi. Không chỉ lọc ở frontend.
