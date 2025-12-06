# 스마트 다운로드 및 파일 관리 시스템 설계

> 작성일: 2025-12-06
> 상태: 설계 단계 (Draft)

## 1. 배경 및 목적

### 1.1 현재 문제점
- 채널 선택 후 Enter를 누르면 **즉시 다운로드가 시작**됨
- 이미 다운로드된 채널도 **중복 다운로드** 발생
- Archived 채널도 **불필요하게 재다운로드** 가능
- **저장 폴더를 선택할 수 없음** (항상 `export/` 루트)
- 다운로드/분석 파일의 **체계적인 관리 부재**

### 1.2 목표
1. **다운로드 효율화**: 중복/불필요한 다운로드 방지
2. **증분 다운로드**: 마지막 다운로드 이후 메시지만 추출
3. **폴더 구조화**: 카테고리별 원본 파일 관리
4. **메타데이터 관리**: 다운로드/분석 이력 추적
5. **LLM 비용 최적화**: 토큰 사용량 및 비용 기록

---

## 2. 요구사항 정의

### 2.1 스마트 다운로드 (Smart Download)

#### FR-SD-01: 다운로드 전 확인 단계
- [ ] 채널 선택 후 Enter 시, 즉시 다운로드하지 않고 **확인 화면** 표시
- [ ] 확인 화면에서 다음 정보 제공:
  - 선택된 채널 수
  - 대상 저장 폴더
  - 이미 다운로드된 채널 목록 (있을 경우)

#### FR-SD-02: 저장 폴더 선택
- [ ] 기존 하위 폴더 목록에서 선택 가능
- [ ] 새 폴더 생성 옵션
- [ ] **스마트 제안 (Smart Suggestion):** 채널명 패턴(예: `proj-`)에 기반하여 적절한 기존 폴더 추천
- [ ] 기본값: `export/` (루트) 또는 마지막 사용 폴더

#### FR-SD-03: 기존 파일 감지 및 정보 표시
- [ ] 선택된 채널 중 이미 다운로드된 파일 감지 (파일명 매칭)
- [ ] 기존 파일 정보 표시:
  - 파일 크기 (KB/MB)
  - 메시지 수 (파일 파싱 또는 메타데이터)
  - 마지막 메시지 날짜
  - Archived 여부
- [ ] **미리보기 (Preview):** 기존 파일의 마지막 3-5줄을 미리 보여주어 덮어쓰기 전 확인 지원

#### FR-SD-04: 다운로드 액션 선택
- [ ] **Skip**: 기존 파일이 있는 채널은 건너뛰고, 새 채널만 다운로드
- [ ] **Incremental**: 마지막 메시지 이후 새 메시지만 추가 다운로드
- [ ] **Overwrite**: 기존 파일 무시하고 전체 재다운로드
- [ ] **Cancel**: 다운로드 취소

#### FR-SD-05: Archived 채널 특별 처리
- [ ] Archived 채널은 내용이 변하지 않으므로:
  - 이미 다운로드된 경우: "이 채널은 아카이브되어 변경사항이 없습니다" 안내
  - Incremental 선택 시: 자동으로 Skip 처리

#### FR-SD-06: 증분 다운로드 구현
- [ ] **타임스탬프 우선순위:**
  1. `meta.json`의 `last_updated` (가장 신뢰)
  2. `.md` 파일 파싱 (메타데이터 없을 경우 Fallback)
- [ ] Slack API의 `oldest` 파라미터로 해당 시점 이후 메시지만 요청
- [ ] 새 메시지를 기존 파일 **끝에 추가** (날짜 순서 유지)

### 2.2 폴더 구조 및 파일 관리

#### FR-FM-01: 폴더 구조
```
export/
├── .meta/                      # 메타데이터 (숨김 폴더)
│   ├── index.json              # 전체 채널 인덱스
│   └── channels/               # 채널별 상세 메타
│       └── {channel_name}.json
├── archived/                   # 사용자 정의 카테고리
│   └── old-project.md
├── sales/
│   └── sales-germany.md
├── project-lg/
│   └── lg-main.md
└── uncategorized/              # 기본 폴더 (미분류)
    └── random.md
```

#### FR-FM-02: 원본 파일 불변성
- [ ] 다운로드된 `.md` 파일은 **원본으로 취급**
- [ ] 증분 다운로드 시에만 내용 추가 (수정 아님)
- [ ] 분석 결과는 별도 폴더(`.analysis/`)에 저장

#### FR-FM-03: 메타데이터 인덱스 (`index.json`)
```json
{
  "version": "1.0",
  "last_updated": "2025-12-06T15:30:00Z",
  "channels": {
    "sales/sales-germany.md": {
      "channel_id": "C12345",
      "channel_name": "sales-germany",
      "is_archived": false,
      "keywords": ["sales", "germany", "q4", "revenue"], // 검색용 키워드
      "short_summary": "독일 영업팀 Q4 실적 논의 및 이슈 트래킹", // 요약 정보
      "download": {
        "first_downloaded": "2025-12-05T10:30:00Z",
        "last_updated": "2025-12-06T14:20:00Z",
        "message_count": 1523,
        "oldest_message": "2024-01-15T09:00:00Z",
        "newest_message": "2025-12-06T12:00:00Z",
        "file_size_bytes": 245678
      },
      "analyses": ["translation", "topics"]
    }
  }
}
```

#### FR-FM-04: 채널별 상세 메타데이터 (`{channel}.json`)
```json
{
  "channel_id": "C12345",
  "channel_name": "sales-germany",
  "source_file": "sales/sales-germany.md",
  "is_archived": false,
  
  "download_history": [
    {
      "timestamp": "2025-12-05T10:30:00Z",
      "type": "full",
      "messages_added": 1500,
      "oldest": "2024-01-15T09:00:00Z",
      "newest": "2025-12-05T10:00:00Z"
    },
    {
      "timestamp": "2025-12-06T14:20:00Z",
      "type": "incremental",
      "messages_added": 23,
      "oldest": "2025-12-05T10:01:00Z",
      "newest": "2025-12-06T12:00:00Z"
    }
  ],
  
  "analyses": [
    {
      "type": "translation",
      "created_at": "2025-12-06T15:00:00Z",
      "output_file": ".analysis/sales-germany/translation.md",
      "llm": {
        "provider": "gemini",
        "model": "gemini-1.5-flash",
        "prompt_version": "v1.2",
        "prompt_hash": "a1b2c3d4",
        "input_tokens": 8500,
        "output_tokens": 4000,
        "estimated_cost_usd": 0.015
      },
      "summary": "영어/네덜란드어 → 한국어 번역"
    }
  ]
}
```

#### FR-FM-05: 메타데이터 동기화 및 복구 (Sync & Repair)
- [ ] **고아 파일 정리:** 메타데이터에는 있지만 실제 파일이 없는 경우 정리
- [ ] **메타 생성:** 파일은 있지만 메타데이터가 없는 경우 스캔하여 생성
- [ ] **무결성 검사:** 파일 크기/수정일이 메타데이터와 크게 다를 경우 경고

### 2.3 LLM 분석 관리

#### FR-LLM-01: 분석 결과 저장 구조
```
export/
└── .analysis/                  # 분석 결과 폴더 (숨김)
    └── {channel_name}/
        ├── meta.json           # 분석 메타데이터
        ├── translation.md      # 번역 결과
        ├── topics.md           # 주제 분석
        ├── sentiment.md        # 감정 분석
        └── contributors.md     # 기여자 분석
```

#### FR-LLM-02: 분석 이력 추적
- [ ] 분석 수행 시 다음 정보 기록:
  - 분석 일시
  - 사용된 LLM Provider 및 모델
  - 프롬프트 버전 또는 핵심 키워드
  - 입력/출력 토큰 수
  - 예상 비용 (USD)

#### FR-LLM-03: 토큰 비용 추적
- [ ] Provider별 토큰 단가 설정 (config)
- [ ] 분석 전 예상 토큰 수 및 비용 표시
- [ ] **예산 제한 (Budget Limit):** 설정된 비용 초과 예상 시 경고 또는 차단
- [ ] 분석 후 실제 사용량 기록

#### FR-LLM-04: 재분석 방지
- [ ] 이미 분석된 채널은 "이미 분석됨" 표시
- [ ] 재분석 시 확인 프롬프트 ("기존 분석을 덮어쓰시겠습니까?")
- [ ] 원본 파일 변경 시에만 재분석 권장

---

## 3. 사용자 인터페이스 설계

### 3.1 다운로드 확인 화면 (TUI)

```
┌─ Download Confirmation ──────────────────────────────────┐
│                                                          │
│  📥 Selected: 5 channels                                 │
│                                                          │
│  📁 Target folder: export/private/                       │
│     [Tab] Change folder  [n] New folder                  │
│                                                          │
│  ─────────────────────────────────────────────────────── │
│                                                          │
│  ⚠️  Already downloaded (3 of 5):                        │
│                                                          │
│   Channel              Size     Messages   Last Updated  │
│   ─────────────────────────────────────────────────────  │
│   sales-germany.md     245 KB   1,523      Dec 5, 14:20  │
│   dev-team.md          89 KB    456        Dec 4, 09:15  │
│   hr-announce.md       12 KB    34         Dec 5, 11:00  │
│                        [ARCHIVED - no new messages]      │
│                                                          │
│  ─────────────────────────────────────────────────────── │
│                                                          │
│  Choose action:                                          │
│   [s] Skip existing    - download 2 new channels only    │
│   [i] Incremental      - add new messages to 3 files     │
│   [o] Overwrite all    - re-download all 5 channels      │
│   [c] Cancel                                             │
│                                                          │
│  💡 Tip: Archived channels have no new messages          │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 3.2 폴더 선택 화면

```
┌─ Select Target Folder ───────────────────────────────────┐
│                                                          │
│  Current: export/                                        │
│                                                          │
│  > archived/          (3 files)                          │
│    sales/             (5 files)                          │
│    project-lg/        (2 files)                          │
│    uncategorized/     (12 files)                         │
│    ────────────────────────────                          │
│    [+] Create new folder...                              │
│                                                          │
│  [Enter] Select  [Esc] Cancel                            │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 3.3 분석 확인 화면 (slack-analyze)

```
┌─ Analysis Options ───────────────────────────────────────┐
│                                                          │
│  📄 File: export/sales/sales-germany.md                  │
│     Size: 245 KB | Messages: 1,523 | Last: Dec 6         │
│                                                          │
│  📊 Previous Analyses:                                   │
│     • Translation (Dec 5) - gemini-1.5-flash             │
│     • Topics (Dec 5) - gpt-4o                            │
│                                                          │
│  💰 Estimated Cost:                                      │
│     Input: ~8,500 tokens ($0.008)                        │
│     Output: ~4,000 tokens ($0.006)                       │
│     Total: ~$0.014                                       │
│                                                          │
│  Select analysis type:                                   │
│   [1] Full Analysis (all)                                │
│   [2] Translation only                                   │
│   [3] Topics only                                        │
│   [4] Sentiment only                                     │
│   [5] Contributors only                                  │
│   [c] Cancel                                             │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## 4. 구현 계획

### Phase 8: 스마트 다운로드 (Smart Download)

#### 8.1 기존 파일 감지 (우선순위: 높음)
- [ ] `export/` 하위 모든 폴더 스캔
- [ ] 채널명과 파일명 매칭 로직
- [ ] 파일 메타정보 추출 (크기, 수정일)
- [ ] `.md` 파일에서 마지막 메시지 타임스탬프 파싱

**예상 작업량:** 2시간

#### 8.2 다운로드 확인 TUI (우선순위: 높음)
- [ ] 새로운 TUI 모드: `confirmDownloadView`
- [ ] 폴더 선택 UI
- [ ] 기존 파일 목록 표시
- [ ] 액션 선택 (s/i/o/c)

**예상 작업량:** 3시간

#### 8.3 증분 다운로드 로직 (우선순위: 중간)
- [ ] `oldest` 파라미터 지원 추가
- [ ] 기존 파일에 새 메시지 병합
- [ ] 날짜별 그룹핑 유지

**예상 작업량:** 2시간

#### 8.4 Archived 채널 처리 (우선순위: 중간)
- [ ] Archived 상태 표시
- [ ] 자동 Skip 로직

**예상 작업량:** 1시간

### Phase 9: 메타데이터 시스템 (Metadata System)

#### 9.1 메타 구조 설계 (우선순위: 중간)
- [ ] `.meta/` 폴더 구조 생성
- [ ] `index.json` 스키마 정의
- [ ] 채널별 메타 파일 스키마 정의

**예상 작업량:** 1시간

#### 9.2 메타데이터 CRUD (우선순위: 중간)
- [ ] `internal/meta/` 패키지 생성
- [ ] 인덱스 읽기/쓰기
- [ ] 채널 메타 읽기/쓰기
- [ ] 다운로드 이력 추가

**예상 작업량:** 2시간

#### 9.3 다운로드와 메타 연동 (우선순위: 중간)
- [ ] 다운로드 완료 시 메타 업데이트
- [ ] 파일 감지 시 메타 참조

**예상 작업량:** 1시간

### Phase 10: LLM 비용 추적 (Cost Tracking)

#### 10.1 토큰 카운팅 (우선순위: 낮음)
- [ ] 입력 텍스트 토큰 추정 (tiktoken 또는 근사치)
- [ ] API 응답에서 실제 토큰 수 추출

**예상 작업량:** 2시간

#### 10.2 비용 계산 (우선순위: 낮음)
- [ ] Provider별 단가 설정 (config)
- [ ] 예상 비용 표시
- [ ] 실제 비용 기록

**예상 작업량:** 1시간

#### 10.3 분석 이력 저장 (우선순위: 낮음)
- [ ] 분석 메타데이터 저장
- [ ] 이전 분석 조회

**예상 작업량:** 1시간

---

## 5. 우선순위 및 일정

| 순위 | Phase | 핵심 기능 | 예상 시간 | 의존성 |
|------|-------|----------|----------|--------|
| 1 | 8.1 | 기존 파일 감지 | 2h | - |
| 2 | 8.2 | 다운로드 확인 TUI | 3h | 8.1 |
| 3 | 8.4 | Archived 채널 처리 | 1h | 8.2 |
| 4 | 8.3 | 증분 다운로드 | 2h | 8.2 |
| 5 | 9.1 | 메타 구조 설계 | 1h | - |
| 6 | 9.2 | 메타데이터 CRUD | 2h | 9.1 |
| 7 | 9.3 | 다운로드-메타 연동 | 1h | 8.3, 9.2 |
| 8 | 10.1 | 토큰 카운팅 | 2h | - |
| 9 | 10.2 | 비용 계산 | 1h | 10.1 |
| 10 | 10.3 | 분석 이력 저장 | 1h | 9.2, 10.2 |

**총 예상 시간:** 16시간

---

## 6. 기술적 고려사항

### 6.1 파일 파싱
- `.md` 파일에서 마지막 메시지 타임스탬프를 추출하려면:
  - 헤더의 메타정보 활용 (권장)
  - 또는 마지막 `###` 헤더의 날짜 파싱

### 6.2 메타데이터 vs 파일 파싱
- **메타데이터 우선**: 빠르고 정확
- **파일 파싱 폴백**: 메타 없을 때 호환성 유지

### 6.3 동시성
- 여러 채널 다운로드 시 메타데이터 동시 쓰기 주의
- 파일 잠금 또는 순차 처리 고려

### 6.4 Git 제외
- `export/` 전체가 `.gitignore`에 포함되어 있으므로:
  - `.meta/`도 자동으로 제외됨
  - `.analysis/`도 자동으로 제외됨

---

## 7. 향후 확장 가능성

- **TUI에서 메타 조회**: 채널 목록에서 다운로드 상태, 분석 상태 표시
- **분석 결과 비교**: 같은 채널의 이전/현재 분석 결과 diff
- **백업/복원**: 메타데이터 기반 export 폴더 백업
- **통계 대시보드**: 총 토큰 사용량, 비용 누적 등
