# 작업 일지 (Work Log)

> 이 문서는 개발 과정에서의 주요 작업 내용, 의사 결정, 이슈 및 해결 방법을 기록합니다.
> `REQUIREMENTS.md`, `PLAN.md`와 함께 참고하여 프로젝트 진행 상황을 파악할 수 있습니다.

---

## 2025-12-06 (Day 2)

### 📋 주요 작업 내용

#### 1. 메타데이터 시스템 구축 (Phase 9)
- **목표:** 채널별 다운로드 및 분석 상태를 체계적으로 관리하여 중복 작업을 방지하고 이력을 추적.
- **구현:**
  - `internal/meta` 패키지 신설.
  - `export/.meta/index.json` 파일에 전체 채널의 메타데이터(ID, 이름, 경로, 메시지 수, 마지막 다운로드/분석 시간) 저장.
  - `slack-extract`: 다운로드 완료 시 메타데이터 업데이트 (`UpdateChannelDownload`).
  - `slack-analyze`: 분석 완료 시 메타데이터 업데이트 (`UpdateChannelAnalysis`).

#### 2. LLM 비용 추적 시스템 (Phase 10)
- **목표:** LLM API 사용량(토큰)과 예상 비용을 추적하여 비용 효율적인 운영 지원.
- **구현:**
  - `internal/llm/client.go`: OpenAI 및 Gemini API 응답에서 `Usage` 정보(Prompt/Completion Tokens) 추출.
  - `internal/llm/cost.go`: 모델별 단가(Pricing) 정의 및 비용 계산 로직 구현.
  - **리포팅:** 분석 결과 리포트(`_analysis.md`) 상단에 예상 비용 및 토큰 사용량 표시.

#### 3. 안전한 캐시 갱신 (Safe Cache Refresh)
- **문제:** `users.json`에 수동으로 입력한 사용자 이름(예: "Guy XXXX" -> "홍길동")이 API 재호출 시 빈 값으로 덮어씌워지는 문제 발생.
- **해결:**
  - `slack-extract`에 `--refresh` 플래그 추가.
  - `FetchUsers` 로직 개선: API에서 가져온 이름이 비어있지 않은 경우에만 기존 캐시를 업데이트하는 **Safe Refresh** 정책 적용.

#### 4. 일괄 분석 및 중복 방지 (Batch Analysis)
- **기능:** `slack-analyze` 실행 시 파일 경로 대신 폴더 경로를 입력하면, 해당 폴더 내 모든 `.md` 파일을 재귀적으로 찾아 분석.
- **최적화:** 분석 시작 전, 이미 분석 결과 파일이 존재하는지 확인하고 있다면 건너뛰는(Skip) 로직 추가.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 메타데이터 저장소 | JSON 파일 (`index.json`) | 별도 DB 없이 파일 시스템 기반으로 가볍게 관리, Git으로 이력 관리 가능 |
| 비용 계산 | 클라이언트 측 계산 | API가 비용을 직접 주지 않으므로, 공개된 단가표를 기반으로 추정치 계산 |
| 사용자 이름 보존 | Safe Merge | API 데이터가 항상 최신/정확하지 않을 수 있음(특히 탈퇴자). 수동 보정 데이터 우선 정책 |

### 🎯 다음 작업 (TODO)

1. **스마트 다운로드 고도화:** 폴더별 자동 분류 및 아카이브 채널 관리 강화.
2. **TUI UX 개선:** 메타데이터를 활용하여 TUI 상에서 다운로드/분석 상태 아이콘 표시.
3. **검색 기능 강화:** 메타데이터 기반의 오프라인 검색 지원.

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

1. **첨부파일 다운로드:** 이미지, 파일 등을 로컬로 다운로드하고 링크 연결.
2. **진행률 표시:** 대량의 메시지 다운로드 시 진행 상황을 TUI에 표시.
3. **스마트 내보내기:** 검색어 기반 폴더 생성, 수동 그룹핑 기능.

---

## 2025-12-05 (Day 1 - 추가 작업 3)

### 📋 주요 작업 내용

#### 1. 스레드(Thread) 다운로드 및 내보내기 구현
- **구조체 변경:** `internal/slack` 패키지에 `Message` 구조체를 정의하여 `slack.Message`를 임베딩하고 `Replies []slack.Message` 필드를 추가.
- **다운로드 로직:** `FetchHistory` 함수에서 메시지의 `ReplyCount > 0`인 경우 `GetConversationReplies` API를 호출하여 댓글을 가져오도록 수정.
- **내보내기 로직:** `internal/export` 패키지에서 댓글이 있는 경우 인용구(`> `) 형식으로 들여쓰기하여 Markdown에 출력하도록 구현.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 스레드 저장 구조 | `Message` 구조체 내 `Replies` 필드 | 메시지와 댓글을 계층적으로 관리하여 내보내기 시 순서 유지 용이 |
| 스레드 표시 방식 | 인용구 (`> `) | Markdown 표준 문법을 사용하여 원본 메시지와 시각적으로 구분 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| `internal/export` 컴파일 에러 | 중복된 코드 붙여넣기 | `replace_string_in_file` 실수로 인한 중복 코드 제거 |

### 🎯 다음 작업 (TODO)

1. **진행률 표시:** 대량의 메시지 다운로드 시 진행 상황을 TUI에 표시.
2. **스마트 내보내기:** 검색어 기반 폴더 생성, 수동 그룹핑 기능.

---

## 2025-12-05 (Day 1 - 추가 작업 4)

### 📋 주요 작업 내용

#### 1. 첨부파일 다운로드 구현
- **기능:** 메시지에 포함된 파일(이미지, 문서 등)을 로컬로 다운로드하고 Markdown에 링크 연결.
- **저장 경로:** `export/attachments/{채널명}/{FileID}_{FileName}` 구조로 저장하여 파일명 충돌 방지.
- **Markdown 처리:**
  - 이미지는 `![Name](Path)` 형식으로 미리보기 표시.
  - 기타 파일은 `[Name](Path)` 형식으로 다운로드 링크 표시.
- **인증 처리:** `slack.NewClient`에서 생성한 `http.Client` (쿠키 포함)를 `export` 패키지로 전달하여 Private URL(`URLPrivateDownload`) 접근 권한 확보.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| 파일 저장 위치 | `export/attachments/{Channel}` | 채널별로 파일을 분리하여 관리 용이성 확보 |
| 파일명 규칙 | `{FileID}_{FileName}` | 동일한 파일명이 있을 경우 덮어쓰기 방지 |
| HTTP 클라이언트 재사용 | `main` -> `export` 전달 | 쿠키 인증 정보를 유지하여 별도 인증 없이 파일 다운로드 가능 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| `export` 패키지 컴파일 에러 | 중복된 함수 선언 및 import 누락 | `create_file`로 파일 전체를 올바르게 덮어써서 해결 |

### 🎯 다음 작업 (TODO)

1. **진행률 표시:** 대량의 메시지 다운로드 시 진행 상황을 TUI에 표시.
2. **스마트 내보내기:** 검색어 기반 폴더 생성, 수동 그룹핑 기능.

---

## 2025-12-05 (Day 1 - Continued)

### 📋 주요 작업 내용

#### 1. Progress Bar 구현
- **목표:** 대량의 메시지 다운로드 시 사용자에게 진행 상황과 예상 시간(ETA)을 시각적으로 제공.
- **구현:**
  - `internal/tui/download.go`: 별도의 Goroutine에서 다운로드를 수행하고 `ProgressMsg`를 통해 UI 업데이트.
  - `internal/tui/model.go`: 진행 상태(현재/전체 개수, 상태 메시지, 시작 시간 등) 필드 추가 및 `View` 렌더링 로직 개선.
  - `internal/slack/service.go`: `FetchHistoryWithProgress` 함수 추가 (콜백 함수를 통해 진행률 보고).
- **결과:** 채널별 진행률 바, 처리 중인 항목 수, ETA가 실시간으로 표시됨.

#### 2. 첨부파일 다운로드 옵션화
- **요구사항:** LLM 번역/요약 목적을 위해 기본적으로는 URL만 저장하고, 필요시에만 파일을 다운로드하도록 변경.
- **구현:**
  - `.env`에 `DOWNLOAD_ATTACHMENTS` 옵션 추가 (기본값: `false`).
  - `internal/config`: 설정 로드 시 해당 옵션 파싱.
  - `internal/export`: 옵션이 `false`일 경우 파일 다운로드를 건너뛰고 원본 URL(`URLPrivate`)을 Markdown에 기록.
- **효과:** 텍스트 추출 속도 향상 및 디스크 용량 절약.

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| Progress UI | 별도 View 모드 전환 | 기존 목록 화면 위에 오버레이하거나 별도 화면으로 전환하여 집중도 향상 |
| 비동기 처리 | `tea.Cmd` + Goroutine | UI 블로킹 없이 긴 작업을 수행하고 메시지로 상태 업데이트 |
| 첨부파일 기본값 | URL Only | LLM 활용 시 텍스트 처리가 우선이며, 파일 다운로드는 선택적 기능으로 전환 |

---

## 2025-12-06 (Day 2)

### 📋 주요 작업 내용

#### 1. 텍스트 정제 (Text Cleaning) 기능 구현
- **목표:** LLM 학습/분석 효율을 높이기 위해 Slack 특유의 포맷을 표준화된 형태로 변환.
- **구현 (`internal/export/cleaner.go`):**
  - `CleanSlackText()` 함수: 채널 링크(`<#C123|general>` → `#general`), URL 포맷(`<url|text>` → `[text](url)`), HTML 엔티티 디코딩(`&lt;`, `&gt;`, `&amp;`)
  - `IsSystemMessage()` 함수: 입/퇴장 메시지 등 시스템 메시지 필터링
- **효과:** LLM에 입력할 텍스트의 노이즈 감소, 토큰 효율성 향상.

#### 2. Rate Limit 자동 재시도 구현
- **목표:** Slack API의 429 Too Many Requests 오류 발생 시 자동 재시도.
- **구현 (`internal/slack/retry.go`):**
  - `withRetry[T]()` 제네릭 함수: Exponential Backoff 알고리즘 적용
  - 최대 5회 재시도, 초기 대기 1초 → 2초 → 4초 → 8초 → 16초
  - `Retry-After` 헤더가 있으면 해당 시간만큼 대기
- **적용:** `FetchHistory`, `FetchThreadReplies`, `FetchUsers` 등 모든 API 호출에 적용.

#### 3. LLM 분석 파이프라인 구현 (Phase 6)
- **목표:** 추출된 Slack 대화를 LLM으로 분석하여 다국어 번역, Topic 분석, 의견 분석, 인물 분석을 수행.
- **구현:**
  - `internal/llm/client.go`: OpenAI API 클라이언트 (GPT-4, GPT-4o 지원)
  - `internal/llm/analyzer.go`: 채널 분석 로직 (`Analyze()` 메인 함수)
    - `translateMessages()`: 비영어 메시지를 영어로 번역 (원문 병기)
    - `extractTopics()`: 주요 Topic 자동 추출 및 중요도 점수 산정
    - `analyzeSentiment()`: 긍정/부정/중립 의견 분류
    - `analyzeContributors()`: 인물별 참여도 통계
    - `generateKoreanSummary()`: 최종 요약을 한국어로 생성
  - `internal/llm/report.go`: Markdown 보고서 생성기
- **CLI 도구:** `cmd/slack-analyze/main.go` - 추출된 Markdown 파일을 분석하는 CLI
- **결과 포맷:** `export/{채널명}_analysis.md` 형태로 분석 결과 저장.


#### 4. Google Gemini API 직접 지원 (Phase 7)
- **배경:** OpenAI 외에 Google Gemini API도 지원하여 사용자 선택의 폭을 넓힘.
- **구현 (`internal/llm/client.go`):**
  - `Provider` 타입 추가: `openai`, `gemini`
  - `chatGemini()` 함수: Gemini API 직접 호출
    - OpenAI 메시지 포맷을 Gemini 포맷으로 변환
    - `system` role → `systemInstruction` 필드
    - `user`/`assistant` → `user`/`model` role 매핑
  - 자동 Provider 감지: `LLM_PROVIDER` 환경 변수 기반
- **지원 모델:**
  - OpenAI: `gpt-4`, `gpt-4o`, `gpt-3.5-turbo`
  - Gemini: `gemini-1.5-flash`, `gemini-1.5-pro`

### 🔧 기술적 결정 사항

| 항목 | 결정 | 이유 |
|------|------|------|
| LLM 클라이언트 | 자체 구현 | 외부 SDK 의존성 최소화, 두 Provider 통합 관리 용이 |
| Gemini 지원 방식 | 직접 API 호출 | OpenAI 호환 프록시 불필요, 네이티브 기능 활용 |
| 분석 결과 포맷 | Markdown | 가독성 높음, 다른 도구에서 활용 용이 |
| 한국어 요약 | 별도 프롬프트 | 번역 품질 향상을 위해 단계적 처리 |

### ⚠️ 이슈 및 해결

| 이슈 | 원인 | 해결 |
|------|------|------|
| Gemini system role 미지원 | Gemini API는 `system` role 직접 미지원 | `systemInstruction` 필드로 분리하여 전달 |
| 대량 메시지 분석 시 토큰 초과 | 긴 대화 기록 전체 전송 불가 | 메시지 청크 분할 및 요약 후 통합 전략 적용 |

### 📁 새로 생성된 파일

```
internal/
├── llm/
│   ├── client.go      # LLM API 클라이언트 (OpenAI + Gemini)
│   ├── analyzer.go    # 채널 분석 로직
│   └── report.go      # Markdown 보고서 생성
├── export/
│   └── cleaner.go     # 텍스트 정제 함수
└── slack/
    └── retry.go       # Rate Limit 재시도 로직

cmd/
└── slack-analyze/
    └── main.go        # LLM 분석 CLI 도구
```

### 🎯 완료된 Phase 요약

| Phase | 상태 | 주요 기능 |
|-------|------|----------|
| Phase 1 | ✅ 완료 | 환경 설정, 요구사항 정의 |
| Phase 2 | ✅ 완료 | TUI 프로토타입, 채널 목록 조회 |
| Phase 3 | ✅ 완료 | 메시지/스레드 다운로드, 사용자 매핑 |
| Phase 3.5 | ✅ 완료 | 사용자/채널 캐싱 |
| Phase 4 | ✅ 완료 | 텍스트 정제, Rate Limit 처리 |
| Phase 5 | ✅ 완료 | Progress Bar, 첨부파일 옵션 |
| Phase 6 | ✅ 완료 | LLM 분석 파이프라인 |
| Phase 7 | ✅ 완료 | Multi-Provider 지원 (OpenAI + Gemini) |

### 🎯 향후 작업 (TODO)

1. **추가 LLM Provider:** Anthropic Claude API 지원 검토
2. **스트리밍 응답:** 긴 분석 결과의 점진적 표시
3. **비용 추적:** API 호출별 토큰 사용량 및 비용 추정
4. **GitHub Release:** 바이너리 배포 자동화

---

## 2025-12-06 (Day 2 - Continued)

### 📋 주요 작업 내용

#### 1. Gemini 2.5 Flash 모델 연동 및 디버깅
- **이슈:** `gemini-2.5-flash` 모델 사용 시 "no response from Gemini" 에러 발생.
- **원인 분석:**
  - `internal/llm/client.go`의 에러 로깅을 강화하여 확인한 결과, `FinishReason: MAX_TOKENS` 에러 발생.
  - Gemini 2.5 Flash는 "Reasoning" 모델로, 내부적인 사고 과정(Thoughts)에 많은 토큰을 소모함.
  - 기존의 기본 토큰 제한(약 2000~4000)으로는 분석 결과를 생성하기에 부족했음.
- **해결:** `internal/llm/analyzer.go`에서 `MaxOutputTokens` 설정을 16,000으로 대폭 상향 조정.
- **결과:** `project-lg-zigbee-liveness.md` 및 `project-lg-thinq-on.md` 파일 분석 성공.

#### 2. 분석 결과 파일 구조 재정비
- **문제:** 분석 결과 파일(`_analysis.md`)이 원본 파일과 동일한 폴더에 생성되어 관리가 어려움.
- **결정:** `docs/DESIGN_SMART_DOWNLOAD.md`의 설계에 따라 분석 결과를 별도 폴더로 분리하기로 함.
- **조치:**
  - 기존 생성된 파일들을 `export/.analysis/{channel_name}/analysis.md` 경로로 이동.
  - 향후 `slack-analyze` 도구도 이 경로에 파일을 생성하도록 수정 예정.

#### 3. Smart Download 설계 검토
- **검토:** `docs/DESIGN_SMART_DOWNLOAD.md` 문서를 재검토하여 현재 구현과의 차이점 확인.
- **계획:** 분석 결과 파일 경로 수정 및 Smart Download 기능(중복 방지, 폴더 선택 등) 구현 착수.

#### 4. Smart Download 기능 구현 (Phase 8)
- **기존 파일 감지:** `internal/manager/scanner.go` 구현. `export/` 폴더를 재귀적으로 스캔하여 기존 다운로드된 파일의 메타데이터(크기, 마지막 메시지 시간 등)를 수집.
- **다운로드 확인 TUI:** `internal/tui/confirm.go` 구현.
  - 채널 선택 후 Enter 입력 시 확인 화면으로 전환.
  - 저장 폴더 선택 (기존 폴더 또는 새 폴더 생성).
  - 기존 파일 충돌 감지 및 경고 표시.
  - 액션 선택: Skip, Incremental, Overwrite, Cancel.
- **증분 다운로드 (Incremental):**
  - `internal/slack/retry.go`에 `FetchHistoryWithRetryAndProgress` 함수 추가.
  - `oldest` 파라미터를 지원하여 기존 파일의 마지막 메시지 이후 데이터만 가져오도록 구현.
- **폴더 구조 지원:** `export/{category}/{channel}.md` 형태의 중첩 폴더 구조 지원.

#### 5. 버그 수정: 증분 다운로드 시 파일 미생성 이슈
- **증상:** `Incremental` 모드로 다운로드 시, 대상 폴더에 파일이 없으면 에러가 발생하거나 파일이 생성되지 않음.
- **원인:** `internal/export/export.go`에서 `appendMode`가 `true`일 경우 `os.OpenFile`을 `os.O_APPEND` 플래그로만 호출하여, 파일이 없으면 에러(`no such file or directory`) 발생.
- **해결:** `appendMode`여도 파일이 존재하지 않으면 `appendMode = false`로 전환하여 새 파일을 생성(`os.Create`)하도록 로직 수정.

