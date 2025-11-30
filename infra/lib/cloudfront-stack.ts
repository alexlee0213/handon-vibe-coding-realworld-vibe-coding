import * as cdk from 'aws-cdk-lib';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import { Construct } from 'constructs';

export interface CloudFrontStackProps extends cdk.StackProps {
  readonly environment: string;
  readonly albDnsName: string;
}

export class CloudFrontStack extends cdk.Stack {
  public readonly distribution: cloudfront.Distribution;
  public readonly distributionDomainName: string;

  constructor(scope: Construct, id: string, props: CloudFrontStackProps) {
    super(scope, id, props);

    const { environment, albDnsName } = props;

    // Create CloudFront distribution with ALB as origin
    this.distribution = new cloudfront.Distribution(this, 'ApiDistribution', {
      comment: `Conduit ${environment} API CloudFront Distribution`,
      defaultBehavior: {
        origin: new origins.HttpOrigin(albDnsName, {
          protocolPolicy: cloudfront.OriginProtocolPolicy.HTTP_ONLY,
          httpPort: 80,
        }),
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
        originRequestPolicy: cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
        responseHeadersPolicy: this.createCorsPolicy(environment),
      },
      priceClass: cloudfront.PriceClass.PRICE_CLASS_100, // North America & Europe only for cost savings
      enabled: true,
      httpVersion: cloudfront.HttpVersion.HTTP2,
    });

    this.distributionDomainName = this.distribution.distributionDomainName;

    // Outputs
    new cdk.CfnOutput(this, 'DistributionId', {
      value: this.distribution.distributionId,
      description: 'CloudFront Distribution ID',
      exportName: `conduit-${environment}-cf-distribution-id`,
    });

    new cdk.CfnOutput(this, 'DistributionDomainName', {
      value: this.distribution.distributionDomainName,
      description: 'CloudFront Distribution Domain Name',
      exportName: `conduit-${environment}-cf-domain`,
    });

    new cdk.CfnOutput(this, 'ApiUrl', {
      value: `https://${this.distribution.distributionDomainName}`,
      description: 'API Base URL (HTTPS via CloudFront)',
      exportName: `conduit-${environment}-api-url-https`,
    });
  }

  private createCorsPolicy(environment: string): cloudfront.ResponseHeadersPolicy {
    return new cloudfront.ResponseHeadersPolicy(this, 'CorsPolicy', {
      responseHeadersPolicyName: `conduit-${environment}-cors-policy`,
      corsBehavior: {
        accessControlAllowCredentials: false,
        accessControlAllowHeaders: ['Authorization', 'Content-Type', 'X-Requested-With'],
        accessControlAllowMethods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        accessControlAllowOrigins: [
          'https://alexlee0213.github.io',
          'http://localhost:5173',
          'http://localhost:3000',
        ],
        accessControlMaxAge: cdk.Duration.seconds(86400),
        originOverride: true,
      },
    });
  }
}
