# AWS Marketplace Listing — ChaosPlane

> This document covers the preparation and configuration for ChaosPlane's AWS Marketplace listing. It includes pricing model design, metering dimensions, listing metadata, Private Offer configuration, and AWS Partner Network (APN) registration requirements.

---

## Listing Overview

| Field | Value |
|---|---|
| Product name | ChaosPlane |
| Tagline | Chaos engineering platform for cloud-native resilience testing |
| Category | Developer Tools > Testing |
| Sub-category | Chaos Engineering / Resilience Testing |
| Delivery method | SaaS subscription |
| Fulfillment URL | https://app.chaosplane.dev/marketplace/aws/activate |
| Support URL | https://docs.chaosplane.dev / support@chaosplane.dev |
| EULA | ChaosPlane SaaS Terms of Service (custom EULA uploaded to Marketplace) |
| Refund policy | 30-day money-back guarantee on annual plans |

---

## Pricing Models

ChaosPlane offers two pricing models on AWS Marketplace: Pay-As-You-Go (PAYG) and Annual Contract. Both use AWS Marketplace Metering Service for usage-based dimensions.

### Model 1: Pay-As-You-Go (PAYG)

Customers pay monthly based on actual usage. No upfront commitment. Billed through AWS consolidated billing.

| Dimension | Unit | Price |
|---|---|---|
| `chaos_experiments` | Per experiment executed | $0.15 |
| `managed_nodes` | Per Kubernetes node monitored (per hour) | $0.008 |
| `ai_analysis_requests` | Per AI assistant analysis call | $0.05 |
| `audit_log_retention_gb` | Per GB of audit log storage beyond 10 GB free | $0.10 |

Monthly minimums: none. Free tier: 10 experiments/month, 2 nodes, 20 AI calls.

PAYG is suited for teams evaluating ChaosPlane or running low-frequency chaos programs. It maps to the Starter and Growth plan tiers.

### Model 2: Annual Contract (SaaS Contract)

Customers commit to an annual subscription with a fixed seat/node count. Discounted vs. PAYG. Billed upfront or quarterly through AWS Marketplace.

| Tier | Included | Annual Price | Overage |
|---|---|---|---|
| Growth | 500 experiments/mo, 50 nodes, 500 AI calls/mo | $18,000/yr | PAYG rates |
| Enterprise | 5,000 experiments/mo, 500 nodes, unlimited AI calls | $72,000/yr | Custom |
| Enterprise+ | Unlimited experiments, unlimited nodes, dedicated support | Custom | N/A |

Annual contract customers receive:
- 20% discount vs. equivalent PAYG spend
- Dedicated Slack channel with ChaosPlane engineering
- Quarterly business reviews
- Priority support SLA (1-hour response for P0/P1)

### Model 3: Private Offer

For enterprise customers with specific pricing, custom terms, or multi-year commitments. Private Offers are created via the AWS Marketplace Management Portal and sent directly to the customer's AWS account.

Private Offer use cases:
- Multi-year deals (2-3 year terms with additional discounts)
- Custom contract terms (data processing addendum, BAA, custom SLA)
- Government customers requiring specific procurement vehicles
- SI partner resale (partner creates Private Offer on behalf of end customer)

Private Offer configuration:
- Offer duration: up to 36 months
- Payment schedule: upfront, quarterly, or monthly
- Custom dimensions: negotiated per deal
- Expiry: 30 days from offer creation (configurable)

---

## AWS Marketplace Metering Service Integration

ChaosPlane uses the AWS Marketplace Metering Service (`aws-marketplace-metering`) to report usage for PAYG and overage billing.

### Metering Dimensions

Registered dimensions in the Marketplace product:

```json
{
  "Dimensions": [
    {
      "Name": "chaos_experiments",
      "Description": "Number of chaos experiments executed",
      "Unit": "Count",
      "Types": ["Metered"]
    },
    {
      "Name": "managed_nodes",
      "Description": "Kubernetes nodes under active chaos management (node-hours)",
      "Unit": "Hrs",
      "Types": ["Metered"]
    },
    {
      "Name": "ai_analysis_requests",
      "Description": "AI assistant analysis and recommendation calls",
      "Unit": "Count",
      "Types": ["Metered"]
    },
    {
      "Name": "audit_log_retention_gb",
      "Description": "Audit log storage beyond 10 GB free tier",
      "Unit": "GB",
      "Types": ["Metered"]
    }
  ]
}
```

### Metering API Call Pattern

ChaosPlane's billing service calls `BatchMeterUsage` once per hour per customer:

```go
// Simplified metering call — runs every hour per tenant
func reportUsage(ctx context.Context, tenantID string, usage UsageSummary) error {
    svc := marketplacemetering.New(sess)
    _, err := svc.BatchMeterUsageWithContext(ctx, &marketplacemetering.BatchMeterUsageInput{
        ProductCode: aws.String(os.Getenv("AWS_MARKETPLACE_PRODUCT_CODE")),
        UsageRecords: []*marketplacemetering.UsageRecord{
            {
                CustomerIdentifier: aws.String(usage.CustomerIdentifier),
                Dimension:          aws.String("chaos_experiments"),
                Quantity:           aws.Int64(usage.ExperimentsThisHour),
                Timestamp:          aws.Time(time.Now()),
            },
            {
                CustomerIdentifier: aws.String(usage.CustomerIdentifier),
                Dimension:          aws.String("managed_nodes"),
                Quantity:           aws.Int64(usage.NodeHours),
                Timestamp:          aws.Time(time.Now()),
            },
        },
    })
    return err
}
```

Metering records are idempotent within a 1-hour window. Duplicate submissions for the same customer/dimension/hour are deduplicated by AWS.

### Customer Registration Flow

When a customer subscribes via AWS Marketplace:

1. AWS sends a `subscribe-success` notification to ChaosPlane's SNS endpoint
2. ChaosPlane receives the `x-amzn-marketplace-token` from the redirect URL
3. ChaosPlane calls `ResolveCustomer` to exchange the token for a `CustomerIdentifier`
4. ChaosPlane creates or links the customer's account with the `CustomerIdentifier`
5. Customer is redirected to the ChaosPlane onboarding flow

```
[Customer clicks Subscribe on Marketplace]
        |
        v
[AWS Marketplace] → POST https://app.chaosplane.dev/marketplace/aws/activate
                    (with x-amzn-marketplace-token)
        |
        v
[ChaosPlane] → ResolveCustomer(token) → CustomerIdentifier
        |
        v
[Account created / linked] → Customer onboarding
```

---

## Listing Metadata

### Short Description (255 chars)

ChaosPlane is a chaos engineering platform for Kubernetes, AWS, Azure, and GCP. Run fault injection experiments, measure resilience, and prevent outages before they happen. 40+ executor types, AI-assisted analysis, enterprise-grade compliance.

### Long Description

ChaosPlane helps engineering teams build more resilient systems by running controlled chaos experiments against their infrastructure. Instead of discovering failure modes during production incidents, teams use ChaosPlane to find and fix weaknesses proactively.

**What ChaosPlane does:**

- Fault injection across Kubernetes (pod kill, network chaos, node drain, eBPF-level), AWS (EC2, RDS, ECS, Lambda, S3), Azure (VM, AKS, Cosmos DB), and GCP (GKE, Cloud SQL, Cloud Run)
- 40+ executor types covering pods, nodes, networks, VMs, cloud services, and kernel-level faults
- Workflow engine with DAG-based experiment orchestration, parallel execution, and conditional branching
- Steady-state validation: define what "healthy" looks like before and after each experiment
- AI assistant: topology analysis, vulnerability detection, experiment recommendations, and result summarization
- Resilience scoring: track your system's resilience over time with quantitative metrics
- Enterprise features: RBAC/ABAC, SAML SSO, SCIM provisioning, audit logs, multi-cluster federation

**Compliance and security:**

ChaosPlane holds SOC 2 Type II, ISO 27001, and ISMS-P certifications. FedRAMP Moderate ATO and DoD IL4/IL5 are available for government customers via the GovCloud deployment. CSAP 상등급 covers Korean public sector customers.

**Deployment:**

ChaosPlane is available as a fully managed SaaS (this listing) or as a self-hosted deployment on your own Kubernetes cluster. The chaos agent runs in your environment and communicates with the ChaosPlane control plane over mTLS.

### Keywords

chaos engineering, resilience testing, fault injection, Kubernetes chaos, GameDay, SRE, reliability, chaos monkey, chaos mesh, litmus chaos, AWS resilience, cloud resilience, eBPF chaos, network chaos, pod kill

### Categories

- Developer Tools > Testing
- Security > Compliance
- Management Tools > Monitoring

---

## AWS Partner Network (APN) Registration

ChaosPlane registers as an AWS ISV Partner to access Marketplace benefits and co-sell programs.

### APN Tier Target

**ISV Accelerate** — the co-sell program for SaaS products listed on AWS Marketplace.

Requirements for ISV Accelerate:
- Active AWS Marketplace listing (SaaS) ✓
- AWS Foundational Technical Review (FTR) passed
- AWS Well-Architected Review completed
- Minimum 3 customer references on AWS

### AWS Foundational Technical Review (FTR) Checklist

The FTR validates that ChaosPlane follows AWS security and operational best practices.

| Area | Requirement | Status |
|---|---|---|
| IAM | No hardcoded credentials, least-privilege roles | Done |
| IAM | MFA on all AWS accounts | Done |
| Logging | CloudTrail enabled in all accounts | Done |
| Logging | VPC Flow Logs enabled | Done |
| Encryption | Encryption at rest (KMS) | Done |
| Encryption | Encryption in transit (TLS 1.2+) | Done |
| Networking | No 0.0.0.0/0 inbound rules except ALB/CloudFront | Done |
| Networking | Private subnets for application and data tiers | Done |
| Vulnerability | Automated CVE scanning (Trivy, Dependabot) | Done |
| Incident response | IR plan documented, contacts registered | Done |
| Marketplace | Metering integration tested | In progress |
| Marketplace | ResolveCustomer flow tested end-to-end | In progress |

### Co-Sell Configuration

Once ISV Accelerate is active, ChaosPlane registers Opportunity records in AWS Partner Central for deals where AWS is involved. This unlocks:

- AWS field seller engagement on qualified opportunities
- AWS co-sell funding for proof-of-concept projects
- AWS Marketplace private pricing for large deals
- Joint go-to-market activities (blog posts, webinars, re:Invent presence)

---

## Pricing Page Copy

### Starter — Free

- 10 experiments/month
- 2 Kubernetes nodes
- 20 AI analysis calls/month
- Community support
- Available via AWS Marketplace free tier

### Growth — $1,500/month (PAYG) or $18,000/year

- 500 experiments/month
- 50 Kubernetes nodes
- 500 AI analysis calls/month
- Email support, 8-hour response SLA
- SAML SSO
- Audit logs (90-day retention)

### Enterprise — from $6,000/month or $72,000/year

- 5,000 experiments/month
- 500 Kubernetes nodes
- Unlimited AI analysis
- Dedicated Slack, 1-hour P0/P1 SLA
- RBAC/ABAC, SCIM provisioning
- Multi-cluster federation
- Custom compliance reports
- Private Offer available

### Government — Custom

- FedRAMP Moderate ATO deployment (GovCloud)
- DoD IL4/IL5 available
- CSAP 상등급 (Korea)
- Air-gap deployment option
- Contact sales for pricing

---

## Launch Checklist

- [ ] AWS Marketplace Management Portal: product created, dimensions registered
- [ ] SNS subscription endpoint deployed and tested (`/marketplace/aws/activate`)
- [ ] `ResolveCustomer` integration tested with sandbox token
- [ ] `BatchMeterUsage` integration tested with sandbox product code
- [ ] Fulfillment URL returns 200 and creates test account
- [ ] Unsubscribe flow tested (account suspension on `unsubscribe-success`)
- [ ] EULA uploaded (ChaosPlane SaaS Terms of Service)
- [ ] Product screenshots uploaded (dashboard, experiment builder, resilience score)
- [ ] AWS FTR submitted and approved
- [ ] APN ISV Accelerate application submitted
- [ ] Private Offer template configured for enterprise deals
- [ ] GovCloud listing created separately (AWS GovCloud Marketplace)

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | In preparation |
| Owner | Business Development / Engineering |
| Last reviewed | April 2026 |
| Next review | Upon Marketplace listing approval |
