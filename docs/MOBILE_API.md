# API mobile: hạng bằng, đề ngẫu nhiên, chủ đề và học biển báo

Triển khai ngày 28/10/2026. Base URL mặc định `http://localhost:8080`.

Mọi API dưới đây cần `Authorization: Bearer <access_token>`. Request có body dùng `Content-Type: application/json`. Nhóm `/api/v1/admin/*` yêu cầu role admin. ID trong ví dụ là minh họa: lấy ID thật từ response tạo dữ liệu.

## 1. Lưu hạng bằng của tài khoản đang đăng nhập

```http
PUT /api/v1/me/license

{"license_type":"A1"}
```

Response 200:

```json
{
  "id": 7,
  "name": "Người học",
  "email": "learner@example.com",
  "role": "learner",
  "current_license_type": "A1"
}
```

- `GET /api/v1/me` và `GET /auth/getProfile` trả `current_license_type`.
- `POST /auth/login` cũng trả trường này trong `data`.
- Giá trị hợp lệ của ứng dụng: `A1`, `A`, `B`, `C1`, `C`, `D`.
- Tài khoản cũ có chuỗi rỗng: mobile hiện màn hình chọn hạng. Không tự chọn A1 cho user cũ.
- Server lấy user_id từ JWT; không nhận user_id/role do client gửi để cập nhật.
- Bộ lọc `license_type` được truyền rõ ràng luôn ưu tiên hơn hạng đã lưu. Hạng đang học là tùy chọn nội dung, không phải quyền đăng ký/gói học.
- Đổi hạng không thay snapshot, thời hạn, điểm đạt hoặc lịch sử của bài đang làm.

### API tự dùng hạng đã lưu khi không truyền license_type

`/practice/questions`, `/practice/wrong-questions`, `/progress`, `/tests`, `/exam-templates`, `/road-signs`, `/road-signs/:id`, `/road-signs/:id/questions`, và danh sách chủ đề `/categories` của learner.

Nếu chưa lưu hạng và không truyền filter, các danh sách vẫn đọc mọi hạng để tương thích client cũ. Riêng `POST /exams/random` yêu cầu có hạng rõ ràng từ request hoặc hồ sơ. API admin không tự áp dụng hạng cá nhân. `/exams` (lịch sử) cũng không tự áp dụng: tránh giấu bài đang làm sau khi đổi hạng.

## 2. Loại chủ đề cho tabbar

```http
GET /api/v1/category-types
GET /api/v1/categories?category_type=SIGN&license_type=A1
```

`category-types` trả `{ "data": [{"code":"SIGN","name":"Biển báo"}, ...] }`.

| type | Tên |
|---|---|
| UNCATEGORIZED | Chưa phân loại |
| CONCEPT | Khái niệm |
| RULE | Quy tắc giao thông |
| SIGN | Biển báo |
| SA_HINH | Sa hình |
| CULTURE | Văn hóa giao thông |
| TECHNIQUE | Kỹ thuật lái xe |
| STRUCTURE | Cấu tạo và sửa chữa |

Tạo/sửa chủ đề bằng API admin hiện có:

```http
POST /api/v1/admin/categories

{"name":"Biển cảnh báo","description":"Học liệu đã được biên tập","type":"SIGN"}
```

Sửa dùng `PUT /api/v1/admin/categories/:id` với name/description/type. Bỏ qua type khi sửa giữ nguyên type cũ; khi tạo mặc định UNCATEGORIZED. Dữ liệu cũ không được phân loại bằng suy đoán từ tên. Không cho đổi chủ đề có bản ghi biển báo sang loại khác SIGN.

Mobile có thể dùng danh sách type làm nhóm tab cố định, sau đó `/categories` để hiển thị chủ đề thực tế bên trong. Learner chỉ nhận chủ đề có câu hỏi hoặc bài học biển báo đã xuất bản phù hợp hạng. Admin vẫn thấy chủ đề rỗng/nháp để biên tập.

```http
GET /api/v1/practice/questions?category_type=SIGN&category_id=12&license_type=A1&page=1&limit=20
GET /api/v1/practice/questions?category_type=RULE&is_critical=true
GET /api/v1/practice/wrong-questions?category_type=SIGN
```

Các điều kiện được kết hợp AND. Lọc trước COUNT/LIMIT/OFFSET, không để frontend ẩn câu sau khi phân trang. Danh sách luyện tập không trả is_correct/lời giải trước khi trả lời.

## 3. Đề ngẫu nhiên do mobile yêu cầu

### Mobile gửi gì?

```http
POST /api/v1/exams/random

{}
```

Body `{}` dùng hạng đã lưu và cấu hình tổng hợp đã xuất bản mới nhất của hạng đó. Hoặc chọn rõ:

```http
POST /api/v1/exams/random

{"license_type":"A1","template_id":5}
```

Thi riêng nhóm biển báo:

```http
POST /api/v1/exams/random

{"category_type":"SIGN"}
```

- `template_id` không có: lấy template đã xuất bản có ID lớn nhất, khớp hạng và category_type. category_type rỗng nghĩa là cấu hình tổng hợp, không tự rơi sang cấu hình SIGN.
- `template_id` có: phải tồn tại, đã xuất bản, khớp hạng. Nếu đồng thời gửi category_type thì phải khớp cả loại.
- Mobile không quyết định quota, đáp án đúng, thời gian hoặc điểm đạt. Chỉ các trường `license_type`, `template_id`, `category_type` được sử dụng; trường bổ sung không thay được quy tắc server.
- Mỗi lần gọi thành công tạo phiên mới. Retry POST sau lỗi mạng có thể tạo thêm phiên; mobile chống nhấn đúp và dùng lịch sử để tìm phiên vừa tạo. Random không bảo đảm hai lần có bộ câu khác nhau, nhất là ngân hàng nhỏ.
- Thiếu câu/điểm liệt hoặc không có template phù hợp trả 400 có lý do; không để lại phiên dở dang và không tự thay số lượng yêu cầu.

### Response 201

```json
{
  "id": 31,
  "test_id": null,
  "template_id": 5,
  "status": "DOING",
  "started_at": "2026-10-28T10:00:00Z",
  "expires_at": "2026-10-28T10:10:00Z",
  "server_time": "2026-10-28T10:00:00Z",
  "total_questions": 2,
  "passing_score": 2,
  "exam": {
    "name": "Ôn biển báo A1 - demo",
    "license_type": "A1",
    "version": "demo-v1",
    "category_type": "SIGN",
    "questions": []
  }
}
```

Ví dụ response đã rút gọn; `exam.questions` thực tế chứa đủ câu được chọn với ID, nội dung, ảnh, category_id, road_sign_id nếu có, phương án và dấu điểm liệt. Không có đáp án đúng/lời giải khi DOING.

**Thay đổi client cần xử lý:** `test_id` có thể null trong response phiên thi/lịch sử đối với bài random. Bài cố định vẫn có test_id là số như trước. Dùng **id của phiên**, không dùng test_id, cho các thao tác tiếp theo:

```http
PUT /api/v1/exams/31/answers/123

{"answer_id":456}
```

```http
GET /api/v1/exams/31
POST /api/v1/exams/31/submit
GET /api/v1/exams?license_type=A1&category_type=SIGN&status=DOING&page=1&limit=20
```

Quyền sở hữu, lưu/bỏ chọn, nộp lại an toàn, timeout, điểm liệt, snapshot, lịch sử và tiến bộ dùng chung với bài cố định. Lịch sử có filter `license_type`, `category_type`, `status=DOING|COMPLETED|TIMEOUT`. Loại/hạng trong lịch sử lấy từ snapshot tại lúc thi.

### Admin cấu hình đề trước

| Method | API | Chức năng |
|---|---|---|
| GET | `/api/v1/admin/exam-templates` | Danh sách cả nháp/xuất bản |
| GET | `/api/v1/admin/exam-templates/:id` | Chi tiết |
| POST | `/api/v1/admin/exam-templates` | Tạo cấu hình |
| PUT | `/api/v1/admin/exam-templates/:id` | Thay cấu hình |
| DELETE | `/api/v1/admin/exam-templates/:id` | Xóa mềm |
| GET | `/api/v1/exam-templates` | Learner chọn template đã xuất bản |

Body tạo/sửa, ví dụ demo **không phải cấu trúc đề thi chính thức**:

```json
{
  "name":"Ôn biển báo A1 - demo",
  "description":"Cấu hình minh họa cho kiểm thử",
  "license_type":"A1",
  "category_type":"SIGN",
  "version":"demo-v1",
  "duration_minutes":10,
  "passing_score":2,
  "critical_fail":true,
  "published":true,
  "quotas":[{"category_id":12,"count":2,"critical_count":1}]
}
```

`count` đã bao gồm `critical_count`. Không gửi question_ids cho template. Có 1–200 câu tổng cộng; từng chủ đề không lặp, category_id tồn tại và phải khớp category_type nếu cấu hình có loại. Có thể lưu cấu hình trước khi đủ câu; khi bắt đầu thi sẽ kiểm tra lại số lượng thật.

Response template có metadata ở ngoài (`id`, `name`, `license_type`, `category_type`, `published`) và toàn bộ cấu hình trong **`config`**. PUT nhận cấu hình phẳng như ví dụ, không bọc trong config. Cấu hình mới/sửa không thay phiên đã bắt đầu. Xóa template không xóa snapshot hoặc lịch sử.

## 4. Học biển báo theo hạng bằng

### Nhóm biển

`GET /api/v1/road-sign-types` trả code/name của `PROHIBITION`, `WARNING`, `MANDATORY`, `INFORMATION`, `SUPPLEMENTARY`, `OTHER`.

### CRUD admin

| Method | API | Chức năng |
|---|---|---|
| GET | `/api/v1/admin/road-signs` | Danh sách, lọc/phân trang |
| GET | `/api/v1/admin/road-signs/:id` | Chi tiết |
| POST | `/api/v1/admin/road-signs` | Tạo bài học biển |
| PUT | `/api/v1/admin/road-signs/:id` | Cập nhật |
| DELETE | `/api/v1/admin/road-signs/:id` | Xóa mềm, từ chối nếu có câu tham chiếu |

```json
{
  "code":"DEMO-01",
  "name":"Biển minh họa",
  "image":"/images/demo.png",
  "meaning":"Ý nghĩa đã được người biên soạn kiểm tra",
  "description":"Ghi chú cách ghi nhớ",
  "sign_type":"WARNING",
  "license_type":"A1",
  "category_id":12,
  "published":true
}
```

Quy tắc:

- Một bản ghi học liệu cho một hạng; cùng mã có thể có bản ghi riêng cho A1/B. Mã không trùng trong cùng hạng với các bản ghi chưa xóa.
- Chủ đề phải có type=SIGN. Biển xuất bản cần ảnh và meaning; ảnh vẫn là URL/đường dẫn, chưa có upload.
- Không được thay hạng/chủ đề của biển khi còn câu liên kết. Không thu hồi xuất bản biển khi còn câu liên kết đã xuất bản: thu hồi các câu trước.
- Không tự nhập nội dung pháp luật hoặc ảnh mẫu vào database thật.

### Learner học và làm câu hỏi

```http
GET /api/v1/road-signs?sign_type=WARNING&page=1&limit=20
GET /api/v1/road-signs/9
GET /api/v1/road-signs/9/questions
```

Dùng hạng đã lưu hoặc truyền `license_type=A1`. Chỉ trả biển đã xuất bản; ID nháp/sai hạng trả 404. Bài học trả meaning để học, còn danh sách câu luyện không trả đáp án đúng.

Liên kết câu với biển bằng trường mới trong body tạo/sửa câu hỏi:

```json
{
  "title":"Câu minh họa về biển",
  "content":"Chọn phương án đúng trong ví dụ?",
  "image":"/images/demo.png",
  "explanation":"Lời giải đã kiểm tra",
  "license_type":"A1",
  "category_id":12,
  "road_sign_id":9,
  "published":true,
  "is_critical":false,
  "answers":[
    {"label":"A","content":"Phương án A","order":1,"is_correct":true},
    {"label":"B","content":"Phương án B","order":2,"is_correct":false}
  ]
}
```

- Câu và biển phải cùng hạng và category_id. Câu xuất bản chỉ được liên kết biển đã xuất bản.
- `road_sign_id: null` bỏ liên kết; mỗi câu có tối đa một biển liên kết. Câu thuộc type SIGN không bắt buộc gắn một biển cụ thể (ví dụ câu so sánh nhiều biển).
- Câu từ crawler/nhập trước đây không tự được gán biển; admin biên tập liên kết nếu cần.
- PUT câu là thay toàn bộ nội dung: giữ road_sign_id nếu vẫn muốn liên kết.

Luyện tập/chấm điểm/lưu lịch sử dùng API hiện có:

```http
GET /api/v1/practice/questions?category_type=SIGN&sign_type=WARNING
GET /api/v1/practice/questions?road_sign_id=9
POST /api/v1/practice/questions/123/answer

{"answer_id":456}
```

Thi thử biển báo dùng template category_type=SIGN và `POST /api/v1/exams/random` với `{"category_type":"SIGN"}`. Đề cố định cũng có trường category_type; nếu đặt SIGN, mọi câu trong đề phải thuộc chủ đề SIGN. Không tự tạo đề khi chỉ tạo một bài học biển.

## 5. Ma trận filter cho mobile

| API | Filter hỗ trợ |
|---|---|
| `/categories` | license_type (learner), category_type, search; trả toàn bộ tab phù hợp |
| `/practice/questions`, `/practice/wrong-questions`, `/admin/questions` | license_type, category_type, category_id, road_sign_id, sign_type, is_critical, search, page, limit; admin thêm published |
| `/tests`, `/admin/tests` | license_type, category_type, search, page, limit; admin thêm published |
| `/road-signs`, `/admin/road-signs` | license_type, category_id, sign_type, search, page, limit; admin thêm published |
| `/exam-templates`, `/admin/exam-templates` | license_type, category_type, search, page, limit; admin thêm published |
| `/exams` | license_type, category_type, status, page, limit; chỉ phiên của user hiện tại |
| `/progress` | license_type, category_type |

`category_type` trong progress lọc lượt luyện theo loại chủ đề hiện tại và các bài thi có snapshot cùng loại. Ví dụ SIGN không tính đề tổng hợp vào tỷ lệ đỗ SIGN. Ôn câu sai vẫn xét cả câu sai trong đề tổng hợp rồi lọc ngân hàng theo chủ đề/biển.

Phân trang mặc định 20, tối đa 100/trang. Enum/số/filter sai trả 400. Không gửi filter ngoài ma trận: chưa phải mọi API đều hiểu mọi filter.

## 6. Web quản trị local

- **Chủ đề:** thêm trường Loại chủ đề.
- **Học biển báo:** danh sách/lọc, tạo/sửa/xóa, xuất bản.
- **Ngân hàng câu hỏi:** chọn liên kết biển theo đúng hạng/chủ đề, lọc category_type.
- **Cấu hình đề random:** tạo/sửa/xóa template và quota, xuất bản cho learner.
- **Đề thi:** phân loại đề, picker lọc category_type trước phân trang.

Khởi động lại Go server để nạp schema/asset mới. Web `/admin/` vẫn là giao diện quản trị, chưa có giao diện learner; mobile dùng API trong tài liệu này.

## 7. Migration, tương thích và quyết định thiết kế

- Thêm users.current_license_type, categories.type, questions.road_sign_id, tests.category_type, bảng road_signs/exam_templates và exam_sessions.template_id.
- exam_sessions.test_id cho phép NULL với phiên random. Không tạo một Test công khai mỗi lần learner thi; tránh làm đầy danh sách đề chung.
- Snapshot và câu trả lời vẫn gắn với phiên/user; không suy quyền từ template_id hoặc test_id.
- Hạng và loại của phiên được lưu trong snapshot; lịch sử không phụ thuộc template hiện tại.
- SQLite cần dựng lại bảng khi thêm foreign key. Migration tắt enforcement trước transaction, thực hiện AutoMigrate, chạy foreign_key_check trước commit, rồi bật lại enforcement. Lỗi rollback schema và dữ liệu; database có orphan cũ phải được sửa trước khi nâng cấp.
- Sao lưu database trước khi chạy phiên bản mới. Chưa có rollback migration theo phiên bản; rollback transaction chỉ bảo vệ lần migration thất bại.
- Bản sao DB hiện có đã được thử nâng cấp: giữ 855 câu và 3 phiên cũ, không có lỗi khóa ngoại trước khi tạo dữ liệu QA.
- Không tự phân loại nội dung cũ từ tên, không tự tạo template chính thức, không tự xuất bản dữ liệu cũ.

## 8. Kiểm thử và đường dẫn code

Nghiệp vụ: `internal/learning/mobile.go`, `road_signs.go`, `templates.go`, `internal/exam/usecase/random.go`.

HTTP: `internal/learning/delivery/mobile.go`, `internal/exam/delivery/router.go`.

Test: `internal/server/mobile_test.go` (phân quyền, preference, filter, sign CRUD/link, random snapshot/ownership/grading/rollback/history), `internal/database/migrate_test.go` (nâng cấp bảng có tham chiếu, idempotence, rollback).

Comment trong code giải thích các điểm quan trọng: hạng là preference thay vì ACL, template server quyết định quy tắc, snapshot riêng không tạo đề dùng chung, migration cần kiểm tra toàn vẹn.

## Tiến độ số câu đã đạt theo hạng bằng

`GET /api/v1/progress?license_type=A1` — yêu cầu `Authorization: Bearer <access_token>`.
Không truyền `license_type` thì dùng hạng đã lưu bằng `PUT /api/v1/me/license`.
Nếu chưa lưu hạng và không truyền query, API thống kê tất cả hạng, trả `license_type: ""`.
User được lấy từ token, không truyền `user_id`.

Các trường mới trong response (ví dụ minh họa; các trường thống kê cũ vẫn trả về):

```json
{
  "license_type": "A1",
  "total_questions": 250,
  "attempted_questions": 100,
  "mastered_questions": 80,
  "wrong_questions": 20,
  "unattempted_questions": 150,
  "completion_percentage": 32
}
```

| Trường | Ý nghĩa |
| --- | --- |
| `total_questions` | Tổng số câu đang xuất bản, chưa bị xóa, thuộc hạng và bộ lọc hiện tại. Không phải số câu cố định theo hạng. |
| `attempted_questions` | Số câu riêng biệt đã luyện hoặc đã xuất hiện trong bài thi kết thúc/hết giờ. |
| `mastered_questions` | Số câu có kết quả gần nhất là đúng. Làm lại sai sẽ giảm số này. |
| `wrong_questions` | Số câu có kết quả gần nhất là sai hoặc bỏ trống trong bài thi đã kết thúc. |
| `unattempted_questions` | `total_questions - attempted_questions`. |
| `completion_percentage` | `100 * mastered_questions / total_questions`; trả 0 nếu tổng bằng 0. Client tự làm tròn khi hiển thị. |

Mobile có thể hiển thị **Đã đạt 80/250 câu** và progress bar **32%** chỉ với một request.
Mỗi câu chỉ tính một lần; luyện lại nhiều lần không tăng tổng câu đã làm.
`total_questions = mastered_questions + wrong_questions + unattempted_questions`.

Kết quả luyện tập dùng thời gian tạo lượt luyện; kết quả thi dùng thời gian kết thúc
(kể cả timeout). Bài đang thi chưa đóng góp vào thống kê. Nếu thời gian trùng nhau
chính xác, ưu tiên kết quả thi; trong cùng nguồn ưu tiên ID lớn hơn.
Các câu bị xóa hoặc ngừng xuất bản không nằm trong các số đếm mới.

Các số đếm mới hỗ trợ `license_type`, `category_type`, `category_id`, `is_critical`,
`road_sign_id`, `sign_type`, `search`. Không phân trang và luôn chỉ tính câu đã
xuất bản, bất kể `page`, `limit`, `published` được gửi lên.
Lọc theo chủ đề/biển báo vẫn tính câu đã làm trong đề tổng hợp.

Các trường cũ `completed_exams`, `passed_exams`, `pass_rate`, `categories`,
`recent_exams`, `suggested_category_id`, `wrong_question_ids` giữ phạm vi thống kê
cũ: theo hạng và `category_type`; loại chủ đề của bài thi dựa trên snapshot.
`categories[].accuracy` tính theo lượt làm, không phải tỷ lệ câu riêng biệt đã đạt.
Không dùng độ dài `wrong_question_ids` để suy ra `wrong_questions`: danh sách cũ
có thể chứa câu lịch sử đã ngừng xuất bản và có phạm vi lọc khác.

## Chọn hạng bằng và lưu lịch sử theo hạng

Các API dưới đây yêu cầu Bearer token của learner.

1. `GET /api/v1/license-types`: lấy các hạng hệ thống hỗ trợ để tạo màn hình chọn bằng.

   ```json
   {"data":[{"code":"A1","name":"Hạng A1"},{"code":"A","name":"Hạng A"},{"code":"B","name":"Hạng B"},{"code":"C1","name":"Hạng C1"},{"code":"C","name":"Hạng C"},{"code":"D","name":"Hạng D"}]}
   ```

   Đây là danh mục mã hệ thống hỗ trợ, không đảm bảo mỗi hạng đã có đề/câu hỏi xuất bản.

2. `PUT /api/v1/me/license` với body `{"license_type":"A1"}`: lưu lựa chọn vào tài khoản. `GET /api/v1/me` trả `current_license_type` để khôi phục lựa chọn sau khi đăng nhập lại.
3. `GET /api/v1/tests`: lấy đề theo hạng đã lưu. Bắt đầu bằng `POST /api/v1/exams` với `{"test_id":123}`, hoặc dùng API random hiện có. Đề cố định lấy hạng từ `test_id`; mobile phải chọn đề từ danh sách đúng hạng. Random không chỉ định hạng/template sẽ dùng hạng đã lưu theo luồng hiện có.
4. `POST /api/v1/exams/:id/submit`: backend lưu kết quả vào phiên của user, không cần client gửi hạng để lưu lại. Hạng được giữ trong snapshot lúc bắt đầu. Đổi hạng trước hoặc sau khi nộp không đổi lịch sử cũ.
5. `GET /api/v1/exams?license_type=A1`: lấy lịch sử theo hạng đang chọn; mỗi phần tử `data` nay có `license_type` bên cạnh `score`, `passing_score`, `is_passed`, `status`.

Lịch sử **không tự áp dụng hạng đã lưu**: bỏ `license_type` vẫn trả tất cả hạng,
để có thể tìm lại bài đang làm sau khi đổi hạng. Mobile truyền rõ mã đang chọn khi
hiển thị tab lịch sử theo hạng. Query `status=COMPLETED` lọc bài đã nộp,
`status=TIMEOUT` lọc bài hết giờ; không gửi `status` sẽ gồm cả bài đang làm.
Danh sách luôn chỉ gồm các phiên của user trong token.
