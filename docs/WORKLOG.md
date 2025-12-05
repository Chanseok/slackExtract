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
2. **스마트 내보내기:** 검색어 기반 폴더 생성 및 그룹핑 기능.
3. **Slack mrkdwn 변환 강화:** Bold, Italic, Strike, Code Block 등 스타일 변환.
4. **채널 링크 변환:** `<#C12345>` 형태를 `#채널명` 텍스트로 변환.

---

## 2025-12-05 (Day 1 - TUI 대폭 개선)

### 📋 주요 작업 내용

#### 1. 코드 리팩토링 (모듈화)
- **문제:** 단일 `main.go` 파일에 모든 로직이 집중되어 유지보수성이 떨어짐.
- **해결:** Go 표준 프로젝트 레이아웃에 맞게 `internal/` 패키지로 분리.
  - `internal/config/`: 환경 변수 로드
  - `internal/slack/`: Slack API 클라이언트, 서비스 (채널/사용자 조회, 캐싱)
  - `internal/export/`: Markdown 파일 저장
  - `internal/tui/`: Bubble Tea 모델 및 스타일

#### 2. Lipgloss 스타일링 적용
- **기능:** `charmbracelet/lipgloss` 패키지를 사용하여 TUI에 컬러 및 스타일 적용.
- **구현 내용:**
  - 채널 유형별 색상 구분: 🔓 Public(초록), 🔒 Private(빨강), 💬 DM(노랑), 🗄️ Archived(회색)
  - 선택된 항목 하이라이트 (청록색 + 볼드)
  - 상태바, 헤더, 도움말 바 스타일 적용

#### 3. 채널 정보 표시 개선
- **기능:** 채널 목록에 더 많은 정보 표시.
- **구현 내용:**
  - 채널 유형별 이모지 아이콘 (🔓/🔒/💬/🗄️)
  - 멤버 수 표시 (예: `(12 members)`)
  - 체크박스 스타일 개선 (`[ ]` / `[✓]`)

#### 4. 검색 기능 구현
- **기능:** `/` 키를 눌러 채널 이름으로 실시간 검색.
- **구현 내용:**
  - 검색 모드 진입/종료 (`/` -> 입력 -> `Enter` 또는 `Esc`)
  - 대소문자 구분 없는 검색 (case-insensitive)
  - 검색 결과만 화면에 표시

#### 5. 필터 메뉴 모드 구현
- **기능:** `f` 키를 눌러 채널 속성별 필터링 메뉴 표시.
- **구현 내용:**
  - 4가지 필터 옵션: Public, Private, Archived, DMs
  - `↑↓`로 옵션 탐색, `Space`로 토글
  - `Esc`/`Enter`/`f`로 메뉴 닫기 및 필터 적용
  - 필터 설정에 따라 채널 목록 동적 갱신

#### 6. 전체 선택/해제 기능
- **기능:** `a` 키로 현재 보이는 (필터링된) 채널을 모두 선택/해제.
- **구현 내용:**
  - 모든 visible 채널이 선택된 상태면 -> 전체 해제
  - 하나라도 미선택 상태면 -> 전체 선택

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 스타일링 라이브러리 | Lipgloss | Bubble Tea와 같은 Charm 생태계, 선언적 스타일 정의 |
| 필터 구현 방식 | 필터 메뉴 모드 | 여러 속성 동시 필터링 가능, 확장성 높음 |
| 필터 적용 시점 | 실시간 (토글 즉시) | UX 개선, 결과 바로 확인 가능 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| Archived 색상 Dark Theme에서 안 보임 | `ColorGray(240)` 너무 어두움 | `ColorGray(246)`으로 밝게 조정 |
| 아이콘-이름 사이 공백 없음 | 이모지 뒤 공백 누락 | `GetChannelIcon` 함수에서 공백 추가 |
| Archived만 선택 시 채널 안 보임 | 필터 로직 오류 (Public/Private 체크 선행) | `shouldShowChannel` 로직 수정: 타입 필터 → Archived 필터 순서로 변경 |

### 📁 파일 구조 변경

```
slackExtract/
├── cmd/
│   └── slack-extract/
│       └── main.go         # 진입점 (간소화됨)
├── internal/
│   ├── config/
│   │   └── config.go       # 환경 변수 로드
│   ├── slack/
│   │   ├── client.go       # Slack 클라이언트 초기화
│   │   └── service.go      # 채널/사용자 조회 및 캐싱
│   ├── export/
│   │   └── markdown.go     # Markdown 저장
│   └── tui/
│       ├── model.go        # Bubble Tea 모델
│       └── styles.go       # Lipgloss 스타일 정의
└── docs/
    └── ...
```

### 🎯 다음 작업 (TODO)

1. **스마트 내보내기:** 검색어 기반 폴더 생성, 수동 그룹핑 기능.
2. **스레드 다운로드:** 메시지에 달린 댓글(Thread)을 계층 구조로 가져오기.
3. **다운로드 진행률 표시:** Progress Bar 시각화.

---

## 2025-12-05 (Day 1 - 추가 작업 2)

### 📋 주요 작업 내용

#### 1. Markdown 내보내기 기능 구현
- **패키지 생성:** `internal/export` 패키지 신설.
- **기능:** `SaveToMarkdown` 함수 구현.
  - 선택된 채널의 메시지를 역순(과거->최신)으로 정렬하여 저장.
  - 파일명: `export/{채널명}.md` (특수문자 `/`는 `_`로 치환).
  - 헤더: 채널명 및 내보낸 시간 표시.
  - 메시지 포맷: `### {사용자명} - {시간}\n\n{내용}\n\n---`

#### 2. 사용자 이름 표시 개선
- **문제:** `users.json`에 없는 사용자(탈퇴자 등)나 봇의 경우 이름이 비어있거나 ID로만 표시되어 가독성이 떨어짐.
- **해결:**
  - **기본:** `userMap`에서 실명 조회.
  - **봇:** `BotID`가 있고 `Username`이 있으면 `Username` 사용.
  - **미확인 사용자:** ID만 있는 경우, 가독성을 위해 **"Guy XXXX"** (ID 마지막 4자리) 형식으로 표시.
  - **메시지 본문 멘션:** `<@U...>` 패턴을 찾아 실명 또는 "Guy XXXX"로 치환.

#### 3. 빌드 설정
- **.gitignore:** 빌드된 실행 파일 `slack-extract` 추가.
- **main.go 수정:** `internal/export` 패키지 연동 및 내보내기 로직 연결.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 미확인 사용자 표시 | "Guy XXXX" | ID 전체 노출보다 가독성이 좋고, 식별 가능함 (마지막 4자리 사용) |
| 멘션 치환 | 정규식 (`regexp`) | 메시지 본문 내의 모든 `<@U...>` 패턴을 효율적으로 찾아 변환 |
| 내보내기 경로 | `./export/` | 프로젝트 루트 하위에 별도 폴더로 관리하여 깔끔함 유지 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| `internal/export` 패키지 누락 | 파일 미생성 | `internal/export/export.go` 파일 생성 및 구현 |
| 이름 없는 사용자 발생 | 탈퇴한 사용자 또는 봇 | "Unknown", "Guy XXXX", 봇 이름 등으로 Fallback 로직 강화 |
| 멘션 치환 미적용 | 단순 텍스트 출력 | `regexp`를 사용하여 본문 내 ID를 이름으로 치환하는 로직 추가 |

### 🎯 다음 작업 (TODO)

1. **스레드(Thread) 지원:** 메시지의 댓글까지 포함하여 내보내기.
2. **첨부파일 다운로드:** 이미지, 파일 등을 로컬로 다운로드하고 링크 연결.
3. **진행률 표시:** 대량의 메시지 다운로드 시 진행 상황을 TUI에 표시.

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
