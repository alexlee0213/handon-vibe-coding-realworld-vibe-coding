import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import { Construct } from 'constructs';

export interface EcsStackProps extends cdk.StackProps {
  readonly environment: string;
  readonly vpc: ec2.IVpc;
  readonly dbSecret: secretsmanager.ISecret;
  readonly dbEndpoint: string;
  readonly dbPort: string;
  readonly jwtSecretArn: string;
}

export class EcsStack extends cdk.Stack {
  public readonly ecrRepository: ecr.IRepository;
  public readonly ecsCluster: ecs.Cluster;
  public readonly fargateService: ecs.FargateService;
  public readonly loadBalancer: elbv2.ApplicationLoadBalancer;

  constructor(scope: Construct, id: string, props: EcsStackProps) {
    super(scope, id, props);

    const { environment, vpc, dbSecret, dbEndpoint, dbPort, jwtSecretArn } = props;

    // Import existing ECR Repository (created manually or by previous deployment)
    // This avoids conflicts when the repository already exists
    this.ecrRepository = ecr.Repository.fromRepositoryName(
      this,
      'ConduitBackendRepo',
      `conduit-${environment}-backend`
    );

    // ECS Cluster
    this.ecsCluster = new ecs.Cluster(this, 'ConduitCluster', {
      clusterName: `conduit-${environment}-cluster`,
      vpc,
      containerInsights: false, // Disabled for cost savings
    });

    // CloudWatch Log Group
    const logGroup = new logs.LogGroup(this, 'BackendLogGroup', {
      logGroupName: `/ecs/conduit-${environment}-backend`,
      retention: logs.RetentionDays.ONE_WEEK, // Minimal retention for cost savings
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Task Execution Role
    const executionRole = new iam.Role(this, 'TaskExecutionRole', {
      roleName: `conduit-${environment}-task-execution-role`,
      assumedBy: new iam.ServicePrincipal('ecs-tasks.amazonaws.com'),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AmazonECSTaskExecutionRolePolicy'),
      ],
    });

    // Grant secrets access to execution role
    dbSecret.grantRead(executionRole);
    const jwtSecret = secretsmanager.Secret.fromSecretCompleteArn(this, 'JwtSecret', jwtSecretArn);
    jwtSecret.grantRead(executionRole);

    // Task Role (for application permissions)
    const taskRole = new iam.Role(this, 'TaskRole', {
      roleName: `conduit-${environment}-task-role`,
      assumedBy: new iam.ServicePrincipal('ecs-tasks.amazonaws.com'),
    });

    // Task Definition - Minimal resources for cost savings
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'BackendTaskDef', {
      family: `conduit-${environment}-backend`,
      cpu: 256, // 0.25 vCPU - minimum
      memoryLimitMiB: 512, // 512MB - minimum for 0.25 vCPU
      executionRole,
      taskRole,
      runtimePlatform: {
        cpuArchitecture: ecs.CpuArchitecture.ARM64, // ARM for cost savings
        operatingSystemFamily: ecs.OperatingSystemFamily.LINUX,
      },
    });

    // Container Definition
    const container = taskDefinition.addContainer('BackendContainer', {
      containerName: 'conduit-backend',
      image: ecs.ContainerImage.fromEcrRepository(this.ecrRepository, 'latest'),
      logging: ecs.LogDrivers.awsLogs({
        streamPrefix: 'backend',
        logGroup,
      }),
      environment: {
        SERVER_ENV: 'production',
        SERVER_PORT: '8080',
        DB_HOST: dbEndpoint,
        DB_PORT: dbPort,
        DB_NAME: 'conduit',
        DB_SSLMODE: 'require',
        JWT_EXPIRY: '72h',
        // CORS: Allow GitHub Pages frontend domain
        // Update this when you have a custom domain
        CORS_ALLOWED_ORIGINS: 'https://alexlee0213.github.io',
      },
      secrets: {
        DB_USERNAME: ecs.Secret.fromSecretsManager(dbSecret, 'username'),
        DB_PASSWORD: ecs.Secret.fromSecretsManager(dbSecret, 'password'),
        JWT_SECRET: ecs.Secret.fromSecretsManager(jwtSecret, 'jwt_secret'),
      },
      portMappings: [
        {
          containerPort: 8080,
          protocol: ecs.Protocol.TCP,
        },
      ],
      healthCheck: {
        command: ['CMD-SHELL', 'curl -f http://localhost:8080/health || exit 1'],
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        retries: 3,
        startPeriod: cdk.Duration.seconds(60),
      },
    });

    // Security Group for Fargate Service
    const serviceSecurityGroup = new ec2.SecurityGroup(this, 'ServiceSecurityGroup', {
      vpc,
      securityGroupName: `conduit-${environment}-service-sg`,
      description: 'Security group for ECS Fargate service',
      allowAllOutbound: true,
    });

    // Note: RDS security group already allows access from private subnets
    // where ECS Fargate runs, so no additional ingress rule needed here

    // Application Load Balancer
    this.loadBalancer = new elbv2.ApplicationLoadBalancer(this, 'BackendAlb', {
      loadBalancerName: `conduit-${environment}-alb`,
      vpc,
      internetFacing: true,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC,
      },
    });

    // ALB Security Group - allow HTTP from anywhere
    this.loadBalancer.connections.allowFromAnyIpv4(ec2.Port.tcp(80), 'Allow HTTP');

    // Allow service to receive traffic from ALB
    serviceSecurityGroup.addIngressRule(
      this.loadBalancer.connections.securityGroups[0],
      ec2.Port.tcp(8080),
      'Allow traffic from ALB'
    );

    // Target Group
    const targetGroup = new elbv2.ApplicationTargetGroup(this, 'BackendTargetGroup', {
      targetGroupName: `conduit-${environment}-tg`,
      vpc,
      port: 8080,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: '/health',
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 3,
        healthyHttpCodes: '200',
      },
    });

    // HTTP Listener
    this.loadBalancer.addListener('HttpListener', {
      port: 80,
      defaultTargetGroups: [targetGroup],
    });

    // Fargate Service - Single instance for cost savings
    this.fargateService = new ecs.FargateService(this, 'BackendService', {
      serviceName: `conduit-${environment}-backend`,
      cluster: this.ecsCluster,
      taskDefinition,
      desiredCount: 1, // Single instance for cost savings
      securityGroups: [serviceSecurityGroup],
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      assignPublicIp: false,
      enableECSManagedTags: true,
      propagateTags: ecs.PropagatedTagSource.SERVICE,
      circuitBreaker: {
        rollback: true,
      },
      minHealthyPercent: 0, // Allow 0% during deployment (single instance)
      maxHealthyPercent: 200,
    });

    // Register service with target group
    this.fargateService.attachToApplicationTargetGroup(targetGroup);

    // Outputs
    new cdk.CfnOutput(this, 'EcrRepositoryUri', {
      value: this.ecrRepository.repositoryUri,
      description: 'ECR Repository URI',
      exportName: `conduit-${environment}-ecr-uri`,
    });

    new cdk.CfnOutput(this, 'ClusterName', {
      value: this.ecsCluster.clusterName,
      description: 'ECS Cluster Name',
      exportName: `conduit-${environment}-cluster-name`,
    });

    new cdk.CfnOutput(this, 'ServiceName', {
      value: this.fargateService.serviceName,
      description: 'ECS Service Name',
      exportName: `conduit-${environment}-service-name`,
    });

    new cdk.CfnOutput(this, 'LoadBalancerDns', {
      value: this.loadBalancer.loadBalancerDnsName,
      description: 'Application Load Balancer DNS',
      exportName: `conduit-${environment}-alb-dns`,
    });

    new cdk.CfnOutput(this, 'ApiUrl', {
      value: `http://${this.loadBalancer.loadBalancerDnsName}`,
      description: 'API Base URL',
      exportName: `conduit-${environment}-api-url`,
    });
  }
}
