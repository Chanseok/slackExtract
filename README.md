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

### 설정 (Configuration)
프로젝트 루트에 `.env` 파일을 생성하여 다음 환경 변수를 설정하세요.

```bash
# Slack 인증 (필수)
SLACK_USER_TOKEN=xoxc-...
SLACK_DS_COOKIE=xoxd-...

# 첨부파일 다운로드 여부 (기본값: false - URL만 저장)
DOWNLOAD_ATTACHMENTS=false 

# ============ LLM 분석 설정 (선택) ============

# OpenAI 사용 시
LLM_PROVIDER=openai
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini        # 또는 gpt-4o, gpt-4-turbo 등

# Gemini 사용 시
LLM_PROVIDER=gemini
LLM_API_KEY=AIza...          # Google AI Studio API 키
LLM_MODEL=gemini-1.5-flash   # 또는 gemini-1.5-pro, gemini-pro 등

# 기타 OpenAI 호환 API 사용 시 (예: Azure, Anthropic via proxy)
LLM_PROVIDER=openai
LLM_API_KEY=your-api-key
LLM_BASE_URL=https://your-api-endpoint/v1
```

## 사용법 (Usage)

### 1. 채널 내보내기
```bash
go run cmd/slack-extract/main.go
# 또는 빌드 후
./slack-extract
```
TUI에서 채널을 선택하고 Enter를 누르면 `export/` 폴더에 Markdown 파일이 생성됩니다.

### 2. LLM 분석
```bash
go run cmd/slack-analyze/main.go export/채널명.md
# 또는 빌드 후
./slack-analyze export/general.md
./slack-analyze export/*.md  # 복수 파일 분석
```
분석 결과는 `export/채널명_analysis.md` 파일로 저장됩니다.

**분석 결과에 포함되는 내용:**
- 한국어 종합 요약
- 주요 토픽 및 중요도 점수
- 토픽별 감정 분석 (긍정/부정/중립)
- 주요 기여자 및 참여 통계
