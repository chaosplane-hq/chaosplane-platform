# Multi-Region Deployment Architecture — ChaosPlane

> ChaosPlane v2.0.0 operates across four geographic regions: US Commercial (us-east-1), EU (eu-west-1), APAC (ap-northeast-2 / ap-southeast-1), and GovCloud (us-gov-west-1). Each region is a fully independent deployment with its own data plane, control plane, and encryption key hierarchy. A global management console provides cross-region visibility without moving customer data across region boundaries.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     ChaosPlane Global Control Plane                         │
│                  (management console — metadata only)                       │
│                                                                             │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│   │  US Region   │  │  EU Region   │  │ APAC Region  │  │  GovCloud    │  │
│   │ us-east-1    │  │ eu-west-1    │  │ap-northeast-2│  │us-gov-west-1 │  │
│   │              │  │              │  │ap-southeast-1│  │              │  │
│   │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │  │
│   │ │ API      │ │  │ │ API      │ │  │ │ API      │ │  │ │ API      │ │  │
│   │ │ Workflow │ │  │ │ Workflow │ │  │ │ Workflow │ │  │ │ Workflow │ │  │
│   │ │ AI Asst  │ │  │ │ AI Asst  │ │  │ │ AI Asst  │ │  │ │ AI Asst  │ │  │
│   │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │  │
│   │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │  │
│   │ │PostgreSQL│ │  │ │PostgreSQL│ │  │ │PostgreSQL│ │  │ │PostgreSQL│ │  │
│   │ │ Redis    │ │  │ │ Redis    │ │  │ │ Redis    │ │  │ │ Redis    │ │  │
│   │ │ S3       │ │  │ │ S3       │ │  │ │ S3       │ │  │ │ S3       │ │  │
│   │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │  │
│   └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
│          |                  |                  |                  |         │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │          Global Metadata Bus (org membership, billing, routing)      │  │
│   │          No customer experiment data crosses region boundaries       │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Region Specifications

### US Commercial — us-east-1

Primary region. Serves North American commercial customers.

| Property | Value |
|---|---|
| AWS Region | us-east-1 |
| Data residency | United States |
| Compliance | SOC 2 Type II, ISO 27001, CMMC Level 2 |
| Availability zones | us-east-1a, us-east-1b, us-east-1c |
| Database | PostgreSQL 15, RDS Multi-AZ |
| KMS | AWS KMS (us-east-1 CMK) |
| CDN | CloudFront (us-east-1 origin) |
| Backup region | us-west-2 (cross-region backup only, no live data) |

### EU — eu-west-1

GDPR-compliant region. All EU customer data stays within the EU. No data egress to non-EU regions without explicit customer consent and a valid legal transfer mechanism.

| Property | Value |
|---|---|
| AWS Region | eu-west-1 (Ireland) |
| Data residency | European Union |
| Compliance | GDPR, SOC 2 Type II, ISO 27001 |
| Availability zones | eu-west-1a, eu-west-1b, eu-west-1c |
| Database | PostgreSQL 15, RDS Multi-AZ |
| KMS | AWS KMS (eu-west-1 CMK, separate from US) |
| CDN | CloudFront (eu-west-1 origin) |
| Backup region | eu-central-1 (Frankfurt, stays in EU) |
| DPA | ChaosPlane EU Data Processing Agreement required |
| SCCs | Standard Contractual Clauses available for non-EU processors |

GDPR-specific controls active in eu-west-1:
- Data subject rights API (access, erasure, portability, rectification) with 30-day SLA
- Consent management with granular purpose tracking
- Breach notification pipeline: CISO alert within 1 hour, DPA notification within 72 hours
- Data retention enforcement: configurable per customer, default 90 days for experiment data
- No EU personal data processed by US-region services

### APAC — ap-northeast-2 / ap-southeast-1

Dual-AZ-region setup for APAC coverage. Korean public sector customers use ap-northeast-2 (Seoul) with CSAP controls. Southeast Asian customers use ap-southeast-1 (Singapore).

| Property | Value |
|---|---|
| AWS Regions | ap-northeast-2 (Seoul), ap-southeast-1 (Singapore) |
| Data residency | Korea (Seoul) or Southeast Asia (Singapore), customer-selected |
| Compliance | CSAP 표준등급 (Seoul), ISMS-P (Seoul), MAS TRM (Singapore) |
| Database | PostgreSQL 15, RDS Multi-AZ per sub-region |
| KMS | AWS KMS per sub-region (separate CMKs) |
| Korean public sector | NHN Cloud / NCP option for CSAP 상등급 |

APAC routing rules:
- Korean public institution customers: ap-northeast-2 mandatory
- Singapore / ASEAN customers: ap-southeast-1 default
- Cross-sub-region data movement: prohibited without customer opt-in

### GovCloud — us-gov-west-1

US federal government region. Completely air-gapped from commercial AWS. Supports FedRAMP Moderate ATO and DoD IL4/IL5.

| Property | Value |
|---|---|
| AWS Region | us-gov-west-1 |
| Data residency | United States (US persons only) |
| Compliance | FedRAMP Moderate ATO, DoD IL4, DoD IL5 (dedicated) |
| Availability zones | us-gov-west-1a, us-gov-west-1b |
| Database | PostgreSQL 15, RDS Multi-AZ (GovCloud) |
| KMS | AWS KMS GovCloud (FIPS 140-2 validated) |
| Crypto | BoringCrypto (FIPS 140-2, GOEXPERIMENT=boringcrypto) |
| CI/CD | Dedicated GovCloud pipeline, no shared infrastructure with commercial |
| Access | US persons only, background check required for ChaosPlane staff |

---

## Data Isolation Model

Customer data never crosses region boundaries except for the following narrow cases, all of which require explicit customer action:

1. Billing metadata (org ID, plan tier, seat count) — synced to global billing service, no experiment data included
2. Audit log exports — customer-initiated, exported to customer-owned storage in their chosen region
3. Cross-region GameDay (v2.0.0 feature) — opt-in, orchestration metadata only, fault injection stays local

### Tenant Isolation Within a Region

Each tenant's data is isolated at the database layer via row-level security (RLS) on all PostgreSQL tables. The application layer enforces tenant context on every request. Cross-tenant queries are architecturally prevented: the API server injects tenant ID from the authenticated session, not from request parameters.

```
Request → Auth middleware (extracts tenant_id from JWT)
        → RLS context set: SET LOCAL app.tenant_id = '<tenant_id>'
        → All queries automatically scoped to tenant
        → Cross-tenant access: impossible without compromising the auth layer
```

---

## Cross-Region Management Console

The management console gives platform operators and enterprise customers a unified view across all regions they have access to. It operates on metadata only.

### What the console shows

- Experiment counts and status per region (aggregated, no raw data)
- Resilience score trends per region
- Active incidents and alerts per region
- User and team membership (synced via global identity service)
- Billing and usage across regions

### What the console does NOT do

- Display experiment payloads, target selectors, or result data from other regions
- Allow cross-region experiment execution from a single API call (each region's API is called independently)
- Store any experiment data centrally

### Console Architecture

```
[Browser]
    |
    | HTTPS
    v
[Global Console API] — reads from:
    ├── US Region: /internal/metrics/summary  (aggregated only)
    ├── EU Region: /internal/metrics/summary  (aggregated only)
    ├── APAC Region: /internal/metrics/summary (aggregated only)
    └── GovCloud: NOT accessible from commercial console
                  (separate GovCloud console instance)
```

GovCloud has its own isolated management console deployed within the GovCloud boundary. It is not accessible from the commercial console.

---

## Region Selection for New Customers

Customers select their home region at organization creation. The selection is permanent and cannot be changed without a data migration process (available on Enterprise plan).

| Customer type | Default region | Override allowed |
|---|---|---|
| US commercial | us-east-1 | Yes, to any commercial region |
| EU / EEA | eu-west-1 | Yes, to other commercial regions |
| Korean public sector | ap-northeast-2 | No (CSAP requirement) |
| US federal agency | us-gov-west-1 | No (FedRAMP requirement) |
| ASEAN | ap-southeast-1 | Yes, to other commercial regions |

---

## Database Strategy

Each region runs an independent PostgreSQL 15 cluster (RDS Multi-AZ). There is no synchronous cross-region replication of customer data.

Global metadata (org routing table, billing records) uses a separate lightweight store with eventual consistency:

- Write: to home region
- Propagation: async replication to global metadata service within 30 seconds
- Read: from global metadata service (may be up to 30 seconds stale for routing lookups)
- Conflict resolution: last-write-wins on non-critical metadata, manual review for billing conflicts

For v2.0.0, CockroachDB is under evaluation as a replacement for the global metadata store to provide stronger consistency guarantees across regions.

---

## Encryption Key Hierarchy

Each region maintains a completely independent encryption key hierarchy. There is no shared key material between regions.

```
AWS KMS (per-region CMK)
    └── Data Encryption Keys (DEKs) — generated per tenant
            └── Tenant data encrypted with tenant DEK
                    └── DEK encrypted with region CMK (envelope encryption)
```

Key rotation schedule:
- CMKs: annual, or immediately on suspected compromise
- DEKs: annual per tenant
- TLS certificates: 90-day rotation via cert-manager

---

## Deployment Configuration

### Region-Specific Helm Values

Each region overrides the base Helm chart with region-specific values:

```yaml
# values-eu-west-1.yaml
global:
  region: eu-west-1
  dataResidency: EU
  gdprMode: true

database:
  endpoint: chaosplane-eu.cluster-xxxx.eu-west-1.rds.amazonaws.com
  kmsKeyId: arn:aws:kms:eu-west-1:xxxx:key/yyyy

compliance:
  gdpr: true
  dataSubjectRightsApi: true
  retentionDays: 90

ingress:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: api.eu.chaosplane.io
```

```yaml
# values-us-gov-west-1.yaml
global:
  region: us-gov-west-1
  govCloud: true
  fipsMode: true

crypto:
  boringCrypto: true
  tlsMinVersion: "1.2"
  cipherSuites:
    - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

database:
  endpoint: chaosplane-gov.cluster-xxxx.us-gov-west-1.rds.amazonaws.com
  kmsKeyId: arn:aws-us-gov:kms:us-gov-west-1:xxxx:key/yyyy

compliance:
  fedramp: true
  dodIL4: true
  fips1402: true
```

---

## Operational Runbooks

### Adding a New Region

1. Create AWS account in target region (or GovCloud account for gov regions)
2. Apply Terraform module `modules/chaosplane-region` with region-specific variables
3. Create region CMK in AWS KMS, document key ARN
4. Deploy ChaosPlane via Helm with region-specific values file
5. Register region in global routing table (org creation will now offer new region)
6. Run smoke tests: `chaosctl region verify --region <region>`
7. Update status page and documentation

### Region Failover

Each region is designed to operate independently. If a region becomes unavailable:

- Customers in that region cannot access the platform until the region recovers
- No automatic failover to another region (data residency requirements prohibit this)
- RTO target: 4 hours. RPO target: 1 hour (from last backup)
- Status page updated within 15 minutes of confirmed outage

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Active |
| Owner | Platform Engineering / CISO |
| Last reviewed | April 2026 |
| Next review | October 2026 |
