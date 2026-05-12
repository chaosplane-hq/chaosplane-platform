# DoD IL4 / IL5 Certification Preparation — ChaosPlane

> ChaosPlane Government Edition is pursuing DoD Impact Level 4 (IL4) and Impact Level 5 (IL5) authorization under the DoD Cloud Computing Security Requirements Guide (CC SRG). IL4 covers Controlled Unclassified Information (CUI). IL5 covers National Security Systems (NSS) and mission-critical DoD workloads. This document describes the technical controls, infrastructure requirements, and DISA evaluation process for both levels.
>
> This document does not represent a completed authorization. IL4 authorization is targeted for Month 24-30. IL5 is targeted for Month 30-36.

---

## Impact Level Overview

| Level | Data Types | Infrastructure | ChaosPlane Target |
|---|---|---|---|
| IL2 | Publicly releasable DoD information | Commercial cloud | Not targeted |
| IL4 | CUI, including Privacy Act data, FOUO, Law Enforcement Sensitive | DoD-approved commercial cloud (AWS GovCloud) | Month 24-30 |
| IL5 | CUI requiring higher protection, NSS, mission-critical systems | DoD-dedicated infrastructure or IL5-approved cloud | Month 30-36 |
| IL6 | Classified (SECRET) | Classified cloud | Not targeted |

---

## IL4 Authorization

### IL4 Scope

ChaosPlane Government Edition IL4 authorization covers:

- ChaosPlane SaaS platform deployed in AWS GovCloud (us-gov-west-1)
- Chaos agent deployed in DoD agency Kubernetes clusters or AWS GovCloud accounts
- All CUI processed, stored, or transmitted by the platform
- ChaosPlane staff access to GovCloud production environment

CUI categories handled by ChaosPlane in IL4 context:
- Infrastructure topology data (network diagrams, system configurations) — may contain CUI
- Experiment results and logs — may reference CUI-bearing system names or configurations
- User identity data — DoD personnel information (name, email, agency affiliation)
- Audit logs — records of CUI system access patterns

### IL4 Technical Requirements

#### CUI Handling

ChaosPlane implements CUI marking and handling per NIST SP 800-171 and the CUI Registry:

- CUI marking: experiment definitions and results that reference CUI systems are tagged with CUI category codes at creation
- CUI dissemination: CUI data is accessible only to users with appropriate clearance and need-to-know, enforced via ABAC policies
- CUI storage: all CUI stored encrypted at rest (AES-256, AWS KMS GovCloud FIPS 140-2 validated)
- CUI transmission: TLS 1.2+ with FIPS-approved cipher suites on all connections
- CUI destruction: data deletion pipeline removes CUI within 30 days of account termination, with certificate of destruction available

```
CUI Marking Flow:
[Experiment created] → [User marks CUI category: e.g., CUI//SP-CTI]
                     → [ABAC policy applied: only cleared users can view]
                     → [Audit log entry: CUI access recorded]
                     → [Export restricted: CUI exports require supervisor approval]
```

#### FIPS 140-2 Cryptography

All cryptographic operations use FIPS 140-2 validated modules (same as FedRAMP Moderate, extended to IL4 requirements):

| Component | FIPS Module | Certificate |
|---|---|---|
| Go services | BoringCrypto | FIPS 140-2 cert #3678 |
| PostgreSQL (RDS) | AWS RDS FIPS endpoints | Inherited from AWS GovCloud |
| Redis (ElastiCache) | AWS ElastiCache FIPS endpoints | Inherited from AWS GovCloud |
| KMS | AWS KMS GovCloud | FIPS 140-2 cert #3139 |
| TLS (all services) | BoringCrypto | FIPS 140-2 cert #3678 |

#### Multi-Factor Authentication

IL4 requires PIV/CAC authentication for all privileged access. ChaosPlane implements:

- End-user authentication: SAML SSO delegated to agency IdP, which enforces PIV/CAC
- ChaosPlane staff access to GovCloud: hardware MFA token (FIDO2/WebAuthn) required
- Emergency access: break-glass procedure with dual-person authorization, full session recording
- Service accounts: no interactive login, certificate-based authentication only

#### Audit Logging

IL4 requires comprehensive audit logging per NIST SP 800-53 AU controls:

- All user actions logged: login, logout, experiment create/modify/delete/execute, CUI access, admin actions
- Log format: structured JSON with timestamp (UTC), actor (user ID + agency), action, resource, outcome, source IP
- Log integrity: append-only storage, SHA-256 hash chain, tamper detection alerts
- Log retention: 3 years minimum (IL4 requirement), 7 years for privileged access logs
- Log export: available to agency ISSO on request, delivered to agency-owned S3 (GovCloud)

#### Network Segmentation

```
┌─────────────────────────────────────────────────────────────────────┐
│  ChaosPlane IL4 Boundary — AWS GovCloud (us-gov-west-1)             │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  DMZ (Public Subnet)                                        │   │
│  │  - CloudFront GovCloud (WAF, TLS termination)               │   │
│  │  - Application Load Balancer                                │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          |                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Application Tier (Private Subnet — CUI processing)         │   │
│  │  - API Server (EKS, BoringCrypto)                           │   │
│  │  - Workflow Engine                                          │   │
│  │  - CUI Marking Service                                      │   │
│  │  - Identity / ABAC Engine                                   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          |                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Data Tier (Private Subnet — CUI at rest)                   │   │
│  │  - PostgreSQL RDS (FIPS endpoints, AES-256)                 │   │
│  │  - Redis ElastiCache (FIPS endpoints)                       │   │
│  │  - S3 (SSE-KMS, GovCloud CMK)                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Management (Restricted Subnet)                             │   │
│  │  - Bastion (PIV/CAC + hardware MFA, session recorded)       │   │
│  │  - CI/CD (GovCloud-dedicated, air-gapped from commercial)   │   │
│  │  - SIEM / Log Aggregation                                   │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
         |
[DoD Agency Users — PIV/CAC via Agency IdP]
[Chaos Agent — mTLS, deployed in agency environment]
```

### DISA Evaluation Process for IL4

DISA (Defense Information Systems Agency) evaluates cloud service providers for IL4 authorization via the DoD CC SRG process.

| Phase | Activity | Owner | Target |
|---|---|---|---|
| Pre-assessment | Gap analysis against CC SRG IL4 controls | ChaosPlane Compliance | Month 22 |
| Documentation | System Security Plan (SSP) for IL4 | ChaosPlane Compliance | Month 23 |
| 3PAO engagement | Select DISA-approved 3PAO for IL4 assessment | ChaosPlane | Month 23 |
| Assessment | 3PAO security assessment against IL4 controls | 3PAO | Month 24-25 |
| SAR review | Security Assessment Report review and POA&M | ChaosPlane + 3PAO | Month 25 |
| DISA review | DISA reviews package, issues Provisional Authorization | DISA | Month 26-27 |
| Agency ATO | Sponsoring DoD agency issues ATO based on PA | Agency AO | Month 27-28 |

Key DISA contacts:
- DISA Cloud Portfolio: cloud.portfolio@disa.mil
- DoD CC SRG: https://public.cyber.mil/dccs/

---

## IL5 Authorization

### IL5 Scope

IL5 extends IL4 to cover National Security Systems and mission-critical DoD workloads. ChaosPlane IL5 is targeted for DoD agencies running chaos engineering against mission-critical infrastructure (weapons systems support, C2 systems, critical infrastructure protection).

IL5 requires dedicated infrastructure — shared multi-tenant cloud is not permitted for IL5 data. ChaosPlane IL5 is a single-tenant dedicated deployment per agency.

### IL5 Infrastructure Requirements

#### Dedicated Infrastructure

Each IL5 customer receives a dedicated ChaosPlane deployment:

- Dedicated AWS GovCloud account (not shared with other customers)
- Dedicated EKS cluster, RDS instance, ElastiCache cluster
- Dedicated KMS CMK (customer-managed, optionally customer-owned HSM)
- No shared compute, storage, or network resources with other tenants

```yaml
# IL5 deployment — dedicated per agency
apiVersion: chaosplane.dev/v1
kind: TenantDeployment
metadata:
  name: agency-dod-il5
spec:
  tier: il5
  dedicated: true
  govCloudAccount: "123456789012"  # agency-dedicated account
  region: us-gov-west-1
  kmsKeyArn: "arn:aws-us-gov:kms:us-gov-west-1:123456789012:key/agency-cmk"
  networkIsolation: strict
  auditLogDestination: "arn:aws-us-gov:s3:::agency-audit-logs"
```

#### Personnel Security

IL5 requires US citizenship and appropriate security clearance for all ChaosPlane staff with access to IL5 environments:

- All staff with IL5 access: US citizens, minimum Secret clearance
- Background investigation: NACLC (National Agency Check with Law and Credit) minimum
- Clearance verification: tracked in ChaosPlane's personnel security system
- Foreign nationals: no access to IL5 environments under any circumstances
- Insider threat program: behavioral monitoring, periodic reinvestigation

#### Supply Chain Risk Management (SCRM)

IL5 requires supply chain risk management per NIST SP 800-161:

- Software Bill of Materials (SBOM): generated for every release, signed with cosign
- Third-party component vetting: all dependencies reviewed for country-of-origin risk
- Open source components: reviewed against CISA Known Exploited Vulnerabilities catalog
- Hardware: AWS GovCloud hardware sourced from US-approved supply chain (inherited from AWS)
- CI/CD pipeline: air-gapped from commercial internet, builds from vetted source mirrors only

#### Mission-Critical Requirements

IL5 deployments supporting mission-critical systems have additional availability requirements:

| Requirement | IL4 | IL5 |
|---|---|---|
| Availability SLA | 99.9% | 99.99% |
| RTO | 4 hours | 1 hour |
| RPO | 1 hour | 15 minutes |
| Backup frequency | Daily | Every 15 minutes (continuous WAL archiving) |
| DR test frequency | Annual | Semi-annual |
| Change window | Maintenance window | Zero-downtime deployments only |

Continuous WAL archiving configuration:
```yaml
# PostgreSQL continuous archiving for IL5
postgresql:
  walArchiving:
    enabled: true
    destination: s3://agency-wal-archive/
    archiveTimeout: 60  # seconds — max data loss
  replication:
    standby: true
    synchronousCommit: "on"  # no async commits for IL5
```

### DISA Evaluation Process for IL5

IL5 evaluation follows the same CC SRG process as IL4 but with additional scrutiny on dedicated infrastructure, personnel security, and supply chain.

| Phase | Activity | Target |
|---|---|---|
| IL4 authorization complete | Prerequisite for IL5 | Month 28 |
| IL5 gap analysis | Delta from IL4 controls | Month 28 |
| Dedicated infrastructure build | Per-agency IL5 deployment | Month 29 |
| Personnel clearance verification | All IL5-access staff cleared | Month 29 |
| 3PAO IL5 assessment | Assessment against IL5 controls | Month 30-31 |
| DISA IL5 Provisional Authorization | DISA review and PA issuance | Month 32-33 |
| First IL5 agency ATO | Sponsoring agency ATO | Month 33-34 |

---

## NIST SP 800-171 CUI Control Mapping

ChaosPlane maps its controls to NIST SP 800-171 Rev 2 for CUI protection (required for IL4):

| Family | Controls | ChaosPlane Implementation |
|---|---|---|
| 3.1 Access Control | 3.1.1 – 3.1.22 | RBAC/ABAC, least privilege, session management, remote access via VPN+MFA |
| 3.2 Awareness and Training | 3.2.1 – 3.2.3 | Annual security training, CUI handling training for all IL4-access staff |
| 3.3 Audit and Accountability | 3.3.1 – 3.3.9 | Append-only audit logs, 3-year retention, tamper detection, log review |
| 3.4 Configuration Management | 3.4.1 – 3.4.9 | IaC (Terraform), drift detection, baseline configurations, change control |
| 3.5 Identification and Authentication | 3.5.1 – 3.5.11 | PIV/CAC via agency IdP, FIPS-compliant password hashing, MFA |
| 3.6 Incident Response | 3.6.1 – 3.6.3 | IR plan, US-CERT reporting within 1 hour, post-incident review |
| 3.7 Maintenance | 3.7.1 – 3.7.6 | Controlled maintenance, remote maintenance via MFA VPN, sanitization |
| 3.8 Media Protection | 3.8.1 – 3.8.9 | No removable media, cloud storage only, CUI marking on exports |
| 3.9 Personnel Security | 3.9.1 – 3.9.2 | Background checks, offboarding within 24 hours |
| 3.10 Physical Protection | 3.10.1 – 3.10.6 | Inherited from AWS GovCloud |
| 3.11 Risk Assessment | 3.11.1 – 3.11.3 | Annual risk assessment, vulnerability scanning, risk treatment |
| 3.12 Security Assessment | 3.12.1 – 3.12.4 | 3PAO assessment, continuous monitoring, POA&M management |
| 3.13 System and Communications | 3.13.1 – 3.13.16 | TLS 1.2+ FIPS, network segmentation, mTLS for agent communication |
| 3.14 System and Information Integrity | 3.14.1 – 3.14.7 | Patch management, malware protection, security alerts, input validation |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | In preparation |
| Owner | CISO / Gov Security Specialist |
| Last reviewed | April 2026 |
| Next review | Upon DISA engagement |
| References | DoD CC SRG v1r4, NIST SP 800-171 Rev 2, NIST SP 800-53 Rev 5 |
