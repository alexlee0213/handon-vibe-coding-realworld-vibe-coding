# Deployment Guide

This guide covers deploying RealWorld Conduit to AWS using CDK.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           GitHub Pages                               │
│                         (Frontend - React)                           │
└────────────────────────────────┬────────────────────────────────────┘
                                 │ HTTPS
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Application Load Balancer                     │
│                           (HTTP - Port 80)                          │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
            ┌────────────────────┴────────────────────┐
            │              Public Subnet              │
            │         ┌─────────────────────┐         │
            │         │    NAT Instance     │         │
            │         │     (t4g.nano)      │         │
            │         └──────────┬──────────┘         │
            └────────────────────┼────────────────────┘
                                 │
            ┌────────────────────┴────────────────────┐
            │             Private Subnet              │
            │         ┌─────────────────────┐         │
            │         │   ECS Fargate       │         │
            │         │ (Go Backend - ARM64)│         │
            │         │  256 CPU / 512MB    │         │
            │         └──────────┬──────────┘         │
            └────────────────────┼────────────────────┘
                                 │
            ┌────────────────────┴────────────────────┐
            │            Isolated Subnet              │
            │         ┌─────────────────────┐         │
            │         │   RDS PostgreSQL    │         │
            │         │    (db.t4g.micro)   │         │
            │         │   PostgreSQL 16     │         │
            │         └─────────────────────┘         │
            └─────────────────────────────────────────┘
```

## Cost Estimate

Estimated monthly cost: **$30-45/month**

| Resource | Specification | Est. Cost |
|----------|--------------|-----------|
| NAT Instance | t4g.nano (spot eligible) | ~$3/month |
| ECS Fargate | 0.25 vCPU, 512MB | ~$8-10/month |
| RDS PostgreSQL | db.t4g.micro, 20GB | ~$12-15/month |
| ALB | Minimal traffic | ~$16/month |
| ECR | 5 images | ~$1/month |
| Secrets Manager | 2 secrets | ~$1/month |

## Prerequisites

### 1. AWS Account Setup

```bash
# Install AWS CLI
brew install awscli  # macOS
# or
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"

# Configure credentials
aws configure
# Enter: Access Key ID, Secret Access Key, Region (us-east-1)
```

### 2. Required Tools

```bash
# Node.js 20+
node --version

# AWS CDK
npm install -g aws-cdk

# Docker (for building images)
docker --version
```

### 3. GitHub Secrets

Configure these secrets in your GitHub repository (Settings > Secrets):

| Secret | Description |
|--------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |

### 4. GitHub Variables

Configure these variables in your GitHub repository (Settings > Variables):

| Variable | Description | Example |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://conduit-alb-xxx.us-east-1.elb.amazonaws.com` |

## Deployment Steps

### Step 1: Bootstrap CDK (First Time Only)

```bash
cd infra

# Install dependencies
npm install

# Bootstrap CDK in your AWS account
npx cdk bootstrap aws://ACCOUNT_ID/us-east-1
```

### Step 2: Deploy Infrastructure

```bash
# Synthesize CloudFormation templates
npx cdk synth

# Deploy all stacks
npx cdk deploy --all --require-approval never

# Or deploy individually in order
npx cdk deploy Conduit-Vpc
npx cdk deploy Conduit-Rds
npx cdk deploy Conduit-Ecs
```

### Step 3: Get Stack Outputs

After deployment, note the following outputs:

```bash
# Get ALB DNS
aws cloudformation describe-stacks \
  --stack-name Conduit-Ecs \
  --query "Stacks[0].Outputs[?OutputKey=='LoadBalancerDns'].OutputValue" \
  --output text

# Get ECR Repository URI
aws cloudformation describe-stacks \
  --stack-name Conduit-Ecs \
  --query "Stacks[0].Outputs[?OutputKey=='EcrRepositoryUri'].OutputValue" \
  --output text
```

### Step 4: Configure GitHub Variable

Update the `VITE_API_URL` variable in GitHub with the ALB DNS:

```
http://conduit-production-alb-xxxxxxxxx.us-east-1.elb.amazonaws.com
```

### Step 5: Deploy Backend

Push to `main` branch or manually trigger the workflow:

```bash
# Push changes to trigger deployment
git push origin main

# Or trigger manually via GitHub UI
# Actions > Deploy Backend > Run workflow
```

### Step 6: Deploy Frontend

The frontend deploys automatically when changes are pushed to `frontend/**`.

```bash
# Or trigger manually
# Actions > Deploy Frontend to GitHub Pages > Run workflow
```

## Monitoring & Troubleshooting

### View ECS Service Logs

```bash
# Using AWS CLI
aws logs tail /ecs/conduit-production-backend --follow

# Or via CloudWatch Console
# CloudWatch > Log groups > /ecs/conduit-production-backend
```

### Check ECS Service Status

```bash
aws ecs describe-services \
  --cluster conduit-production-cluster \
  --services conduit-production-backend \
  --query "services[0].{status:status,running:runningCount,desired:desiredCount}"
```

### Health Check

```bash
# Get ALB DNS
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name Conduit-Ecs \
  --query "Stacks[0].Outputs[?OutputKey=='LoadBalancerDns'].OutputValue" \
  --output text)

# Check health endpoint
curl http://${ALB_DNS}/health
```

### Common Issues

#### 1. ECS Task Failing to Start

Check CloudWatch logs for errors:
```bash
aws logs tail /ecs/conduit-production-backend --since 1h
```

Common causes:
- Database connection issues (check RDS security group)
- Invalid secrets (check Secrets Manager)
- Image pull failures (check ECR permissions)

#### 2. Database Connection Issues

Verify RDS is accessible from ECS:
```bash
# Check RDS endpoint
aws rds describe-db-instances \
  --db-instance-identifier conduit-production-db \
  --query "DBInstances[0].Endpoint"
```

#### 3. Frontend Not Connecting to Backend

1. Verify `VITE_API_URL` is set correctly in GitHub Variables
2. Check CORS configuration in backend
3. Verify ALB security group allows HTTP (port 80)

## Cleanup

To destroy all resources:

```bash
cd infra

# Destroy in reverse order
npx cdk destroy Conduit-Ecs
npx cdk destroy Conduit-Rds
npx cdk destroy Conduit-Vpc

# Or destroy all at once
npx cdk destroy --all
```

**Warning**: This will delete all data including the database!

## CI/CD Pipeline

### Backend Pipeline (`.github/workflows/deploy-backend.yml`)

Triggers:
- Push to `main` branch with changes in `backend/**`
- Manual trigger via workflow_dispatch

Steps:
1. Build Docker image (ARM64)
2. Push to ECR with commit SHA and `latest` tags
3. Update ECS task definition
4. Deploy to ECS with rolling update
5. Verify health check

### Frontend Pipeline (`.github/workflows/deploy-frontend.yml`)

Triggers:
- Push to `main` branch with changes in `frontend/**`
- Manual trigger via workflow_dispatch

Steps:
1. Install dependencies
2. Run tests
3. Build with `VITE_API_URL`
4. Deploy to GitHub Pages

## Security Considerations

1. **Secrets Management**: All secrets stored in AWS Secrets Manager
2. **Network Isolation**: RDS in isolated subnet, only accessible from private subnet
3. **No SSH Access**: NAT instance has no SSH key for security
4. **Minimal Permissions**: IAM roles follow least privilege principle
5. **SSL/TLS**: Consider adding HTTPS with ACM certificate

## Adding HTTPS (Recommended for Production)

1. Register a domain in Route 53
2. Request an ACM certificate
3. Update ALB listener to use HTTPS
4. Update `VITE_API_URL` to use HTTPS

```typescript
// In ecs-stack.ts, add HTTPS listener
const certificate = acm.Certificate.fromCertificateArn(
  this, 'Cert', 'arn:aws:acm:...'
);

this.loadBalancer.addListener('HttpsListener', {
  port: 443,
  certificates: [certificate],
  defaultTargetGroups: [targetGroup],
});
```
