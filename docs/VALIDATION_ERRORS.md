# Cách đọc lỗi validation API

Các request không hợp lệ trả HTTP **400**. Thông báo giữ tiếng Anh như API hiện có; tên trường khớp JSON để bạn tìm trực tiếp trong Postman. Các ví dụ ID bên dưới chỉ minh họa.

## 1. Thiếu trường, sai định dạng hoặc giới hạn

Áp dụng bộ xử lý JSON chung cho auth, câu hỏi/đáp án cũ và các API dùng `httpapi.Bind` (admin, luyện tập, thi).

Ví dụ đăng ký với password ngắn hơn 8 ký tự:

```json
{
  "status_code": 400,
  "message": "password must be at least 8 characters",
  "errors": [
    {
      "field": "password",
      "rule": "min",
      "message": "password must be at least 8 characters"
    }
  ]
}
```

- `errors` liệt kê các lỗi do bộ kiểm tra DTO phát hiện trong lượt bind; `message` ghép các thông báo để đọc trực tiếp.
- Mảng dùng chỉ số từ **0**: `answers[1].content` là nội dung phương án thứ hai; import có thể là `questions[0].answers[1].content`.
- `rule`: `required`, `email`, `min`, `max`, `type`, `json`, `max_size`.
- Số có giới hạn giá trị; chuỗi có giới hạn ký tự; mảng có giới hạn số phần tử.
- Sai JSON trả vị trí byte nếu bộ giải mã cung cấp; thiếu body được báo riêng. Giới hạn body là 5 MiB.
- Sai kiểu cần sửa kiểu JSON: `category_id` gửi số `12`, không phải chuỗi `"12"`. Bộ giải mã có thể chỉ trả đường dẫn trường mà không có chỉ số mảng cho lỗi kiểu dữ liệu.
- Không đưa giá trị mật khẩu/token đã gửi vào lỗi.

## 2. Lỗi quy tắc nghiệp vụ khi tạo đề

Các lỗi nghiệp vụ tiếp tục dùng `message`, trả lỗi đầu tiên theo thứ tự kiểm tra. Không phải mọi lỗi 400 đều có mảng `errors`; frontend nên dùng `errors` nếu có và hiển thị `message` nếu không có.

### Câu chưa xuất bản

```json
{
  "message": "invalid request: question_ids[0]: question 55 is unpublished; set published=true before adding it to a test"
}
```

Rà soát rồi cập nhật câu 55 với `published: true` qua API quản trị; GET lại câu trước khi tạo đề. PUT câu hỏi vẫn yêu cầu body đầy đủ.

### Sai hạng bằng lái

```json
{
  "message": "invalid request: question_ids[1]: question 56 has license_type=B; test requires A1"
}
```

Thay câu thứ hai bằng câu A1, hoặc sửa hạng của đề và chọn toàn bộ câu tương ứng. Không đổi hạng câu chỉ để vượt validation nếu nội dung không phù hợp.

### Điểm đạt vượt số câu

```json
{
  "message": "invalid request: passing_score must be between 1 and 2 (number of questions); received 21"
}
```

Đề hai câu chỉ nhận điểm đạt 1 hoặc 2. Không dùng nguyên cấu hình 21 điểm của đề lớn.

### Những lỗi khác

| Tình huống | Thông tin trả về |
|---|---|
| ID câu bằng 0 | Vị trí `question_ids[i]` và yêu cầu số nguyên dương |
| ID không tồn tại/đã xóa | Vị trí mảng và ID câu |
| Trùng câu trong đề | Vị trí bị trùng và ID câu |
| Câu có bộ đáp án không hợp lệ | Vị trí, ID câu và lỗi đáp án |
| `name`, `version` chỉ chứa khoảng trắng | Tên trường không được để trống |
| Hạng không hỗ trợ | Danh sách A1, A, B, C1, C, D |
| Thời gian không hợp lệ | `duration_minutes` phải trong 1–180 |

ID tài nguyên con không hợp lệ trong body tạo đề (ví dụ câu không tồn tại) nay trả **400**, thay vì 404 chung trước đây. GET một tài nguyên không tồn tại vẫn giữ cách xử lý 404.

## 3. Câu hỏi, đáp án và đề ngẫu nhiên

- Phương án thiếu nội dung, nhãn/thứ tự sai hoặc trùng: chỉ rõ `answers[i].content`, `.label`, `.order`.
- Sai số đáp án đúng: `answers must have exactly one is_correct=true; received 2`.
- Chủ đề không tồn tại: chỉ rõ `category_id` và ID chủ đề.
- Quota sai: chỉ rõ `quotas[i].category_id`, `.count` hoặc `.critical_count`.
- Ngân hàng không đủ câu để sinh đề: trả chủ đề, hạng, loại điểm liệt, số cần và số hiện có.

Khi lưu đáp án bài thi, lỗi tách rõ câu không thuộc phiên hoặc phương án không thuộc câu trong snapshot. Dùng ID từ response phiên thi; đáp án ngân hàng có thể đã được thay sau khi sửa câu.

## 4. Phạm vi và kiểm thử

Giữ nguyên các điều kiện nghiệp vụ, quyền truy cập và response thành công. Thông báo 500 vẫn chung để không lộ SQL hoặc chi tiết hệ thống; đăng nhập sai vẫn báo email hoặc mật khẩu không hợp lệ, không tiết lộ tài khoản có tồn tại hay không.

Test nằm ở `internal/common/httpapi/http_test.go` và `internal/server/integration_test.go`, kiểm tra đường dẫn lồng nhau, JSON lỗi, auth, tạo đề không hợp lệ không ghi dữ liệu, câu hỏi và quota. Test dùng SQLite tạm, không ghi database dự án.
