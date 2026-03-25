# Learning English — Kế hoạch dự án toàn diện

> React + TypeScript + Vite · Go API · PostgreSQL · Markdown-driven content

## 1. Mục tiêu dự án

### 1.1. Mục tiêu sản phẩm

Xây dựng một website học tiếng Anh với 4 module chính:

- Auth: đăng nhập bằng username/password, quản lý session an toàn
- Vocabulary: học từ vựng bằng flashcards, trắc nghiệm, viết đáp án, tùy biến chế độ học
- Grammar: nhập nội dung ngữ pháp từ Markdown, bài tập theo chủ đề, highlight từ khóa
- Reading: đọc song ngữ Anh - Việt, hỗ trợ nội dung dài dạng docs, highlight, import Markdown/text

### 1.2. Mở rộng mục tiêu theo hướng sản phẩm thực tế

#### Learner experience

- vào học nhanh, ít friction
- có thể tiếp tục bài đang học dở
- nội dung grammar/reading dài vẫn mượt trên mobile
- phản hồi đúng/sai rõ ràng
- không mất input hoặc mất tiến độ khi gặp lỗi mạng ngắn

#### Content operations

- có luồng import, preview, validate, commit, publish cho nội dung
- tách learner APIs khỏi admin content APIs
- import lỗi phải truy được nguyên nhân và record lỗi
- không phụ thuộc sửa DB thủ công để đưa nội dung lên hệ thống

#### Business outcome

- rút ngắn thời gian từ "có nội dung" đến "publish"
- tăng khả năng ship MVP chạy thật thay vì chỉ dừng ở demo
- giữ cửa mở cho spaced repetition, audio, analytics, CMS sau này

### 1.3. Tech stack chính

#### Frontend

- React + TypeScript + Vite
- React Router
- TanStack Query cho server state
- React Hook Form cho forms
- Zustand hoặc Context cho UI state cục bộ
- `react-markdown` + `remark-gfm` cho grammar/reading
- Tailwind CSS là lựa chọn ưu tiên để đi nhanh; nếu team muốn strict boundaries hơn, có thể thay bằng CSS Modules
- animation chỉ dùng ở nơi thật sự tạo giá trị, ví dụ flip card hoặc transition nhỏ

#### Backend

- Go
- HTTP router: giữ đúng stack gốc thì dùng Gorilla/mux; nếu greenfield hoàn toàn có thể thay bằng router nhẹ tương đương
- GORM
- Swagger/OpenAPI
- clean layered architecture: handler -> service -> repository
- bổ sung `policies`, `queries`, `jobs` như các package phụ trợ, nhưng vẫn giữ layered architecture làm trục chính

#### Database

- PostgreSQL

#### Storage

- S3-compatible object storage cho media upload khi ảnh/audio trở thành first-class feature
- giai đoạn sớm có thể chấp nhận external URLs cho ảnh từ vựng nếu muốn giảm scope hạ tầng

#### Tooling / Delivery

- Docker Compose cho local development
- SQL migrations
- Makefile cho common tasks
- CI cho lint, test, build
- CD hoặc script deploy chuẩn hóa cho staging/prod

## 2. Mục tiêu kỹ thuật

### 2.1. Functional goals

- người dùng đăng nhập và học theo tài khoản riêng
- theo dõi tiến độ học từ vựng theo từng từ
- import nội dung học từ file Markdown/text/CSV/JSON theo module
- hỗ trợ UI học tập tối ưu cho mobile và desktop
- dễ mở rộng thêm:
  - spaced repetition
  - audio pronunciation
  - admin CMS
  - analytics học tập

### 2.2. Non-functional goals

- API rõ ràng, có versioning
- codebase dễ bảo trì, tách layer rõ ràng
- hỗ trợ test tốt ở cả frontend và backend
- query database hiệu quả cho bài học lớn
- UX mượt khi chuyển tab, lật thẻ, làm quiz, đọc dài
- auth an toàn, dễ revoke session
- import nội dung có preview/validation trước khi publish
- hệ thống có logging, monitoring, health checks
- dev setup đủ chuẩn để nhiều người cùng làm mà ít lệch môi trường

### 2.3. Delivery principles

- giữ learner APIs tối giản, rõ ràng, ổn định
- các thao tác CRUD/import/publish nằm sau admin APIs thay vì trộn vào learner flows
- ưu tiên ship một "content-to-learning loop" hoàn chỉnh hơn là nhiều feature rời rạc
- chỉ chọn các tối ưu thật sự tăng xác suất triển khai, test, deploy và vận hành thành công
- learner read path chỉ đọc published snapshot, không đọc draft hoặc mutable state đang chỉnh sửa

## 3. Kiến trúc hệ thống

### 3.1. Kiến trúc tổng thể

```text
[ React App ]
   |
   | HTTPS / JSON
   v
[ Go API Server ]
   - router
   - auth middleware
   - handlers
   - services
   - repositories
   - import service
   - job dispatcher
   - progress tracking
   - admin content ops
   - asset storage abstraction
   - health/readiness endpoints
   |
   +--------------------------+
   |                          |
   v                          v
[ PostgreSQL ]         [ Object Storage / CDN ]
   - users                  - word images
   - sessions               - imported media
   - vocabulary             - audio assets (future)
   - grammar
   - reading
   - imports
   - content revisions
   - publications
   - content audit logs
   - jobs
   - job attempts
   - auth events
```

### 3.2. Kiến trúc backend đề xuất

```text
learning-english/
├── frontend/
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── vite.config.ts
├── backend/
│   ├── cmd/api/main.go
│   ├── internal/
│   ├── migrations/
│   ├── docs/
│   ├── go.mod
│   └── go.sum
├── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

```text
/internal
  /config
  /database
  /middleware
  /models
  /repositories
  /queries
  /policies
  /services
  /handlers
  /dto
  /utils
  /importers
  /jobs
  /storage
  /router
  /swagger
/cmd/api
/migrations
/docs
```

#### Trách nhiệm từng layer

- handlers: nhận request/response HTTP, mapping DTO, translate error
- services: business logic, orchestration, transaction boundary
- repositories: truy cập DB qua GORM
- queries: read-model/query tối ưu cho dashboard, reading docs, admin history
- policies: authorization rules, publish rules, validation gates
- models: entity DB
- dto: request/response schema
- middleware: auth, logging, recovery, CORS, request ID
- importers: parse Markdown/text/CSV thành dữ liệu có cấu trúc
- jobs: background workers cho commit/publish/unpublish, lease-based execution và retry
- storage: upload, signed URL, asset metadata
- router: phân vùng public/private/admin routes

#### Khuyến nghị local/dev

- PostgreSQL chạy bằng Docker Compose
- Swagger UI bật ở dev và staging
- migrations chạy qua `make migrate-up` / `make migrate-down`
- seed data riêng cho vocabulary, grammar, reading để QA test nhanh

### 3.3. Kiến trúc frontend đề xuất

```text
/src
  /app
  /routes
  /pages
  /components
  /features
    /auth
    /vocabulary
    /grammar
    /reading
    /admin
  /services
  /hooks
  /store
  /types
  /utils
  /styles
```

#### Gợi ý thư viện frontend

- React Router: routing
- Axios hoặc fetch wrapper: gọi API
- TanStack Query: cache server state
- Zustand hoặc Context: UI/session state
- React Hook Form: forms
- `react-markdown`: render Markdown
- animation library: chỉ dùng cho flashcard và transition có chủ đích
- icon/toast libs là utility tùy chọn, không phải kiến trúc cốt lõi

#### Nguyên tắc frontend

- server state tách khỏi UI state
- có resume session và optimistic UX vừa đủ
- reading mobile ưu tiên toggle/sync theo section; swipe là mode phụ chứ không ép buộc
- tách learner screens và admin screens từ đầu để route tree không bị lẫn
- không làm pixel-perfect sync scroll nếu chi phí cao; sync theo section là đủ cho MVP

## 4. Kiến trúc module

### 4.1. Auth module

#### Scope

- login bằng username/password
- logout current device
- logout all devices
- lấy thông tin user hiện tại
- refresh access token

#### Cơ chế session khuyến nghị

- access token ngắn hạn lưu in-memory ở frontend
- refresh token dài hạn lưu trong HttpOnly Secure SameSite cookie
- refresh token lưu hash trong DB để revoke được
- mỗi refresh thành công sẽ rotate refresh token

#### Auth transport hybrid

- `POST /auth/login`: trả về access token và set refresh cookie
- `POST /auth/refresh`: đọc refresh cookie, rotate token, trả access token mới
- `POST /auth/logout`: revoke session hiện tại và clear cookie
- `POST /auth/logout-all`: revoke tất cả session của user
- `GET /auth/me`: lấy user hiện tại

#### Khuyến nghị

- password hash bằng bcrypt
- không lưu access token trong `localStorage` hoặc `sessionStorage`
- nếu refresh flow chạy cross-site thì bật `withCredentials`, siết CORS và bổ sung CSRF protection cho refresh/logout
- nếu deployment cùng site thì giữ cookie strategy đơn giản hơn, ít rủi ro hơn

### 4.2. Vocabulary module

#### Chức năng chính

- danh sách sách từ vựng
- chọn sách và mở phiên học
- 2 tab chính:
  - đoán nghĩa
  - học từng từ
- 2 study mode chính:
  - flip
  - write
- tùy biến:
  - random / sequential / repeat
  - en-vi / vi-en
  - bật/tắt IPA
  - lọc trạng thái: new / learning / mastered
  - auto-next on/off
- tiếp tục phiên học gần nhất
- quick resume từ dashboard

#### Mô hình logic nên tách

- Vocabulary Book
- Vocabulary Word
- Word Media / Examples
- User Progress per Word
- Study Session
- Quiz Attempt

#### Refinement cho logic học

- `random`: dùng strategy riêng có test
- `sequential`: theo `sort_order` ổn định
- `repeat`: ưu tiên từ sai gần đây, từ có streak thấp, hoặc từ bị tụt khỏi mastered
- distractor cho MCQ nên ưu tiên cùng part of speech hoặc cùng độ khó để đáp án nhiễu hợp lý hơn
- write mode nên normalize input:
  - trim
  - lowercase
  - tùy chọn bỏ punctuation
  - hỗ trợ strict/lenient mode
- mỗi session nên snapshot item/question đã phát ra để retry, resume và analytics deterministic

#### Learner UX refinement

- đúng: hiển thị feedback rõ và auto-next với delay ngắn nếu user bật
- sai: giữ câu hiện tại, hiện đáp án đúng và ví dụ
- hiển thị progress bar theo session
- có session summary cuối phiên:
  - total answered
  - accuracy
  - words moved to learning/mastered

### 4.3. Grammar module

#### Chức năng chính

- import bài ngữ pháp từ Markdown
- tổ chức theo chủ đề
- render nội dung bài học
- highlight từ khóa
- bài tập theo lesson/topic

#### Mô hình nội dung

Markdown nên parse thành:

- metadata: `title`, `slug`, `topic`, `level`
- sections
- blocks:
  - heading
  - paragraph
  - list
  - example
  - note
- keywords canonical list

#### Refinement

- import theo 2 bước: preview -> commit; publish là action riêng
- frontmatter validation bắt buộc
- lesson có `status` rõ ràng: `draft`, `published`, `archived`
- exercises tách khỏi lesson content để có thể reuse theo topic
- live history đi qua publication records; learner chỉ đọc active publication

#### Highlight strategy

- phase đầu: keyword-based highlight, escape ký tự đặc biệt trước khi match
- chỉ highlight theo word boundary khi phù hợp để tránh tô sai chuỗi con
- không highlight bên trong code block hoặc inline code
- phase sau mới cân nhắc range-based highlight nếu thật sự cần độ chính xác cao

### 4.4. Reading module

#### Chức năng chính

- import Markdown hoặc plain text
- hiển thị song ngữ Anh - Việt
- hỗ trợ docs mode cho nội dung dài
- highlight từ khóa
- lưu progress theo section

#### Mô hình hiển thị

Reading content phải lưu theo sections để:

- render docs dài hiệu quả
- lazy load được
- đồng bộ EN/VI theo đoạn
- highlight theo đoạn dễ hơn
- resume theo section chính xác hơn

#### Reading UX refinement

- Desktop:
  - 2 cột EN/VI song song
  - sync scroll theo section
  - sticky toolbar
- Mobile:
  - mặc định 1 ngôn ngữ + toggle EN/VI
  - swipe là mode phụ, không phải mode bắt buộc
- Toolbar:
  - mode switcher
  - language toggle
  - highlight on/off
  - font size
  - resume from last section

#### Nguyên tắc quan trọng

- không render toàn bộ bài dài thành một blob nếu document lớn
- không theo đuổi pixel-perfect sync scroll ở MVP
- section pairing lỗi phải block publish
- learner chỉ đọc published projection thuộc active publication, không đụng draft/source state
- highlights và progress phải được gắn với publication context để rollback/publish mới không làm lệch dữ liệu

### 4.5. Admin / Content Operations module

#### Chức năng chính

- import grammar Markdown
- import reading Markdown/plain text
- import vocabulary CSV/JSON
- preview parse result trước khi commit
- xem validation errors
- xem diff giữa bản live và revision sắp publish
- yêu cầu change summary khi publish/rollback
- publish/unpublish nội dung
- xem import history
- xem publication history và audit log

#### Mục tiêu

- learner APIs ổn định và gọn
- admin workflows không làm bẩn learner domain
- content team có thể vận hành mà không chạm DB trực tiếp
- publication model rõ ràng để rollback, audit và cache không bị nhập nhằng draft/live

## 5. Database schema

### 5.1. Danh sách bảng tổng thể

#### Auth / User

- users
- user_sessions

#### Vocabulary

- vocab_books
- vocab_words
- vocab_word_examples
- vocab_word_media
- user_vocab_progress
- vocab_study_sessions
- vocab_session_items
- vocab_quiz_attempts

#### Grammar

- grammar_topics
- grammar_lessons
- grammar_exercises
- grammar_exercise_options
- grammar_attempts

#### Reading

- reading_documents
- reading_sections
- reading_highlights
- reading_progress

#### Common / Import / Ops

- imports
- import_errors
- content_revisions
- publications
- content_audit_logs
- jobs
- job_attempts
- idempotency_keys
- auth_events

### 5.2. Database schema chi tiết

#### 5.2.1. `users`

```text
id                  uuid pk
username            varchar unique not null
password_hash       varchar not null
display_name        varchar
role                varchar not null default 'learner'   -- learner/admin
status              varchar not null default 'active'
created_at          timestamptz
updated_at          timestamptz
last_login_at       timestamptz
```

Notes:

- unique index cho username
- không lưu password plain text
- thêm `role` ngay từ đầu để tách learner/admin APIs sạch hơn

#### 5.2.2. `user_sessions`

```text
id                  uuid pk
user_id             uuid fk -> users.id
refresh_token_hash  varchar not null
token_family        varchar
user_agent          text
ip_address          inet
expires_at          timestamptz
revoked_at          timestamptz null
last_used_at        timestamptz null
replaced_by_session_id uuid null fk -> user_sessions.id
created_at          timestamptz
updated_at          timestamptz
```

Purpose:

- quản lý refresh token
- logout từng thiết bị hoặc toàn bộ
- phục vụ refresh token rotation
- hỗ trợ refresh token reuse detection như một phần chính thức của session model

#### 5.2.3. `vocab_books`

```text
id                  uuid pk
title               varchar not null
slug                varchar unique not null
description         text
cover_image_url     text
level               varchar
language_from       varchar default 'en'
language_to         varchar default 'vi'
status              varchar not null default 'draft'   -- draft/published/archived
latest_import_id    uuid null fk -> imports.id
draft_revision_id   uuid null fk -> content_revisions.id
active_publication_id uuid null fk -> publications.id
word_count          integer default 0
sort_order          integer default 0
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

#### 5.2.4. `vocab_words`

```text
id                  uuid pk
book_id             uuid fk -> vocab_books.id
word                varchar not null
normalized_word     varchar not null
ipa                 varchar
part_of_speech      varchar
meaning_vi          text
meaning_en          text
context             text
tips                text
image_url           text
status_default      varchar default 'new'
sort_order          integer default 0
created_at          timestamptz
updated_at          timestamptz
```

Indexes:

- `(book_id, sort_order)`
- `unique (book_id, normalized_word)`

#### 5.2.5. `vocab_word_examples`

```text
id                  uuid pk
word_id             uuid fk -> vocab_words.id
example_en          text not null
example_vi          text
sort_order          integer default 0
created_at          timestamptz
```

Rule:

- mỗi từ nên có ít nhất 2 examples

#### 5.2.6. `vocab_word_media`

```text
id                  uuid pk
word_id             uuid fk -> vocab_words.id
media_type          varchar not null   -- image/audio
media_url           text not null
sort_order          integer default 0
created_at          timestamptz
```

Notes:

- phase đầu có thể chỉ dùng image
- để sẵn khả năng mở rộng audio pronunciation

#### 5.2.7. `user_vocab_progress`

```text
id                  uuid pk
user_id             uuid fk -> users.id
word_id             uuid fk -> vocab_words.id
learning_status     varchar not null default 'new'   -- new/learning/mastered
last_studied_at     timestamptz
correct_count       integer default 0
wrong_count         integer default 0
streak_count        integer default 0
next_review_at      timestamptz null
is_favorite         boolean default false
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

Unique:

- `unique (user_id, word_id)`

Purpose:

- filter trạng thái
- hỗ trợ repetition/spaced repetition sau này

#### 5.2.8. `vocab_study_sessions`

```text
id                  uuid pk
user_id             uuid fk -> users.id
book_id             uuid fk -> vocab_books.id
mode                varchar not null      -- guess_meaning/learn_word
study_mode          varchar not null      -- flip/write
direction           varchar not null      -- en_vi/vi_en
ordering            varchar not null      -- random/sequential/repeat
show_ipa            boolean default true
status_filter       varchar               -- all/new/learning/mastered
auto_next           boolean default true
current_question_seq integer default 0
started_at          timestamptz
ended_at            timestamptz null
total_items         integer default 0
completed_items     integer default 0
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

#### 5.2.9. `vocab_session_items`

```text
id                  uuid pk
session_id          uuid fk -> vocab_study_sessions.id
word_id             uuid fk -> vocab_words.id
question_seq        integer not null
question_type       varchar not null      -- mcq/write/flip
prompt_language     varchar not null      -- en/vi
prompt_text         text not null
options_json        jsonb null
correct_answer      text not null
status              varchar not null default 'pending'   -- pending/answered/skipped
created_at          timestamptz
```

Unique:

- `unique (session_id, question_seq)`

#### 5.2.10. `vocab_quiz_attempts`

```text
id                  uuid pk
session_id          uuid fk -> vocab_study_sessions.id
session_item_id     uuid fk -> vocab_session_items.id
user_id             uuid fk -> users.id
word_id             uuid fk -> vocab_words.id
question_type       varchar not null      -- mcq/write/flip
prompt_language     varchar not null      -- en/vi
answer_text         text
selected_option_index integer null
is_correct          boolean
answered_at         timestamptz
request_idempotency_key varchar null
```

Unique:

- `unique (session_id, session_item_id)`

#### 5.2.11. `grammar_topics`

```text
id                  uuid pk
title               varchar not null
slug                varchar unique not null
description         text
level               varchar
sort_order          integer default 0
created_at          timestamptz
updated_at          timestamptz
```

#### 5.2.12. `grammar_lessons`

```text
id                  uuid pk
topic_id            uuid fk -> grammar_topics.id
title               varchar not null
slug                varchar unique not null
summary             text
status              varchar not null default 'draft'   -- draft/published/archived
latest_import_id    uuid null fk -> imports.id
draft_revision_id   uuid null fk -> content_revisions.id
active_publication_id uuid null fk -> publications.id
sort_order          integer default 0
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

#### 5.2.13. `grammar_exercises`

```text
id                  uuid pk
lesson_id           uuid fk -> grammar_lessons.id
topic_id            uuid null fk -> grammar_topics.id
exercise_type       varchar not null      -- mcq_single/fill_blank/matching
question_text       text not null
explanation         text
difficulty          varchar
sort_order          integer default 0
created_at          timestamptz
updated_at          timestamptz
```

Rules / invariants:

- nếu `lesson_id` khác null thì `topic_id` phải null hoặc phải khớp với `grammar_lessons.topic_id`
- nếu cần exercise reusable ở cấp topic, cho phép `lesson_id = null` và `topic_id != null`

#### 5.2.14. `grammar_exercise_options`

```text
id                  uuid pk
exercise_id         uuid fk -> grammar_exercises.id
option_text         text not null
is_correct          boolean default false
sort_order          integer default 0
created_at          timestamptz
```

Rules / indexes:

- với `exercise_type = mcq_single`, chỉ cho phép tối đa 1 option `is_correct = true`
- khuyến nghị triển khai bằng partial unique index hoặc trigger, không chỉ dựa vào service validation

#### 5.2.15. `grammar_attempts`

```text
id                  uuid pk
user_id             uuid fk -> users.id
exercise_id         uuid fk -> grammar_exercises.id
answer_text         text
selected_option_id  uuid null fk -> grammar_exercise_options.id
is_correct          boolean
answered_at         timestamptz
created_at          timestamptz
```

#### 5.2.16. `reading_documents`

```text
id                  uuid pk
title               varchar not null
slug                varchar unique not null
description         text
is_bilingual        boolean default true
default_display_mode varchar default 'docs'   -- default hint cho viewer
level               varchar
status              varchar not null default 'draft'   -- draft/published/archived
latest_import_id    uuid null fk -> imports.id
draft_revision_id   uuid null fk -> content_revisions.id
active_publication_id uuid null fk -> publications.id
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

#### 5.2.17. `reading_sections`

```text
id                  uuid pk
publication_id      uuid fk -> publications.id
section_order       integer default 0
heading             varchar
content_en          text
content_vi          text
keywords_json       jsonb
blocks_json         jsonb not null
content_hash        varchar null
created_at          timestamptz
updated_at          timestamptz
```

Lý do tách section:

- đọc dài hiệu quả hơn
- render song ngữ theo đoạn
- hỗ trợ highlight từng section
- dễ sync scroll hoặc resume

Unique:

- `unique (publication_id, section_order)`

#### 5.2.18. `reading_highlights`

```text
id                  uuid pk
publication_id      uuid fk -> publications.id
section_order       integer not null
keyword             varchar not null
color               varchar default 'yellow'
start_index         integer null
end_index           integer null
created_at          timestamptz
```

Notes:

- có 2 hướng:
  - highlight theo keyword text
  - highlight theo index range
- khuyến nghị phase đầu:
  - dùng keyword text
  - nâng cấp sang range-based nếu sau này có yêu cầu chính xác cao

#### 5.2.19. `reading_progress`

```text
id                  uuid pk
user_id             uuid fk -> users.id
document_id         uuid fk -> reading_documents.id
last_section_id     uuid null fk -> reading_sections.id
last_section_order  integer default 0
publication_id      uuid null fk -> publications.id
completion_percent  numeric(5,2) default 0
last_read_at        timestamptz
created_at          timestamptz
updated_at          timestamptz
version             integer not null default 1
```

Unique:

- `unique (user_id, document_id)`

#### 5.2.20. `imports`

```text
id                  uuid pk
module_name         varchar not null      -- grammar/reading/vocabulary
source_type         varchar not null      -- markdown/text/csv/json
source_name         varchar
source_checksum     varchar
source_artifact_url text null
source_inline_text  text null
status              varchar not null      -- uploaded/validated/validation_failed/commit_queued/committed/commit_failed/cancelled
total_records       integer default 0
success_records     integer default 0
failed_records      integer default 0
preview_payload     jsonb
validation_report   jsonb
preview_artifact_url text null
validation_artifact_url text null
validated_at        timestamptz null
started_at          timestamptz
finished_at         timestamptz null
created_by          uuid fk -> users.id
created_at          timestamptz
```

Purpose:

- tránh import trùng ngoài ý muốn
- preview trước commit
- lưu validation report để debug nhanh
- source của import phải rõ ràng: file artifact hoặc inline text
- payload/report lớn nên chuyển sang artifact storage thay vì nhồi hết vào DB
- đúng một trong hai field `source_artifact_url` hoặc `source_inline_text` nên được set

#### 5.2.21. `import_errors`

```text
id                  uuid pk
import_id           uuid fk -> imports.id
record_ref          varchar
error_message       text
payload             jsonb
created_at          timestamptz
```

#### 5.2.22. `content_revisions`

```text
id                  uuid pk
entity_type         varchar not null      -- vocab_book/grammar_lesson/reading_document
entity_id           uuid not null
version_no          integer not null
import_id           uuid null fk -> imports.id
source_type         varchar not null      -- markdown/plain_text/csv/json
source_payload      text not null
parsed_payload_json jsonb not null
keywords_json       jsonb null
render_artifact_url text null
change_summary      text null
created_by          uuid fk -> users.id
created_at          timestamptz
unique (entity_type, entity_id, version_no)
```

Notes:

- `parsed_payload_json` là canonical structure sau parse/normalize
- với vocabulary ở MVP, có thể lưu full draft dataset ở cấp `vocab_book` revision; nếu payload quá lớn thì phase sau tối ưu thêm

#### 5.2.23. `publications`

```text
id                  uuid pk
entity_type         varchar not null      -- vocab_book/grammar_lesson/reading_document
entity_id           uuid not null
revision_id         uuid not null fk -> content_revisions.id
action_type         varchar not null      -- publish/unpublish/rollback
status              varchar not null      -- active/superseded/reverted
change_summary      text not null
created_by          uuid fk -> users.id
created_at          timestamptz
```

Rules / indexes:

- chỉ có tối đa 1 publication `status = active` cho mỗi `(entity_type, entity_id)`
- learner read path resolve từ `active_publication_id` -> `publications.revision_id`

#### 5.2.24. `content_audit_logs`

```text
id                  uuid pk
actor_id            uuid fk -> users.id
action_type         varchar not null      -- import_commit/publish/unpublish/rollback
entity_type         varchar not null
entity_id           uuid not null
revision_id         uuid null fk -> content_revisions.id
publication_id      uuid null fk -> publications.id
metadata_json       jsonb null
created_at          timestamptz
```

#### 5.2.25. `jobs`

```text
id                  uuid pk
job_type            varchar not null      -- import_commit/publish/unpublish/rollback
status              varchar not null      -- queued/running/succeeded/failed/cancelled/dead_lettered
payload_json        jsonb not null
attempt_count       integer default 0
max_attempts        integer default 5
dedupe_key          varchar null
next_run_at         timestamptz not null default now()
lease_until         timestamptz null
last_heartbeat_at   timestamptz null
dead_letter_reason  text null
error_message       text null
created_by          uuid null fk -> users.id
started_at          timestamptz null
finished_at         timestamptz null
created_at          timestamptz
updated_at          timestamptz
```

Rules / indexes:

- worker chỉ được claim job khi `next_run_at <= now()` và lease còn trống/hết hạn
- `dedupe_key` nên có unique/partial index phù hợp để chặn enqueue trùng ngoài ý muốn

#### 5.2.26. `job_attempts`

```text
id                  uuid pk
job_id              uuid fk -> jobs.id
attempt_no          integer not null
worker_id           varchar null
started_at          timestamptz
finished_at         timestamptz null
error_message       text null
created_at          timestamptz
```

Unique:

- `unique (job_id, attempt_no)`

#### 5.2.27. `idempotency_keys`

```text
id                  uuid pk
actor_id            uuid not null fk -> users.id
scope               varchar not null      -- vocab_answer/reading_progress/import_commit/publish
idempotency_key     varchar not null
request_hash        varchar not null
status              varchar not null      -- started/completed/failed
response_snapshot   jsonb null
expires_at          timestamptz null
created_at          timestamptz
unique (actor_id, scope, idempotency_key)
```

#### 5.2.28. `auth_events`

```text
id                  uuid pk
user_id             uuid null fk -> users.id
session_id          uuid null fk -> user_sessions.id
event_type          varchar not null      -- login_success/login_failed/refresh_reuse/logout/logout_all
ip_address          inet
user_agent          text
metadata            jsonb
created_at          timestamptz
```

### 5.3. Nguyên tắc schema quan trọng

- dùng một nguồn sự thật cho publish state: `status`, không giữ thêm `is_published` song song
- `imports.preview_payload` chỉ nên lưu dữ liệu preview vừa phải; nếu payload lớn, chuyển sang artifact storage và chỉ lưu reference
- `content_revisions` là nguồn sự thật cho draft/source history
- `publications` là nguồn sự thật cho live history
- learner tables chỉ chứa published projection bất biến theo `publication_id`
- learner read path resolve từ `active_publication_id` -> `publications.revision_id`
- `imports` phải chỉ rõ source là artifact hay inline text; nếu preview/report lớn thì dùng artifact URLs
- thêm CHECK constraints cho các enum-like fields: `status`, `learning_status`, `direction`, `ordering`, `default_display_mode`
- thêm composite/partial indexes cho query nóng: active publications, active sessions, non-revoked sessions, queued jobs
- thêm invariant ở DB khi phù hợp, ví dụ: `status = published => active_publication_id is not null`
- khai báo rõ `ON DELETE` strategy cho các quan hệ chính thay vì để implicit
- các mutation từ client phải có idempotency boundary rõ ràng

## 6. Quan hệ dữ liệu chính

```text
users 1---n user_sessions
users 1---n vocab_study_sessions
users 1---n user_vocab_progress
users 1---n grammar_attempts
users 1---n reading_progress
users 1---n jobs (created_by)
users 1---n auth_events
users 1---n publications (created_by)
users 1---n content_audit_logs (actor_id)

vocab_books 1---n vocab_words
vocab_books 1---n content_revisions
vocab_books 1---n publications
vocab_words 1---n vocab_word_examples
vocab_words 1---n vocab_word_media
vocab_study_sessions 1---n vocab_session_items
vocab_study_sessions 1---n vocab_quiz_attempts

grammar_topics 1---n grammar_lessons
grammar_topics 1---n grammar_exercises
grammar_lessons 1---n grammar_exercises
grammar_exercises 1---n grammar_exercise_options
grammar_lessons 1---n content_revisions
grammar_lessons 1---n publications

reading_documents 1---n content_revisions
reading_documents 1---n publications
publications 1---n reading_sections
publications 1---n reading_highlights
imports 1---n content_revisions
content_revisions 1---n publications
publications 1---n content_audit_logs
```

## 7. API design

### 7.1. API conventions

- base path: `/api/v1`
- auth transport:
  - access token: `Authorization: Bearer <token>`
  - refresh token: HttpOnly Secure SameSite cookie
- mọi request có side effect nên hỗ trợ `Idempotency-Key`
- các resource mutable nên trả `version` hoặc `ETag` để hỗ trợ optimistic concurrency
- published learner content endpoints nên hỗ trợ `ETag` / `Last-Modified`
- list/history lớn nên ưu tiên cursor pagination ngoài `page/limit`
- publish/rollback requests nên kèm `change_summary` và precondition check phù hợp
- response format thống nhất:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

- error response:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body",
    "details": {
      "username": "required"
    }
  }
}
```

#### API phân vùng

- public auth endpoints
- private learner endpoints
- private admin endpoints dưới `/api/v1/admin/...`

#### Health endpoints

- `GET /api/v1/healthz`
- `GET /api/v1/readyz`

### 7.2. Auth endpoints

#### `POST /api/v1/auth/login`

Request:

```json
{
  "username": "long",
  "password": "secret123"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "user": {
      "id": "uuid",
      "username": "long",
      "display_name": "Long",
      "role": "learner"
    }
  }
}
```

Notes:

- refresh token không trả raw trong body ở production flow
- refresh token được set qua HttpOnly cookie

#### Các endpoint còn lại

- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/logout-all`
- `GET /api/v1/auth/me`

Tùy mô hình sản phẩm:

- nếu self-service registration là requirement thật, thêm `POST /api/v1/auth/register`
- nếu đây là hệ thống private/internal, account do admin tạo và registration bị tắt

### 7.3. Dashboard endpoints

- `GET /api/v1/dashboard/summary`

Response nên gồm:

- resume candidate cho vocabulary
- recent grammar lessons
- recent reading documents
- progress snapshot

### 7.4. Vocabulary endpoints

#### Books

- `GET /api/v1/vocab/books`
- `GET /api/v1/vocab/books/:bookId`
- `GET /api/v1/vocab/books/:bookId/words`

Query params ví dụ:

- `page`
- `limit`
- `status`
- `ordering=sequential|random`
- `direction=en_vi|vi_en`

#### Study session

- `POST /api/v1/vocab/books/:bookId/sessions`
- `GET /api/v1/vocab/sessions/:sessionId`
- `GET /api/v1/vocab/sessions/:sessionId/next`
- `GET /api/v1/vocab/sessions/:sessionId/items/:questionSeq`
- `POST /api/v1/vocab/sessions/:sessionId/answer`
- `POST /api/v1/vocab/sessions/:sessionId/complete`
- `GET /api/v1/vocab/sessions/resume?book_id=...`

Request ví dụ:

```json
{
  "mode": "guess_meaning",
  "study_mode": "flip",
  "direction": "en_vi",
  "ordering": "random",
  "show_ipa": true,
  "status_filter": "learning",
  "auto_next": true
}
```

Answer submit nên mang theo:

- `session_item_id`
- `Idempotency-Key`

Response ví dụ cho `next`:

```json
{
  "success": true,
  "data": {
    "session_item_id": "uuid",
    "question_seq": 7,
    "word_id": "uuid",
    "prompt": "abandon",
    "ipa": "/əˈbændən/",
    "options": [
      "từ bỏ",
      "xây dựng",
      "lặp lại",
      "giấu đi"
    ],
    "question_type": "mcq"
  }
}
```

Response ví dụ cho `answer`:

```json
{
  "success": true,
  "data": {
    "session_item_id": "uuid",
    "is_correct": true,
    "correct_answer": "abandon",
    "updated_progress": {
      "learning_status": "learning",
      "correct_count": 5,
      "wrong_count": 1
    }
  }
}
```

#### User progress

- `GET /api/v1/vocab/progress?book_id=...`
- `PATCH /api/v1/vocab/words/:wordId/progress`

#### Admin / import

- `POST /api/v1/admin/vocab/import/preview`
- `POST /api/v1/admin/vocab/import/commit`   -- enqueue job, trả `job_id`
- `PATCH /api/v1/admin/vocab/books/:bookId/publish`   -- enqueue publish job
- `PATCH /api/v1/admin/vocab/books/:bookId/unpublish`
- `GET /api/v1/admin/vocab/books/:bookId/revisions`
- `POST /api/v1/admin/vocab/books/:bookId/rollback`

### 7.5. Grammar endpoints

#### Topics / lessons

- `GET /api/v1/grammar/topics`
- `GET /api/v1/grammar/topics/:topicId`
- `GET /api/v1/grammar/lessons/:lessonId`

Response nên gồm:

- metadata
- parsed blocks
- rendered content
- keywords highlight
- related exercises
- publication metadata (`publication_id`, `revision_id`, `published_at`, `etag`)

#### Exercises

- `GET /api/v1/grammar/topics/:topicId/exercises`
- `POST /api/v1/grammar/exercises/:exerciseId/answer`

#### Admin import / publish

- `POST /api/v1/admin/grammar/import/preview`
- `POST /api/v1/admin/grammar/import/commit`   -- enqueue job, trả `job_id`
- `PATCH /api/v1/admin/grammar/lessons/:lessonId/publish`   -- enqueue publish job
- `PATCH /api/v1/admin/grammar/lessons/:lessonId/unpublish`
- `GET /api/v1/admin/grammar/lessons/:lessonId/revisions`
- `POST /api/v1/admin/grammar/lessons/:lessonId/rollback`

### 7.6. Reading endpoints

#### Documents

- `GET /api/v1/reading/documents`
- `GET /api/v1/reading/documents/:documentId`
- `GET /api/v1/reading/documents/:documentId/sections?cursor=...&limit=...`

#### Progress

- `PATCH /api/v1/reading/documents/:documentId/progress`

Request ví dụ:

```json
{
  "base_version": 3,
  "publication_id": "uuid",
  "last_section_id": "uuid",
  "last_section_order": 12,
  "completion_percent": 42.5
}
```

Rules:

- server chỉ chấp nhận update nếu `base_version` còn mới
- `publication_id` phải khớp active publication hoặc policy migrate progress đã được định nghĩa rõ

#### Admin import / publish

- `POST /api/v1/admin/reading/import/preview`
- `POST /api/v1/admin/reading/import/commit`   -- enqueue job, trả `job_id`
- `PATCH /api/v1/admin/reading/documents/:documentId/publish`   -- enqueue publish job
- `PATCH /api/v1/admin/reading/documents/:documentId/unpublish`
- `GET /api/v1/admin/reading/documents/:documentId/revisions`
- `POST /api/v1/admin/reading/documents/:documentId/rollback`

Notes:

- learner endpoints chỉ resolve từ `active_publication_id`
- preview/draft chỉ có ở admin routes

### 7.7. Admin ops utility endpoints

- `GET /api/v1/admin/imports`
- `GET /api/v1/admin/imports/:importId`
- `GET /api/v1/admin/jobs/:jobId`
- `GET /api/v1/admin/publications?entity_type=...&entity_id=...`
- `GET /api/v1/admin/content/diff?entity_type=...&entity_id=...&target_revision_id=...`

Mục đích:

- xem import history
- xem validation report cũ
- theo dõi commit/publish jobs
- xem publication history của từng content entity
- xem diff live vs revision trước publish
- debug import lỗi nhanh hơn

### 7.8. Swagger coverage

Cần generate Swagger cho:

- auth
- dashboard
- vocabulary
- grammar
- reading
- admin content ops
- error schema
- pagination schema
- health/readiness schema

Nên có các model docs:

- `LoginRequest`
- `LoginResponse`
- `ErrorResponse`
- `VocabSessionRequest`
- `VocabAnswerRequest`
- `GrammarLessonResponse`
- `ReadingDocumentResponse`
- `ImportPreviewResponse`
- `ImportValidationError`
- `SessionSummaryResponse`

## 8. UI flows

### 8.1. Auth flow

```text
Login Page
  -> nhập username/password
  -> submit
  -> thành công: redirect Dashboard
  -> lỗi: hiển thị inline error / toast
```

UI states:

- idle
- submitting
- invalid credentials
- network error
- logged in
- refreshing session

### 8.2. Dashboard flow

```text
Dashboard
  -> Continue last vocab session
  -> Recommended next lesson
  -> Recent grammar lessons
  -> Recent reading documents
  -> Progress snapshot
```

Mục tiêu:

- giảm số click để quay lại học
- tăng tỷ lệ user hoàn thành first successful learning session

### 8.3. Vocabulary flow

#### Flow A: học theo sách

```text
Dashboard
  -> Vocabulary Books Page
  -> chọn 1 book
  -> Book Detail
  -> chọn tab
  -> chọn custom settings
  -> Start Session
  -> học từng item
  -> xem kết quả hoặc resume nếu còn unfinished session
```

#### Flow B: tab Đoán nghĩa

```text
Book Detail
  -> Tab Đoán nghĩa
  -> hiển thị 1 từ
  -> 4 đáp án
  -> chọn 1 đáp án
  -> feedback đúng/sai
  -> next question
```

#### Flow C: tab Học từng từ

```text
Book Detail
  -> Tab Học từng từ
  -> hiển thị word, IPA, part of speech, nghĩa, image, context, examples, tips
  -> chọn chế độ flip hoặc write
  -> next / previous
```

#### Flow D: custom study settings

```text
Settings Drawer / Modal
  -> ordering: random/sequential/repeat
  -> direction: en-vi / vi-en
  -> show IPA on/off
  -> filter status
  -> auto-next on/off
  -> apply
  -> session reload theo config
```

#### Flow E: session summary

```text
Session complete
  -> accuracy
  -> words moved to learning/mastered
  -> continue difficult words
  -> back to book
```

UX details:

- quiz options hiển thị grid gọn, dễ tap
- đúng: highlight xanh + auto-next delay ngắn
- sai: highlight đỏ + hiện đáp án đúng + ví dụ
- write mode không mất input nếu submit lỗi mạng

### 8.4. Grammar flow

```text
Dashboard
  -> Grammar Topics
  -> chọn topic
  -> Lesson Detail
  -> đọc nội dung markdown
  -> từ khóa được highlight
  -> kéo xuống bài tập
  -> làm bài
  -> xem giải thích
```

Import flow nội bộ/admin:

```text
Admin Import Page
  -> upload markdown
  -> parse preview
  -> validate metadata
  -> fix lỗi nếu có
  -> commit (enqueue job)
  -> theo dõi job status
  -> xem diff live vs revision mới
  -> nhập change summary
  -> publish
```

Validation failures cần hiển thị rõ:

- thiếu frontmatter bắt buộc
- topic/slug không hợp lệ
- keyword parse lỗi
- exercise payload lỗi

### 8.5. Reading flow

```text
Dashboard
  -> Reading Documents
  -> chọn tài liệu
  -> Reading Viewer
      - docs mode cho nội dung dài
      - desktop 2-pane sync theo section
      - mobile language toggle
      - highlight
  -> lưu progress theo section
```

Hai mode hiển thị nên hỗ trợ:

- song ngữ song song
- toggle language cho màn hình nhỏ

Admin flow cho reading:

```text
Admin Import Page
  -> upload markdown/plain text
  -> parse preview
  -> kiểm tra section pairing EN/VI
  -> commit (enqueue job)
  -> theo dõi job status
  -> xem diff live vs revision mới
  -> nhập change summary
  -> publish
```

## 9. Component tree

### 9.1. App-level

```text
App
 ├── AppRouter
 ├── AuthProvider
 ├── QueryProvider
 ├── Layout
 │    ├── Header
 │    ├── Sidebar
 │    ├── MainContent
 │    └── GlobalToast
 ├── ErrorBoundary
 └── SessionRefreshGate
```

### 9.2. Auth components

```text
LoginPage
 ├── LoginForm
 │    ├── UsernameInput
 │    ├── PasswordInput
 │    ├── SubmitButton
 │    └── ErrorMessage
```

### 9.3. Vocabulary components

```text
VocabularyPage
 ├── BookList
 │    └── BookCard
 └── BookDetailPage
      ├── BookHeader
      ├── StudySettingsPanel
      ├── ResumeSessionBanner
      ├── VocabularyTabs
      │    ├── GuessMeaningTab
      │    │    ├── QuestionCard
      │    │    ├── OptionList
      │    │    ├── ResultFeedback
      │    │    └── NextButton
      │    └── LearnWordTab
      │         ├── Flashcard
      │         ├── WordHeader
      │         ├── IPASection
      │         ├── MeaningSection
      │         ├── ImageSection
      │         ├── ContextSection
      │         ├── ExampleList
      │         ├── TipsSection
      │         ├── FlipModeControls
      │         ├── WriteAnswerPanel
      │         └── NavigationControls
      ├── ProgressSummary
      └── SessionSummaryModal
```

### 9.4. Grammar components

```text
GrammarPage
 ├── TopicList
 │    └── TopicCard
 └── GrammarLessonPage
      ├── LessonHeader
      ├── MarkdownRenderer
      ├── HighlightedKeyword
      ├── ExerciseSection
      │    ├── ExerciseCard
      │    ├── OptionList
      │    ├── FillBlankInput
      │    └── ExplanationPanel
      └── LessonFooter
```

### 9.5. Reading components

```text
ReadingPage
 ├── DocumentList
 │    └── DocumentCard
 └── ReadingDetailPage
      ├── ReadingHeader
      ├── ReadingToolbar
      │    ├── ModeSwitcher
      │    ├── LanguageToggle
      │    ├── HighlightToggle
      │    └── FontSizeControl
      ├── ReadingDocsViewer
      │    └── ReadingSectionCard
      │         ├── Heading
      │         ├── EnglishPane
      │         ├── VietnamesePane
      │         └── HighlightedText
      └── ReadingProgressBar
```

### 9.6. Admin components

```text
AdminImportPage
 ├── ImportSourcePicker
 ├── FileUploadBox
 ├── RawTextInput
 ├── ParsePreviewPanel
 ├── ValidationReport
 ├── PublishControls
 └── ImportHistoryTable
```

## 10. State management design

### 10.1. Server state

Dùng TanStack Query cho:

- dashboard summary
- books list
- book detail
- vocab session question
- grammar lesson
- grammar exercises
- reading documents
- reading sections
- progress data
- import preview data

### 10.2. Client/UI state

Dùng Zustand hoặc React Context cho:

- auth session
- flashcard local state
- study settings
- reading viewer preferences
- display mode preference theo device/user
- theme / font size / highlight toggle
- resume session info
- admin import draft state
- local persistence qua IndexedDB hoặc abstraction tương đương cho:
  - write-mode input chưa submit xong
  - pending vocab answer submissions
  - pending reading progress updates
  - admin import draft recovery

Ví dụ state Vocabulary:

```ts
type StudySettings = {
  ordering: 'random' | 'sequential' | 'repeat';
  direction: 'en_vi' | 'vi_en';
  showIPA: boolean;
  statusFilter: 'all' | 'new' | 'learning' | 'mastered';
  studyMode: 'flip' | 'write';
  autoNext: boolean;
};
```

## 11. Import pipeline design

### 11.1. Grammar import từ Markdown

Format gợi ý:

```md
---
title: Present Perfect
slug: present-perfect
topic: tenses
level: beginner
keywords:
  - present perfect
  - have/has + V3
---

# Present Perfect

Present perfect is used to...

## Structure
Have/Has + V3

## Example
- I have finished my homework.
- She has gone to school.
```

Pipeline:

```text
Upload Markdown
  -> parse frontmatter
  -> validate required fields
  -> normalize slug/topic
  -> extract keywords
  -> render preview
  -> validation report
  -> enqueue commit job
  -> tạo draft revision
  -> link topic/exercises
  -> xem diff live vs draft revision
  -> publish/unpublish qua publish job + change summary
```

### 11.2. Reading import từ Markdown/text

Markdown format gợi ý:

```md
---
title: A Visit to Da Lat
slug: a-visit-to-da-lat
default_display_mode: docs
is_bilingual: true
---

## Section 1
[EN]
Da Lat is a beautiful city...

[VI]
Đà Lạt là một thành phố đẹp...

## Section 2
[EN]
The weather is cool all year round...

[VI]
Thời tiết mát mẻ quanh năm...
```

Plain text format gợi ý:

```text
# Title: A Visit to Da Lat

=== SECTION ===
EN: Da Lat is a beautiful city...
VI: Đà Lạt là một thành phố đẹp...

=== SECTION ===
EN: The weather is cool all year round...
VI: Thời tiết mát mẻ quanh năm...
```

Pipeline:

```text
Upload file
  -> detect type markdown/plain_text
  -> parse title/metadata
  -> split sections
  -> map EN/VI pairs
  -> extract highlights
  -> validation report
  -> enqueue commit job
  -> tạo draft revision
  -> materialize sections/highlights theo `publication_id` khi publish
  -> xem diff live vs draft revision
  -> publish/unpublish qua publish job + change summary
```

### 11.3. Vocabulary import từ CSV/JSON

Format gợi ý:

```text
word,ipa,part_of_speech,meaning_vi,meaning_en,context,example_en_1,example_vi_1,image_url
```

Pipeline:

```text
Upload file
  -> validate required columns
  -> normalize word/slug
  -> detect duplicates trong file và trong DB
  -> preview N records đầu
  -> enqueue commit job
  -> commit vào draft revision / draft dataset
  -> validation report
  -> materialize `vocab_words` / examples / media khi publish
  -> xem diff live vs draft revision
  -> publish/unpublish book + change summary
  -> thống kê success/failed rows
```

### 11.4. Import rules chung

- preview trước commit
- publish là bước tách riêng sau commit
- import source phải là uploaded artifact hoặc inline text có lưu dấu vết rõ ràng
- giới hạn file size
- kiểm tra MIME type
- ghi `import_errors` đủ chi tiết để sửa được
- dùng checksum để tránh import trùng ngoài ý muốn
- preview nên synchronous cho payload vừa phải để feedback nhanh
- commit/publish nên chạy qua background jobs để tránh timeout và double-submit
- UI phải poll `job status` hoặc có cơ chế theo dõi tương đương
- retry commit/publish phải idempotent
- preview/report lớn nên đẩy sang artifact storage và chỉ giữ reference trong DB

## 12. Business rules

### 12.1. Auth

- username là duy nhất
- password tối thiểu 8 ký tự
- login sai nhiều lần có thể rate limit
- access token hết hạn ngắn, refresh token dài hơn
- refresh token rotate sau mỗi refresh thành công
- session có thể revoke riêng từng thiết bị
- refresh token reuse detection phải revoke được token family liên quan

### 12.2. Vocabulary

Mỗi word nên có:

- word
- meaning
- part_of_speech
- tối thiểu 2 examples

Rules:

- session phải lưu config lúc bắt đầu để replay đúng
- khi trả lời đúng/sai phải cập nhật `user_vocab_progress`
- `repeat` ưu tiên từ sai gần đây và từ có streak thấp
- MCQ phải tránh 4 options trùng hoặc quá dễ phân biệt
- learner app chỉ hiển thị books có `status = published`
- import thành công không tự động làm book live; publish là bước tách riêng
- learner reads của vocabulary đi qua published projection được build từ `active_publication_id`
- answer submit phải retry-safe khi client retry sau lỗi mạng ngắn
- câu hỏi đã phát ra phải được snapshot qua `vocab_session_items` để resume/retry deterministic

Logic cập nhật trạng thái đề xuất:

- `new`: chưa học hoặc học rất ít
- `learning`: đã học nhưng chưa ổn định
- `mastered`: đúng liên tiếp hoặc vượt ngưỡng

Ví dụ rule:

- đúng liên tiếp `>= 5` -> `mastered`
- đúng `1-4` lần -> `learning`
- sai sau `mastered` -> có thể hạ xuống `learning`

### 12.3. Grammar

- lesson phải thuộc 1 topic
- exercise có thể gắn với topic và lesson
- keyword highlight phải parse được từ markdown/frontmatter
- learner app chỉ hiển thị lesson có `status = published`
- learner chỉ đọc active publication, không đọc draft/source revision
- publish phải trỏ sang revision đã validate
- rollback không được làm mất lịch sử revision

### 12.4. Reading

- nếu `is_bilingual = true`, mỗi section nên có cả `content_en` và `content_vi`
- tài liệu dài phải tách sections
- progress lưu theo section cuối cùng đã đọc
- section pairing lỗi phải block publish
- learner chỉ đọc active publication
- section API phải hỗ trợ load theo cửa sổ cho tài liệu dài
- progress update phải gửi `base_version`, `publication_id`, `last_section_order` để chặn stale overwrite

### 12.5. Content ops

- import preview lỗi không được commit một phần âm thầm
- publish chỉ hợp lệ khi validation pass
- unpublish không được làm mất lịch sử import
- publish phải tạo publication record mới và cập nhật `active_publication_id`, không overwrite trực tiếp live history
- cần hỗ trợ rollback sang revision đã publish trước đó
- commit/publish là job-backed operations có retry và auditability rõ ràng
- diff preview và change summary là guardrail mặc định trước publish/rollback

## 13. Error handling strategy

### 13.1. Backend error handling

Middleware cần có:

- request logger
- panic recovery
- auth middleware
- request ID middleware
- CORS middleware

Chuẩn hóa error code:

- `UNAUTHORIZED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `NOT_FOUND`
- `CONFLICT`
- `PRECONDITION_FAILED`
- `IDEMPOTENCY_CONFLICT`
- `RATE_LIMITED`
- `INTERNAL_SERVER_ERROR`
- `IMPORT_PARSE_ERROR`
- `IMPORT_VALIDATION_ERROR`
- `SESSION_REVOKED`

Ví dụ mapping:

- sai password -> `401 UNAUTHORIZED`
- username thiếu -> `400 VALIDATION_ERROR`
- lesson không tồn tại -> `404 NOT_FOUND`
- lỗi parse markdown -> `422 IMPORT_PARSE_ERROR`
- refresh token reuse hoặc revoked -> `401 SESSION_REVOKED`
- panic hoặc DB down -> `500 INTERNAL_SERVER_ERROR`

### 13.2. Frontend error handling

Nguyên tắc:

- lỗi form hiển thị inline
- lỗi network hiển thị toast + retry
- lỗi session hết hạn -> refresh token tự động
- refresh fail -> redirect login
- loading skeleton cho page chính
- empty state riêng cho list rỗng

Cases cụ thể:

- Login page:
  - sai tài khoản -> hiển thị dưới form
  - API timeout -> toast "Không thể kết nối máy chủ"
- Vocabulary:
  - không tải được next question -> cho retry
  - submit answer fail -> giữ nguyên câu hiện tại, không mất input
  - nếu lỗi mạng ngắn -> queue submit cục bộ và retry an toàn bằng `Idempotency-Key`
- Grammar:
  - markdown render fail -> hiển thị fallback raw text nếu có
- Reading:
  - section load lỗi -> retry riêng từng section
  - progress update fail -> giữ pending progress cục bộ và flush lại khi online
- Admin import:
  - preview fail -> giữ file/raw input, hiện validation report, không xóa draft

## 14. Security strategy

### 14.1. Auth security

- bcrypt hash password
- không log password hoặc token raw
- refresh token lưu hash
- JWT secret lưu env
- HTTPS only trên production
- rate limit login endpoint
- account lock tạm thời nếu brute force nhiều lần
- refresh cookie phải có `HttpOnly`, `Secure`, `SameSite` phù hợp deployment model
- access token chỉ giữ in-memory
- logout phải clear cookie và revoke session DB
- refresh token reuse detection phải được test và audit log
- nếu dùng cookie-authenticated refresh/logout, CSRF protection là mặc định
- signing key nên hỗ trợ rotation

### 14.2. API security

- validate request body
- sanitize input Markdown/text nếu render HTML
- tránh XSS khi render content import
- giới hạn file upload size
- validate MIME type khi import
- phân quyền learner/admin ở middleware hoặc policy layer
- không expose admin import/publish routes cho learner

### 14.3. Content security

Khi render Markdown:

- không cho phép raw HTML tùy ý nếu chưa sanitize
- whitelist markdown features
- strip `script` và `style`

### 14.4. Asset security

- media upload qua signed URL hoặc backend proxy
- giới hạn content type và size
- không cho client ghi trực tiếp bucket public không kiểm soát
- dùng CDN/object storage cho asset tĩnh để giảm tải app server

### 14.5. Admin-sensitive actions

- publish/unpublish/rollback nên yêu cầu recent re-auth hoặc equivalent step-up auth nếu deployment model có privileged admin accounts
- mọi live content change phải có `change_summary`
- mọi live content change phải ghi `content_audit_logs`

## 15. Performance strategy

### 15.1. Backend

- index cho các cột filter/query nhiều
- phân trang cho list APIs
- preload hợp lý với GORM
- tránh N+1 query ở:
  - vocab words + examples
  - grammar lesson + exercises
  - reading document + sections
- có thể precompute hoặc cache nhẹ cho distractor pool nếu cần
- data integrity nên được khóa ở DB bằng unique/check/index phù hợp, không chỉ ở service validation

### 15.2. Frontend

- code split theo route
- lazy load module nặng
- virtualized list nếu số lượng words lớn
- cache query với TanStack Query
- debounce search input
- lazy load image trong vocabulary
- skeleton states cho dashboard/list/detail

### 15.3. Reading docs dài

- render theo section
- không render toàn bộ một lần nếu nội dung rất dài
- lưu progress theo section thay vì pixel scroll
- desktop sync theo section thay vì pixel-perfect sync
- section API phải hỗ trợ load theo cửa sổ để tránh trả toàn bộ tài liệu trong một response

## 16. Testing strategy

### 16.1. Backend testing

#### Unit tests

Test cho:

- auth service
- vocab session generator
- progress update rules
- grammar markdown parser
- reading section parser
- repository query filters
- highlight matcher / keyword escaper
- refresh token rotation logic
- parser fuzz tests cho markdown/reading section parsing
- strategy determinism tests cho random/sequential/repeat
- publication activation / supersede / rollback logic
- job lease / heartbeat / retry scheduler logic

#### Integration tests

Dùng PostgreSQL test DB hoặc container:

- login flow
- create vocab session
- answer question và update progress
- import grammar markdown
- import reading markdown
- get reading sections
- publish/unpublish flow
- revoke session flow
- rollback flow không làm leak draft sang learner
- job worker crash rồi recover không chạy đôi cùng một publish job

#### API tests

- status code
- response schema
- auth middleware
- invalid payload
- not found cases
- admin permission cases
- idempotency/retry behavior
- optimistic concurrency / version conflict

Gợi ý công cụ:

- Go testing
- `httptest`
- `testify`
- Docker test database

### 16.2. Frontend testing

#### Unit tests

- form validation
- flashcard state transitions
- study settings reducer/store
- highlight renderer
- reading mode switcher
- auth refresh gate

#### Component tests

- LoginForm
- GuessMeaningTab
- LearnWordTab
- GrammarExerciseCard
- ReadingSectionCard
- ImportPreviewPanel

#### Contract / client sync

- generated TypeScript client từ OpenAPI
- contract tests để phát hiện drift giữa backend spec và frontend usage

#### E2E tests

- login thành công/thất bại
- vào vocab book và học 1 session
- đổi setting random / sequential
- làm bài grammar
- mở reading doc và đổi mode hiển thị
- resume unfinished session
- import preview -> commit -> publish
- submit answer lỗi mạng nhưng không mất input
- reading progress lỗi mạng ngắn rồi replay thành công
- learner không nhìn thấy draft revision chưa publish

Gợi ý công cụ:

- Vitest
- React Testing Library
- Playwright

### 16.3. Manual verification / smoke checklist

- Auth:
  - login
  - refresh
  - logout current device
  - logout all devices
- Vocabulary:
  - start session
  - wrong/correct answer flow
  - resume session
  - summary cuối phiên
  - same session resume vẫn thấy đúng câu hỏi/options đã phát ra
- Grammar:
  - import preview
  - validation errors
  - publish lesson
  - diff preview + change summary trước publish
  - publication history / rollback
- Reading:
  - section pairing
  - desktop 2-pane
  - mobile language toggle
  - progress gửi version/publication_id đúng
- Ops:
  - `healthz` / `readyz`
  - migrate up/down
  - `docker-compose up` chạy được end-to-end

## 17. Test cases trọng điểm

### 17.1. Auth

- login đúng tài khoản
- login sai password
- access token hết hạn
- refresh token revoked
- refresh token reuse bị phát hiện và token family bị revoke
- logout xong không refresh được nữa
- logout-all xong session cũ không dùng lại được

### 17.2. Vocabulary

- lấy book list thành công
- book không có words
- guess meaning trả đúng 4 options
- option đúng không bị lặp
- write answer không phân biệt hoa thường khi ở lenient mode
- lọc status hoạt động đúng
- ordering random không lặp bất thường
- ordering sequential đúng sort order
- repeat mode lặp lại từ sai nhiều hơn
- session summary tính đúng accuracy
- retry submit answer không tạo duplicate attempt
- resume session vẫn giữ đúng prompt/options đã phát cho mỗi `question_seq`

### 17.3. Grammar

- import markdown hợp lệ
- thiếu frontmatter bắt buộc
- parse keywords đúng
- render highlight đúng từ khóa
- submit exercise lưu attempt
- lesson draft không lộ ra learner API
- `selected_option_id` luôn trỏ tới option hợp lệ
- topic-level exercise và lesson-level exercise không drift topic

### 17.4. Reading

- import markdown song ngữ thành sections đúng
- plain text import đúng format
- docs dài hiển thị section-based
- mobile toggle hoạt động đúng
- progress update đúng section cuối
- section pairing lỗi thì không publish được
- replay pending progress sau lỗi mạng không ghi đè sai progress mới hơn
- rollback publication không làm highlight trỏ nhầm sang section khác

## 18. Logging, monitoring, observability

### 18.1. Logging

Backend nên log:

- `request_id`
- `trace_id`
- route
- method
- `status_code`
- latency
- `user_id`
- `session_id`
- `job_id`
- `worker_id`
- `import_id`
- `publication_id`
- `error_code`

Không log:

- password
- token raw
- nội dung nhạy cảm

### 18.2. Monitoring

Theo dõi:

- login fail rate
- API latency
- import fail rate
- vocab answer submit fail rate
- reading section load fail rate
- refresh failure rate
- publish failure rate
- DB connection saturation
- job retry rate
- dead-letter job rate
- import/publish completion latency
- offline replay failure rate
- API contract drift failure rate trong CI

### 18.3. Product success metrics

- activation:
  - % user hoàn thành session vocab đầu tiên
- engagement:
  - sessions/user/week
  - reading completion rate
- content ops:
  - time từ import đến publish
  - import validation failure rate
- reliability:
  - p95 latency
  - 5xx rate

## 19. Triển khai thư mục backend đề xuất

```text
backend/
├── cmd/api/main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── request_id.go
│   ├── models/
│   │   ├── user.go
│   │   ├── vocab.go
│   │   ├── grammar.go
│   │   ├── reading.go
│   │   └── publication.go
│   ├── queries/
│   ├── policies/
│   ├── dto/
│   │   ├── auth_dto.go
│   │   ├── vocab_dto.go
│   │   ├── grammar_dto.go
│   │   ├── reading_dto.go
│   │   └── admin_import_dto.go
│   ├── repositories/
│   │   ├── user_repository.go
│   │   ├── vocab_repository.go
│   │   ├── grammar_repository.go
│   │   ├── reading_repository.go
│   │   ├── import_repository.go
│   │   ├── publication_repository.go
│   │   ├── audit_repository.go
│   │   ├── job_repository.go
│   │   ├── job_attempt_repository.go
│   │   ├── idempotency_repository.go
│   │   └── auth_event_repository.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── vocab_service.go
│   │   ├── grammar_service.go
│   │   ├── reading_service.go
│   │   ├── import_service.go
│   │   └── publish_service.go
│   ├── jobs/
│   │   ├── dispatcher.go
│   │   └── workers.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── vocab_handler.go
│   │   ├── grammar_handler.go
│   │   ├── reading_handler.go
│   │   └── admin_handler.go
│   ├── importers/
│   │   ├── markdown_parser.go
│   │   ├── reading_parser.go
│   │   └── vocab_csv_parser.go
│   ├── storage/
│   │   └── asset_storage.go
│   ├── router/
│   └── utils/
│       ├── jwt.go
│       ├── password.go
│       └── validator.go
├── migrations/
└── docs/
```

## 20. Triển khai thư mục frontend đề xuất

```text
frontend/
├── src/
│   ├── app/
│   │   ├── App.tsx
│   │   └── providers.tsx
│   ├── routes/
│   │   ├── index.tsx
│   │   ├── ProtectedRoute.tsx
│   │   └── AdminRoute.tsx
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── VocabularyPage.tsx
│   │   ├── VocabularyBookDetailPage.tsx
│   │   ├── GrammarPage.tsx
│   │   ├── GrammarLessonPage.tsx
│   │   ├── ReadingPage.tsx
│   │   ├── ReadingDetailPage.tsx
│   │   └── AdminImportPage.tsx
│   ├── features/
│   │   ├── auth/
│   │   ├── vocabulary/
│   │   ├── grammar/
│   │   ├── reading/
│   │   └── admin/
│   ├── components/
│   │   ├── common/
│   │   ├── layout/
│   │   └── ui/
│   ├── services/
│   │   ├── apiClient.ts
│   │   ├── authApi.ts
│   │   ├── vocabApi.ts
│   │   ├── grammarApi.ts
│   │   ├── readingApi.ts
│   │   └── adminApi.ts
│   ├── store/
│   │   ├── authStore.ts
│   │   ├── studySettingsStore.ts
│   │   ├── importDraftStore.ts
│   │   ├── persistedDraftStore.ts
│   │   └── mutationQueueStore.ts
│   ├── hooks/
│   ├── utils/
│   └── types/
├── public/
└── vite.config.ts
```

Root-level delivery files:

- `docker-compose.yml`
- `Makefile`
- `.env.example`
- `README.md`

## 21. Roadmap triển khai

### Phase 0: Delivery foundation

- repo split frontend/backend
- Docker Compose
- PostgreSQL
- migrations
- Makefile
- CI: lint + unit test + build
- `healthz` / `readyz`

Exit criteria:

- developer mới clone repo và chạy local được trong dưới 30 phút

### Phase 1: Auth & session foundation

- login/logout/me
- refresh token rotation
- revoke current/all devices
- protected routes
- base layout + auth guard

Exit criteria:

- session flow an toàn và không phụ thuộc `localStorage` cho token dài hạn

### Phase 2: Content operations MVP

- grammar import preview/commit/publish
- reading import preview/commit/publish
- vocab CSV/JSON import preview/commit
- import error reporting
- baseline reliability cho content ops:
  - async commit/publish jobs
  - import/publish idempotency
  - revision history tối thiểu
  - publication history + content audit logs

Exit criteria:

- content team có thể tự import/publish mà không sửa DB thủ công

### Phase 3: Vocabulary MVP

- book list/detail
- guess meaning
- learn word
- flip/write mode
- progress tracking
- resume session
- summary cuối phiên
- baseline reliability cho learner writes:
  - answer submission idempotency
  - local draft / pending queue cho write mode
  - deterministic session items cho vocab

Exit criteria:

- user hoàn thành 1 session vocab trọn vẹn và xem lại tiến độ

### Phase 4: Grammar MVP

- topic/lesson pages
- markdown render
- keyword highlight
- exercise basic

Exit criteria:

- lesson import -> publish -> learner làm bài được end-to-end

### Phase 5: Reading MVP

- document list/detail
- section-based bilingual reading
- desktop 2-pane + mobile toggle
- progress save/resume

Exit criteria:

- tài liệu dài vẫn đọc mượt trên mobile và desktop

### Phase 6: Hardening, scale & deployment

- E2E smoke tests
- API contract tests
- generated client sync check trong CI
- observability
- performance optimization
- asset storage/CDN
- staging/prod deployment

Exit criteria:

- demo ổn định với dữ liệu thật và quan sát được lỗi/latency

## 22. MVP scope khuyến nghị

### Bắt buộc

- login
- vocabulary books
- flashcard learn mode
- guess meaning MCQ
- grammar lesson import từ markdown
- grammar exercise MCQ
- reading bilingual section-based
- highlight cơ bản
- progress cơ bản
- resume session
- admin import preview + commit + publish cho grammar/reading/vocab
- revisioned publish flow cho grammar/reading
- publication history + content audit logs cho content ops
- async commit/publish jobs cơ bản
- idempotency cho learner/admin writes quan trọng
- deterministic session items cho vocabulary
- Docker Compose + migrations + seed data
- `healthz` / `readyz` + logging cơ bản

### Để phase sau

- repeat mode thông minh hơn
- audio
- analytics dashboard
- admin CMS đầy đủ
- drag-and-drop import UI nâng cao
- note cá nhân / bookmark

### Ghi chú chiến lược

- admin import preview/publish là một phần của MVP, không phải nice-to-have
- nếu không có content ops vận hành được, learner product sẽ dừng ở mức demo
- không mở thêm product surfaces lớn cho tới khi vòng `import -> validate -> publish -> learner consume -> progress/retry-safe` ổn định

## 23. Rủi ro kỹ thuật và cách xử lý

### 23.1. Markdown parser không ổn định

Rủi ro:

- format file import không thống nhất, dễ lỗi parse

Giải pháp:

- định nghĩa format chuẩn có frontmatter rõ ràng
- validate trước khi commit
- có preview import

### 23.2. Reading song ngữ bị lệch đoạn

Rủi ro:

- EN/VI không map đúng section

Giải pháp:

- bắt buộc import theo section pairs
- pairing lỗi phải báo validation error và block publish

### 23.3. Vocabulary random/repeat phức tạp

Rủi ro:

- logic chọn từ khó debug

Giải pháp:

- tách strategy rõ ràng:
  - sequential strategy
  - random strategy
  - repeat strategy
- mỗi strategy có test riêng
- submit/retry phải deterministic và không tạo double attempt

### 23.4. XSS khi render Markdown

Rủi ro:

- người nhập nội dung chèn HTML nguy hiểm

Giải pháp:

- sanitize Markdown/HTML trước khi render
- không render raw HTML bừa bãi

### 23.5. Session management dễ lỗi hoặc kém an toàn

Rủi ro:

- refresh/access token flow dễ sai, dẫn tới logout bất thường hoặc hở bảo mật

Giải pháp:

- access token in-memory
- refresh token HttpOnly cookie
- refresh rotation
- revoke DB-backed sessions
- test đầy đủ refresh/revoke/logout-all

### 23.6. Dev environment lệch giữa các máy

Rủi ro:

- frontend/backend/database chạy khác nhau giữa các thành viên

Giải pháp:

- Docker Compose
- `.env.example`
- migrations
- Makefile
- README local setup

### 23.7. Content ops nghẽn cổ chai

Rủi ro:

- nếu không có preview/validation/publish flow, team sẽ phải sửa file hoặc DB thủ công

Giải pháp:

- admin import preview
- validation report
- import history
- publish/unpublish state

### 23.8. Content publish dễ ghi đè hoặc khó rollback

Rủi ro:

- import mới hoặc publish mới có thể vô tình ghi đè content đang live nếu không có revision boundary

Giải pháp:

- dùng `content_revisions`
- publish tạo publication record mới rồi cập nhật `active_publication_id`
- rollback về revision cũ mà không mất lịch sử

### 23.9. Retry/double-submit làm sai progress hoặc duplicate import

Rủi ro:

- mạng chập chờn hoặc double click có thể tạo duplicate attempt, duplicate progress update hoặc duplicate import job

Giải pháp:

- `Idempotency-Key`
- optimistic concurrency cho resource mutable
- background jobs có retry có kiểm soát

### 23.10. Job bị kẹt hoặc bị nhiều worker chạy trùng

Rủi ro:

- worker chết giữa chừng hoặc 2 worker cùng claim một job có thể làm publish/import chạy đôi hoặc kẹt `running`

Giải pháp:

- jobs có `lease_until`, `last_heartbeat_at`, `next_run_at`, `dedupe_key`
- lưu `job_attempts` để debug retry/failure rõ ràng
- có cơ chế requeue hoặc dead-letter cho job lỗi lặp lại

### 23.11. Reading progress bị stale overwrite sau publish mới hoặc replay muộn

Rủi ro:

- client replay progress cũ có thể ghi đè progress mới hơn nếu không gửi publication/version context

Giải pháp:

- PATCH progress phải kèm `base_version`, `publication_id`, `last_section_order`
- server từ chối update stale bằng precondition check rõ ràng

## 24. Kết luận kiến trúc đề xuất

Kiến trúc phù hợp nhất cho dự án này là:

- frontend React SPA
- backend Go theo layered architecture
- layered architecture được bổ sung `policies`, `queries`, `jobs`
- PostgreSQL với schema tách rõ theo module
- object storage cho media khi asset upload vào scope
- import pipeline riêng cho Markdown/text/CSV với `preview -> commit -> publish`
- commit/publish đi qua background jobs khi operation dài hoặc nặng
- session-based learning logic cho Vocabulary
- deterministic session model cho Vocabulary qua `vocab_session_items`
- section-based content model cho Reading
- revisioned content lifecycle cho source history + publication model riêng cho live history
- auth hybrid: access token in-memory + refresh token HttpOnly cookie
- idempotency và optimistic concurrency cho write paths quan trọng
- Swagger/OpenAPI + migrations + Docker Compose + CI-ready delivery
- test strategy gồm unit + integration + E2E + manual smoke checklist

Cách tổ chức này có các ưu điểm:

- dễ maintain
- dễ mở rộng thêm module mới
- phù hợp với React + Go + PostgreSQL
- hỗ trợ import nội dung học tốt
- tối ưu cho trải nghiệm học tập thực tế
- triển khai thật nhanh hơn
- an toàn hơn về session/auth
- vận hành nội dung thực tế tốt hơn
- boundary giữa draft/source history và live history rõ hơn
- giảm overwrite nhầm, retry lỗi và race conditions ở các write path
- giảm rủi ro "demo chạy được nhưng đội không ship nổi"
