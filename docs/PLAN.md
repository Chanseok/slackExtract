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
- [x] **메시지 다운로드:** 선택된 채널의 히스토리 가져오기 (`GetConversationHistory` + Pagination)
- [ ] **스레드 다운로드:** 각 메시지의 댓글(Thread) 가져오기 (`GetConversationReplies`)
- [x] **사용자 매핑:** User ID를 실제 이름으로 변환 (`GetUsersPaginated` + JSON 캐싱)
- [x] **Markdown 변환:** 기본 Markdown 포맷 적용 (사용자 멘션 치환 포함)
- [x] **파일 저장:** `export/{채널명}.md` 구조로 저장

## Phase 3.5: 로컬 캐싱 시스템 (Local Caching)
- [x] **사용자 캐싱:** `users.json` 파일에 사용자 목록 캐싱 (완료)
- [x] **채널 캐싱:** `channels.json` 파일에 채널 목록 및 속성 캐싱 (완료)
  - 저장 속성: ID, Name, IsArchived, IsPrivate, IsMember, NumMembers, Topic, Purpose, Created 등
  - 캐시 파일이 있으면 API 호출 생략, 없으면 API에서 가져와 저장
- [ ] **캐시 갱신 명령:** 사용자가 명시적으로 캐시를 갱신할 수 있는 옵션 제공
  - 예: `--refresh-users`, `--refresh-channels` 또는 TUI 메뉴 (`r` 키)

## Phase 4: TUI 개선 및 사용성 (TUI & Usability) ⬅️ **현재 단계**

### 4.1 시각적 개선 (Visual Enhancement)
- [ ] **Lipgloss 스타일링:** 컬러 테마, 테두리, 하이라이트 적용
- [ ] **채널 유형별 색상 구분:**
  - 🔓 Public Channel: 초록색
  - 🔒 Private Channel: 빨간색
  - 💬 DM (Direct Message): 노란색
  - 📁 Archived: 회색 (흐리게)
- [ ] **선택 상태 시각화:** 체크박스 스타일 `[ ]` / `[x]` + 하이라이트
- [ ] **상태 아이콘:** 캐시됨(📦), 다운로드됨(⬇️), 아카이브됨(🗄️)

### 4.2 정보 표시 강화 (Information Display)
- [ ] **채널 정보 표시:** 멤버 수, 유형, 상태를 한 줄에 표시
  - 예: `[x] 🔒 sales-germany    (12 members)  ⬇️ Downloaded`
- [ ] **상태바 (Status Bar):**
  - Workspace 이름
  - 전체/참여/선택된 채널 수
  - 캐시 상태 및 갱신 시간
- [ ] **하단 도움말 바:** 사용 가능한 키보드 단축키 표시

### 4.3 인터랙션 개선 (Interaction)
- [ ] **키보드 단축키:**
  | 키 | 기능 |
  |----|------|
  | `/` | 채널 검색 (실시간 필터링) |
  | `Tab` | 뷰 전환 (All / Joined / Archived / Downloaded) |
  | `a` | 전체 선택 / 해제 |
  | `r` | 캐시 새로고침 |
  | `Enter` | 선택된 채널 다운로드 시작 |
  | `i` | 채널 상세 정보 보기 |
  | `?` | 도움말 표시 |
- [ ] **채널 필터링:** Archived 숨기기, 참여 채널만 보기 등
- [ ] **채널 정렬:** 이름순, 멤버 수 순 등

### 4.4 UI 레이아웃
```
┌─ Slack Extract ──────────────────────────────────────┐
│ Workspace: MyCompany (mycompany.slack.com)           │
│ Channels: 245 total | 89 joined | 3 selected         │
│ Cache: Users ✓ (2h ago) | Channels ✓ (5m ago)        │
├──────────────────────────────────────────────────────┤
│ [Filter: All ▼] [Sort: Name ▼] [Search: ________]    │
├──────────────────────────────────────────────────────┤
│  [ ] 🔓 general         156 members   📦             │
│  [x] 🔒 sales-germany    12 members   ⬇️             │
│  ...                                                 │
├──────────────────────────────────────────────────────┤
│ ↑↓: Navigate | Space: Select | Enter: Download       │
│ /: Search | Tab: Filter | r: Refresh | ?: Help       │
└──────────────────────────────────────────────────────┘
```

### 4.5 구현 우선순위
1. **[높음]** Lipgloss 스타일링 (색상, 테두리, 하이라이트)
2. **[높음]** 채널 정보 표시 개선 (유형, 멤버 수, 상태)
3. **[높음]** 검색 및 필터링 기능 (스마트 내보내기의 기반)
4. **[중간]** 스마트 내보내기 (폴더 그룹핑)
5. **[중간]** 상태바 및 도움말 바
6. **[낮음]** 다운로드 진행률 표시

### 4.6 스마트 내보내기 워크플로우 (Smart Export)
- [ ] **검색어 기반 폴더 생성:** 검색 필터가 적용된 상태에서 다운로드 시, 검색어 이름으로 하위 폴더 생성 제안
- [ ] **수동 그룹핑 (Ad-hoc Grouping):** 선택된 채널들을 특정 그룹명(폴더)으로 묶어서 내보내기 (`g` 키)
- [ ] **필터 결과 일괄 선택:** 검색된 결과만 한 번에 선택/해제하는 기능

## Phase 5: 고급 기능 및 최적화 (Advanced & Optimization)
- [ ] **스레드 다운로드:** 각 메시지의 댓글(Thread) 가져오기 (`GetConversationReplies`)
- [ ] **다운로드 진행률 표시:** Progress Bar로 진행 상황 시각화
- [ ] **채널 요약:** 채널별 주요 논의 내용 키워드 추출 또는 요약 (LLM 연동 고려)
- [ ] **첨부파일 처리:** 이미지 및 파일 다운로드 옵션
- [ ] **에러 핸들링:** Rate Limit 자동 재시도 및 부분 실패 처리
- [ ] **설정 관리:** 토큰, 쿠키, 저장 경로 등을 설정 파일로 관리
- [ ] **채널 그룹핑(Preset):** 자주 사용하는 채널 그룹을 설정 파일(`config.yaml`)로 저장/로드

## Phase 6: 배포 및 문서화 (Release)
- [ ] `README.md` 작성 (사용법, 설치법)
- [ ] GitHub Release를 통한 바이너리 배포
