import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

export interface VpcStackProps extends cdk.StackProps {
  readonly environment: string;
}

export class VpcStack extends cdk.Stack {
  public readonly vpc: ec2.Vpc;
  public readonly natInstance: ec2.Instance;

  constructor(scope: Construct, id: string, props: VpcStackProps) {
    super(scope, id, props);

    const { environment } = props;

    // Create VPC without NAT Gateway (we'll use NAT Instance for cost savings)
    this.vpc = new ec2.Vpc(this, 'ConduitVpc', {
      vpcName: `conduit-${environment}-vpc`,
      maxAzs: 2,
      natGateways: 0, // No NAT Gateway - using NAT Instance instead
      subnetConfiguration: [
        {
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
          cidrMask: 24,
        },
        {
          name: 'Private',
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
          cidrMask: 24,
        },
        {
          name: 'Isolated',
          subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
          cidrMask: 24,
        },
      ],
    });

    // NAT Instance Security Group
    const natSecurityGroup = new ec2.SecurityGroup(this, 'NatSecurityGroup', {
      vpc: this.vpc,
      securityGroupName: `conduit-${environment}-nat-sg`,
      description: 'Security group for NAT Instance',
      allowAllOutbound: true,
    });

    // Allow all traffic from Private Subnets
    this.vpc.privateSubnets.forEach((subnet, index) => {
      natSecurityGroup.addIngressRule(
        ec2.Peer.ipv4(subnet.ipv4CidrBlock),
        ec2.Port.allTraffic(),
        `Allow all traffic from Private Subnet ${index + 1}`
      );
    });

    // NAT Instance - t4g.nano (ARM, ~$3/month)
    // Using Amazon Linux 2023 with NAT configuration
    this.natInstance = new ec2.Instance(this, 'NatInstance', {
      instanceName: `conduit-${environment}-nat`,
      vpc: this.vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC,
      },
      instanceType: ec2.InstanceType.of(ec2.InstanceClass.T4G, ec2.InstanceSize.NANO),
      machineImage: ec2.MachineImage.latestAmazonLinux2023({
        cpuType: ec2.AmazonLinuxCpuType.ARM_64,
      }),
      securityGroup: natSecurityGroup,
      sourceDestCheck: false, // Required for NAT functionality
      associatePublicIpAddress: true,
    });

    // Configure NAT Instance with user data
    this.natInstance.addUserData(
      '#!/bin/bash',
      'set -e',
      '',
      '# Enable IP forwarding',
      'echo 1 > /proc/sys/net/ipv4/ip_forward',
      'echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf',
      '',
      '# Configure iptables for NAT',
      'yum install -y iptables-services',
      'systemctl enable iptables',
      'systemctl start iptables',
      '',
      '# Get the primary network interface',
      'PRIMARY_IF=$(ip route | grep default | awk \'{print $5}\')',
      '',
      '# Set up NAT masquerading',
      'iptables -t nat -A POSTROUTING -o $PRIMARY_IF -j MASQUERADE',
      'iptables -F FORWARD',
      'iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT',
      'iptables -A FORWARD -i $PRIMARY_IF -j ACCEPT',
      '',
      '# Save iptables rules',
      'service iptables save',
      '',
      'echo "NAT Instance configuration complete"'
    );

    // Add routes from Private Subnets to NAT Instance
    this.vpc.privateSubnets.forEach((subnet, index) => {
      const routeTable = subnet.routeTable;
      new ec2.CfnRoute(this, `PrivateSubnetRoute${index + 1}`, {
        routeTableId: routeTable.routeTableId,
        destinationCidrBlock: '0.0.0.0/0',
        instanceId: this.natInstance.instanceId,
      });
    });

    // Outputs
    new cdk.CfnOutput(this, 'VpcId', {
      value: this.vpc.vpcId,
      description: 'VPC ID',
      exportName: `conduit-${environment}-vpc-id`,
    });

    new cdk.CfnOutput(this, 'PublicSubnetIds', {
      value: this.vpc.publicSubnets.map((s) => s.subnetId).join(','),
      description: 'Public Subnet IDs',
      exportName: `conduit-${environment}-public-subnet-ids`,
    });

    new cdk.CfnOutput(this, 'PrivateSubnetIds', {
      value: this.vpc.privateSubnets.map((s) => s.subnetId).join(','),
      description: 'Private Subnet IDs',
      exportName: `conduit-${environment}-private-subnet-ids`,
    });

    new cdk.CfnOutput(this, 'IsolatedSubnetIds', {
      value: this.vpc.isolatedSubnets.map((s) => s.subnetId).join(','),
      description: 'Isolated Subnet IDs',
      exportName: `conduit-${environment}-isolated-subnet-ids`,
    });

    new cdk.CfnOutput(this, 'NatInstanceId', {
      value: this.natInstance.instanceId,
      description: 'NAT Instance ID',
      exportName: `conduit-${environment}-nat-instance-id`,
    });
  }
}
