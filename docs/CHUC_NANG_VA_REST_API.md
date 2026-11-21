# Chức năng hiện có và REST API tương ứng

Danh mục **59 endpoint** đối chiếu với router. Base URL `http://localhost:8080`.

Công khai không cần access token; Đăng nhập dùng `Authorization: Bearer <access_token>`; Admin cần thêm role admin trong DB. `:id` là tham số cần thay bằng ID thực. Body JSON dùng `Content-Type: application/json`.

Hướng dẫn mobile và ví dụ request/response: [MOBILE_API.md](MOBILE_API.md).

## Tài khoản và hệ thống

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| POST | `/auth/register` | Công khai | Đăng ký bằng name, email, password; luôn tạo learner, không cho tự cấp admin. |
| POST | `/auth/login` | Công khai | Đăng nhập bằng email/password; trả access_token, refresh_token và role trong data. |
| POST | `/auth/refresh` | Công khai | Gửi refresh_token để lấy cặp token mới; refresh token cũ bị thu hồi. |
| GET | `/health` | Công khai | Kiểm tra HTTP server; không kiểm tra toàn bộ trạng thái database. |
| GET | `/auth/getProfile` | Đăng nhập | Lấy ID, tên, email, role và current_license_type của người hiện tại. |

## Chủ đề

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/categories` | Đăng nhập | Danh sách chủ đề có type; lọc category_type/search. Learner chỉ nhận chủ đề có học liệu xuất bản phù hợp license_type (mặc định hạng đã lưu). Admin thấy cả chủ đề rỗng. |
| POST | `/api/v1/admin/categories` | Admin | Tạo chủ đề: name, description, type; mặc định UNCATEGORIZED. |
| PUT | `/api/v1/admin/categories/:id` | Admin | Sửa tên/mô tả/type; không bỏ SIGN khi còn bản ghi biển báo. |
| DELETE | `/api/v1/admin/categories/:id` | Admin | Xóa mềm chủ đề; từ chối nếu còn câu hỏi. |

## Ngân hàng câu hỏi

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/admin/questions` | Admin | Cả nháp/xuất bản và khóa đáp án; lọc license_type/category_type/category_id/road_sign_id/sign_type/published/is_critical/search, phân trang. |
| GET | `/api/v1/admin/questions/:id` | Admin | Chi tiết câu hỏi, chủ đề, phương án và đáp án đúng. |
| POST | `/api/v1/admin/questions` | Admin | Tạo câu kèm đáp án; road_sign_id tùy chọn, phải khớp hạng/chủ đề. Ít nhất hai đáp án, đúng một đáp án đúng. |
| PUT | `/api/v1/admin/questions/:id` | Admin | Thay toàn bộ nội dung và tập đáp án. Đặt published để xuất bản/thu hồi; xuất bản cần lời giải. ID đáp án có thể đổi. |
| DELETE | `/api/v1/admin/questions/:id` | Admin | Xóa mềm câu hỏi; từ chối nếu đang thuộc một đề chưa xóa. |
| POST | `/api/v1/admin/questions/publish` | Admin | Body {question_ids:[...]}, 1–500 ID duy nhất. Xuất bản câu hợp lệ, trả published_ids và failed theo từng ID. Giữ nguyên nội dung và ID đáp án; câu lỗi không thay đổi. |
| POST | `/api/v1/admin/questions/import` | Admin | Body {questions:[...]}; 1–500 câu, tối đa 5 MiB, lỗi một dòng hoàn tác toàn bộ. Nhập lại có thể tạo trùng. |
| POST | `/api/v1/admin/questions/crawl` | Admin | Không body/query. Cào A1/B từ nguồn cố định, tạo chủ đề và câu nháp; chống trùng theo nguồn/hạng; cooldown 60 giây sau thành công. |

## Quản lý đề thi

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/admin/tests` | Admin | Danh sách đề nháp và xuất bản; lọc license_type, phân trang page/limit. |
| GET | `/api/v1/admin/tests/:id` | Admin | Chi tiết cấu hình và câu hỏi: questions[i].question, gồm chủ đề và đáp án đúng. |
| POST | `/api/v1/admin/tests` | Admin | Tạo đề cố định; category_type tùy chọn để giới hạn chủ đề. 1–200 câu đã xuất bản cùng hạng. |
| PUT | `/api/v1/admin/tests/:id` | Admin | Thay cấu hình và danh sách câu; published điều khiển xuất bản, critical_fail bật/tắt quy tắc điểm liệt. |
| DELETE | `/api/v1/admin/tests/:id` | Admin | Xóa mềm đề; giữ snapshot và lịch sử của phiên thi đã tạo. |
| POST | `/api/v1/admin/tests/generate` | Admin | Sinh và lưu đề bằng quotas: category_id, count, critical_count; không lặp câu, báo lỗi khi thiếu dữ liệu. |

## Thi thử và lịch sử

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/tests` | Đăng nhập | Đề cố định đã xuất bản; lọc hạng (mặc định đã lưu), category_type, search và phân trang. |
| POST | `/api/v1/exams` | Đăng nhập | Body {test_id}; tạo phiên thi mới và snapshot. Thi lại cũng gọi API này, nhận exam_id mới. |
| GET | `/api/v1/exams` | Đăng nhập | Lịch sử của chính user; lọc license_type/category_type/status, phân trang. Phiên random có test_id=null và template_id. |
| GET | `/api/v1/exams/:id` | Đăng nhập | Lấy bài đang làm để tiếp tục, hoặc xem kết quả/lời giải khi kết thúc. :id là exam_id, không phải test_id. |
| PUT | `/api/v1/exams/:id/answers/:question_id` | Đăng nhập | Body {answer_id}; lưu/thay lựa chọn, answer_id=0 để bỏ chọn. Chỉ chủ phiên được sửa trước khi hết giờ. |
| POST | `/api/v1/exams/:id/submit` | Đăng nhập | Nộp và chấm ở server; xử lý điểm đạt, điểm liệt và hết giờ; nộp lại không tạo kết quả mới. |

## Ôn tập và tiến bộ

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/practice/questions` | Đăng nhập | Câu đã xuất bản, không lộ đáp án; lọc hạng (mặc định đã lưu), category_type/category_id/road_sign_id/sign_type/is_critical/search, phân trang. |
| POST | `/api/v1/practice/questions/:id/answer` | Đăng nhập | Body {answer_id}; lưu lượt luyện, trả đúng/sai cùng câu hỏi và lời giải. |
| GET | `/api/v1/practice/wrong-questions` | Đăng nhập | Câu sai/bỏ trống theo lần làm gần nhất; câu đã sửa đúng được loại ra; có bộ lọc và phân trang. |
| GET | `/api/v1/progress` | Đăng nhập | Tiến bộ riêng theo license_type (mặc định hạng đã lưu) và category_type; tỷ lệ đỗ, chủ đề và 20 lượt thi gần nhất. |

## API legacy — tương thích hệ thống cũ

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/question/getList` | Admin | API danh sách câu hỏi cũ; ưu tiên GET /api/v1/admin/questions. |
| POST | `/question/create` | Admin | API tạo câu hỏi cũ; giao diện mới dùng POST /api/v1/admin/questions để tạo cùng đáp án. |
| POST | `/questions/:question_id/answers` | Admin | Thêm phương án; phải giữ quy tắc đáp án hợp lệ. |
| GET | `/questions/:question_id/answers` | Admin | Danh sách phương án của một câu, có khóa đáp án đúng. |
| PUT | `/questions/:question_id/answers/:answer_id` | Admin | Sửa một phương án; bị chặn nếu làm câu có 0 hoặc 2 đáp án đúng. Đổi đáp án đúng nên PUT toàn câu hỏi. |
| DELETE | `/questions/:question_id/answers/:answer_id` | Admin | Xóa phương án; phải còn ít nhất hai phương án và đúng một đáp án đúng. |
| GET | `/answer/list_answer` | Admin | Danh sách đáp án theo API cũ; chỉ admin được truy cập. |

## Mobile: hạng bằng, random và biển báo

| Method | Endpoint | Quyền | Chú thích |
|---|---|---|---|
| GET | `/api/v1/me` | Đăng nhập | Hồ sơ hiện tại, gồm current_license_type. |
| PUT | `/api/v1/me/license` | Đăng nhập | Body {license_type}; lưu hạng bằng mặc định cho người trong JWT, không thay phiên thi đang làm. |
| GET | `/api/v1/category-types` | Đăng nhập | Mã loại chủ đề và tên hiển thị cho tab mobile. |
| GET | `/api/v1/road-sign-types` | Đăng nhập | Mã nhóm biển báo và tên hiển thị. |
| POST | `/api/v1/exams/random` | Đăng nhập | Body {} hoặc {license_type,template_id,category_type}; tạo phiên riêng từ template đã xuất bản, test_id=null. Thiếu câu báo lỗi, không tạo đề công khai. |
| GET | `/api/v1/exam-templates` | Đăng nhập | Cấu hình random đã xuất bản; lọc license_type/category_type, phân trang; mặc định hạng đã lưu. |
| GET | `/api/v1/road-signs` | Đăng nhập | Học liệu biển đã xuất bản; lọc hạng, nhóm biển, chủ đề, tìm kiếm và phân trang. |
| GET | `/api/v1/road-signs/:id` | Đăng nhập | Chi tiết biển đã xuất bản theo hạng đã lưu hoặc license_type explicit. |
| GET | `/api/v1/road-signs/:id/questions` | Đăng nhập | Câu luyện đã xuất bản liên kết với biển, không trả đáp án đúng; dùng API practice để trả lời. |
| GET | `/api/v1/admin/road-signs` | Admin | Danh sách biển báo, gồm nháp; lọc/phân trang. |
| GET | `/api/v1/admin/road-signs/:id` | Admin | Chi tiết biển báo. |
| POST | `/api/v1/admin/road-signs` | Admin | Tạo biển báo; xem MOBILE_API.md về body và quy tắc. |
| PUT | `/api/v1/admin/road-signs/:id` | Admin | Cập nhật toàn bộ biển báo. |
| DELETE | `/api/v1/admin/road-signs/:id` | Admin | Xóa mềm biển báo; biển có câu tham chiếu bị chặn, template không ảnh hưởng snapshot cũ. |
| GET | `/api/v1/admin/exam-templates` | Admin | Danh sách cấu hình đề ngẫu nhiên, gồm nháp; lọc/phân trang. |
| GET | `/api/v1/admin/exam-templates/:id` | Admin | Chi tiết cấu hình đề ngẫu nhiên. |
| POST | `/api/v1/admin/exam-templates` | Admin | Tạo cấu hình đề ngẫu nhiên; xem MOBILE_API.md về body và quy tắc. |
| PUT | `/api/v1/admin/exam-templates/:id` | Admin | Cập nhật toàn bộ cấu hình đề ngẫu nhiên. |
| DELETE | `/api/v1/admin/exam-templates/:id` | Admin | Xóa mềm cấu hình đề ngẫu nhiên; biển có câu tham chiếu bị chặn, template không ảnh hưởng snapshot cũ. |

## Quy tắc và chức năng không có endpoint riêng

- Tiếp tục thi dùng GET phiên; thi lại dùng POST tạo phiên mới.
- Chấm điểm, kiểm tra thời hạn, snapshot và chống đọc bài người khác nằm trong luồng thi, không có API cho client gửi điểm.
- Timeout được xử lý khi có request liên quan, chưa có worker tự đóng toàn bộ phiên.
- Xuất bản nhiều câu dùng POST `/api/v1/admin/questions/publish`. Xuất bản/thu hồi từng câu hoặc đề cũng có thể dùng PUT với đầy đủ body và trường `published`.
- Cấp/thu hồi admin dùng `go run ./cmd/admin -email ... -role admin|learner`, chưa có API quản lý quyền.
- Chưa có API upload ảnh, sửa hồ sơ, đổi/quên mật khẩu, logout/thu hồi phiên hoặc quản trị người dùng.
- Đăng xuất ở giao diện local chỉ xóa token trong tab; không thu hồi token đã cấp ở server.
- Phân trang mới trả `{data,total,page,limit}`, mặc định 20, tối đa 100 bản ghi/trang.
- Mã lỗi thường gặp: 400 dữ liệu không hợp lệ, 401 thiếu/hết hạn token, 403 thiếu quyền, 404 không tìm thấy, 409 xung đột, 429 cooldown crawler.

## Giao diện local mới

`GET /admin` chuyển đến `GET /admin/`. Trang HTML và các asset `app.js`, `style.css`, `api.json` là tài nguyên công khai, không chứa dữ liệu hay khóa bí mật. Dữ liệu quản trị vẫn đi qua API có JWT/admin middleware.

Xem [hướng dẫn chạy trang quản trị](ADMIN_LOCAL.md), [body và ví dụ API](API.md), [lộ trình và tiến độ](../work_follow.md).
