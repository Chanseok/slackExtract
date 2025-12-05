# 작업 일지 (Work Log)

> 이 문서는 개발 과정에서의 주요 작업 내용, 의사 결정, 이슈 및 해결 방법을 기록합니다.
> `REQUIREMENTS.md`, `PLAN.md`와 함께 참고하여 프로젝트 진행 상황을 파악할 수 있습니다.

---

## 2025-12-05 (Day 1)

### 📋 주요 작업 내용

#### 1. 프로젝트 기획 및 요구사항 정의
- **목표:** Slack 대화 내용을 채널별로 추출하여 Markdown 파일로 저장하는 개인용 Interactive CLI 도구 개발.
- **런타임 선택:** 처음에는 Bun/Deno를 고려했으나, 실행 파일 크기, 메모리 효율성, 멀티 플랫폼 배포 편의성을 고려하여 **Go 언어**로 결정.
- **기존 도구 분석:** `slackdump`(Go), `slackprep`(Python) 등 기존 도구를 조사. `slackdump`의 강력한 기능을 참고하되, TUI는 새로 구현하기로 결정.

#### 2. 개발 환경 설정
- **Go 설치:** `brew install go` (v1.25.5)
- **프로젝트 초기화:** `go mod init github.com/chanseok/slackExtract`
- **주요 라이브러리 설치:**
  - `github.com/charmbracelet/bubbletea` (TUI 프레임워크)
  - `github.com/slack-go/slack` (공식 Slack SDK)
  - `github.com/joho/godotenv` (.env 파일 로더)
- **GitHub Repository 생성:** [https://github.com/Chanseok/slackExtract](https://github.com/Chanseok/slackExtract)

#### 3. 인증 방식 결정
- **문제:** Slack App을 만들어 `xoxp-` 토큰을 발급받으려 했으나, 워크스페이스 관리자 승인이 필요한 상황 발생 ("Request to Install" 상태).
- **해결:** 브라우저 쿠키(`d` cookie)와 클라이언트 토큰(`xoxc-`)을 이용한 인증 방식으로 전환. 이 방식은 관리자 승인 없이 개인 데이터에 접근 가능.
- **구현:** `net/http/cookiejar`를 사용하여 커스텀 HTTP 클라이언트에 쿠키를 심어 Slack API 호출.

#### 4. TUI 프로토타입 구현
- **Bubble Tea 모델 구현:**
  - `model` 구조체: 채널 목록, 커서 위치, 선택된 채널 집합 관리.
  - `Init()`, `Update()`, `View()` 메서드 구현.
- **기능:**
  - 화살표 키(↑/↓) 또는 j/k로 채널 탐색.
  - 스페이스바로 채널 다중 선택/해제.
  - Enter로 확인, q로 종료.
- **채널 목록 조회:** `GetConversations` API를 사용하여 Public, Private, MPIM, IM 채널 모두 조회.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 언어 | Go | 바이너리 크기 작음, 메모리 효율적, 크로스 컴파일 용이 |
| TUI 라이브러리 | Bubble Tea | Go 생태계에서 가장 인기 있고, Elm 아키텍처 기반으로 깔끔함 |
| Slack SDK | slack-go/slack | 공식 SDK, 문서화 잘 되어 있음 |
| 인증 방식 | Cookie (d + xoxc) | 관리자 승인 불필요, slackdump도 동일 방식 사용 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| `go: command not found` | Go 미설치 | `brew install go` 실행 |
| `Request to Install` (OAuth) | 워크스페이스 앱 설치 정책 | Cookie 인증 방식으로 전환 |
| `replace_string_in_file` 실패 | 파일 내용 불일치 | `read_file`로 내용 확인 후 정확한 문자열로 재시도 |

### 📁 생성된 파일 목록

```
slackExtract/
├── .env                    # 환경변수 (Git 제외)
├── .env.example            # 환경변수 템플릿
├── .gitignore              # Git 제외 파일 목록
├── README.md               # 프로젝트 소개
├── go.mod                  # Go 모듈 정의
├── go.sum                  # 의존성 체크섬
├── cmd/
│   └── slack-extract/
│       └── main.go         # 메인 엔트리 포인트 (TUI 포함)
└── docs/
    ├── REQUIREMENTS.md     # 요구사항 정의서
    ├── PLAN.md             # 개발 계획
    └── WORKLOG.md          # 작업 일지 (이 파일)
```

### 🎯 다음 작업 (TODO)

1. **메시지 다운로드 구현:** 선택된 채널의 메시지를 `GetConversationHistory` API로 가져오기.
2. **스레드 처리:** 각 메시지의 `ThreadTimestamp`를 확인하여 `GetConversationReplies`로 댓글 가져오기.
3. **사용자 ID 매핑:** `GetUsers` API로 사용자 목록을 캐싱하고, 메시지의 `UserID`를 실제 이름으로 변환.
4. **Markdown 변환기:** Slack mrkdwn 포맷을 표준 Markdown으로 변환하는 로직 구현.
5. **파일 저장:** `export/{날짜}/{채널명}.md` 구조로 파일 저장.

---

## 2025-12-05 (Day 1 - 추가 작업)

### 📋 주요 작업 내용

#### 1. 채널 목록 조회 개선 (Pagination)
- **문제:** `GetConversations` API 호출 시 `Limit: 1000` 설정으로 인해 1000개 이상의 채널이 있는 워크스페이스에서 일부 채널이 누락됨.
- **해결:** `next_cursor`를 확인하여 모든 페이지의 채널을 가져오도록 반복문(Loop) 처리 구현.

#### 2. TUI 뷰 개선 (Scrolling & AltScreen)
- **문제:** 채널 목록이 터미널 높이를 초과할 경우, 스크롤 기능이 없어 아래쪽 채널을 볼 수 없거나 화면이 잘리는 현상 발생.
- **해결:**
  - `tea.WithAltScreen()` 옵션을 추가하여 터미널 전체 화면 모드 적용.
  - `windowMin`, `height` 변수를 도입하여 스크롤 가능한 뷰포트(Viewport) 로직 구현.
  - `tea.WindowSizeMsg`를 처리하여 터미널 크기 변경에 동적으로 대응.

#### 3. 개발 환경 이슈 해결 (Go Version Mismatch)
- **문제:** Dev Container 환경에서 `go run` 실행 시 `compile: version mismatch` 오류 발생 (Go 1.25.4 vs 1.25.5).
- **원인:** 이전 빌드 캐시와 현재 설치된 Go 툴체인 버전 간의 불일치.
- **해결:**
  - `go clean -cache`로 빌드 캐시 삭제.
  - `go run -a ...` 옵션으로 강제 재빌드.
  - `export GOTOOLCHAIN=local` 설정으로 로컬 툴체인 사용 강제.

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| 채널 목록 누락 | API Pagination 미구현 | `cursor` 기반 반복 호출 로직 추가 |
| 터미널 화면 잘림 | TUI 스크롤 미구현 | Viewport 로직 및 `AltScreen` 적용 |
| `compile: version mismatch` | Go 버전 업데이트 후 캐시 잔존 | `go clean -cache` 및 `go run -a` 실행 |

---

## 2025-12-05 (Day 1 - 핵심 기능 구현)

### 📋 주요 작업 내용

#### 1. 메시지 다운로드 구현
- **기능:** `GetConversationHistory` API를 사용하여 선택된 채널의 모든 메시지를 가져오기.
- **Pagination:** `cursor` 기반 반복 호출로 메시지가 많은 채널도 빠짐없이 수집.
- **결과:** 채널 선택 후 Enter 키를 누르면 메시지 다운로드가 시작됨.

#### 2. Markdown 파일 저장
- **저장 경로:** `export/{채널명}.md`
- **포맷:** 메시지별로 `### 사용자 (시간)\n내용` 형식의 Markdown으로 저장.
- **정렬:** 과거 -> 최신 순서로 읽기 편하게 역순 정렬.

#### 3. 사용자 ID -> 이름 매핑
- **문제:** 메시지에 User ID (`U12345`)가 표시되어 가독성이 떨어짐.
- **해결:** `GetUsersPaginated` API로 전체 사용자 목록을 가져와 매핑 테이블 생성.
- **멘션 변환:** 본문 내 `<@U12345>` 형태의 멘션을 `@홍길동` 형태로 자동 변환.

#### 4. 사용자 목록 캐싱 (users.json)
- **문제:** 대규모 워크스페이스에서 매 실행마다 사용자 목록을 가져오는 것은 비효율적이며, API Rate Limit에 걸릴 위험이 있음.
- **해결:** `users.json` 파일에 캐싱하여 다음 실행 시에는 파일에서 즉시 로드.
- **갱신:** 캐시 파일 삭제 시 API에서 다시 가져옴.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 사용자 캐싱 | JSON 파일 | SQLite보다 가볍고 CGO 의존성 없음, 직접 편집 가능 |
| 사용자 API | `GetUsersPaginated` | 기존 `GetUsers`는 대규모 워크스페이스에서 일부 사용자 누락 |
| 시간 포맷 | `YYYY-MM-DD HH:MM:SS` | Unix Timestamp보다 사람이 읽기 쉬움 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| 일부 사용자 매핑 실패 (`U03G0JYGC3B`) | `GetUsers` API Pagination 미지원 | `GetUsersPaginated`로 변경 및 캐시 갱신 필요 |

### 📁 생성된 파일 목록 (추가)

```
slackExtract/
├── users.json              # 사용자 목록 캐시 (Git 제외 권장)
├── export/
│   └── {채널명}.md         # 추출된 대화 내용
└── docs/
    └── DEV_TIPS.md         # 개발 팁 & 트러블슈팅 가이드
```

### 🎯 다음 작업 (TODO)

1. **스레드(Thread) 다운로드:** 메시지에 달린 댓글(Thread)을 가져와 계층 구조로 표시.
2. **TUI 디자인 개선:** `Lipgloss`를 활용한 컬러 테마 적용으로 시인성 향상.
3. **Slack mrkdwn 변환 강화:** Bold, Italic, Strike, Code Block 등 스타일 변환.
4. **채널 링크 변환:** `<#C12345>` 형태를 `#채널명` 텍스트로 변환.

---

## 템플릿 (복사용)

```markdown
## YYYY-MM-DD (Day N)

### 📋 주요 작업 내용
- 

### 🔧 기술적 결정 사항
- 

### ⚠️ 이슈 및 해결
- 

### 🎯 다음 작업 (TODO)
- 
```
