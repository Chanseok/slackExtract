# Slack Extract

Slack의 대화 내용을 안전하게 로컬 Markdown 파일로 백업하는 Interactive CLI 도구입니다.

## 특징
- **Interactive CLI:** 터미널에서 화살표 키로 간편하게 채널을 선택하고 백업할 수 있습니다.
- **Markdown Export:** Slack의 독자적인 포맷을 읽기 쉬운 표준 Markdown으로 변환합니다.
- **Private & Threads:** 내가 참여한 비공개 채널과 스레드 댓글까지 모두 수집합니다.
- **Secure:** 모든 데이터는 로컬에만 저장되며, 토큰은 안전하게 관리됩니다.
- **Cross Platform:** Go 언어로 작성되어 macOS와 Windows에서 단일 실행 파일로 동작합니다.

## 개발 환경 (Dev Environment)
- **Language:** Go 1.21+
- **Runtime:** Native Binary
- **Key Libraries:**
    - [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI)
    - [Slack Go SDK](https://github.com/slack-go/slack) (Official SDK) or Custom implementation based on `slackdump`

## 시작하기 (Getting Started)

### 요구사항
- Go 1.21 이상 설치
- Slack User Token (`xoxp-...`)

### 설치 및 실행
```bash
git clone https://github.com/chanseok/slackExtract.git
cd slackExtract
go mod tidy
go run cmd/slack-extract/main.go
```
