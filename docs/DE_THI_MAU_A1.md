# Đề mẫu A1 đã tạo trong database

Đề demo dùng để thử chức năng, không phải cấu hình kỳ thi chính thức. Nội dung/đáp án/lời giải sao chép từ ngân hàng đã crawl đang có trong DB; lần tạo này không xác minh lại nguồn pháp luật. Ngân hàng gốc giữ nguyên.

- `test_id`: **2**.
- `category_id`: **8**.
- 5 câu A1, 10 phút, đạt từ 4 câu đúng; sai hoặc bỏ trống câu điểm liệt thì trượt.
- Đề và cả 5 câu đã xuất bản.
- [JSON đầy đủ dành cho dev/admin](de_thi_mau_a1.json) có đáp án chuẩn. Đây là bản xuất dữ liệu, không phải body import trực tiếp. Không dùng file này làm nguồn câu hỏi trên màn thi người học.

## 1. Xem đề bằng tài khoản người học

```bash
curl --request GET '{{base_url}}/api/v1/tests?license_type=A1' \
  --header 'Authorization: Bearer {{learner_token}}'
```

## 2. Bắt đầu thi

```bash
curl --request POST '{{base_url}}/api/v1/exams' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"test_id":2}'
```

Lấy `id` của response đặt vào biến `exam_id` trong Postman. Mỗi lần gọi bắt đầu là một lượt mới. Khi thi lại, gọi lại chính request này rồi lưu exam_id mới.

## 3. Lưu lựa chọn từng câu

Các request sau chọn đúng cả 5 câu, dùng để kiểm tra đường đi đậu. Đợi mỗi request thành công trước khi nộp. Dùng ID trong response lượt thi nếu câu đã được cập nhật sau lúc viết tài liệu.

### Câu 1 — ID 854

Chọn B (ID 2656).

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/854' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":2656}'
```

### Câu 2 — ID 855

Chọn B (ID 2659).

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/855' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":2659}'
```

### Câu 3 — ID 856

Chọn A (ID 2661).

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/856' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":2661}'
```

### Câu 4 — ID 857

Chọn B (ID 2665).

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/857' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":2665}'
```

### Câu 5 — ID 858 — điểm liệt

Chọn B (ID 2668).

```bash
curl --request PUT '{{base_url}}/api/v1/exams/{{exam_id}}/answers/858' \
  --header 'Authorization: Bearer {{learner_token}}' \
  --header 'Content-Type: application/json' \
  --data-raw '{"answer_id":2668}'
```

## 4. Nộp và xem kết quả

```bash
curl --request POST '{{base_url}}/api/v1/exams/{{exam_id}}/submit' \
  --header 'Authorization: Bearer {{learner_token}}'
```

Nếu lưu đúng cả 5 câu trước hạn, mong đợi `score=5`, `correct_answers=5`, `is_passed=true`.

```bash
curl --request GET '{{base_url}}/api/v1/exams/{{exam_id}}' \
  --header 'Authorization: Bearer {{learner_token}}'
```

## 5. Thử các kết quả khác

Mỗi tình huống bắt đầu một lượt mới:

- **Đậu sát điểm:** đúng câu 2–5, để trống câu 1 → 4/5 và đậu.
- **Đủ điểm nhưng trượt liệt:** đúng câu 1–4, để trống câu 5 → 4/5, trượt với `critical_question`.
- **Thiếu điểm:** đúng câu 3–5, để trống câu 1–2 → 3/5, trượt với `insufficient_score`.
- **Hết giờ:** lưu đáp án, chờ hết 10 phút rồi GET bài; `TIMEOUT`, chấm các lựa chọn đã lưu. Không cần chờ để thử các case còn lại.
- **Ôn câu sai:** sau bài trượt gọi GET `/api/v1/practice/wrong-questions?license_type=A1`, trả lời qua API ôn tập. Ôn đúng không đổi điểm bài thi cũ.

Không dùng token admin cho các bước thi nếu bạn muốn lịch sử nằm ở tài khoản người học. Tài liệu [tutorial_dev.md](tutorial_dev.md) giải thích đầy đủ cách dùng các API.
