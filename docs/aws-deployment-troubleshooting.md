# AWS CDK Deployment Troubleshooting Guide

RealWorld Conduit 프로젝트의 AWS CDK 배포 과정에서 발생한 오류와 해결책을 정리한 문서입니다.

## 목차

1. [CDK 스택 구조](#cdk-스택-구조)
2. [배포 순서](#배포-순서)
3. [오류 및 해결책](#오류-및-해결책)
   - [Cyclic Dependency 오류](#1-cyclic-dependency-오류)
   - [ECR Repository 충돌](#2-ecr-repository-충돌)
   - [Security Group 간 참조 문제](#3-security-group-간-참조-문제)
   - [ECS Task 시작 실패](#4-ecs-task-시작-실패)
   - [RDS 연결 오류](#5-rds-연결-오류)
   - [Docker 이미지 빌드 오류](#6-docker-이미지-빌드-오류)
4. [배포 명령어 모음](#배포-명령어-모음)
5. [스택 삭제 및 재배포](#스택-삭제-및-재배포)

---

## CDK 스택 구조

```
┌─────────────────────────────────────────────────────┐
│                    Conduit-Ecs                       │
│  - ECS Cluster, Fargate Service                     │
│  - Application Load Balancer                         │
│  - ECR Repository (기존 사용)                        │
│  - CloudWatch Log Group                              │
└───────────────────────┬─────────────────────────────┘
                        │ depends on
┌───────────────────────▼─────────────────────────────┐
│                    Conduit-Rds                       │
│  - RDS PostgreSQL (db.t4g.micro)                    │
│  - DB Security Group                                 │
│  - Secrets Manager (DB credentials)                  │
└───────────────────────┬─────────────────────────────┘
                        │ depends on
┌───────────────────────▼─────────────────────────────┐
│                    Conduit-Vpc                       │
│  - VPC with 2 AZs                                    │
│  - Public, Private, Isolated Subnets                │
│  - NAT Instance (t4g.nano)                          │
│  - Internet Gateway                                  │
└─────────────────────────────────────────────────────┘
```

---

## 배포 순서

스택은 의존성에 따라 다음 순서로 배포됩니다:

1. **Conduit-Vpc** (약 3-5분)
2. **Conduit-Rds** (약 10-15분, RDS 인스턴스 생성 시간)
3. **Conduit-Ecs** (약 5-10분)

---

## 오류 및 해결책

### 1. Cyclic Dependency 오류

#### 증상
```
Error: 'Conduit-Rds' depends on 'Conduit-Ecs'. Adding this dependency
would create a cyclic reference.
```

#### 원인
EcsStack에서 RdsStack의 Security Group 객체를 직접 참조하면, CDK가 cross-stack reference를 생성합니다. 이때 EcsStack이 RdsStack의 Security Group에 ingress rule을 추가하려고 하면 양방향 의존성이 발생합니다.

```typescript
// 문제가 되는 코드 (infra.ts)
const ecsStack = new EcsStack(app, 'Conduit-Ecs', {
  dbSecurityGroup: rdsStack.dbSecurityGroup,  // 객체 직접 전달 → 순환 참조 발생
});
```

#### 해결책
Security Group ID만 전달하고, EcsStack에서 `fromSecurityGroupId`로 import합니다.

**infra/bin/infra.ts:**
```typescript
const ecsStack = new EcsStack(app, 'Conduit-Ecs', {
  // Security Group ID만 전달 (문자열)
  dbSecurityGroupId: rdsStack.dbSecurityGroup.securityGroupId,
});
```

**infra/lib/ecs-stack.ts:**
```typescript
export interface EcsStackProps extends cdk.StackProps {
  readonly dbSecurityGroupId: string;  // ID만 받음
}

// EcsStack 내부에서 import
const dbSecurityGroup = ec2.SecurityGroup.fromSecurityGroupId(
  this,
  'ImportedDbSecurityGroup',
  dbSecurityGroupId
);

// 그 후 ingress rule 추가
dbSecurityGroup.addIngressRule(
  serviceSecurityGroup,
  ec2.Port.tcp(5432),
  'Allow ECS Fargate to connect to RDS PostgreSQL'
);
```

---

### 2. ECR Repository 충돌

#### 증상
```
Error: Resource handler returned message: "Repository already exists"
```

#### 원인
ECR Repository가 이미 존재하는데 CDK가 새로 생성하려고 시도할 때 발생합니다.

#### 해결책
ECR Repository는 수동으로 생성하고, CDK에서는 기존 repository를 import합니다.

```typescript
// 새로 생성하지 않고 기존 repository 사용
this.ecrRepository = ecr.Repository.fromRepositoryName(
  this,
  'ConduitBackendRepo',
  `conduit-${environment}-backend`
);
```

**ECR Repository 수동 생성:**
```bash
aws ecr create-repository \
  --repository-name conduit-production-backend \
  --region us-east-1
```

---

### 3. Security Group 간 참조 문제

#### 증상
ECS 태스크가 RDS에 연결하지 못하고 타임아웃 발생

#### 원인
1. RDS Security Group에 ECS Service의 Security Group으로부터의 ingress rule이 없음
2. 서로 다른 서브넷에 있어 라우팅 문제 발생

#### 해결책

**ECS Service Security Group → RDS Security Group 허용:**
```typescript
// ecs-stack.ts
dbSecurityGroup.addIngressRule(
  serviceSecurityGroup,
  ec2.Port.tcp(5432),
  'Allow ECS Fargate to connect to RDS PostgreSQL'
);
```

**서브넷 배치 확인:**
- ECS: `PRIVATE_WITH_EGRESS` (NAT를 통한 인터넷 접근 가능)
- RDS: `PRIVATE_ISOLATED` (인터넷 접근 불가)
- 같은 VPC 내에서는 Security Group만 허용되면 통신 가능

---

### 4. ECS Task 시작 실패

#### 증상
```
Essential container in task exited
Task failed ELB health checks
```

#### 원인
1. Health check 실패 (curl이 설치되지 않음)
2. 환경 변수 누락
3. Secrets Manager 권한 없음

#### 해결책

**Dockerfile에 curl 설치:**
```dockerfile
RUN apk --no-cache add ca-certificates curl
```

**Health check 설정:**
```typescript
healthCheck: {
  command: ['CMD-SHELL', 'curl -f http://localhost:8080/health || exit 1'],
  interval: cdk.Duration.seconds(30),
  timeout: cdk.Duration.seconds(5),
  retries: 3,
  startPeriod: cdk.Duration.seconds(60),  // 초기 시작 시간 여유
},
```

**Task Execution Role에 Secrets 권한 부여:**
```typescript
dbSecret.grantRead(executionRole);
jwtSecret.grantRead(executionRole);
```

---

### 5. RDS 연결 오류

#### 증상
```
dial tcp: connection refused
FATAL: password authentication failed
```

#### 원인
1. Security Group 미설정
2. SSL 모드 불일치
3. 잘못된 credential

#### 해결책

**환경 변수 확인:**
```typescript
environment: {
  DB_HOST: dbEndpoint,
  DB_PORT: dbPort,
  DB_NAME: 'conduit',
  DB_SSLMODE: 'require',  // RDS는 SSL 필수
},
secrets: {
  DB_USERNAME: ecs.Secret.fromSecretsManager(dbSecret, 'username'),
  DB_PASSWORD: ecs.Secret.fromSecretsManager(dbSecret, 'password'),
},
```

**개별 환경변수로 분리 (DATABASE_URL 대신):**
Go 백엔드에서 개별 환경변수를 사용하도록 수정:
```go
// DATABASE_URL 대신 개별 변수 사용
dbHost := os.Getenv("DB_HOST")
dbPort := os.Getenv("DB_PORT")
dbUser := os.Getenv("DB_USERNAME")
dbPass := os.Getenv("DB_PASSWORD")
dbName := os.Getenv("DB_NAME")
sslMode := os.Getenv("DB_SSLMODE")
```

---

### 6. Docker 이미지 빌드 오류

#### 증상
```
src/test/setup.ts: error TS2304: Cannot find name 'global'
```

#### 원인
TypeScript 빌드 시 테스트 파일이 포함되어 빌드 실패

#### 해결책

**tsconfig.json에서 테스트 파일 제외:**
```json
{
  "exclude": ["src/test/**", "**/*.test.ts", "**/*.test.tsx"]
}
```

**별도의 tsconfig.build.json 사용:**
```json
{
  "extends": "./tsconfig.json",
  "exclude": ["src/test/**", "**/*.test.*"]
}
```

---

## 배포 명령어 모음

### 전체 배포
```bash
cd infra
npx cdk deploy --all --require-approval never
```

### 개별 스택 배포
```bash
npx cdk deploy Conduit-Vpc --require-approval never
npx cdk deploy Conduit-Rds --require-approval never
npx cdk deploy Conduit-Ecs --require-approval never
```

### 스택 상태 확인
```bash
aws cloudformation describe-stacks --stack-name Conduit-Vpc --query 'Stacks[0].StackStatus'
aws cloudformation describe-stacks --stack-name Conduit-Rds --query 'Stacks[0].StackStatus'
aws cloudformation describe-stacks --stack-name Conduit-Ecs --query 'Stacks[0].StackStatus'
```

### ECS 서비스 상태 확인
```bash
aws ecs describe-services \
  --cluster conduit-production-cluster \
  --services conduit-production-backend \
  --query 'services[0].{status:status,runningCount:runningCount,desiredCount:desiredCount}'
```

### CloudWatch 로그 확인
```bash
aws logs tail /ecs/conduit-production-backend --follow
```

---

## 스택 삭제 및 재배포

### 개별 스택 삭제 (역순)
```bash
# ECS 먼저 삭제
aws cloudformation delete-stack --stack-name Conduit-Ecs
aws cloudformation wait stack-delete-complete --stack-name Conduit-Ecs

# RDS 삭제
aws cloudformation delete-stack --stack-name Conduit-Rds
aws cloudformation wait stack-delete-complete --stack-name Conduit-Rds

# VPC 마지막 삭제
aws cloudformation delete-stack --stack-name Conduit-Vpc
aws cloudformation wait stack-delete-complete --stack-name Conduit-Vpc
```

### 진행 중인 배포 중단 및 재시작
```bash
# 배포 중인 스택 삭제
aws cloudformation delete-stack --stack-name Conduit-Ecs

# 삭제 완료 대기
aws cloudformation wait stack-delete-complete --stack-name Conduit-Ecs

# 재배포
npx cdk deploy Conduit-Ecs --require-approval never
```

---

## 비용 최적화 설정

이 프로젝트는 비용 최소화를 위해 다음 설정을 사용합니다:

| 리소스 | 설정 | 예상 비용/월 |
|--------|------|-------------|
| NAT Instance | t4g.nano | ~$3 |
| RDS | db.t4g.micro | ~$12 |
| ECS Fargate | 0.25 vCPU, 512MB | ~$9 |
| ALB | 단일 AZ | ~$16 |
| **총계** | | **~$40** |

---

## 참고 사항

1. **RDS 생성 시간**: RDS 인스턴스 생성에는 약 10-15분이 소요됩니다.
2. **NAT Instance vs NAT Gateway**: 비용 절감을 위해 NAT Gateway 대신 NAT Instance를 사용합니다.
3. **ARM64 아키텍처**: ECS Fargate와 RDS 모두 ARM64(Graviton)를 사용하여 비용을 절감합니다.
4. **단일 인스턴스**: 프로덕션이지만 비용 절감을 위해 ECS 서비스는 단일 인스턴스로 운영합니다.
