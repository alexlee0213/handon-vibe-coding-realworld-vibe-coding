# Issue #27: Phase 7.3 프로덕션 배포 구현 계획

## 비용 최적화 전략

> **목표**: 최소 비용으로 Production 환경만 운용 (Staging 환경 없음)

### 예상 월 비용 (최소 구성)
| 서비스 | 구성 | 예상 비용 |
|--------|------|----------|
| RDS PostgreSQL | db.t4g.micro (Free Tier) | $0 ~ $15/월 |
| ECS Fargate | 0.25 vCPU / 0.5GB | ~$9/월 |
| NAT Instance | t4g.nano | ~$3/월 |
| ALB | 1개 | ~$16/월 |
| ECR | 최소 사용 | ~$1/월 |
| CloudWatch Logs | 최소 보관 | ~$1/월 |
| **합계** | | **~$30-45/월** |

> ✅ NAT Gateway 대신 NAT Instance 사용으로 ~$30/월 절감

---

## 현재 상태 분석

### 완료된 항목
| 항목 | 상태 | 파일 |
|------|------|------|
| VPC Stack | ✅ 완료 | `infra/lib/vpc-stack.ts` |
| Backend Dockerfile | ✅ 완료 | `backend/Dockerfile` |
| CI Workflow | ✅ 완료 | `.github/workflows/ci.yml` |
| Frontend Deploy Workflow | ✅ 완료 | `.github/workflows/deploy-frontend.yml` |

### 미완료 항목
| 항목 | 상태 | 설명 |
|------|------|------|
| RDS PostgreSQL Stack | ❌ 미완료 | 데이터베이스 인프라 |
| ECS Fargate Stack | ❌ 미완료 | 백엔드 서버 인프라 |
| CloudWatch 로그 설정 | ❌ 미완료 | ECS 스택에 포함 예정 |
| Backend 배포 워크플로우 | ❌ 미완료 | GitHub Actions |
| Secrets Manager 설정 | ❌ 미완료 | JWT_SECRET, DB 비밀번호 |

---

## 구현 계획

### Step 1: RDS PostgreSQL Stack 생성

**파일**: `infra/lib/rds-stack.ts`

**구성 요소**:
```
RDS Stack
├── Security Group (Postgres 5432 from Private Subnets only)
├── RDS Instance
│   ├── Engine: PostgreSQL 16
│   ├── Instance Class: db.t4g.micro (ARM, Free Tier 적용 가능)
│   ├── Storage: 20GB (gp3)
│   ├── Subnet: Isolated Subnet
│   ├── Multi-AZ: false (비용 절감)
│   ├── Backup Retention: 1일
│   └── Deletion Protection: false
├── Secrets Manager Secret (DB credentials)
└── CloudFormation Outputs
```

**비용 최적화 설정**:
| 설정 | 값 | 비고 |
|------|-----|------|
| Instance Class | db.t4g.micro | ARM 기반, Free Tier 12개월 |
| Multi-AZ | false | 단일 AZ로 비용 절감 |
| Backup Retention | 1일 | 최소 보관 |
| Storage | 20GB gp3 | 최소 용량 |
| Deletion Protection | false | 개발/테스트 용이성 |

---

### Step 2: ECS Fargate Stack 생성

**파일**: `infra/lib/ecs-stack.ts`

**구성 요소**:
```
ECS Stack
├── ECR Repository (conduit-backend)
├── ECS Cluster
├── Task Definition
│   ├── CPU: 256 (0.25 vCPU) - 최소 사양
│   ├── Memory: 512MB - 최소 사양
│   ├── Container Definition
│   │   ├── Image: ECR Repository
│   │   ├── Port: 8080
│   │   ├── Environment Variables
│   │   └── Secrets (from Secrets Manager)
│   └── Logging (CloudWatch Logs)
├── Fargate Service
│   ├── Desired Count: 1 (단일 인스턴스)
│   ├── Subnets: Private Subnets
│   └── Health Check: /health
├── Application Load Balancer
│   ├── Subnets: Public Subnets
│   ├── Listener: HTTP 80 (HTTPS는 ACM 인증서 필요 시 추가)
│   └── Target Group
└── CloudFormation Outputs
```

**비용 최적화 설정**:
| 설정 | 값 | 비고 |
|------|-----|------|
| Fargate CPU | 256 (0.25 vCPU) | 최소 사양 |
| Fargate Memory | 512MB | 최소 사양 |
| Desired Count | 1 | 단일 인스턴스 |
| Auto Scaling | 없음 | 비용 절감 |
| HTTPS | 선택 | ACM 인증서 필요 시 추가 |

---

### Step 3: Secrets Manager 설정

**파일**: `infra/lib/secrets-stack.ts` (또는 ECS 스택에 통합)

**Secrets**:
1. **DB Credentials** (RDS 스택에서 자동 생성)
   - Username
   - Password
   - Endpoint

2. **Application Secrets**
   - JWT_SECRET

---

### Step 4: VPC Stack 수정 (NAT Instance 사용)

**파일**: `infra/lib/vpc-stack.ts`

**변경 사항**:
- NAT Gateway 제거 → NAT Instance (t4g.nano) 사용
- 비용 절감: ~$32/월 → ~$3/월

**NAT Instance 구성**:
```
NAT Instance
├── Instance Type: t4g.nano (ARM, 2 vCPU, 0.5GB)
├── AMI: Amazon Linux 2023 (ARM)
├── Subnet: Public Subnet
├── Security Group
│   ├── Inbound: All traffic from Private Subnets
│   └── Outbound: All traffic to Internet
├── Source/Dest Check: Disabled
└── Route Table: Private Subnet → NAT Instance
```

**주의사항**:
- NAT Instance는 단일 인스턴스이므로 고가용성 없음
- 인스턴스 장애 시 Private Subnet의 인터넷 연결 불가
- 필요 시 Auto Recovery 설정 가능

---

### Step 5: infra/bin/infra.ts 업데이트

**변경 사항**:
```typescript
// Production Only - 스택 의존성 체인
VpcStack (Conduit-Vpc)
    ↓
RdsStack (Conduit-Rds, depends on VpcStack)
    ↓
EcsStack (Conduit-Ecs, depends on VpcStack, RdsStack)
```

> Staging 환경 제거, Production 환경만 배포

---

### Step 6: Backend 배포 워크플로우 생성

**파일**: `.github/workflows/deploy-backend.yml`

**트리거**:
- `main` 브랜치 push (backend/** 변경 시)
- Manual dispatch

**단계**:
1. Build Docker Image
2. Push to ECR
3. Update ECS Service (Force New Deployment)
4. Wait for Deployment Stable
5. Health Check Verification

---

### Step 7: Frontend 배포 설정 업데이트

**변경 사항**:
- `VITE_API_URL` 환경 변수를 ALB DNS로 설정
- GitHub Repository Variables 사용

---

## 파일 구조 (최종)

```
infra/
├── bin/
│   └── infra.ts              # CDK App 진입점 (Production only)
├── lib/
│   ├── vpc-stack.ts          # 🔄 수정 (NAT Instance 사용)
│   ├── rds-stack.ts          # 🆕 신규
│   └── ecs-stack.ts          # 🆕 신규
└── ...

.github/workflows/
├── ci.yml                    # ✅ 기존
├── deploy-frontend.yml       # ✅ 기존 (업데이트)
└── deploy-backend.yml        # 🆕 신규
```

---

## 구현 순서

| 단계 | 작업 | 예상 시간 |
|------|------|----------|
| 1 | VPC Stack 수정 (NAT Instance) | 30분 |
| 2 | RDS Stack 생성 | 30분 |
| 3 | ECS Stack 생성 | 45분 |
| 4 | infra.ts 업데이트 (Production only) | 15분 |
| 5 | Backend 배포 워크플로우 생성 | 20분 |
| 6 | Frontend 배포 설정 업데이트 | 10분 |
| 7 | CDK synth 검증 | 10분 |
| **합계** | | **~2.5시간** |

---

## 비용 최적화 요약

### 최소 비용 구성
| 구성 요소 | 설정 | 월 비용 |
|-----------|------|---------|
| RDS | db.t4g.micro, Single-AZ | ~$0-15 |
| Fargate | 0.25 vCPU, 512MB, 1개 | ~$9 |
| NAT Instance | t4g.nano | ~$3 |
| ALB | 1개 | ~$16 |
| 기타 | ECR, CloudWatch | ~$2 |
| **합계** | | **~$30-45/월** |

### 추가 비용 절감 옵션
1. **Fargate Spot**: 최대 70% 할인 (단, 중단 가능성)
2. **Reserved Capacity**: 1년 약정 시 할인
3. **RDS Free Tier**: 신규 AWS 계정 12개월 무료

### 비용 vs 가용성 트레이드오프
| 항목 | 현재 설정 | 트레이드오프 |
|------|----------|-------------|
| NAT | Instance (t4g.nano) | 인스턴스 장애 시 인터넷 연결 불가 |
| Multi-AZ | 비활성화 | 단일 AZ 장애 시 DB 다운 |
| ECS 인스턴스 | 1개 | 배포 시 잠시 다운타임 |
| Auto Scaling | 없음 | 트래픽 증가 시 수동 대응 |

---

## 보안

- RDS: Isolated Subnet 배치, VPC 내부에서만 접근
- ECS: Private Subnet 배치, ALB를 통해서만 외부 접근
- Secrets: AWS Secrets Manager 사용
- HTTPS: HTTP로 시작, 필요 시 ACM 인증서 추가

---

## 검증 방법

1. **CDK Synth**
   ```bash
   cd infra && npx cdk synth
   ```

2. **CDK Deploy**
   ```bash
   npx cdk deploy --all
   ```

3. **Health Check**
   ```bash
   curl http://<alb-dns>/health
   ```

4. **API 테스트**
   ```bash
   curl http://<alb-dns>/api/articles
   ```

---

## 의존성

- AWS 계정 및 IAM 권한
- GitHub Secrets 설정:
  - `AWS_ACCESS_KEY_ID`
  - `AWS_SECRET_ACCESS_KEY`
  - `AWS_REGION` (ap-northeast-2)
