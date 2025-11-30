import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as rds from 'aws-cdk-lib/aws-rds';
import * as secretsmanager from 'aws-cdk-lib/aws-secretsmanager';
import { Construct } from 'constructs';

export interface RdsStackProps extends cdk.StackProps {
  readonly environment: string;
  readonly vpc: ec2.IVpc;
}

export class RdsStack extends cdk.Stack {
  public readonly dbInstance: rds.DatabaseInstance;
  public readonly dbSecret: secretsmanager.ISecret;
  public readonly dbSecurityGroup: ec2.SecurityGroup;

  constructor(scope: Construct, id: string, props: RdsStackProps) {
    super(scope, id, props);

    const { environment, vpc } = props;

    // Security Group for RDS - Only allow access from Private Subnets
    this.dbSecurityGroup = new ec2.SecurityGroup(this, 'DbSecurityGroup', {
      vpc,
      securityGroupName: `conduit-${environment}-db-sg`,
      description: 'Security group for RDS PostgreSQL',
      allowAllOutbound: false,
    });

    // Allow PostgreSQL access from Private Subnets only
    vpc.privateSubnets.forEach((subnet, index) => {
      this.dbSecurityGroup.addIngressRule(
        ec2.Peer.ipv4(subnet.ipv4CidrBlock),
        ec2.Port.tcp(5432),
        `Allow PostgreSQL from Private Subnet ${index + 1}`
      );
    });

    // RDS PostgreSQL Instance - Minimal cost configuration
    this.dbInstance = new rds.DatabaseInstance(this, 'ConduitDb', {
      instanceIdentifier: `conduit-${environment}-db`,
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_16,
      }),

      // Minimal instance class - ARM-based for cost savings
      // t4g.micro: 2 vCPU, 1GB RAM, Free Tier eligible for 12 months
      instanceType: ec2.InstanceType.of(ec2.InstanceClass.T4G, ec2.InstanceSize.MICRO),

      // Storage configuration - minimal
      allocatedStorage: 20, // Minimum for gp3
      storageType: rds.StorageType.GP3,
      maxAllocatedStorage: 100, // Allow auto-scaling if needed

      // Network configuration
      vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_ISOLATED, // Isolated subnet for security
      },
      securityGroups: [this.dbSecurityGroup],

      // Database configuration
      databaseName: 'conduit',
      port: 5432,

      // High availability - disabled for cost savings
      multiAz: false,

      // Backup configuration - minimal
      backupRetention: cdk.Duration.days(1),
      deleteAutomatedBackups: true,

      // Maintenance
      preferredMaintenanceWindow: 'Sun:03:00-Sun:04:00',
      preferredBackupWindow: '02:00-03:00',

      // Deletion configuration - allow deletion for dev/test purposes
      deletionProtection: false,
      removalPolicy: cdk.RemovalPolicy.DESTROY,

      // Monitoring
      enablePerformanceInsights: false, // Disabled for cost savings
      cloudwatchLogsExports: ['postgresql'], // Export PostgreSQL logs

      // Credentials - auto-generated and stored in Secrets Manager
      credentials: rds.Credentials.fromGeneratedSecret('conduit_admin', {
        secretName: `conduit-${environment}-db-credentials`,
      }),
    });

    // Store the secret reference
    this.dbSecret = this.dbInstance.secret!;

    // Create application secret for JWT
    const jwtSecret = new secretsmanager.Secret(this, 'JwtSecret', {
      secretName: `conduit-${environment}-jwt-secret`,
      description: 'JWT secret for Conduit application',
      generateSecretString: {
        secretStringTemplate: JSON.stringify({}),
        generateStringKey: 'jwt_secret',
        excludeCharacters: '"@/\\',
        passwordLength: 64,
      },
    });

    // Outputs
    new cdk.CfnOutput(this, 'DbEndpoint', {
      value: this.dbInstance.dbInstanceEndpointAddress,
      description: 'Database endpoint',
      exportName: `conduit-${environment}-db-endpoint`,
    });

    new cdk.CfnOutput(this, 'DbPort', {
      value: this.dbInstance.dbInstanceEndpointPort,
      description: 'Database port',
      exportName: `conduit-${environment}-db-port`,
    });

    new cdk.CfnOutput(this, 'DbSecretArn', {
      value: this.dbSecret.secretArn,
      description: 'Database credentials secret ARN',
      exportName: `conduit-${environment}-db-secret-arn`,
    });

    new cdk.CfnOutput(this, 'DbSecurityGroupId', {
      value: this.dbSecurityGroup.securityGroupId,
      description: 'Database security group ID',
      exportName: `conduit-${environment}-db-sg-id`,
    });

    new cdk.CfnOutput(this, 'JwtSecretArn', {
      value: jwtSecret.secretArn,
      description: 'JWT secret ARN',
      exportName: `conduit-${environment}-jwt-secret-arn`,
    });
  }
}
