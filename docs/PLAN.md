# 개발 계획 (Development Plan)

## Phase 1: 환경 설정 및 분석 (Setup & Analysis)
- [x] 프로젝트 초기화 (`go mod init`)
- [x] 요구사항 정의서 작성 (`REQUIREMENTS.md`)
- [ ] `.gitignore` 및 `.env` 설정
- [ ] `slackdump` 소스 코드 분석 (구조, API 호출 방식, 라이선스 확인)
- [ ] 기술 스택 확정 (TUI 라이브러리: Bubble Tea 등)

## Phase 2: 프로토타입 (Prototype)
- [ ] Slack API 연동 테스트 (User Token으로 채널 목록 가져오기)
- [ ] Bubble Tea를 이용한 기본 TUI 구현 (채널 목록 표시 및 선택)
- [ ] 선택된 채널 ID 출력 기능 구현

## Phase 3: 핵심 기능 구현 (Core Implementation)
- [ ] 메시지 다운로드 로직 구현 (Pagination, Rate Limit 처리)
- [ ] 스레드(Thread) 다운로드 로직 구현
- [ ] 사용자 정보(User List) 캐싱 및 ID 매핑 로직 구현
- [ ] Slack mrkdwn -> Markdown 변환기 구현

## Phase 4: 고도화 및 최적화 (Refinement)
- [ ] 파일 저장 시스템 구현 (폴더 구조화)
- [ ] 첨부파일/이미지 다운로드 처리 (선택 사항)
- [ ] 진행률 표시 (Progress Bar) 및 에러 핸들링 강화
- [ ] Cross-compile 빌드 설정 (Mac/Windows)

## Phase 5: 배포 및 문서화 (Release)
- [ ] `README.md` 작성 (사용법, 설치법)
- [ ] GitHub Release를 통한 바이너리 배포
