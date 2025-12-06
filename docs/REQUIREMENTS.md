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
- [x] **스레드(Thread) 지원:** 각 메시지에 달린 댓글(Thread)을 빠짐없이 가져와 계층 구조를 유지한다.
- [x] **Rate Limit 처리:** Slack API의 속도 제한을 준수하여 차단되지 않도록 Exponential Backoff 재시도 로직 적용.

### 2.4 변환 및 저장 (Conversion & Storage)
- [x] **Markdown 변환:** Slack 전용 포맷(mrkdwn)을 표준 Markdown으로 변환한다.
    - [x] 사용자 멘션 (@User) -> 이름으로 치환 (미확인 시 "Guy XXXX"로 표시)
    - [x] **텍스트 정제 (Text Cleaning):**
        - 채널 링크 (`<#C123|general>`) -> 텍스트(`#general`)로 치환
        - URL 링크 (`<http://...|Text>`) -> Markdown 링크(`[Text](http://...)`)로 치환
        - HTML 엔티티 (`&lt;`, `&gt;`, `&amp;`) 디코딩
    - [x] **LLM 최적화:**
        - 시스템 메시지(입/퇴장 등) 필터링
        - 날짜별 헤더 그룹핑
- [x] **파일 저장:**
    - 구조: `./export/{ChannelName}.md`
    - (옵션) 이미지/첨부파일 다운로드 및 로컬 링크 처리
        - 기본값: URL만 저장 (LLM 활용 효율성 증대)
        - 설정: `DOWNLOAD_ATTACHMENTS=true` 시 파일 다운로드

### 2.5 사용자 관리 (User Management)
- [x] **로컬 캐싱:** 사용자 목록을 `users.json` 파일에 캐싱하여 API 호출을 최소화한다.
- [x] **Pagination 지원:** 대규모 워크스페이스의 모든 사용자를 빠짐없이 가져온다.
- [x] **수동 갱신:** `--refresh` 플래그를 통해 사용자 목록을 갱신할 수 있다.
    - **Safe Refresh:** 기존에 수동으로 입력된 이름이 있고 API 응답이 비어있으면 덮어쓰지 않고 보존한다.

### 2.6 LLM 분석 (LLM Analysis)
- [x] **다국어 처리:**
    - 언어 자동 감지 (영어, 네덜란드어, 기타)
    - 비영어 메시지는 영어 번역과 원문을 함께 병기
    - 최종 분석 결과는 한국어로 출력
- [x] **Topic 분석:**
    - 주요 Topic 자동 추출 및 클러스터링
    - 중요도 점수 산정 기준: 스레드 길이, 이모지 반응 수, 특정 키워드("urgent", "important", "blocker" 등)
    - Topic별 핵심 요약 생성
- [x] **의견 분석:**
    - 긍정/부정/중립 의견 분류 (Sentiment Analysis)
    - 주요 논쟁점 및 합의점 하이라이트
- [x] **인물 분석:**
    - 인물별 Topic 참여도 및 발언 빈도 통계
    - 주요 기여자(Key Contributors) 식별

### 2.7 LLM Provider 지원 (Multi-Provider)
- [x] **OpenAI API:** GPT-4, GPT-4o, GPT-3.5-turbo 지원
- [x] **Google Gemini API:** Gemini 1.5 Flash, Gemini 1.5 Pro 지원
- [x] **설정:** `.env` 파일을 통한 Provider 선택 (`LLM_PROVIDER=openai|gemini`)

### 2.8 스마트 다운로드 (Smart Download) - 완료
> 상세 설계: `docs/DESIGN_SMART_DOWNLOAD.md`

- [x] **다운로드 전 확인:** 채널 선택 후 즉시 다운로드하지 않고 확인 화면 표시
- [x] **저장 폴더 선택:** 기존 하위 폴더 선택 또는 새 폴더 생성
- [x] **기존 파일 감지:** 이미 다운로드된 채널 표시 (크기, 메시지 수, 마지막 날짜)
- [x] **액션 선택:**
    - Skip: 새 채널만 다운로드
    - Incremental: 마지막 메시지 이후만 추가 다운로드
    - Overwrite: 전체 재다운로드
    - Cancel: 취소
- [x] **Archived 채널 처리:** 변경 없음 안내, 자동 Skip 권장
- [x] **증분 다운로드:** 마지막 타임스탬프 이후 메시지만 추출하여 기존 파일에 병합

### 2.9 파일 관리 및 메타데이터 (File Management) - 완료

- [x] **폴더 구조화:** `export/{category}/{channel}.md` 형태로 카테고리별 관리
- [x] **원본 불변성:** 다운로드된 `.md` 파일은 원본으로 취급, 분석 결과는 별도 저장
- [x] **메타데이터 관리:**
    - `.meta/index.json`: 전체 채널 인덱스 (다운로드 및 분석 상태 추적)
    - **자동 등록 (Auto-Registration):** 인덱스에 없는 파일 분석 시 자동으로 채널 등록
- [x] **LLM 비용 추적:**
    - 분석 시 토큰 사용량(Prompt/Completion) 및 예상 비용 기록
    - Provider별 단가 설정 (GPT-4o, Gemini 1.5 Flash 등)
    - 분석 이력 저장 및 중복 분석 방지 (Skip Logic)
- [x] **일괄 분석 (Batch Analysis):**
    - 폴더 경로를 입력받아 해당 폴더 내 모든 `.md` 파일 일괄 분석
    - 이미 분석된 파일은 자동으로 건너뛰기 (Skip)

## 3. 비기능적 요구사항 (Non-Functional Requirements)

### 3.1 성능 및 효율성
- [ ] **경량화:** 실행 파일 크기를 최소화한다 (Go 언어 특성 활용).
- [ ] **메모리 효율:** 대용량 대화 로그 처리 시 메모리 사용량을 최적화한다.
- [ ] **멀티 플랫폼:** macOS와 Windows에서 동일하게 동작하는 단일 바이너리를 제공한다.

### 3.2 사용자 경험 (UX)
- [x] **TUI (Text User Interface):** `Bubble Tea` 라이브러리를 활용하여 미려하고 직관적인 CLI 환경을 제공한다.
- [x] **Lipgloss 스타일링:** 채널 유형별 색상 구분, 선택 상태 하이라이트 등 시각적 개선 적용.
- [x] **필터 메뉴 모드:** `f` 키로 필터 설정 메뉴를 열어 채널 속성별 표시 여부를 토글할 수 있다.
- [x] **진행 상황 표시:** 다운로드 및 변환 진행률(Progress Bar)을 시각적으로 보여준다. (ETA 포함)

### 3.3 아키텍처 및 코드 품질 (Architecture & Code Quality)
- [ ] **Standard Go Project Layout:** Go 표준 프로젝트 구조(`cmd`, `internal`, `pkg`)를 준수하여 코드의 모듈화와 재사용성을 높인다.
    - `cmd/`: 메인 애플리케이션 진입점
    - `internal/`: 비즈니스 로직 (외부에서 import 불가)
    - `pkg/`: 라이브러리 코드로 사용 가능한 패키지 (선택 사항)
- [ ] **Clean Code:** 함수 분리, 명확한 변수명, 주석 작성을 통해 가독성을 유지한다.
- [ ] **Error Handling:** 모든 에러를 명시적으로 처리하고, 사용자에게 친절한 에러 메시지를 제공한다.
- [ ] **Configuration Management:** 설정(Config)과 코드(Code)를 분리하여 유지보수성을 높인다.

```
