# Slack Conversation Extractor (CLI)

## 1. 개요 (Overview)
Slack의 대화 내용을 채널별로 추출하여 Markdown 파일로 저장하는 개인용 Interactive CLI 도구입니다.
기존 오픈소스인 `slackdump`의 강력한 데이터 수집 능력을 참고하고, Go 언어와 Bubble Tea(TUI)를 사용하여 사용자 친화적인 인터페이스를 제공합니다.

## 2. 기능적 요구사항 (Functional Requirements)

### 2.1 인증 (Authentication)
- [x] **Cookie 인증 지원:** Slack의 보안 정책 강화로 인해 User Token(`xoxp-`) 발급이 어려워짐에 따라, 브라우저 쿠키(`d` cookie)와 클라이언트 토큰(`xoxc-`)을 이용한 인증 방식을 사용한다.
- [x] **보안:** 토큰과 쿠키는 소스코드에 저장하지 않으며, `.env` 파일을 통해 로컬에서만 관리한다. (`.gitignore`에 `.env` 포함됨)

### 2.2 채널 탐색 및 관리 (Discovery & Management)
- [x] **채널 목록 조회:** Public Channels, Private Channels, Direct Messages(MPIM/IM) 목록을 가져온다.
- [x] **Interactive Selection:** CLI 화면에서 화살표 키로 추출할 채널을 다중 선택(Multi-select)하거나 전체 선택할 수 있다.
- [x] **로컬 캐싱 (Channel Caching):** 채널 목록 및 속성(Archived 여부, 멤버 수 등)을 `channels.json`에 캐싱하여 실행 속도를 높이고 API 호출을 최소화한다.
- [x] **필터링:** 필터 메뉴(`f` 키)를 통해 채널 유형별(Public/Private/Archived/DM) 표시 여부를 토글할 수 있다.
- [x] **검색:** `/` 키를 눌러 채널 이름으로 실시간 검색할 수 있다.

### 2.3 데이터 추출 (Extraction)
- [x] **메시지 수집:** 선택된 채널의 메시지 히스토리를 가져온다. (`GetConversationHistory` + Pagination)
- [ ] **스레드(Thread) 지원:** 각 메시지에 달린 댓글(Thread)을 빠짐없이 가져와 계층 구조를 유지한다.
- [ ] **Rate Limit 처리:** Slack API의 속도 제한을 준수하여 차단되지 않도록 요청 속도를 조절한다.

### 2.4 변환 및 저장 (Conversion & Storage)
- [x] **Markdown 변환:** Slack 전용 포맷(mrkdwn)을 표준 Markdown으로 변환한다.
    - [x] 사용자 멘션 (@User) -> 이름으로 치환 (미확인 시 "Guy XXXX"로 표시)
    - [ ] 채널 링크 (#Channel) -> 텍스트로 치환
    - [ ] 스타일 (Bold, Italic, Strike, Code) 유지
- [x] **파일 저장:**
    - 구조: `./export/{ChannelName}.md`
    - (옵션) 이미지/첨부파일 다운로드 및 로컬 링크 처리

### 2.5 사용자 관리 (User Management)
- [x] **로컬 캐싱:** 사용자 목록을 `users.json` 파일에 캐싱하여 API 호출을 최소화한다.
- [x] **Pagination 지원:** 대규모 워크스페이스의 모든 사용자를 빠짐없이 가져온다.
- [ ] **수동 갱신:** CLI 명령어 또는 캐시 파일 삭제를 통해 사용자 목록을 갱신할 수 있다.

## 3. 비기능적 요구사항 (Non-Functional Requirements)

### 3.1 성능 및 효율성
- [ ] **경량화:** 실행 파일 크기를 최소화한다 (Go 언어 특성 활용).
- [ ] **메모리 효율:** 대용량 대화 로그 처리 시 메모리 사용량을 최적화한다.
- [ ] **멀티 플랫폼:** macOS와 Windows에서 동일하게 동작하는 단일 바이너리를 제공한다.

### 3.2 사용자 경험 (UX)
- [x] **TUI (Text User Interface):** `Bubble Tea` 라이브러리를 활용하여 미려하고 직관적인 CLI 환경을 제공한다.
- [x] **Lipgloss 스타일링:** 채널 유형별 색상 구분, 선택 상태 하이라이트 등 시각적 개선 적용.
- [x] **필터 메뉴 모드:** `f` 키로 필터 설정 메뉴를 열어 채널 속성별 표시 여부를 토글할 수 있다.
- [ ] **진행 상황 표시:** 다운로드 및 변환 진행률(Progress Bar)을 시각적으로 보여준다.

### 3.3 아키텍처 및 코드 품질 (Architecture & Code Quality)
- [ ] **Standard Go Project Layout:** Go 표준 프로젝트 구조(`cmd`, `internal`, `pkg`)를 준수하여 코드의 모듈화와 재사용성을 높인다.
    - `cmd/`: 메인 애플리케이션 진입점
    - `internal/`: 비즈니스 로직 (외부에서 import 불가)
    - `pkg/`: 라이브러리 코드로 사용 가능한 패키지 (선택 사항)
- [ ] **Clean Code:** 함수 분리, 명확한 변수명, 주석 작성을 통해 가독성을 유지한다.
- [ ] **Error Handling:** 모든 에러를 명시적으로 처리하고, 사용자에게 친절한 에러 메시지를 제공한다.
- [ ] **Configuration Management:** 설정(Config)과 코드(Code)를 분리하여 유지보수성을 높인다.

```
