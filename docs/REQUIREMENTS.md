# Slack Conversation Extractor (CLI)

## 1. 개요 (Overview)
Slack의 대화 내용을 채널별로 추출하여 Markdown 파일로 저장하는 개인용 Interactive CLI 도구입니다.
기존 오픈소스인 `slackdump`의 강력한 데이터 수집 능력을 참고하고, Go 언어와 Bubble Tea(TUI)를 사용하여 사용자 친화적인 인터페이스를 제공합니다.

## 2. 기능적 요구사항 (Functional Requirements)

### 2.1 인증 (Authentication)
- [ ] **User Token 지원:** `xoxp-`로 시작하는 사용자 토큰을 입력받아 인증한다.
- [ ] **보안:** 토큰은 소스코드에 저장하지 않으며, 환경변수(`.env`) 또는 로컬 보안 저장소를 통해 관리한다.

### 2.2 채널 탐색 (Discovery)
- [ ] **채널 목록 조회:** Public Channels, Private Channels, Direct Messages(MPIM/IM) 목록을 가져온다.
- [ ] **Interactive Selection:** CLI 화면에서 화살표 키로 추출할 채널을 다중 선택(Multi-select)하거나 전체 선택할 수 있다.

### 2.3 데이터 추출 (Extraction)
- [ ] **메시지 수집:** 선택된 채널의 메시지 히스토리를 가져온다.
- [ ] **스레드(Thread) 지원:** 각 메시지에 달린 댓글(Thread)을 빠짐없이 가져와 계층 구조를 유지한다.
- [ ] **Rate Limit 처리:** Slack API의 속도 제한을 준수하여 차단되지 않도록 요청 속도를 조절한다.

### 2.4 변환 및 저장 (Conversion & Storage)
- [ ] **Markdown 변환:** Slack 전용 포맷(mrkdwn)을 표준 Markdown으로 변환한다.
    - 사용자 멘션 (@User) -> 이름으로 치환
    - 채널 링크 (#Channel) -> 텍스트로 치환
    - 스타일 (Bold, Italic, Strike, Code) 유지
- [ ] **파일 저장:**
    - 구조: `./export/{YYYY-MM-DD_HHMM}/{ChannelName}.md` (예시)
    - (옵션) 이미지/첨부파일 다운로드 및 로컬 링크 처리

## 3. 비기능적 요구사항 (Non-Functional Requirements)

### 3.1 성능 및 효율성
- [ ] **경량화:** 실행 파일 크기를 최소화한다 (Go 언어 특성 활용).
- [ ] **메모리 효율:** 대용량 대화 로그 처리 시 메모리 사용량을 최적화한다.
- [ ] **멀티 플랫폼:** macOS와 Windows에서 동일하게 동작하는 단일 바이너리를 제공한다.

### 3.2 사용자 경험 (UX)
- [ ] **TUI (Text User Interface):** `Bubble Tea` 라이브러리를 활용하여 미려하고 직관적인 CLI 환경을 제공한다.
- [ ] **진행 상황 표시:** 다운로드 및 변환 진행률(Progress Bar)을 시각적으로 보여준다.

### 3.3 개발 및 유지보수
- [ ] **언어:** Go (Golang)
- [ ] **VCS:** GitHub Public Repo (민감 정보 제외 필수)
- [ ] **구조:** Go 표준 프로젝트 레이아웃 준수 (`cmd`, `internal`, `pkg`)
