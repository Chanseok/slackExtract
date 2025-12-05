# 개발 계획 (Development Plan)

## Phase 1: 환경 설정 및 분석 (Setup & Analysis)
- [x] 프로젝트 초기화 (`go mod init`)
- [x] 요구사항 정의서 작성 (`REQUIREMENTS.md`)
- [x] `.gitignore` 및 `.env` 설정 (Cookie 인증 방식 적용)
- [x] GitHub Repository 생성 및 연동
- [x] `slack-go/slack` 라이브러리 연동 (공식 SDK 사용, `slackdump` 로직 참고)

## Phase 2: 프로토타입 (Prototype)
- [x] Slack API 연동 테스트 (Cookie 인증 성공)
- [x] Bubble Tea를 이용한 기본 TUI 구현 (채널 목록 표시 및 선택)
- [x] 채널 목록 가져오기 (`GetConversations` API 사용)
- [x] TUI 개선: 페이지네이션(Pagination), 스크롤링(Scrolling), 전체화면(AltScreen) 적용

## Phase 3: 핵심 기능 구현 (Core Implementation)
- [ ] **TUI 디자인 개선:** `Lipgloss`를 활용한 컬러 테마 및 스타일 적용 (시인성 향상)
- [x] **메시지 다운로드:** 선택된 채널의 히스토리 가져오기 (`GetConversationHistory` + Pagination)
- [ ] **스레드 다운로드:** 각 메시지의 댓글(Thread) 가져오기 (`GetConversationReplies`)
- [x] **사용자 매핑:** User ID를 실제 이름으로 변환 (`GetUsersPaginated` + JSON 캐싱)
- [x] **Markdown 변환:** 기본 Markdown 포맷 적용 (사용자 멘션 치환 포함)
- [x] **파일 저장:** `export/{채널명}.md` 구조로 저장

## Phase 4: 사용성 개선 (Usability Enhancement)
- [ ] **채널 그룹핑(Preset):** 자주 사용하는 채널 그룹을 설정 파일(`config.yaml`)로 저장/로드
- [ ] **검색 및 필터링:** TUI 내에서 채널 이름으로 검색 및 필터링 기능
- [ ] **설정 관리:** 토큰, 쿠키, 저장 경로 등을 설정 파일로 관리

## Phase 5: 고급 분석 및 최적화 (Advanced & Optimization)
- [ ] **채널 요약:** 채널별 주요 논의 내용 키워드 추출 또는 요약 (LLM 연동 고려)
- [ ] **첨부파일 처리:** 이미지 및 파일 다운로드 옵션
- [ ] **진행률 표시:** 다운로드 진행 상황 Progress Bar 표시
- [ ] **에러 핸들링:** Rate Limit 자동 재시도 및 부분 실패 처리

## Phase 6: 배포 및 문서화 (Release)
- [ ] `README.md` 작성 (사용법, 설치법)
- [ ] GitHub Release를 통한 바이너리 배포
