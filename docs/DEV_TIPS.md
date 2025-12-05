# 개발 팁 & 트러블슈팅 (Development Tips & Troubleshooting)

이 문서는 `slackExtract` 프로젝트를 개발하고 실행할 때 유용한 팁과 자주 발생하는 문제에 대한 해결 방법을 정리합니다.

---

## 1. 개발 환경 설정 (Setup)

### 필수 요구사항
- **Go:** 1.25.5 이상
- **Slack 인증 정보:** `.env` 파일 설정 필요

### 의존성 설치
프로젝트를 처음 클론하거나 `go.mod`가 변경되었을 때 실행합니다.

```bash
go mod tidy
```

### 환경 변수 설정 (.env)
프로젝트 루트에 `.env` 파일을 생성하고 아래 내용을 채워주세요. 이 방식은 `slackdump` 도구에서 사용하는 인증 방식과 동일하며, 관리자 권한 없이 개인 계정 권한으로 데이터를 추출하기 위해 사용됩니다.

```ini
SLACK_USER_TOKEN=xoxc-...
SLACK_DS_COOKIE=xoxd-...
```

#### 토큰 및 쿠키 추출 방법
1. 웹 브라우저(Chrome 등)에서 Slack 워크스페이스에 로그인합니다.
2. 개발자 도구(`F12`)를 엽니다.
3. **Application** 탭으로 이동합니다.
4. 왼쪽 사이드바에서 **Cookies** > `https://app.slack.com` (또는 워크스페이스 URL)을 선택합니다.
5. `d` 라는 이름의 쿠키 값을 찾아 복사하여 `SLACK_DS_COOKIE`에 입력합니다. (반드시 `xoxd-`로 시작해야 하며, URL 인코딩된 값이어도 상관없습니다.)
6. **Network** 탭으로 이동합니다.
7. 필터 창에 `client.counts`를 입력하고, Slack 화면을 새로고침하거나 다른 채널을 클릭합니다.
8. 목록에 나타난 요청 중 하나를 클릭하고 **Payload** 탭(또는 **Form Data**)에서 `token` 항목을 찾습니다.
   - 값은 `xoxc-`로 시작해야 합니다. 이를 복사하여 `SLACK_USER_TOKEN`에 입력합니다.

---

## 2. 실행 및 빌드 (Run & Build)

### 소스 코드 바로 실행
개발 중에는 `go run` 명령어를 사용하여 빠르게 실행해 볼 수 있습니다.

```bash
go run cmd/slack-extract/main.go
```

### 실행 파일 빌드
배포하거나 반복적으로 실행할 때는 바이너리로 빌드하는 것이 좋습니다.

```bash
# 빌드
go build -o slack-extract cmd/slack-extract/main.go

# 실행
./slack-extract
```

---

## 3. 트러블슈팅 (Troubleshooting)

### Q1. `compile: version "..." does not match go tool version "..."` 오류가 발생해요.

**원인:**
Go 버전이 업데이트되었거나, 이전에 빌드된 캐시 파일(`go-build`)의 버전과 현재 설치된 Go 툴체인의 버전이 일치하지 않을 때 발생합니다. 특히 Dev Container 환경에서 자주 발생할 수 있습니다.

**해결 방법:**
아래 명령어들을 순서대로 시도해 보세요.

1. **빌드 캐시 삭제 및 강제 재빌드 (권장)**
   ```bash
   go clean -cache
   go run -a cmd/slack-extract/main.go
   ```
   - `-a` 옵션은 모든 패키지를 강제로 다시 컴파일합니다.

2. **로컬 툴체인 강제 사용**
   ```bash
   export GOTOOLCHAIN=local
   go run cmd/slack-extract/main.go
   ```

3. **모듈 캐시까지 삭제 (최후의 수단)**
   ```bash
   go clean -modcache
   go mod tidy
   go run -a cmd/slack-extract/main.go
   ```
   > **주의:** `go clean -modcache` 실행 시 툴체인 파일이 삭제되어 `go` 명령어가 작동하지 않을 수 있습니다. 이 경우 `export GOTOOLCHAIN=local`을 먼저 실행하거나, Dev Container를 Rebuild 해야 합니다.

### Q2. 터미널에서 채널 목록이 잘려서 보여요.

**원인:**
채널이 매우 많은 경우 터미널 화면 높이를 초과하여 윗부분이 잘릴 수 있습니다.

**해결 방법:**
최신 버전의 코드는 **스크롤 기능**과 **전체 화면 모드(AltScreen)** 를 지원합니다.
- `go run`으로 최신 코드를 실행하세요.
- 방향키 `↑`, `↓` 또는 `k`, `j` 키를 사용하여 목록을 스크롤할 수 있습니다.

### Q3. Slack 앱에 보이는 채널과 목록이 달라요.

**원인:**
- **페이지네이션:** 채널이 1000개를 넘는 경우, 초기 코드에서는 1000개까지만 가져왔었습니다. (현재는 수정됨)
- **참여 여부:** Slack 앱의 사이드바는 "내가 참여한 채널" 위주로 보여주지만, API(`GetConversations`)는 기본적으로 "모든 공개 채널"을 가져옵니다. 따라서 내가 참여하지 않은 채널도 목록에 포함될 수 있습니다.

**해결 방법:**
- 현재 코드는 모든 채널을 가져오도록 수정되었습니다.
- 원하는 채널이 보이지 않는다면 스크롤을 내려보거나, 정렬 순서(이름순)를 확인해 보세요.
