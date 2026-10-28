# Work follow — Theo dõi phát triển backend

## Mục tiêu đợt này

Hoàn thiện vòng học/thi trong backend hiện có, bổ sung phân quyền và quản trị nội dung, kèm comment giải thích quyết định và tài liệu API.

## Đã triển khai — 23/10/2026

- [x] Quyền `learner`/`admin`; khóa API ghi nội dung và API cũ chứa khóa đáp án.
- [x] CLI cấp/thu hồi quyền, không cho client tự nâng quyền qua đăng ký.
- [x] Sửa profile để trả đúng ID và role.
- [x] Quản lý chủ đề; tạo/sửa/xóa câu hỏi, lời giải và trạng thái nháp/xuất bản.
- [x] Import JSON nguyên tử, báo dòng lỗi, giới hạn 500 câu và 5 MiB/request.
- [x] Quy tắc đáp án dùng chung; bảo vệ cả API thêm/sửa/xóa đáp án cũ.
- [x] ID trong DTO; chuẩn hóa khóa JSON cho câu hỏi/đáp án/đề.
- [x] Lọc/tìm kiếm/phân trang ngân hàng câu hỏi.
- [x] Quản lý đề cố định; cấu hình hạng bằng, phiên bản, thời gian, điểm đạt, điểm liệt.
- [x] Sinh đề theo quota chủ đề, số câu điểm liệt; không lặp, báo thiếu dữ liệu.
- [x] Bắt đầu phiên thi, lưu/bỏ chọn/tiếp tục bài, thời hạn phía server.
- [x] Snapshot đề và đáp án, không lộ đáp án trong phiên đang làm.
- [x] Chấm điểm, lỗi điểm liệt, timeout, nộp lại, giới hạn chủ sở hữu.
- [x] Lịch sử và xem lại bài với lời giải.
- [x] Ôn theo chủ đề, hạng bằng, câu điểm liệt; lưu từng lượt luyện.
- [x] Ôn câu sai/bỏ trống; tự bỏ khỏi danh sách khi lần làm gần nhất đúng.
- [x] Tiến bộ theo chủ đề, tỷ lệ đỗ, 20 lượt thi gần nhất, gợi ý chủ đề yếu.
- [x] Đăng ký các module vào router/server thật.
- [x] Migration các model/bảng mới, cấu hình DB qua môi trường.
- [x] README, tài liệu API và ví dụ request.
- [x] Integration test cho các quy tắc quan trọng bằng SQLite tạm.

## Quyết định và lý do

| Quyết định | Lý do | Nơi xem |
|---|---|---|
| Snapshot JSON cho mỗi phiên | Giữ nguyên kết quả và nội dung khi admin sửa ngân hàng | `internal/exam/model/snapshot.go`, `internal/exam/usecase/usecase.go` |
| Chấm điểm và thời gian ở server | Không tin điểm, đáp án đúng hoặc đồng hồ client gửi | `internal/exam/usecase/usecase.go` |
| Không cho đọc bài người khác | Scope `user_id` ngay trong query | `internal/exam/usecase/usecase.go` |
| Transaction + unique session/question | Lưu lại không trùng câu trả lời, nộp không sửa nửa chừng | `internal/exam/model/user_answer.go` |
| Một SQLite connection | Tuần tự hóa ghi của backend hiện tại, tránh nâng mức song song khi chưa có khóa phù hợp | `internal/database/database.go` |
| Thay cả tập đáp án trong PUT question | Đổi đáp án đúng nguyên tử; không có giai đoạn 0/2 đáp án đúng | `internal/learning/service.go` |
| Xóa mềm câu hỏi/đáp án, snapshot giữ nội dung cũ | Không phá quan hệ phiên đã lưu; nhãn/thứ tự đáp án cũ chuyển sang giá trị lưu trữ để tái sử dụng | `internal/learning/service.go`, `internal/answer/repository/repository.go` |
| Quyền admin lấy từ DB mỗi request | Thu hồi quyền không cần chờ JWT hết hạn | `internal/middleware/admin.go` |
| Câu hỏi mới/cũ mặc định nháp | Chỉ nội dung đã được kiểm tra mới xuất hiện cho người học | `internal/question/model/question.go` |
| Quota điểm liệt theo từng chủ đề | Sinh đề đúng số lượng được cấu hình, không ngẫu nhiên sai cấu trúc | `internal/learning/service.go` |
| Theo lần làm gần nhất cho ôn sai | Câu đã sửa đúng không tiếp tục nằm trong danh sách cần ôn | `internal/learning/progress.go` |

## Kiểm chứng

Bộ test trong `internal/server/integration_test.go` chạy trên router và database thật nhưng ở thư mục tạm:

- Phân quyền route mới/cũ, thu hồi quyền với token còn hạn.
- Đăng ký không nhận role admin, mật khẩu được hash, profile có ID, refresh token cũ không tái sử dụng.
- Không lộ khóa đáp án; không đọc/lưu/nộp bài của người khác.
- Không nhận đáp án của câu khác; snapshot vẫn chấm đúng sau khi sửa nội dung/đáp án gốc.
- Sai/bỏ trống điểm liệt, tắt quy tắc điểm liệt, đúng thời điểm hết hạn.
- Lưu lại, bỏ chọn, tiếp tục bài, nộp lại và nộp đồng thời.
- Import rollback, không phá quy tắc một đáp án đúng, tái sử dụng nhãn đáp án đã xóa.
- Không sinh đề thiếu câu; bản nháp, sai hạng bằng, trùng câu hoặc điểm đạt không hợp lệ bị chặn.
- Lịch sử tự xử lý hết giờ; ôn sai và tiến bộ chỉ thuộc người hiện tại.

Migration đã được thử trên bản sao SQLite hiện có ở `/tmp`, không cập nhật file dữ liệu dự án. File `cmd/data/gorm.db` đã có thay đổi cục bộ từ trước đợt làm việc này và được giữ nguyên.

Kết quả kiểm tra cuối:

- `GOCACHE=/tmp/driving-go-build-cache go test ./...`: đạt.
- `GOCACHE=/tmp/driving-go-build-cache go test -race ./...`: đạt, gồm 10 integration test; không phát hiện data race.
- `GOCACHE=/tmp/driving-go-build-cache go vet ./...`: đạt; đã dọn DTO xóa đáp án cũ có trường JSON không export.
- `git diff --check`: đạt.

## Giới hạn/phần vận hành tiếp theo

- Phạm vi đợt này là **backend**; repo chưa có frontend để triển khai màn hình.
- Chưa tự nhập ngân hàng câu hỏi hay xác nhận quy định thi chính thức. Cần chủ dự án chuẩn bị nội dung và cấu hình theo tài liệu đã kiểm chứng.
- Dữ liệu hiện có cần rà soát chủ đề/đáp án trước khi xuất bản; migration không tự suy đoán các quan hệ thiếu.
- Timeout được hoàn tất khi có request liên quan; thêm worker nếu cần chuyển trạng thái ngay cả khi không có request.
- Thống kê đọc lịch sử để tổng hợp; thêm bảng tổng hợp khi dữ liệu lớn.
- AutoMigrate chưa có migration rollback theo phiên bản; sao lưu trước nâng cấp.
- Chưa bổ sung upload ảnh, reset mật khẩu qua email, CORS theo domain triển khai hoặc hệ thống gửi thông báo. Đây là các tích hợp ngoài luồng học/thi của đợt này.

## Bổ sung tài liệu — 24/10/2026

- Thêm `docs/HUONG_DAN_API_TUNG_BUOC.md`: giải thích luồng quản trị, thi thử, ôn tập và tiến bộ bằng một ví dụ xuyên suốt.
- Có sơ đồ, cách lấy token/ID, request/response minh họa, dữ liệu thay đổi sau mỗi bước, cách nối API với màn hình và xử lý lỗi.
- Đối chiếu tài liệu với router, DTO và logic chấm điểm hiện có; không thay đổi hành vi backend trong đợt bổ sung này.

## API cào dữ liệu — 25/10/2026

- [x] Tìm và kiểm tra nguồn công khai: HocLaixeOnline, Dữ Liệu Pháp Luật; liên kết nguồn nhà nước để đối chiếu.
- [x] `POST /api/v1/admin/questions/crawl`: chỉ token admin, không body/query.
- [x] Mặc định đọc HTML `/hang/a1` và `/hang/b`, tương ứng 250/600 câu; không chạy JavaScript.
- [x] Giữ dấu điểm liệt theo từng hạng (nguồn hiện có 20/60), kiểm tra nội dung/đáp án chung giữa hai bộ.
- [x] Tự tạo chủ đề, nhập câu/đáp án/lời giải/URL ảnh dưới dạng nháp, lưu `source_url`.
- [x] `crawl_records` chống trùng theo nguồn + hạng + ID nguồn; không ghi đè chỉnh sửa/xuất bản hoặc hồi sinh câu đã xóa.
- [x] Đọc robots, giới hạn host/redirect, dung lượng/thời gian, chặn chạy đồng thời và cooldown 60 giây sau thành công.
- [x] Kiểm tra nguồn đầy đủ trước khi ghi; transaction toàn lô, báo nguồn thay đổi mà không ghi đè.
- [x] Test offline và smoke test Internet thật trên SQLite tạm: nhập 850 bản ghi, 6 chủ đề; lần hai tạo 0, bỏ qua 850.
- [x] Thêm `docs/CRAWL_DU_LIEU.md`, cập nhật tài liệu API, ví dụ HTTP và README.

Phạm vi: chỉ adapter HocLaixeOnline A1/B; ảnh là URL từ xa, chưa tải file ảnh; câu mới cần được rà soát/xuất bản qua API quản trị hiện có. 850 bản ghi là theo hạng, không phải 850 nội dung độc lập. Chống trùng không so sánh với câu đã nhập thủ công không có nguồn.

Trong đợt này không chạy import vào database dự án; các lần kiểm thử dùng database tạm.

Kiểm tra cuối cho crawler: `go test -race ./...` và `go vet ./...` đạt (dùng `GOCACHE=/tmp/driving-go-build-cache` trong môi trường giới hạn ghi cache). Smoke test `CRAWLER_LIVE=1` gọi nguồn thật đạt; JSON mẫu, liên kết file trong tài liệu và `git diff --check` đạt.

## Kiểm tra domain ảnh — 25/10/2026

- Kiểm tra toàn bộ trường `questions.image` trong database hiện tại: 445 bản ghi dùng `hoclaixeonline.com`, 3 bản ghi dùng `example.com`.
- Ba bản ghi ID 24, 25, 26 đều có URL mẫu `https://example.com/sign.png`, nội dung chung “Người lái xe phải làm gì khi gặp biển báo này?” và không có `source_url`. Không đủ thông tin xác định ảnh biển báo tương ứng; chưa thay bằng ảnh phỏng đoán.
- Kiểm tra trực tiếp `https://hoclaixeonline.com/cau-hoi/cau_36.webp`: HTTP 200, Content-Type `image/webp`; đối chiếu trang câu 36 của nguồn.
- Crawler đã dùng domain thật `https://hoclaixeonline.com` để chuẩn hóa đường dẫn ảnh. Không phát hiện cấu hình `example.com` trong đường đi nhập dữ liệu thật; các URL example trong test dùng kiểm tra từ chối host ngoài nguồn.
- Không thay đổi database trong lần kiểm tra này. Để sửa chính xác ba câu mẫu, cần ảnh gốc hoặc số câu/link nguồn tương ứng. Việc chỉ thay domain nhưng giữ `/sign.png` không xác định được ảnh đúng.

## Review model và schema — 25/10/2026

- Đối chiếu 11 bảng được migrate với model: không phát hiện cột scalar cần thiết bị thiếu. `Question.Answers` là association, lưu tại `answers.question_id`, không phải cột `questions.answers`.
- Thêm comment giải thích quan hệ trong Question/Answer và tài liệu `docs/MODEL_VA_DATABASE.md` có sơ đồ, SQL kiểm tra và ví dụ từ dữ liệu hiện tại.
- Phát hiện câu mẫu 24/26 không có đáp án, câu 25 có hai đáp án đúng; cả ba tham chiếu chủ đề không tồn tại. Không tự sửa/xóa dữ liệu chưa xác định nội dung đúng.
- Ghi nhận để cải thiện: FK cho PracticeAttempt/CrawlRecord, unique email. Đợt review không đổi schema hoặc hành vi backend.

## Cập nhật hướng dẫn thực hành Postman — 25/10/2026

- Viết lại `docs/HUONG_DAN_API_TUNG_BUOC.md` theo router/DTO/usecase hiện tại, với luồng 0–15 từ chạy server tới tạo đề hai câu và xem kết quả thi.
- Thêm 38 cURL theo từng tiêu đề thao tác API và 17 script Post-response tự lưu token/ID, kiểm tra trạng thái và kết quả.
- Tách rõ token admin/người học, đề/phiên thi và command terminal/cURL Postman.
- Nhánh dùng crawler có GET chọn câu, dựng đầy đủ body PUT, rà soát/xuất bản và lấy lại ID đáp án sau cập nhật trước khi tạo đề.
- Kiểm tra cú pháp các cURL, JSON sau thay biến, JavaScript bằng `node --check`, các liên kết nội bộ và `git diff --check`. Chưa chạy các request hướng dẫn lên database dự án; đợt này chỉ sửa tài liệu.

## Bổ sung ERD — 25/10/2026

- Tạo `docs/ERD.md` và mã Mermaid riêng `docs/ERD.mmd` cho đủ 11 bảng, cột và quan hệ.
- Đối chiếu model với SQLite chỉ đọc; phân biệt FK thật với liên kết logic của practice_attempts/crawl_records.
- Giải thích snapshot JSON, bảng trung gian, unique index và tên cột khác JSON API. Không thay đổi code hoặc database.

## Chi tiết hóa lỗi validation — 25/10/2026

- Dùng chung bộ bind JSON có giới hạn 5 MiB cho auth, question/answer cũ và API học/thi; trả `errors` gồm đường dẫn JSON, rule và thông báo, giữ `message` để đọc trên Postman.
- Tách lỗi cấu hình đề, ID câu bằng 0/trùng/không tồn tại, câu nháp, sai hạng; nêu vị trí và ID cụ thể. ID câu không tồn tại trong body tạo đề trả 400 thay cho 404 chung.
- Tách lỗi nội dung/nhãn/thứ tự đáp án; nêu số đáp án đúng thực tế. Quota chỉ rõ trường sai, thiếu ngân hàng nêu số cần/số có; xác minh chủ đề tồn tại.
- Kiểm tra cấu hình đề ngẫu nhiên trước truy vấn ngân hàng; giữ validation ở service để import/direct callers cũng được kiểm tra.
- Tách lỗi câu và phương án không thuộc phiên thi; bổ sung comment giải thích, `docs/VALIDATION_ERRORS.md` và liên kết trong hướng dẫn thực hành.
- Test mới: JSON sai kiểu/cú pháp/body rỗng, đường dẫn nested/embedded, auth không echo password, lỗi tạo đề không ghi dữ liệu, câu hỏi/đáp án/quota. Các test liên quan đạt; `go vet ./...` và `git diff --check` đạt.
- `go test ./...`: lần đầu test crawler có sẵn `TestConcurrentCrawlAndCancellation` lỗi `context cancellation lost: <nil>`, lần chạy lại đạt. Ghi nhận tính không ổn định của test cancellation (timer 0 và context canceled có thể cùng sẵn sàng); không sửa crawler trong thay đổi validation này. Sau bổ sung cuối, chạy lại các package bị ảnh hưởng đều đạt.
- Chỉ test trên SQLite tạm; không sửa database dự án.

## Tutorial cho dev — 25/10/2026

- Tạo mới `docs/tutorial_dev.md`, không sửa file hướng dẫn từng bước trong đợt này.
- Giải thích theo màn hình và tình huống: bắt đầu, tiếp tục, thi lại, đổi/bỏ chọn, nộp lặp, hết giờ, mất mạng, snapshot, ôn câu sai, tiến bộ và quản trị ngân hàng.
- Mỗi API có điều kiện tiên quyết, quyền, cURL, kết quả và trường hợp cần xử lý; phân biệt API chính và legacy, ghi rõ chức năng chưa có.
- Khởi tạo router với SQLite in-memory để đối chiếu: đủ 39 method/route thực tế, không thiếu hoặc thừa endpoint. Kiểm tra cú pháp 39 cURL, JSON body sau thay biến và liên kết tài liệu đạt. Không chạy request ghi lên DB dự án.

## Tạo đề mẫu A1 — 25/10/2026

- Theo yêu cầu tạo đề mẫu, đã thêm đề ID 2, chủ đề demo ID 8 và câu 854–858 vào `cmd/data/gorm.db` qua learning service trong transaction. Đề 5 câu, 10 phút, đạt 4/5, bật điểm liệt; câu thứ năm là điểm liệt.
- Sao chép nội dung/đáp án/lời giải từ ngân hàng A1 hiện có, giữ source_url; giữ nguyên các câu nguồn. Đề và câu bản sao đã xuất bản, ghi rõ demo, không phải cấu hình thi chính thức.
- Sao lưu trước thao tác tại `tmp/db-backups/gorm-before-sample-20261025-230046.db`.
- Thêm `docs/DE_THI_MAU_A1.md` với cURL dùng ID thật và `docs/de_thi_mau_a1.json` chứa dữ liệu/đáp án dành cho dev.
- Đối chiếu dữ liệu đã lưu; smoke test start/save/submit trên bản sao DB cho ba trường hợp đậu 5/5, trượt liệt 4/5 và thiếu điểm 3/5. Không tạo lịch sử thi thử vào DB chính.

## Sửa GET chi tiết đề trả câu rỗng — 26/10/2026

- Response người dùng gửi có question_id thật nhưng question.id=0, title/content rỗng, answers=null vì chỉ preload bảng nối Questions.
- GET `/api/v1/admin/tests/:id` nay preload cả `Questions.Question.Category` và `Questions.Question.Answers`, giữ thứ tự câu trong đề và order của phương án; không đổi cấu trúc JSON bên ngoài.
- Thêm comment giải thích và cập nhật tutorial_dev.md/mục 20 hướng dẫn từng bước về đường dẫn đọc dữ liệu `questions[i].question`.
- Test tích hợp xác minh nội dung, chủ đề, đáp án đúng, thứ tự, 401/403 và 404; dùng SQLite tạm, không thay đổi DB dự án.

## Hướng dẫn riêng tạo đề, tiếp tục thi và thi lại

- Tạo mới `docs/TAO_DE_TIEP_TUC_THI_VA_THI_LAI.md` theo yêu cầu xuất Markdown, giữ nguyên các tài liệu trước.
- Trình bày kiểu hỏi/đáp, tạo đề demo hai câu từ đầu, phân biệt test_id/exam_id, tiếp tục bằng GET và thi lại bằng POST; thêm điều kiện, tình huống lỗi, sơ đồ và cURL Postman.
- Đối chiếu route/DTO và kiểm tra cú pháp cURL, body JSON, response ví dụ, liên kết nội bộ. Chỉ viết tài liệu, không gọi API ghi dữ liệu hay tạo thêm đề trong DB.
