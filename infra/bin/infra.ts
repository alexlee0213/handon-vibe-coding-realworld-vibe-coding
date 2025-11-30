#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { VpcStack } from '../lib/vpc-stack';
import { RdsStack } from '../lib/rds-stack';
import { EcsStack } from '../lib/ecs-stack';
import { CloudFrontStack } from '../lib/cloudfront-stack';

const app = new cdk.App();

// Production Only - No staging environment for cost savings
const environment = 'production';

// AWS environment configuration
const env: cdk.Environment = {
  account: process.env.CDK_DEFAULT_ACCOUNT || process.env.AWS_ACCOUNT_ID,
  region: process.env.CDK_DEFAULT_REGION || process.env.AWS_REGION || 'us-east-1',
};

const commonTags = {
  Project: 'conduit',
  Environment: environment,
  ManagedBy: 'CDK',
};

// Stack 1: VPC with NAT Instance
const vpcStack = new VpcStack(app, 'Conduit-Vpc', {
  env,
  environment,
  description: 'RealWorld Conduit VPC Stack with NAT Instance',
  tags: commonTags,
});

// Stack 2: RDS PostgreSQL (depends on VPC)
const rdsStack = new RdsStack(app, 'Conduit-Rds', {
  env,
  environment,
  vpc: vpcStack.vpc,
  description: 'RealWorld Conduit RDS PostgreSQL Stack',
  tags: commonTags,
});
rdsStack.addDependency(vpcStack);

// Stack 3: ECS Fargate (depends on VPC and RDS)
const ecsStack = new EcsStack(app, 'Conduit-Ecs', {
  env,
  environment,
  vpc: vpcStack.vpc,
  dbSecret: rdsStack.dbSecret,
  dbEndpoint: rdsStack.dbInstance.dbInstanceEndpointAddress,
  dbPort: rdsStack.dbInstance.dbInstanceEndpointPort,
  jwtSecretArn: cdk.Fn.importValue(`conduit-${environment}-jwt-secret-arn`),
  dbSecurityGroupId: rdsStack.dbSecurityGroup.securityGroupId,
  description: 'RealWorld Conduit ECS Fargate Stack',
  tags: commonTags,
});
ecsStack.addDependency(rdsStack);

// Stack 4: CloudFront for HTTPS (depends on ECS for ALB)
const cloudFrontStack = new CloudFrontStack(app, 'Conduit-CloudFront', {
  env,
  environment,
  albDnsName: ecsStack.loadBalancer.loadBalancerDnsName,
  description: 'RealWorld Conduit CloudFront Distribution for HTTPS',
  tags: commonTags,
});
cloudFrontStack.addDependency(ecsStack);

app.synth();
