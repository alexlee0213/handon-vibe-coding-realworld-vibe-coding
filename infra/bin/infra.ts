#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { VpcStack } from '../lib/vpc-stack';

const app = new cdk.App();

// Get environment from context or default to 'staging'
const environment = app.node.tryGetContext('environment') || 'staging';

// AWS environment configuration
const env: cdk.Environment = {
  account: process.env.CDK_DEFAULT_ACCOUNT || process.env.AWS_ACCOUNT_ID,
  region: process.env.CDK_DEFAULT_REGION || process.env.AWS_REGION || 'ap-northeast-2',
};

// VPC Stack
new VpcStack(app, `Conduit-${environment}-Vpc`, {
  env,
  environment,
  description: `RealWorld Conduit VPC Stack (${environment})`,
  tags: {
    Project: 'conduit',
    Environment: environment,
  },
});

app.synth();
