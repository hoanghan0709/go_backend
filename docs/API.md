# API backend ôn thi bằng lái

Base URL mặc định: `http://localhost:8080`. Body JSON, header `Content-Type: application/json`.

## Xác thực và quyền

| Method | Path | Quyền | Chức năng |
|---|---|---|---|
| POST | `/auth/register` | Công khai | `{ "name", "email", "password" }`; mật khẩu 8–72 ký tự theo validator hiện có |
| POST | `/auth/login` | Công khai | `{ "email", "password" }`; trả access/refresh token trong `data` |
| POST | `/auth/refresh` | Công khai | `{ "refresh_token" }`; token cũ bị thu hồi sau khi đổi thành công |
| GET | `/auth/getProfile` | Đăng nhập | ID, tên, email và role của chính mình |
| GET | `/health` | Công khai | Kiểm tra HTTP server |

Access token hiện có thời hạn 12 giờ, refresh token 7 ngày. Các API bên dưới yêu cầu `Authorization: Bearer <access_token>`; nhóm `/admin` yêu cầu role `admin`.

## Quản trị nội dung

| Method | Path (sau `/api/v1`) | Chức năng |
|---|---|---|
| GET | `/categories` | Danh sách chủ đề, mọi tài khoản đăng nhập |
| POST | `/admin/categories` | Tạo chủ đề `{ "name", "description" }` |
| PUT | `/admin/categories/:id` | Thay tên/mô tả |
| DELETE | `/admin/categories/:id` | Xóa mềm chủ đề; từ chối nếu còn câu hỏi |
| GET | `/admin/questions` | Danh sách cả bản nháp và xuất bản, có khóa đáp án |
| GET | `/admin/questions/:id` | Xem chi tiết câu hỏi và đáp án |
| POST | `/admin/questions` | Tạo câu hỏi kèm đáp án |
| PUT | `/admin/questions/:id` | Thay toàn bộ câu hỏi và tập đáp án |
| DELETE | `/admin/questions/:id` | Xóa mềm; từ chối nếu còn thuộc đề chưa xóa |
| POST | `/admin/questions/crawl` | Cào bộ A1 và B từ nguồn cố định; không body/query; xem [hướng dẫn crawler](CRAWL_DU_LIEU.md) |
| POST | `/admin/questions/import` | Import JSON 1–500 câu trong một transaction |
| GET | `/admin/tests` | Danh sách đề, gồm bản nháp |
| GET | `/admin/tests/:id` | Cấu hình và thứ tự question_id của đề |
| POST | `/admin/tests` | Tạo đề cố định |
| PUT | `/admin/tests/:id` | Thay cấu hình và danh sách câu hỏi |
| DELETE | `/admin/tests/:id` | Xóa mềm đề; giữ nguyên phiên thi đã bắt đầu |
| POST | `/admin/tests/generate` | Sinh và lưu một đề theo quota |

Tạo/sửa câu hỏi:

```json
{
  "title": "Câu hỏi minh họa",
  "content": "Chọn phương án đúng trong ví dụ này?",
  "image": "",
  "explanation": "Phương án A được đánh dấu đúng trong dữ liệu minh họa.",
  "license_type": "A1",
  "category_id": 1,
  "is_critical": false,
  "published": true,
  "answers": [
    {"content": "Phương án A", "label": "A", "order": 1, "is_correct": true},
    {"content": "Phương án B", "label": "B", "order": 2, "is_correct": false}
  ]
}
```

- Hạng bằng được code hỗ trợ: `A1`, `A`, `B`, `C1`, `C`, `D`. Đây là danh sách của ứng dụng; không phải tài liệu quy định pháp luật.
- Chủ đề phải tồn tại; tiêu đề/nội dung không được chỉ chứa khoảng trắng.
- Có ít nhất hai đáp án; nhãn, nội dung và thứ tự không trùng; đúng một đáp án đúng. Nhãn chuẩn hóa thành chữ hoa; nội dung được trim.
- Câu hỏi xuất bản cần lời giải. Có thể dùng PUT để xuất bản hoặc đưa lại về bản nháp.
- Muốn đổi đáp án đúng, dùng PUT cả câu hỏi để thay đổi nguyên tử. API sửa riêng đáp án sẽ từ chối thao tác khiến câu hỏi tạm thời có 0 hoặc 2 đáp án đúng.
- Import body: `{ "questions": [<question>, <question>] }`. Một dòng lỗi sẽ rollback cả lô và báo số dòng. Body tối đa 5 MiB. Nhập lại cùng lô sẽ tạo câu mới, không tự chống trùng nội dung giữa các câu.

Tạo/sửa đề cố định:

```json
{
  "name": "Đề minh họa",
  "description": "Không dùng làm cấu hình thi chính thức",
  "license_type": "A1",
  "version": "demo-v1",
  "duration_minutes": 10,
  "passing_score": 1,
  "critical_fail": true,
  "published": true,
  "question_ids": [1, 2]
}
```

- Đề có 1–200 câu, không lặp ID; tất cả đã xuất bản và cùng hạng bằng.
- Thời gian 1–180 phút, điểm đạt từ 1 đến số câu; mỗi câu đúng tính 1 điểm.
- Khi sửa câu hỏi đã nằm trong đề sang hạng bằng khác/bản nháp, cần cập nhật đề tương ứng. Server kiểm tra lại ngân hàng câu hỏi lúc bắt đầu thi và từ chối đề không còn hợp lệ.
- Bài đã bắt đầu không bị ảnh hưởng bởi thay đổi đề/câu hỏi vì có snapshot riêng.

Sinh đề: giữ các trường cấu hình trên, thay `question_ids` bằng:

```json
{
  "quotas": [
    {"category_id": 1, "count": 2, "critical_count": 1},
    {"category_id": 2, "count": 3, "critical_count": 0}
  ]
}
```

Mỗi chủ đề chỉ khai báo một lần; `critical_count` nằm trong `0..count`. Ví dụ trên cần chính xác 1 câu điểm liệt và 1 câu thường ở chủ đề 1; 3 câu thường ở chủ đề 2. Thiếu dữ liệu thì trả lỗi, không sinh đề thiếu hoặc lặp câu. Mỗi lần gọi tạo một đề mới. `question_ids` nếu gửi kèm sẽ được thay bằng kết quả sinh.

## Ôn tập và tiến bộ

| Method | Path (sau `/api/v1`) | Chức năng |
|---|---|---|
| GET | `/practice/questions` | Câu hỏi xuất bản; không có khóa đáp án/lời giải |
| POST | `/practice/questions/:id/answer` | `{ "answer_id": 123 }`; lưu lượt luyện và trả lời giải/đáp án |
| GET | `/practice/wrong-questions` | Câu có lần làm gần nhất sai hoặc bỏ trống |
| GET | `/progress` | Tỷ lệ đỗ, độ chính xác từng chủ đề, 20 lượt thi gần nhất, chủ đề nên ôn |
| GET | `/tests` | Danh sách đề đã xuất bản, không kèm khóa đáp án |

Danh sách câu hỏi hỗ trợ `page`, `limit`, `license_type`, `category_id`, `is_critical=true|false`, `search`. Mặc định trang 1, 20 phần tử; tối đa 100 phần tử/trang. Danh sách đề hỗ trợ `page`, `limit`, `license_type`. Progress hỗ trợ `license_type`.

Response phân trang:

```json
{"data": [], "total": 0, "page": 1, "limit": 20}
```

Luyện tập trả `{ "attempt": {...}, "question": {...} }`, trong đó `question` có lời giải và đáp án đúng. Đây là chế độ học mở; API không phải hệ thống chống gian lận cho kỳ thi có giám sát.

Câu sai được tổng hợp từ bài thi đã hoàn tất và lượt luyện tập. Trả lời đúng ở lượt gần nhất sẽ loại câu khỏi danh sách; câu đã gỡ xuất bản/xóa không được trả về để ôn. Có thể lặp GET câu sai → POST câu trả lời để thực hiện một lượt ôn; chưa tạo một loại phiên ôn riêng. `wrong_question_ids` trong progress phản ánh lịch sử và có thể chứa ID nội dung đã gỡ.

`categories[].accuracy` và `pass_rate` là phần trăm từ 0 đến 100; thứ tự chủ đề tăng theo độ chính xác. Chủ đề yếu nhất được gợi ý nếu độ chính xác chưa đạt 100%. Các lượt làm lại đều tính vào thống kê, không phải tỷ lệ câu đã thuộc. Danh sách `recent_exams` mới nhất trước.

## Phiên thi

| Method | Path (sau `/api/v1`) | Chức năng |
|---|---|---|
| POST | `/exams` | `{ "test_id": 1 }`; tạo phiên mới và snapshot đề |
| GET | `/exams` | Lịch sử của mình; `page`, `limit` |
| GET | `/exams/:id` | Tiếp tục bài đang làm hoặc xem lại bài hoàn tất |
| PUT | `/exams/:id/answers/:question_id` | `{ "answer_id": 123 }`; lưu/đổi lựa chọn; 0 để bỏ chọn |
| POST | `/exams/:id/submit` | Nộp và chấm bài, không cần body |

Response phiên thi có `id`, `test_id`, `status`, `started_at`, `expires_at`, `server_time`, `finished_at`, `score`, `correct_answers`, `total_questions`, `passing_score`, `is_passed`, `failure_reason`, `exam`.

`exam` gồm tên đề, hạng bằng, phiên bản và câu hỏi kèm lựa chọn. Frontend dùng **ID đáp án trong phiên thi**, không lấy lại ID từ ngân hàng câu hỏi đang thay đổi.

Quy tắc:

1. Chỉ chủ phiên đọc/lưu/nộp được. Truy cập phiên của người khác trả 404.
2. Khi `DOING`, không trả `is_correct` hoặc lời giải. `score=0`, `is_passed=false` lúc này là giá trị chưa chấm.
3. `expires_at` do server xác định. Frontend có thể dùng `server_time` để hiệu chỉnh đồng hồ hiển thị; đồng hồ phía client không quyết định hạn nộp.
4. Đáp án phải thuộc đúng câu và câu phải thuộc snapshot. `answer_id=0` xóa lựa chọn; thiếu trường này là lỗi.
5. Nộp sớm chuyển sang `COMPLETED`. Đã đến hạn chuyển sang `TIMEOUT`, chấm trên đáp án đã lưu; timeout vẫn có thể đạt nếu đủ điểm và thỏa quy tắc.
6. Khi bật `critical_fail`, sai **hoặc bỏ trống** câu điểm liệt sẽ trượt. `failure_reason` là `critical_question` hoặc `insufficient_score` nếu không đạt.
7. Nộp lặp trả kết quả đã chấm; không tạo điểm mới. Lưu đáp án vào bài đã nộp/hết giờ trả 409.
8. Nộp bài không tự gửi các lựa chọn chưa lưu. Client cần chờ lưu thành công trước khi gọi submit.
9. GET phiên/lịch sử/progress/ôn câu sai xử lý bài đã hết hạn. Chưa có background worker; không có request thì trạng thái DB có thể còn `DOING`, nhưng không thể lưu quá hạn.
10. POST bắt đầu luôn tạo phiên mới. Khi mở lại app, dùng lịch sử để lấy ID phiên cần tiếp tục thay vì POST lần nữa.

## Mã lỗi

- `400`: request, cấu hình hoặc quan hệ câu hỏi/đáp án không hợp lệ.
- `401`: thiếu/hỏng/hết hạn access token.
- `403`: không đủ quyền admin.
- `404`: không tìm thấy tài nguyên hoặc không sở hữu phiên thi.
- `409`: phiên thi không còn cho lưu đáp án.
- `500`: lỗi nội bộ; API mới không trả chi tiết SQL cho client.

API cũ vẫn giữ dạng response của module cũ và một số mã lỗi cũ; nên dùng `/api/v1` cho phần học/thi và quản trị mới.

Chi tiết phản hồi validation (HTTP 400), đường dẫn trường JSON và lỗi nghiệp vụ: [VALIDATION_ERRORS.md](VALIDATION_ERRORS.md).
