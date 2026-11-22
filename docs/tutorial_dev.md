# Tutorial cho dev — API ứng dụng thi bằng lái xe

Đối chiếu code ngày 25/10/2026. Tài liệu này nói về **những gì backend hiện làm**, không mô tả một app lý tưởng chưa được viết. Bạn có thể đọc theo tình huống hoặc tìm đúng tên API cần dùng. File hướng dẫn cũ được giữ nguyên.

**Đọc theo nhu cầu:**

- Muốn hiểu app vận hành: đọc mục 1–3.
- Đang viết màn thi lại, tiếp tục bài, hết giờ: đọc mục 4.
- Đang viết ôn tập và câu sai: đọc mục 5.
- Cần gửi request cụ thể: tra mục 6 (API chính) hoặc 7 (API cũ).
- Gặp lỗi hoặc tìm chức năng chưa có: mục 8–9.

## 1. Hiểu sản phẩm trước khi gọi API

Bạn đang làm hai phần: màn quản trị cho người soạn câu và màn học/thi cho người học.

Admin chuẩn bị **chủ đề → câu hỏi → đề thi**. Người học chọn đề, bắt đầu một lượt thi, chọn đáp án rồi nộp. Nếu chỉ muốn học từng câu thì vào ôn tập, không cần tạo lượt thi.

| Khái niệm | Hiểu như đang dùng app | Biến Postman |
|---|---|---|
| Chủ đề | Một nhóm câu, ví dụ biển báo | `category_id` |
| Câu hỏi | Nội dung câu và các phương án A/B/C | `question_id` |
| Phương án | Một lựa chọn cụ thể của một câu | `answer_id` |
| Đề thi | Danh sách câu, thời gian và điểm đạt | `test_id` |
| Lượt thi / phiên thi | Một lần một người làm một đề | `exam_id` |
| Xuất bản | Admin cho phép dữ liệu được dùng | `published: true` |
| Điểm liệt | Câu có thể khiến người học trượt khi sai/bỏ trống | `is_critical` trên câu |
| Áp dụng điểm liệt | Đề có bật quy tắc trượt vì câu liệt hay không | `critical_fail` trên đề |

Ví dụ đề số 3 được bạn làm hai lần: lần đầu là lượt 10, lần sau là lượt 11. Cùng `test_id=3`, khác `exam_id`. ID là mã định danh, **không phải số lượng bản ghi**. Các số trong ví dụ chỉ minh họa, không đảm bảo có trong DB của bạn.

Một câu có `is_critical=true` không tự làm mọi đề trượt: đề phải bật `critical_fail`. Câu đã xuất bản cũng không tự được thêm vào đề nào.

## 2. Chuẩn bị một lần, dùng cho toàn bộ tài liệu

### Chạy backend

Mở terminal tại thư mục dự án. Nếu chưa có `.env`, sao chép `.env.example` thành `.env`; không ghi đè file đang có. Đặt `JWT_SECRET` riêng, `DATABASE_PATH=cmd/data/gorm.db` và `LISTEN_ADDR=:8080` nếu chạy mặc định.

```bash
go run ./cmd/server
```

Server tự migrate cấu trúc bảng. DB rỗng không tự có tài khoản admin, câu hỏi hoặc đề. Sau reset DB, đăng ký lại và lấy lại tất cả ID/token; không dùng biến Postman còn từ DB cũ. Nên khởi động lại server với JWT secret mới khi cần vô hiệu hóa cả access token cũ sau reset; chỉ xóa dòng DB không tự thay khóa ký JWT.

### Quy ước Postman

Tạo Environment, chọn Environment đó trước khi Send:

| Biến | Lấy ở đâu |
|---|---|
| `base_url` | Nhập `http://localhost:8080`, không có dấu `/` cuối |
| `admin_token` | `data.access_token` khi login tài khoản admin |
| `learner_token` | `data.access_token` khi login người học |
| `learner_refresh_token` | `data.refresh_token` của người học |
| `category_id` | `id` của chủ đề đã tạo/chọn |
| `question_id` | `id` của câu đang thao tác |
| `answer_id` | `id` phương án đang chọn; khi thi phải lấy từ lượt thi |
| `test_id` | `id` của đề đang chọn |
| `exam_id` | `id` khi bắt đầu lượt thi |

Import từng cURL bằng **Import → Raw text**, rồi kiểm tra URL, token và body. Các mẫu dùng biến Postman, không phải biến shell. Nếu chạy terminal, thay chúng bằng giá trị thật. Body số như `{{test_id}}` không bọc dấu nháy; token là chuỗi nên có nháy trong JSON.

Không có script tự đổi biến trong tài liệu này. Bạn chủ động lấy ID từ response và nhập vào Environment để biết mình đang thao tác dữ liệu nào. Khi đổi đề hoặc bắt đầu lại, nhớ thay biến liên quan.

`Bearer` là cách gửi thẻ đăng nhập: header `Authorization: Bearer <access_token>`. Không dùng refresh token thay access token. Admin cũng có thể gọi API người học, nhưng lịch sử lúc đó thuộc tài khoản admin; để thử như người dùng thật, dùng tài khoản learner riêng.

### Đọc response cho đúng

| API | Dữ liệu nằm ở đâu? |
|---|---|
| Register/login/refresh | Trong `data`; token là `data.access_token` |
| Profile | Object trực tiếp, ví dụ `role` |
| Tạo/sửa/xem câu, chủ đề, đề; bắt đầu/xem/lưu/nộp lượt thi | Object trực tiếp: `id`, không phải `data.id` |
| Danh sách v1 có phân trang | `data`, `total`, `page`, `limit` |
| Danh sách chủ đề | `{ "data": [...] }`, không có phân trang |
| Trả lời ôn tập | `attempt` và `question` |
| Thống kê | Object trực tiếp |
| DELETE v1 thành công | HTTP 204, không có body |

## 3. Nối các màn hình thành luồng chạy được

### Mới reset DB: làm gì trước?

1. Register tài khoản quản trị. Chạy CLI cấp admin theo mục API đăng ký bên dưới, rồi login.
2. Register/login tài khoản người học riêng.
3. Admin tạo chủ đề, tạo ít nhất một câu hợp lệ đã xuất bản. Hoặc crawl, rà soát rồi xuất bản câu đã nhập.
4. Admin tạo đề chứa các ID câu đó và đặt `published=true`.
5. Người học lấy danh sách đề, bắt đầu một lượt, lưu lựa chọn, nộp và xem lại.

Để thử nhanh nhất, một đề một câu, điểm đạt 1, thời gian 10 phút là đủ. Đây là dữ liệu thử API, không phải cấu hình kỳ thi chính thức.

### Người học bấm “Vào thi”

GET danh sách đề → chọn một `test_id` → POST `/api/v1/exams` → lưu `id` trả về thành `exam_id`. Màn làm bài hiển thị `exam.questions` từ response vừa nhận. Không gọi API quản trị để lấy đáp án đúng cho màn thi.

Mỗi lần chọn phương án, gửi PUT đáp án. Đợi lưu thành công rồi mới báo “Đã lưu”. Khi bấm nộp, đợi các request lưu đang chạy kết thúc rồi mới POST submit. Server chỉ chấm lựa chọn đã lưu trên server, không biết lựa chọn chỉ còn ở màn hình điện thoại.

### Muốn thi lại cùng đề

POST `/api/v1/exams` với **cùng `test_id`**, rồi lưu `exam_id` mới. Bài cũ và điểm cũ vẫn còn. Không gọi submit bài cũ để thi lại; không có API reset một lượt thi về ban đầu.

Thi lại dùng phiên bản đề hiện tại tại thời điểm bắt đầu, nên có thể khác bài cũ nếu admin đã sửa. Nếu đề bị ẩn/xóa hoặc câu bên trong không còn hợp lệ, không bắt đầu lại được dù vẫn xem lịch sử cũ được.

### Muốn tiếp tục bài đang làm

Dùng `exam_id` đã lưu, GET `/api/v1/exams/:id`. Nếu `status=DOING`, còn giờ thì dựng lại màn làm bài với `selected_answer_id`. Câu không có lựa chọn thì không đánh dấu phương án nào.

Nếu mất `exam_id`, gọi GET lịch sử, tìm lượt `DOING` mình muốn tiếp tục rồi GET chi tiết. Lịch sử có cả bài đang làm, không chỉ bài đã nộp. Backend chưa có endpoint “lượt đang làm duy nhất”; có thể có nhiều lượt cùng đang mở.

### Muốn làm đề khác hoặc đề ngẫu nhiên

Chọn `test_id` khác rồi bắt đầu mới. “Thi lại” không tự trộn câu. Một đề cố định giữ thứ tự câu đã cấu hình. Sinh đề ngẫu nhiên là POST admin `/tests/generate`, tạo và lưu một đề mới; sau đó người học bắt đầu đề ấy. Hiện không có API người học tự xin đề ngẫu nhiên riêng.

### Muốn ôn bài thay vì thi

GET `/practice/questions` → người học chọn phương án → POST `/practice/questions/:id/answer`. Nhận đúng/sai và lời giải ngay sau mỗi câu. Không cần đề, không có `exam_id`, không có giờ thi, không phải nộp toàn bộ bài.

### Thi trượt rồi ôn lại câu sai

GET `/practice/wrong-questions` → trả lời từng câu bằng API ôn bình thường → tải lại danh sách câu sai. Trả lời đúng ở lượt gần nhất thì câu được loại khỏi danh sách sai. Nếu sau đó làm sai lại, câu có thể xuất hiện lại.

Ôn đúng **không biến bài thi trượt cũ thành đậu**. Muốn có kết quả thi mới thì bắt đầu một lượt thi mới.

## 4. Những tình huống phải xử lý ở màn thi

| Tình huống | Backend hiện xử lý thế nào? | App nên làm gì? |
|---|---|---|
| Bấm bắt đầu hai lần | Mỗi POST tạo một lượt mới; chưa có khóa chống tạo lặp | Khóa nút lúc đang gửi; đừng tự retry POST vô hạn |
| Đã có bài đang làm nhưng bắt đầu thêm | Backend vẫn cho tạo lượt khác | Nếu sản phẩm chỉ muốn một bài đang làm, hỏi người học tiếp tục hay bắt đầu mới ở app |
| POST bắt đầu bị mất response | Không biết chắc server đã tạo chưa | Đọc lịch sử để đối chiếu trước khi gửi lại |
| Chọn A rồi đổi sang B | PUT cập nhật lựa chọn của cùng câu trong cùng lượt | Gửi ID B; không tạo lượt mới |
| Bỏ chọn | PUT `answer_id: 0` xóa lựa chọn | Dùng 0, không gửi thiếu trường hoặc null |
| Chọn nhiều phương án | Chỉ một `answer_id` mỗi câu | UI dùng chọn đơn, không gửi mảng |
| Bỏ trống một số/toàn bộ câu rồi nộp | Vẫn chấm; câu trống tính sai | Có thể hỏi xác nhận trên app, backend không bắt trả lời hết |
| Lưu lại đúng một đáp án nhiều lần | Cập nhật cùng bản ghi, không nhân số câu trả lời | Retry được khi còn giờ, nhưng không gửi nhiều lựa chọn khác nhau song song |
| A và B gửi gần đồng thời | Lựa chọn cuối theo thứ tự server xử lý | Xếp hàng request theo câu để tránh lựa chọn cũ ghi đè lựa chọn mới |
| Mất mạng khi chọn | Server chỉ có những gì đã nhận thành công | GET lại bài để đối chiếu; đánh dấu lựa chọn chưa đồng bộ |
| Tắt app, đổi tab, đăng xuất | Đồng hồ không dừng | Mở lại thì GET lượt cũ; không tính lại thời gian từ đầu |
| Hết giờ nhưng app không mở | Hạn đã cố định; trạng thái DB có thể chưa cập nhật ngay | Đọc lại sẽ xử lý hết giờ; không dựa riêng trạng thái cũ trong DB |
| Đúng thời điểm hết hạn | Server coi là hết giờ | Không cho thêm một giây vì đồng hồ client chậm |
| Lưu sau hết giờ | Chấm timeout rồi trả 409, không nhận lựa chọn mới | GET chi tiết và chuyển sang màn kết quả |
| Nộp sau hết giờ | Trả kết quả `TIMEOUT` từ lựa chọn đã lưu | Hiển thị kết quả; không gửi điểm do app tự tính |
| Nộp lại / mất response lúc nộp | Trả kết quả đã chấm, không tạo bài mới | Có thể retry submit hoặc GET chi tiết |
| Đổi đáp án sau nộp | Trả 409 | Khóa màn làm bài; muốn làm lại thì tạo lượt mới |
| Câu/đáp án không thuộc lượt | Trả 400 | Lấy ID từ `exam.questions` của đúng lượt, không từ ngân hàng mới |
| Người khác đọc/sửa/nộp bài của bạn | Trả 404 theo phạm vi chủ sở hữu | Đăng nhập đúng người đã bắt đầu, admin không tự xem được bài người khác qua API này |
| Đề/câu bị admin sửa sau khi bạn bắt đầu | Lượt hiện tại dùng bản chụp lúc bắt đầu | Tiếp tục dùng dữ liệu phiên, không trộn dữ liệu ngân hàng |
| Đề bị xóa sau khi bắt đầu | Bản chụp của lượt còn; có thể tiếp tục/nộp theo trạng thái và hạn | Không xóa màn lịch sử chỉ vì không còn đề trong danh sách |
| Token hết hạn giữa bài | Request cần đăng nhập bị từ chối; giờ thi vẫn chạy | Refresh hoặc login, rồi GET cùng `exam_id` |
| Bấm “thoát/hủy thi” | Chưa có API hủy hay pause | Rời màn hình không xóa lượt; sẽ hết giờ hoặc người học quay lại nộp |

### Đọc điểm và trạng thái

`DOING` = đang làm. `COMPLETED` = đã nộp trước hạn. `TIMEOUT` = hết giờ. Hai trạng thái sau đều là bài đã chấm, không đồng nghĩa đậu hay rớt.

Mỗi câu đúng được 1 điểm; `score` hiện bằng `correct_answers`, không phải thang 10 hay phần trăm. Trường trọng số trong bảng nối chưa được dùng để chấm trọng số. `score=0`, `is_passed=false` khi DOING là giá trị chưa chấm, không phải kết luận trượt.

| Kết quả | Điều kiện |
|---|---|
| Đậu | Điểm >= `passing_score` và không vi phạm quy tắc điểm liệt đang bật |
| Trượt điểm liệt | `critical_fail=true` và ít nhất một câu liệt sai hoặc trống; `failure_reason=critical_question` |
| Trượt thiếu điểm | Không vi phạm điểm liệt nhưng điểm dưới ngưỡng; `failure_reason=insufficient_score` |
| Vừa thiếu điểm vừa sai liệt | Lý do trả về ưu tiên `critical_question` |
| Hết giờ nhưng đủ điểm | Vẫn có thể đậu; thời gian hết không tự làm trượt |

Server giữ `started_at`, `expires_at`, `server_time`. UI dùng chúng để dựng đồng hồ còn lại và đồng bộ khi đọc lại; server mới là nơi quyết định còn hạn hay không. Hiện chưa có worker quét liên tục. GET/lưu/nộp bài, lịch sử, tiến bộ và câu sai có thể kích hoạt chấm các lượt đã hết hạn.

Bản chụp (snapshot) nghĩa là backend lưu bản sao nội dung, đáp án và quy tắc chấm lúc bắt đầu. Admin sửa sau đó không thay điểm đạt hay đáp án đúng của lượt đang làm. Khi DOING, API che đáp án chuẩn và lời giải; khi đã chấm, `exam.questions[i].is_correct` nói **người học làm đúng không**, còn `answers[j].is_correct` nói **phương án đó có phải đáp án đúng không**.

## 5. Những tình huống ở màn ôn và thống kê

| Tình huống | Cách hiểu/cách dùng |
|---|---|
| Chưa từng học/thi | Lịch sử và câu sai rỗng, thống kê bằng 0 là bình thường |
| Ôn cùng một câu nhiều lần | Mỗi POST tạo lượt ôn mới và tính vào thống kê |
| Mất response ôn rồi gửi lại | Có thể ghi hai lượt; API chưa chống gửi lặp bằng mã request |
| Ôn muốn “bỏ chọn” | Không có thao tác bỏ chọn lượt đã ghi; `answer_id=0` bị từ chối, khác API thi |
| Muốn sửa lượt ôn cũ | Chưa có API; gửi lựa chọn mới là một lượt mới |
| Admin vừa sửa đáp án | Lấy lại câu hiện tại để lấy ID mới; ôn dùng ngân hàng hiện tại |
| Câu bị ẩn hoặc xóa | Không ôn được bằng endpoint hiện tại; không còn trong danh sách câu sai để học |
| Bài đang làm chưa chấm | Chưa góp kết quả đúng/sai vào tiến bộ; khi chấm mới tính |
| Bài đã hết giờ | Được tính cùng các bài hoàn thành sau khi hệ thống xử lý timeout |
| Bỏ trống trong bài đã chấm | Được tính sai và có thể vào danh sách câu sai |
| Ôn đúng sau lần thi sai | Có thể bỏ câu khỏi danh sách sai; bài thi cũ vẫn trượt |
| Thi sai lại sau khi ôn đúng | Lần làm gần nhất sai nên câu có thể trở lại danh sách |
| `wrong_question_ids` nhiều hơn danh sách ôn câu sai | Progress giữ dấu vết lịch sử; danh sách ôn lọc thêm câu hiện còn xuất bản/tồn tại |
| Có lịch sử thuộc nhiều hạng | Truyền `license_type=A1` để xem riêng; bỏ lọc để xem tất cả |

`completed_exams` tính cả COMPLETED và TIMEOUT; `passed_exams` đếm bài đậu; `pass_rate` là phần trăm từ 0 đến 100. Trong `categories`, `attempts` là tổng lượt trả lời/câu đã chấm, `correct` là lượt đúng, `accuracy` là phần trăm. `recent_exams` tối đa 20 bài đã chấm. `suggested_category_id` là chủ đề có độ chính xác thấp nhất trong dữ liệu đã làm và dưới 100%; không phải dự đoán AI và không gợi ý chủ đề chưa có dữ liệu.

“Gần nhất” đang so theo thời gian lượt ôn hoặc thời gian hoàn thành bài thi. Nếu nhiều sự kiện trùng chính xác thời gian, code chưa có quy tắc phụ rõ ràng để phân thắng thua. Không nên dựa vào thứ tự tuyệt đối khi gửi nhiều lượt đồng thời.

## 6. Danh mục API — mở đúng mục rồi dùng

Mỗi mục có điều kiện, cURL và cách đọc kết quả. API ghi JSON cần `Content-Type: application/json`. Các yêu cầu quyền trong mục là bắt buộc; không có token cho route bảo vệ thì 401, learner gọi route admin thì 403.

### 6.1. Kiểm tra server — GET /health

**Dùng khi / cần trước:** Server đang chạy; dùng trước khi kiểm tra lỗi token/DB.

**Quyền:** Công khai.

```bash
curl --request GET '{{base_url}}/health'
```

**Kết quả và chú ý:** 200 `{ "status": "ok" }`. Đây là kiểm tra tiến trình HTTP, không phải kiểm tra sâu tất cả dữ liệu.

### 6.2. Đăng ký — POST /auth/register

**Dùng khi / cần trước:** Email chưa đăng ký. `name` 2–100 ký tự, email hợp lệ tối đa 255, password 8–72 ký tự theo binding. Bcrypt giới hạn 72 byte, nên tránh mật khẩu nhiều ký tự Unicode vượt 72 byte.

**Quyền:** Công khai.

```bash
curl --request POST '{{base_url}}/auth/register' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "name": "Dev Admin",
  "email": "admin@example.com",
  "password": "DemoPass123!"
}'
```

**Kết quả và chú ý:** 201, người dùng trong `data`, mặc định learner. Trùng email trả 409. Đăng ký không tự login và không tự nâng quyền admin.

Để cấp admin cho tài khoản vừa tạo, mở terminal tại cùng thư mục/cùng cấu hình DB với server:

```bash
go run ./cmd/admin -email admin@example.com -role admin
```

Đây là lệnh quản trị máy chủ, không phải API Postman. Có thể dùng `-role learner` để hạ quyền. Tạo tài khoản học riêng bằng register với email khác. Role admin được middleware đọc từ DB, không tự nhận role do client gửi.

### 6.3. Đăng nhập — POST /auth/login

**Dùng khi / cần trước:** Đã register. Dùng đúng email/password.

**Quyền:** Công khai.

```bash
curl --request POST '{{base_url}}/auth/login' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "email": "admin@example.com",
  "password": "DemoPass123!"
}'
```

**Kết quả và chú ý:** 200, `data.access_token`, `data.refresh_token`, `data.expires_in`. Access token hiện 12 giờ, refresh 7 ngày. Sai thông tin trả 401. Login admin lưu vào `admin_token`; login learner lưu `learner_token` và `learner_refresh_token`. Login không tạo lượt thi.

### 6.4. Xem tài khoản đang dùng — GET /auth/getProfile

**Dùng khi / cần trước:** Có access token. Muốn kiểm tra admin thì thay header bằng `admin_token`.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/auth/getProfile' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 object trực tiếp với ID/tên/email/role. Dùng để biết có đang cầm nhầm token của người khác không.

### 6.5. Lấy token mới — POST /auth/refresh

**Dùng khi / cần trước:** Có refresh token còn hạn, chưa bị thu hồi. Không cần access token còn hạn.

**Quyền:** Công khai.

```bash
curl --request POST '{{base_url}}/auth/refresh' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "refresh_token": "{{learner_refresh_token}}"
}'
```

**Kết quả và chú ý:** 200, lưu cả `data.access_token` và `data.refresh_token` mới. Token refresh cũ bị thu hồi sau khi đổi thành công, không dùng lại. Hết hạn/không hợp lệ trả 401; login lại. Tránh nhiều request refresh song song; mất response sau khi đổi có thể phải login lại. Admin dùng refresh token của admin và lưu đúng biến admin.

### 6.6. Danh sách chủ đề — GET /api/v1/categories

**Dùng khi / cần trước:** Đã login.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/categories' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 `{data:[...]}`, theo ID tăng dần, không phân trang. Chủ đề có thể chưa có câu xuất bản.

### 6.7. Tạo chủ đề — POST /api/v1/admin/categories

**Dùng khi / cần trước:** Có token admin; tên không được trống hoặc chỉ khoảng trắng.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/categories' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "name": "Chủ đề thử API",
  "description": "Dữ liệu minh họa để thử luồng"
}'
```

**Kết quả và chú ý:** 201 object trực tiếp; lưu `id` thành `category_id`. Tên trùng không có quy tắc unique ở API này.

### 6.8. Sửa chủ đề — PUT /api/v1/admin/categories/:id

**Dùng khi / cần trước:** Chủ đề tồn tại. Gửi tên và mô tả muốn giữ.

**Quyền:** Admin.

```bash
curl --request PUT '{{base_url}}/api/v1/admin/categories/{{category_id}}' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "name": "Chủ đề thử đã sửa",
  "description": "Mô tả mới"
}'
```

**Kết quả và chú ý:** 200 object sau sửa; không có chủ đề thì 404. Không gửi description sẽ nhận giá trị rỗng trong cập nhật.

### 6.9. Xóa chủ đề — DELETE /api/v1/admin/categories/:id

**Dùng khi / cần trước:** Chủ đề tồn tại và không còn câu chưa xóa. Chỉ dùng khi thực sự muốn dọn dữ liệu.

**Quyền:** Admin.

```bash
curl --request DELETE '{{base_url}}/api/v1/admin/categories/{{category_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 204 rỗng; còn câu thì 400; không tồn tại thì 404. Đây là xóa mềm.

### 6.10. Tạo câu — POST /api/v1/admin/questions

**Dùng khi / cần trước:** Có chủ đề. Title/content không trống; hạng thuộc A1, A, B, C1, C, D. Ít nhất hai phương án, đúng một `is_correct=true`. Nhãn tối đa 2 ký tự theo DTO (service còn kiểm tra 2 byte); nên dùng A/B/C/D. Order là số dương. Nhãn, thứ tự và nội dung không trùng sau chuẩn hóa. Xuất bản phải có explanation.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/questions' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "title": "Câu thử API",
  "content": "Trong ví dụ thử này, hãy chọn A.",
  "image": "",
  "explanation": "Ví dụ quy ước A là phương án đúng.",
  "license_type": "A1",
  "category_id": {{category_id}},
  "is_critical": true,
  "published": true,
  "answers": [
    {"content":"Phương án A","label":"A","order":1,"is_correct":true},
    {"content":"Phương án B","label":"B","order":2,"is_correct":false}
  ]
}'
```

**Kết quả và chú ý:** 201 object câu và answers có ID thật. Lưu `id` thành `question_id`. `published=false` hoặc bỏ trường này thì câu là nháp. `position` do server tự cấp trong chủ đề, không phải số câu toàn hệ thống. Image là URL/chuỗi, không phải upload ảnh; rỗng là không ảnh. API này không tự kiểm tra URL ảnh có tải được hay không.

### 6.11. Tìm câu — GET /api/v1/admin/questions

**Dùng khi / cần trước:** Đã login admin, không yêu cầu ngân hàng có dữ liệu.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/api/v1/admin/questions?license_type=A1&page=1&limit=20' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 phân trang, gồm cả nháp và đã xuất bản, có đáp án chuẩn. Lọc `license_type`, `category_id`, `is_critical`, `search` (title/content). Thứ tự chủ đề → position → ID, nên câu mới không nhất thiết ở trang 1. Không hỗ trợ filter `published` riêng; không tự giả định query lạ có tác dụng.

### 6.12. Xem chính xác câu vừa tạo — GET /api/v1/admin/questions/:id

**Dùng khi / cần trước:** ID lấy từ response tạo/list, không dùng tổng số câu làm ID.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/api/v1/admin/questions/{{question_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 object đầy đủ có category và answers; không tồn tại/đã xóa thì 404. Đây là cách kiểm tra câu mới nhanh nhất, không bị phân trang che khuất.

### 6.13. Sửa hoặc xuất bản câu — PUT /api/v1/admin/questions/:id

**Dùng khi / cần trước:** Câu và chủ đề tồn tại; các điều kiện như tạo câu. Gửi **toàn bộ** dữ liệu muốn giữ, không chỉ `{published:true}`.

**Quyền:** Admin.

```bash
curl --request PUT '{{base_url}}/api/v1/admin/questions/{{question_id}}' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "title": "Câu thử API",
  "content": "Trong ví dụ thử này, hãy chọn A.",
  "image": "",
  "explanation": "Ví dụ quy ước A là phương án đúng.",
  "license_type": "A1",
  "category_id": {{category_id}},
  "is_critical": true,
  "published": true,
  "answers": [
    {"content":"Phương án A","label":"A","order":1,"is_correct":true},
    {"content":"Phương án B","label":"B","order":2,"is_correct":false}
  ]
}'
```

**Kết quả và chú ý:** 200; thay cả bộ đáp án nên ID đáp án thay đổi. Lấy lại ID từ response. ID câu vẫn giữ; source_url của câu crawl được giữ. Đổi chủ đề thì cấp position trong chủ đề mới. Bỏ trường boolean sẽ thành false; bỏ image/explanation có thể làm rỗng hoặc lỗi nếu xuất bản. Câu đang được đề dùng vẫn có thể sửa; lượt thi cũ giữ snapshot, lượt mới kiểm tra ngân hàng hiện tại.

### 6.14. Xóa câu — DELETE /api/v1/admin/questions/:id

**Dùng khi / cần trước:** Câu không còn được đề chưa xóa tham chiếu. Xóa/gỡ câu khỏi tất cả đề liên quan trước.

**Quyền:** Admin.

```bash
curl --request DELETE '{{base_url}}/api/v1/admin/questions/{{question_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 204 rỗng, 400 nếu còn được đề dùng, 404 nếu không có. Không có API khôi phục xóa mềm. Lịch sử thi dùng snapshot nên không mất nội dung cũ.

### 6.15. Nhập nhiều câu — POST /api/v1/admin/questions/import

**Dùng khi / cần trước:** Có các chủ đề cần dùng. Body gồm 1–500 câu, mỗi câu theo điều kiện tạo câu; body tối đa 5 MiB.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/questions/import' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"questions": [{
  "title": "Câu thử API",
  "content": "Trong ví dụ thử này, hãy chọn A.",
  "image": "",
  "explanation": "Ví dụ quy ước A là phương án đúng.",
  "license_type": "A1",
  "category_id": {{category_id}},
  "is_critical": true,
  "published": true,
  "answers": [
    {"content":"Phương án A","label":"A","order":1,"is_correct":true},
    {"content":"Phương án B","label":"B","order":2,"is_correct":false}
  ]
}]}'
```

**Kết quả và chú ý:** 201 `{data:[...], imported:N}`. Cả lô dùng một transaction: một câu lỗi thì không ghi cả lô. Lỗi service nêu hàng (từ 1), lỗi binding có đường dẫn mảng (từ 0). Gửi lại cùng dữ liệu có thể tạo câu trùng nội dung; đây không phải cơ chế chống trùng của crawler.

### 6.16. Cào ngân hàng — POST /api/v1/admin/questions/crawl

**Dùng khi / cần trước:** Server truy cập được nguồn mặc định hoclaixeonline.com. Không gửi body, kể cả `{}`, và không gửi query/URL tùy chọn.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/questions/crawl' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 báo `received`, `created`, `skipped`, `source_changed`, `categories_created` và thông tin nguồn. Ngân hàng mặc định A1/B, dự kiến tổng 850 bản ghi theo hạng khi nguồn đúng cấu trúc; mới nhập là nháp. 850 không có nghĩa 850 nội dung độc lập. Ảnh là URL từ xa, không tải file về server. Bận 409; gọi sớm sau lần thành công 429 (chờ 60 giây); nguồn/robots/cấu trúc lỗi 502; timeout 504. Không có nhập dở dang khi lô thất bại.

Crawl lại bỏ qua bản ghi đã nhận theo nguồn + hạng + ID nguồn; không ghi đè biên tập admin, không tự xuất bản, không khôi phục câu đã xóa. Nguồn đổi nội dung được đếm `source_changed` để rà soát, không tự cập nhật câu cũ. Chống trùng này không so với câu nhập thủ công. Sau crawl: GET admin câu → rà soát → PUT đầy đủ với `published=true` → đưa ID vào đề. Tìm hai câu đúng hạng và hợp lệ là đủ để thử, không bắt xuất bản cả kho.

### 6.17. Tạo đề — POST /api/v1/admin/tests

**Dùng khi / cần trước:** Có 1–200 ID câu khác nhau, dương, tồn tại, đã xuất bản, cùng hạng của đề, bộ đáp án hợp lệ. Name/version không trống; duration_minutes 1–180; passing_score từ 1 đến số câu.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/tests' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "name":"Đề thử một câu",
  "description":"Thử luồng API, không phải cấu hình thi chính thức",
  "license_type":"A1",
  "version":"demo-v1",
  "duration_minutes":10,
  "passing_score":1,
  "critical_fail":true,
  "published":true,
  "question_ids":[{{question_id}}]
}'
```

**Kết quả và chú ý:** 201 object đề, lưu `id` thành `test_id`. `question_ids` quyết định thứ tự câu; tổng câu tự tính. `published=false` ẩn đề khỏi danh sách người học; **đề nháp vẫn yêu cầu câu đã xuất bản**. Trường version là nhãn, không tạo kho phiên bản tự động.

### 6.18. Danh sách đề cho admin — GET /api/v1/admin/tests

**Dùng khi / cần trước:** Có token admin.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/api/v1/admin/tests?page=1&limit=20&license_type=A1' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 phân trang cả nháp/xuất bản, ID giảm dần. Filter thực dùng: page, limit, license_type. Không trả danh sách nội dung câu ở endpoint list.

### 6.19. Xem cấu hình đề — GET /api/v1/admin/tests/:id

**Dùng khi / cần trước:** Đề tồn tại.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/api/v1/admin/tests/{{test_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 object cấu hình và `questions` là các dòng liên kết gồm `question_id`, `test_id`, `order`, `score`... Mỗi dòng có `question` chứa đầy đủ `title`, `content`, `image`, `explanation`, `published`, chủ đề và `answers` có `is_correct`. Đọc nội dung tại `questions[i].question.content`, không phải `questions[i].content`. Câu theo thứ tự trong đề, phương án theo order. Đây không phải bài làm của người học; 404 nếu đề không còn.

### 6.20. Sửa đề — PUT /api/v1/admin/tests/:id

**Dùng khi / cần trước:** Đề tồn tại; gửi toàn bộ cấu hình hợp lệ và danh sách câu, như lúc tạo.

**Quyền:** Admin.

```bash
curl --request PUT '{{base_url}}/api/v1/admin/tests/{{test_id}}' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "name":"Đề thử một câu",
  "description":"Thử luồng API, không phải cấu hình thi chính thức",
  "license_type":"A1",
  "version":"demo-v1",
  "duration_minutes":10,
  "passing_score":1,
  "critical_fail":true,
  "published":true,
  "question_ids":[{{question_id}}]
}'
```

**Kết quả và chú ý:** 200; thay danh sách câu và thứ tự. Muốn thử quy tắc chấm mới phải bắt đầu lượt mới; các lượt đã bắt đầu giữ cấu hình trước đó.

### 6.21. Xóa đề — DELETE /api/v1/admin/tests/:id

**Dùng khi / cần trước:** Đề tồn tại và bạn muốn gỡ khỏi hệ thống đang sử dụng.

**Quyền:** Admin.

```bash
curl --request DELETE '{{base_url}}/api/v1/admin/tests/{{test_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 204 rỗng; xóa mềm đề, gỡ các liên kết câu của đề này. Câu ngân hàng không bị xóa, lịch sử thi vẫn có snapshot. Nếu câu còn dùng ở đề khác thì vẫn chưa xóa câu được.

### 6.22. Sinh đề ngẫu nhiên — POST /api/v1/admin/tests/generate

**Dùng khi / cần trước:** Các điều kiện cấu hình đề như tạo thủ công. Mỗi quota có category_id tồn tại và không trùng quota khác, count 1–200, critical_count từ 0 đến count; tổng count 1–200. Có đủ câu published đúng hạng, đủ riêng câu liệt và thường trong từng chủ đề.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/api/v1/admin/tests/generate' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"name":"Đề ngẫu nhiên","description":"Demo","license_type":"A1","version":"v1","duration_minutes":10,"passing_score":2,"critical_fail":true,"published":true,"quotas":[{"category_id":{{category_id}},"count":2,"critical_count":1}]}'
```

**Kết quả và chú ý:** 201 một đề mới đã lưu. `count=2,critical_count=1` nghĩa một liệt + một thường, không phải ba câu. Chọn ngẫu nhiên và trộn thứ tự khi sinh; bắt đầu lại cùng đề không sinh lại. Question_ids nếu gửi sẽ bị thay bởi kết quả quota, nên bỏ trường đó. Thiếu câu trả lỗi chỉ rõ số cần và số có; không tự mượn câu chủ đề/hạng khác. Mỗi POST thành công tạo thêm đề, kể cả nội dung ngẫu nhiên trùng lần trước.

### 6.23. Chọn đề để thi — GET /api/v1/tests

**Dùng khi / cần trước:** Admin đã tạo và xuất bản đề nếu muốn danh sách có dữ liệu.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/tests?license_type=A1&page=1&limit=20' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 phân trang đề published, ID giảm dần. Không trả sẵn nội dung câu. Lấy một `id` làm `test_id`. Đề có thể xuất hiện nhưng start bị từ chối nếu câu trong đề đã bị admin ẩn/sửa sai hạng sau khi tạo.

### 6.24. Bắt đầu / thi lại — POST /api/v1/exams

**Dùng khi / cần trước:** Đề published, tồn tại, cấu hình và mọi câu còn hợp lệ. Đây là hành động tạo lượt mới, không phải resume.

**Quyền:** Người đã đăng nhập.

```bash
curl --request POST '{{base_url}}/api/v1/exams' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"test_id":{{test_id}}}'
```

**Kết quả và chú ý:** 201 object lượt, lấy `id` lưu `exam_id`. `exam.questions` chứa nội dung và ID phương án cho màn thi, chưa có đáp án đúng/lời giải. Đề ẩn/không còn trả 404; cấu hình/câu không hợp lệ trả 400. Hiện không giới hạn số lượt thi lại hay số bài đang làm.

### 6.25. Lịch sử — GET /api/v1/exams

**Dùng khi / cần trước:** Dùng tài khoản đã thi; tài khoản mới nhận danh sách rỗng.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/exams?page=1&limit=20' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 phân trang, chỉ của người có token, mới nhất trước; gồm cả DOING và bài đã chấm. Không có snapshot/đáp án trong list; chọn ID rồi GET chi tiết. Hiện chỉ phân trang, không hỗ trợ lọc theo test_id/status/license_type ở truy vấn lịch sử, dù bộ đọc query chung có thể nhận một số tên này. Đọc lịch sử cũng xử lý bài của bạn đã hết hạn.

### 6.26. Tiếp tục / xem kết quả — GET /api/v1/exams/:id

**Dùng khi / cần trước:** Lượt tồn tại và thuộc người có token.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/exams/{{exam_id}}' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 chi tiết. DOING còn giờ: làm tiếp; COMPLETED/TIMEOUT: xem kết quả và lời giải. Hết hạn khi đọc sẽ được chấm. Không tạo lượt mới, không gia hạn. 404 nếu không tồn tại hoặc thuộc người khác.

### 6.27. Chọn / đổi / bỏ chọn — PUT /api/v1/exams/:id/answers/:question_id

**Dùng khi / cần trước:** Lượt của bạn, DOING và còn hạn. Question_id thuộc snapshot; answer_id thuộc phương án của câu ấy trong snapshot.

**Quyền:** Người đã đăng nhập.

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/{{question_id}}' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":{{answer_id}}}'
```

**Kết quả và chú ý:** 200 toàn bộ lượt với lựa chọn đã lưu. Đổi phương án bằng ID khác; bỏ chọn bằng `answer_id:0`. Không gửi score/is_correct. 400 nếu sai ID/missing body; 404 nếu không có quyền sở hữu; 409 nếu bài đóng/hết giờ. GET lại bài để xem kết quả sau 409.

### 6.28. Nộp bài — POST /api/v1/exams/:id/submit

**Dùng khi / cần trước:** Lượt thuộc bạn. Chờ các lần lưu đáp án thành công trước; không cần trả lời hết.

**Quyền:** Người đã đăng nhập.

```bash
curl --request POST '{{base_url}}/api/v1/exams/{{exam_id}}/submit' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 kết quả. Không cần body, server không dùng điểm/chọn lựa client gửi trong submit. Nộp trước hạn là COMPLETED, sau hạn TIMEOUT; xem is_passed để biết đậu. Nộp lặp trả kết quả cũ, không chấm thành bài khác.

### 6.29. Lấy câu để ôn — GET /api/v1/practice/questions

**Dùng khi / cần trước:** Ngân hàng có câu published nếu muốn dữ liệu. Không cần đề/lượt thi.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/practice/questions?license_type=A1&page=1&limit=20' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 phân trang, chỉ câu published; ẩn đáp án chuẩn/lời giải. Lọc category_id, is_critical, search, license_type. Theo chủ đề/position/ID. Lấy question_id và answer_id từ dữ liệu hiện tại.

### 6.30. Trả lời một câu ôn — POST /api/v1/practice/questions/:id/answer

**Dùng khi / cần trước:** Câu hiện tồn tại và published; answer_id dương, thuộc bộ đáp án hiện tại.

**Quyền:** Người đã đăng nhập.

```bash
curl --request POST '{{base_url}}/api/v1/practice/questions/{{question_id}}/answer' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":{{answer_id}}}'
```

**Kết quả và chú ý:** 200 `{attempt,question}`. `attempt.is_correct` nói lượt này đúng/sai; `question` có lời giải và đáp án chuẩn. Mỗi POST ghi một lượt mới; không sửa lịch sử thi. ID đáp án sai thuộc tính trả 400; câu không tồn tại/nháp thì 404. Không có answer_id=0 để xóa lượt ôn.

### 6.31. Danh sách câu cần ôn lại — GET /api/v1/practice/wrong-questions

**Dùng khi / cần trước:** Có lịch sử thi đã chấm hoặc lượt ôn sai nếu muốn danh sách có câu.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/practice/wrong-questions?license_type=A1&page=1&limit=20' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 phân trang câu có sự kiện gần nhất sai/trống và hiện còn published; ID tăng dần. Hỗ trợ các filter câu như ôn thường. Đáp án chuẩn bị ẩn. Trả lời qua API 6.30, không có API answer riêng cho wrong-questions. Danh sách rỗng không phải lỗi.

### 6.32. Tiến bộ — GET /api/v1/progress

**Dùng khi / cần trước:** Có token, không cần có lịch sử.

**Quyền:** Người đã đăng nhập.

```bash
curl --request GET '{{base_url}}/api/v1/progress?license_type=A1' \
  --header 'Authorization: Bearer {{learner_token}}'
```

**Kết quả và chú ý:** 200 object thống kê như phần 5. Filter thực dùng là license_type; không phân trang thống kê. Phản ánh toàn bộ lịch sử thuộc phạm vi, không chỉ lượt mới nhất; có thể làm mới sau submit/ôn.

## 7. API cũ vẫn tồn tại — chỉ dành cho admin

Các route này vẫn được đăng ký. Khi viết app mới, dùng nhóm `/api/v1/admin/questions` để quản lý cả câu và đáp án cho dễ nhất. Giữ phần này để bạn không bỏ sót API đang có hoặc nhầm chúng là API người học. Tất cả đều yêu cầu token admin; danh sách có thể lộ đáp án chuẩn đúng theo quyền quản trị.

### 7.1. GET /question/getList

**Dùng khi / cần trước:** Có token admin.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/question/getList' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 mảng câu trực tiếp, không envelope phân trang; có category/answers và câu nháp. Không có page/limit chuẩn như v1.

### 7.2. POST /question/create

**Dùng khi / cần trước:** Chủ đề tồn tại, câu/đáp án hợp lệ như kiểm tra dữ liệu chung. DTO cũ không nhận explanation/published.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/question/create' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "title": "Câu thử API",
  "content": "Trong ví dụ thử này, hãy chọn A.",
  "image": "",
  "license_type": "A1",
  "category_id": {{category_id}},
  "is_critical": true,
  "answers": [
    {
      "content": "Phương án A",
      "label": "A",
      "order": 1,
      "is_correct": true
    },
    {
      "content": "Phương án B",
      "label": "B",
      "order": 2,
      "is_correct": false
    }
  ]
}'
```

**Kết quả và chú ý:** 201 thông báo thành công, không trả object câu/ID như v1. Câu tạo mặc định nháp; muốn xuất bản phải đọc lại và PUT bằng API v1 với explanation. Không nên dùng route này cho luồng mới cần lưu ID ngay.

### 7.3. POST /questions/:question_id/answers

**Dùng khi / cần trước:** Câu tồn tại. Sau khi thêm, cả bộ vẫn >=2 phương án, đúng một phương án đúng, không trùng nhãn/thứ tự/nội dung.

**Quyền:** Admin.

```bash
curl --request POST '{{base_url}}/questions/{{question_id}}/answers' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "content": "Phương án bổ sung C",
  "label": "C",
  "order": 3,
  "is_correct": false
}'
```

**Kết quả và chú ý:** 201 thông báo, không trả object phương án. GET lại để lấy ID. Nếu câu đã có một phương án đúng, thêm is_correct=true nữa sẽ bị từ chối.

### 7.4. GET /questions/:question_id/answers

**Dùng khi / cần trước:** Có ID câu muốn tra.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/questions/{{question_id}}/answers' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 wrapper `data` chứa danh sách đáp án theo order. Query danh sách này có thể trả mảng rỗng khi không có câu/đáp án thay vì xác nhận 404; dùng GET admin câu nếu cần xác minh câu tồn tại.

### 7.5. PUT /questions/:question_id/answers/:answer_id

**Dùng khi / cần trước:** Câu và phương án thuộc câu tồn tại. Gửi đủ nội dung/nhãn/order/is_correct muốn giữ. Bộ đáp án sau sửa vẫn hợp lệ.

**Quyền:** Admin.

```bash
curl --request PUT '{{base_url}}/questions/{{question_id}}/answers/{{answer_id}}' \
  --header 'Authorization: Bearer {{admin_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "content": "Phương án bổ sung C",
  "label": "C",
  "order": 3,
  "is_correct": false
}'
```

**Kết quả và chú ý:** 200 thông báo; GET lại để xem. Muốn đổi đáp án đúng từ A sang B thì nên PUT cả câu bằng v1: sửa riêng A thành sai sẽ tạm không còn đáp án đúng, sửa riêng B thành đúng sẽ tạm có hai đáp án đúng, cả hai đều bị từ chối.

### 7.6. DELETE /questions/:question_id/answers/:answer_id

**Dùng khi / cần trước:** Sau xóa vẫn ít nhất hai phương án và đúng một phương án đúng.

**Quyền:** Admin.

```bash
curl --request DELETE '{{base_url}}/questions/{{question_id}}/answers/{{answer_id}}' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 thông báo (khác DELETE v1 trả 204). Không xóa phương án đúng duy nhất hoặc làm còn một phương án. Xóa mềm, không có route khôi phục.

### 7.7. GET /answer/list_answer

**Dùng khi / cần trước:** Có token admin.

**Quyền:** Admin.

```bash
curl --request GET '{{base_url}}/answer/list_answer' \
  --header 'Authorization: Bearer {{admin_token}}'
```

**Kết quả và chú ý:** 200 wrapper `data` chứa tất cả đáp án còn hoạt động theo ID, không phân trang. Với ngân hàng lớn nên đọc theo câu thay vì dùng danh sách này cho màn học.

## 8. Quy tắc dùng chung và cách xử lý lỗi ở frontend

### Phân trang và filter

Danh sách v1 mặc định page=1, limit=20. Limit lớn hơn 100 được giảm về 100. Page/limit phải là số nguyên dương khi gửi; page tối đa 1.000.000. Không có kết quả hoặc vượt trang thì data rỗng, total vẫn biểu thị tổng theo filter. `total` không phải ID cuối.

Filter câu `search` tìm title/content, không tìm theo ID. Biến query gửi nhưng endpoint không dùng sẽ không tự có hiệu lực; tra đúng phần kết quả mỗi API. API GET detail không cần page/limit. Không có query sắp “mới nhất trước” cho danh sách câu hiện tại.

### Mã HTTP

| Mã | Hiểu như người dùng | Xử lý |
|---|---|---|
| 200 | Đọc/cập nhật/xử lý thành công | Đọc đúng cấu trúc response |
| 201 | Tạo mới thành công | Lưu ID nếu response có; legacy có thể chỉ thông báo |
| 204 | Xóa thành công, không nội dung | Không gọi parse JSON trên body rỗng |
| 400 | Dữ liệu hoặc điều kiện nghiệp vụ sai | Hiện message và đánh dấu field nếu có errors |
| 401 | Chưa đăng nhập/token không hợp lệ/hết hạn hoặc sai thông tin login | Refresh/login phù hợp, không thử lại vô hạn |
| 403 | Không đủ quyền admin | Kiểm tra tài khoản/role, không chỉ login lặp |
| 404 | Tài nguyên không có, đã xóa, đang ẩn trong luồng học hoặc không thuộc bạn | Kiểm tra ID, token, trạng thái xuất bản |
| 409 | Xung đột: email trùng, lượt thi đóng, crawler đang chạy | Đọc message; mỗi API có cách xử lý khác nhau |
| 429 | Crawl trong thời gian chờ | Đợi theo Retry-After |
| 500 | Lỗi phía server | Ghi nhận request và kiểm tra log, không coi là nhập sai thông thường |
| 502/504 | Crawl gặp lỗi nguồn/mạng/timeout | Thử lại khi nguồn ổn; có thể dùng câu nhập tay để tiếp tục thử |

Lỗi binding mới có `message` và `errors` gồm `field`, `rule`, `message`; ví dụ field `answers[0].content`. Lỗi nghiệp vụ thường chỉ có `message`, ví dụ câu 55 chưa xuất bản. Frontend không giả định mọi lỗi đều có errors, và không nên phụ thuộc toàn bộ câu chữ message để điều khiển nghiệp vụ. Một số handler legacy vẫn có cách trả lỗi riêng, không phải tất cả response đã thống nhất.

```json
{
  "status_code":400,
  "message":"password must be at least 8 characters",
  "errors":[{"field":"password","rule":"min","message":"password must be at least 8 characters"}]
}
```

### Gửi lại request khi mạng chập chờn

GET thường có thể gọi lại; GET lượt thi/lịch sử còn có thể hoàn tất bài hết giờ. PUT cùng lựa chọn cập nhật một lựa chọn, nhưng phải kiểm tra còn giờ và thứ tự gửi. Submit lặp trả kết quả đã chấm. POST bắt đầu/ôn/tạo đề/tạo câu/import có thể tạo thêm dữ liệu khi gọi lặp. Crawl chống trùng theo nguồn, không đồng nghĩa mọi POST đều chống trùng. DELETE gọi lại sau khi xóa có thể 404; không tự coi đó là lần xóa đầu thất bại.

## 9. Những chức năng chưa có — đừng tìm một API chưa được viết

- Chưa có API logout/thu hồi mọi phiên đăng nhập, quên/đổi mật khẩu, chỉnh hồ sơ hoặc tự cấp admin. “Logout” ở app hiện là bỏ token phía client, không tự thu hồi token trên server.
- Chưa có pause/gia hạn/hủy/xóa lượt thi, xóa lịch sử ôn, sửa điểm thủ công, bảng xếp hạng hoặc dashboard xem bài của mọi người. Admin dùng route lịch sử vẫn chỉ thấy bài của chính admin.
- Chưa có API retry riêng: thi lại dùng POST start cùng test_id. Không có giới hạn lượt thi hoặc bắt buộc chỉ một lượt đang làm.
- Chưa có lưu nhiều đáp án trong một request, nhiều lựa chọn đúng mỗi câu, điểm trọng số hoặc tự sinh đề ngẫu nhiên cho learner.
- Chưa có API PATCH chỉ published, xuất bản hàng loạt, upload ảnh, tải ảnh nguồn về máy chủ, quản lý nguồn crawl tùy ý hoặc phục hồi câu đã xóa.
- Chưa có GET chi tiết câu ôn riêng trước trả lời hay GET chi tiết đề đầy đủ dành cho learner trước bắt đầu. Dùng list câu ôn và response start theo luồng đã mô tả.
- Chưa có GET chi tiết chủ đề riêng; đọc danh sách chủ đề. Không có filter lịch sử theo status/test_id và không có sort mới nhất cho danh sách câu.
- Chưa có scheduler tự crawl hoặc worker liên tục đóng bài hết giờ. Deadline vẫn được áp dụng khi các API liên quan xử lý.
- Chuẩn response còn khác giữa auth, v1 và legacy. Đây là thực trạng code, không nên viết một hàm luôn lấy `response.data.id` cho mọi API.

## 10. Checklist cho dev trước khi nối UI

- Đăng ký và phân quyền admin chạy được với đúng file DB; token admin/learner không nhầm.
- Tạo một câu, GET lại bằng ID; xác nhận đủ phương án và đúng một đáp án đúng.
- Tạo đề published đúng hạng; learner thấy đề trong danh sách.
- Start nhận exam_id mới; lưu ID phiên riêng khỏi test_id.
- Chọn, đổi, bỏ chọn rồi GET lại thấy đúng lựa chọn đã lưu.
- Submit sau khi lưu xong; xem đúng điểm và lý do trượt.
- Thi lại cùng đề có exam_id khác; lịch sử cũ không bị mất.
- Đóng/mở app tiếp tục đúng phiên; hết giờ không sửa được.
- Ôn từng câu nhận lời giải; ôn đúng câu sai cập nhật danh sách nhưng không sửa điểm cũ.
- Kiểm tra token hết hạn, sở hữu bài, submit lặp, mất response, bank thiếu quota và dữ liệu bị admin sửa.
- Không gọi endpoint xóa chỉ vì đang làm theo ví dụ; đó là thao tác dọn dữ liệu thật.

## 11. Nguồn code để đối chiếu khi thay đổi backend

- [Router tổng](../internal/server/router.go)
- [Auth](../internal/auth/delivery/router.go)
- [Ngân hàng, đề, luyện tập](../internal/learning/delivery/router.go)
- [Validation và tạo đề](../internal/learning/service.go)
- [Lượt thi: route](../internal/exam/delivery/router.go), [logic chấm/hết giờ](../internal/exam/usecase/usecase.go)
- [Thống kê và câu sai](../internal/learning/progress.go)
- [Crawler](../internal/crawler/router.go)
- [Route câu cũ](../internal/question/delivery/router.go), [route đáp án cũ](../internal/answer/delivery/router.go)
- [Giải thích validation](VALIDATION_ERRORS.md), [ERD](ERD.md)

Tài liệu bao quát các route hiện đăng ký và các nhánh nghiệp vụ được nêu trong code đã đối chiếu. Khi thêm API hoặc thay đổi logic, cập nhật tài liệu này cùng thay đổi đó; không xem đây là cam kết mọi tình huống hạ tầng hoặc lỗi chưa phát hiện đều đã được xử lý.
