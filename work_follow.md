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
