# FedRAMP Moderate System Security Plan (SSP) — ChaosPlane

> This document is a draft System Security Plan (SSP) for ChaosPlane's FedRAMP Moderate authorization. It is prepared in accordance with NIST SP 800-18 (Guide for Developing Security Plans for Federal Information Systems) and the FedRAMP SSP template requirements.
>
> ChaosPlane is preparing for FedRAMP Moderate authorization. This document does not represent a completed or approved authorization package. A Third Party Assessment Organization (3PAO) will conduct the formal Security Assessment (SA) prior to submission to the FedRAMP Program Management Office (PMO).

---

## 1. System Identification

| Field | Value |
|---|---|
| System Name | ChaosPlane Government Edition |
| System Abbreviation | CPGE |
| System Owner | ChaosPlane, Inc. |
| Authorizing Official (AO) | To be designated by sponsoring Agency |
| System Type | Major Application — SaaS |
| FedRAMP Authorization Type | Agency Authorization (path to JAB P-ATO in Phase 5) |
| Impact Level | Moderate |
| Deployment Model | Government Community Cloud (AWS GovCloud us-gov-west-1) |
| Service Model | Software as a Service (SaaS) |
| Status | In Development — preparing for 3PAO assessment |

---

## 2. System Description

ChaosPlane Government Edition is a chaos engineering platform that enables federal agency engineering teams to design, execute, and analyze fault injection experiments against cloud-native and hybrid infrastructure. The system enables agencies to proactively test the resilience of their systems before failures occur in production.

### 2.1 System Purpose

ChaosPlane Government Edition provides:

- Experiment design and execution: fault injection against Kubernetes workloads, VMs, bare-metal servers, and cloud services
- Resilience scoring: quantitative measurement of system resilience over time
- Audit logging: tamper-evident, append-only logs of all experiment activity
- AI-assisted analysis: topology analysis, vulnerability detection, experiment recommendations, and result summarization
- Role-based and attribute-based access control (RBAC/ABAC) for multi-team environments
- SAML SSO integration with agency identity providers (including CAC/PIV via agency IdP)
- CUI marking system for controlled unclassified information handling

### 2.2 System Components

| Component | Technology | Deployment |
|---|---|---|
| Web Application | React (TypeScript), served via CloudFront (GovCloud) | AWS GovCloud |
| API Server | Go, deployed on AWS EKS (GovCloud) | AWS GovCloud |
| Chaos Agent | Go binary, deployed in agency Kubernetes clusters or AWS GovCloud accounts | Agency environment or AWS GovCloud |
| Workflow Engine | Go, deployed on AWS EKS (GovCloud) | AWS GovCloud |
| AI Assistant | Go service, calls FIPS-compliant LLM endpoint | AWS GovCloud |
| Identity Service | Go, SAML/SCIM, RBAC/ABAC engine | AWS GovCloud |
| Audit Log Service | Go, append-only, tamper-evident | AWS GovCloud |
| Primary Database | PostgreSQL 15, AWS RDS (GovCloud), Multi-AZ | AWS GovCloud |
| Session Store | Redis 7, AWS ElastiCache (GovCloud) | AWS GovCloud |
| Object Storage | AWS S3 (GovCloud), server-side encryption | AWS GovCloud |
| Key Management | AWS KMS (GovCloud), FIPS 140-2 validated | AWS GovCloud |
| Container Registry | AWS ECR (GovCloud) | AWS GovCloud |
| CI/CD | Dedicated GovCloud CI/CD pipeline, air-gapped from commercial | AWS GovCloud |

---

## 3. System Boundary

### 3.1 Authorization Boundary Description

The ChaosPlane Government Edition authorization boundary encompasses all components deployed within AWS GovCloud (us-gov-west-1) that are operated and maintained by ChaosPlane, Inc. The boundary includes:

**In Boundary:**
- All ChaosPlane application components listed in Section 2.2
- AWS GovCloud infrastructure managed by ChaosPlane (EKS clusters, RDS instances, ElastiCache, S3 buckets, KMS keys, VPC, subnets, security groups, IAM roles)
- ChaosPlane's GovCloud CI/CD pipeline and build infrastructure
- ChaosPlane's GovCloud monitoring and logging stack
- ChaosPlane staff access to GovCloud environment (via dedicated GovCloud IAM accounts, MFA required)

**Out of Boundary (Leveraged Systems):**
- AWS GovCloud infrastructure services (compute, networking, storage) — covered by AWS GovCloud FedRAMP High P-ATO
- Agency identity providers (PIV/CAC authentication infrastructure) — agency responsibility
- Agency Kubernetes clusters where the chaos agent is deployed — agency responsibility
- ChaosPlane's commercial AWS environment (separate, no data flow to/from GovCloud)

### 3.2 System Boundary Diagram (Text Representation)

```
┌─────────────────────────────────────────────────────────────────────┐
│  ChaosPlane Government Edition — Authorization Boundary             │
│  AWS GovCloud (us-gov-west-1)                                       │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Public-Facing Tier                                         │   │
│  │  - CloudFront (GovCloud CDN) — Web Application              │   │
│  │  - Application Load Balancer                                │   │
│  │  - WAF (AWS WAF, GovCloud)                                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          |                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Application Tier (Private Subnet)                          │   │
│  │  - API Server (EKS)                                         │   │
│  │  - Workflow Engine (EKS)                                    │   │
│  │  - AI Assistant Service (EKS)                               │   │
│  │  - Identity Service (EKS)                                   │   │
│  │  - Audit Log Service (EKS)                                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          |                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Data Tier (Private Subnet)                                 │   │
│  │  - PostgreSQL (RDS Multi-AZ)                                │   │
│  │  - Redis (ElastiCache)                                      │   │
│  │  - S3 (audit exports, backups)                              │   │
│  │  - KMS (FIPS 140-2 validated)                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Management / Operations (Restricted Subnet)                │   │
│  │  - Bastion Host (MFA required, session recorded)            │   │
│  │  - CI/CD Pipeline (GovCloud-dedicated)                      │   │
│  │  - Monitoring Stack (CloudWatch, GovCloud)                  │   │
│  │  - Log Aggregation (CloudWatch Logs, GovCloud)              │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
         |                                    |
[Agency Users]                    [Chaos Agent in Agency Env]
(PIV/CAC via Agency IdP)          (deployed by agency, calls API
                                   over TLS 1.2+, mTLS)
```

---

## 4. Data Flow

### 4.1 Data Flow Description

**User Authentication Flow:**
1. Agency user navigates to ChaosPlane Government Edition web application (CloudFront GovCloud)
2. User initiates SAML SSO — redirected to agency IdP (which handles PIV/CAC authentication)
3. Agency IdP returns SAML assertion to ChaosPlane Identity Service
4. Identity Service validates assertion, creates session token (256-bit random, stored in Redis)
5. Session token returned to browser via HTTPS (TLS 1.2+, FIPS-approved cipher suites)

**Experiment Execution Flow:**
1. Authenticated user creates experiment definition via API (TLS 1.2+)
2. API Server validates request, checks RBAC/ABAC permissions, writes to PostgreSQL
3. Workflow Engine picks up experiment, dispatches to Chaos Agent
4. Chaos Agent (in agency environment) executes fault injection, reports results via mTLS
5. Results written to PostgreSQL, audit log written to Audit Log Service (append-only)
6. Real-time status updates delivered to user via WebSocket (TLS 1.2+, re-authenticated every 5 minutes)

**Audit Log Export Flow:**
1. Authorized user requests audit log export via API
2. Audit Log Service generates export file (JSON or CSV)
3. Export uploaded to agency-designated S3 bucket (GovCloud or agency-owned) via TLS
4. User notified via email, time-limited download link provided (24-hour expiry)
5. Export file deleted from ChaosPlane storage after download or expiry

### 4.2 Data Flow Diagram (Text Representation)

```
[Agency User / PIV-CAC]
        |
        | HTTPS TLS 1.2+ (FIPS cipher suites)
        v
[CloudFront GovCloud] → [ALB] → [API Server / Identity Service]
                                        |
                          ┌─────────────┼─────────────┐
                          v             v             v
                    [PostgreSQL]    [Redis]      [Audit Log]
                    (RDS Multi-AZ)  (ElastiCache) (append-only)
                          |
                    [Workflow Engine]
                          |
                          | mTLS
                          v
                    [Chaos Agent]
                    (Agency Env)
                          |
                    [Agency Infrastructure]
                    (Kubernetes / VM / Bare Metal)
```

---

## 5. FIPS 140-2 Cryptography

All cryptographic operations within the ChaosPlane Government Edition authorization boundary use FIPS 140-2 validated modules.

### 5.1 Cryptographic Implementations

| Component | Algorithm | FIPS Module | Use |
|---|---|---|---|
| Go API Server | AES-256-GCM, SHA-256, RSA-2048 | BoringCrypto (Google, FIPS 140-2 cert #3678) | TLS, token signing, data encryption |
| Go Chaos Agent | AES-256-GCM, SHA-256 | BoringCrypto | mTLS, payload encryption |
| PostgreSQL | AES-256 | AWS RDS FIPS endpoints | Data at rest |
| Redis | TLS 1.2+ | AWS ElastiCache FIPS endpoints | Data in transit |
| S3 | AES-256 (SSE-KMS) | AWS KMS (FIPS 140-2 validated) | Data at rest |
| KMS | AES-256, RSA-2048 | AWS KMS GovCloud (FIPS 140-2 cert #3139) | Key management |
| TLS (all services) | TLS 1.2+, FIPS cipher suites only | BoringCrypto | Data in transit |

### 5.2 FIPS Mode Configuration

ChaosPlane's Go services are compiled with BoringCrypto enabled (`GOEXPERIMENT=boringcrypto`). This replaces the standard Go crypto library with FIPS 140-2 validated BoringSSL bindings. Non-FIPS algorithms are disabled at compile time.

TLS configuration enforces FIPS-approved cipher suites only:
- TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
- TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
- TLS_RSA_WITH_AES_128_GCM_SHA256
- TLS_RSA_WITH_AES_256_GCM_SHA384

TLS 1.0 and 1.1 are disabled. TLS 1.3 is supported where FIPS-approved cipher suites are available.

### 5.3 Key Management

- All encryption keys managed via AWS KMS GovCloud (FIPS 140-2 validated)
- Key rotation: annual for KMS CMKs, immediate on suspected compromise
- Key access restricted to authorized IAM roles via least-privilege policies
- Key usage logged via AWS CloudTrail (GovCloud)
- No plaintext key material ever leaves KMS

---

## 6. AWS GovCloud Architecture

### 6.1 GovCloud Account Structure

ChaosPlane maintains a dedicated AWS GovCloud account structure, completely separate from the commercial AWS environment:

| Account | Purpose |
|---|---|
| govcloud-prod | Production ChaosPlane Government Edition |
| govcloud-staging | Staging / pre-production testing |
| govcloud-cicd | CI/CD pipeline and build infrastructure |
| govcloud-security | Security tooling, log aggregation, SIEM |
| govcloud-management | AWS Organizations management account |

No data flows between GovCloud accounts and commercial AWS accounts. The CI/CD pipeline in govcloud-cicd builds and deploys only to govcloud-prod and govcloud-staging.

### 6.2 Network Architecture

- VPC with three subnet tiers: public (ALB, WAF), private-app (EKS workloads), private-data (RDS, ElastiCache)
- No direct internet access from private subnets. Outbound traffic via NAT Gateway (GovCloud)
- VPC Flow Logs enabled, retained in CloudWatch Logs (GovCloud) for 1 year
- AWS PrivateLink used for all AWS service endpoints (S3, KMS, ECR, etc.) — no traffic traverses public internet
- Security groups follow least-privilege: each service allows only necessary inbound ports from specific source security groups

### 6.3 Identity and Access Management

- Dedicated GovCloud IAM accounts for all ChaosPlane staff with GovCloud access
- MFA required for all GovCloud IAM users (hardware token or virtual MFA)
- IAM roles follow least privilege. No wildcard permissions in production
- AWS root account credentials stored in hardware security module, not used for day-to-day operations
- CloudTrail enabled in all GovCloud accounts, logs retained for 7 years
- AWS Config enabled for continuous compliance monitoring

### 6.4 Separation from Commercial Environment

- GovCloud accounts have no trust relationships with commercial AWS accounts
- No shared IAM users, roles, or credentials between GovCloud and commercial
- Separate CI/CD pipeline for GovCloud — no shared build infrastructure
- GovCloud container images built from scratch in govcloud-cicd, not pulled from commercial ECR
- Separate monitoring and alerting stack in GovCloud

---

## 7. FedRAMP Moderate Control Implementation Summary

ChaosPlane Government Edition implements controls from NIST SP 800-53 Rev 5 at the Moderate baseline. The following table summarizes key control families:

| Control Family | Key Controls | ChaosPlane Implementation |
|---|---|---|
| AC — Access Control | AC-2, AC-3, AC-6, AC-17 | RBAC/ABAC, least privilege, MFA, VPN for remote access |
| AU — Audit and Accountability | AU-2, AU-3, AU-9, AU-11 | Append-only audit logs, tamper-evident, 3-year retention, CloudTrail |
| CA — Assessment, Authorization | CA-2, CA-7, CA-8 | 3PAO assessment, continuous monitoring, penetration testing |
| CM — Configuration Management | CM-2, CM-6, CM-7, CM-8 | IaC (Terraform), drift detection, asset inventory, minimal functionality |
| CP — Contingency Planning | CP-2, CP-9, CP-10 | BCP documented, automated backups, DR tested annually |
| IA — Identification and Authentication | IA-2, IA-5, IA-8 | MFA, FIPS-compliant password hashing, PIV/CAC via agency IdP |
| IR — Incident Response | IR-2, IR-4, IR-5, IR-6 | IR plan, detection, handling, reporting to US-CERT |
| MA — Maintenance | MA-2, MA-4 | Controlled maintenance, remote maintenance via MFA VPN |
| MP — Media Protection | MP-2, MP-6, MP-7 | No removable media, cloud storage only, AWS handles media disposal |
| PE — Physical and Environmental | PE-1 through PE-20 | Inherited from AWS GovCloud FedRAMP High P-ATO |
| PL — Planning | PL-2, PL-8 | This SSP, security architecture |
| PS — Personnel Security | PS-3, PS-4, PS-5 | Background checks, offboarding within 24 hours, role changes |
| RA — Risk Assessment | RA-3, RA-5 | Annual risk assessment, vulnerability scanning |
| SA — System and Services Acquisition | SA-3, SA-8, SA-11 | SDLC security, secure design, developer testing |
| SC — System and Communications Protection | SC-8, SC-12, SC-13, SC-28 | TLS 1.2+, FIPS 140-2 crypto, encryption at rest |
| SI — System and Information Integrity | SI-2, SI-3, SI-7, SI-10 | Patch management, malware protection, integrity verification, input validation |

---

## 8. Roles and Responsibilities

| Role | Responsibility |
|---|---|
| System Owner (ChaosPlane CISO) | Overall accountability for system security. Signs SSP. |
| Authorizing Official (Agency AO) | Grants ATO based on 3PAO assessment and risk acceptance |
| 3PAO | Independent security assessment, SAR production |
| FedRAMP PMO | Reviews package, maintains FedRAMP Marketplace listing |
| ChaosPlane Compliance Manager | Day-to-day compliance operations, evidence collection, POA&M management |
| ChaosPlane Security Engineers | Control implementation, vulnerability management, incident response |
| Agency ISSO | Agency-side security oversight, CUECs |

---

## 9. Continuous Monitoring Plan

Post-authorization, ChaosPlane Government Edition will maintain a continuous monitoring program per FedRAMP requirements:

| Activity | Frequency | Responsible |
|---|---|---|
| Vulnerability scanning (OS/container) | Monthly | Security Engineer |
| Vulnerability scanning (web application) | Monthly | Security Engineer |
| Penetration testing | Annual | 3PAO |
| Security control assessment | Annual (subset) | 3PAO |
| POA&M review and update | Monthly | Compliance Manager |
| Incident reporting to US-CERT | Within 1 hour of P0/P1 | Security Engineer |
| Significant change notification to AO | Per change | CISO |
| Annual assessment | Annual | 3PAO |
| ConMon report to AO | Monthly | Compliance Manager |

---

## 10. Plan of Action and Milestones (POA&M) — Open Items

| Item | Control | Finding | Target Date | Status |
|---|---|---|---|---|
| 3PAO selection | CA-2 | 3PAO not yet engaged | Month 15 | In progress |
| Agency sponsor identification | CA-1 | No sponsoring agency yet | Month 16 | In progress |
| GovCloud infrastructure build-out | SC-7 | GovCloud accounts created, workloads not yet migrated | Month 16 | In progress |
| FIPS mode validation | SC-13 | BoringCrypto build pipeline in progress | Month 15 | In progress |
| PIV/CAC integration testing | IA-8 | Requires agency IdP for testing | Month 17 | Planned |
| Air-gap deployment testing | CP-2 | Air-gap bundle script complete, full test pending | Month 17 | Planned |
| ConMon tooling setup | CA-7 | Tooling selection in progress | Month 16 | In progress |

---

## 11. Leveraged FedRAMP Authorizations

| System | Provider | Authorization | Impact Level |
|---|---|---|---|
| AWS GovCloud (us-gov-west-1) | Amazon Web Services | JAB P-ATO | High |

ChaosPlane Government Edition inherits physical and environmental controls (PE family) and a subset of infrastructure controls from AWS GovCloud's FedRAMP High P-ATO. The Customer Responsibility Matrix (CRM) documents which controls are inherited vs. ChaosPlane-implemented.

---

## Document Control

| Field | Value |
|---|---|
| Version | 0.1.0 (Draft) |
| Status | Draft — preparing for 3PAO engagement |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | Upon 3PAO engagement |
| Standard reference | NIST SP 800-18, NIST SP 800-53 Rev 5, FedRAMP Moderate Baseline |
| Template reference | FedRAMP SSP Template v3.3 |
