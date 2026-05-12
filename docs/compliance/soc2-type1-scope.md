# SOC 2 Type I Scope and Control Descriptions — ChaosPlane

> ChaosPlane is preparing for a SOC 2 Type I examination. This document defines the scope of that examination and describes the controls in place as of the report date. It is intended for use by the independent service auditor, enterprise customers conducting vendor due diligence, and the internal compliance team.
>
> A SOC 2 Type I report provides an auditor's opinion on whether controls are suitably designed at a point in time. It does not assess operating effectiveness over a period (that is SOC 2 Type II, planned for a subsequent engagement).

---

## Service Organization

**Organization:** ChaosPlane, Inc.
**Service:** ChaosPlane SaaS Platform — chaos engineering platform for cloud-native systems
**Report type:** SOC 2 Type I
**Examination period:** Point-in-time (date to be confirmed with auditor)
**Applicable Trust Service Criteria:** Security (CC), Availability (A), Processing Integrity (PI), Confidentiality (C), Privacy (P)

---

## System Description

ChaosPlane is a SaaS chaos engineering platform that enables engineering teams to design, execute, and analyze fault injection experiments against cloud-native infrastructure. The system consists of:

- Web application (React frontend, served via CDN)
- REST and WebSocket API (Go backend, deployed on AWS EKS)
- Chaos agent (deployed in customer Kubernetes clusters or AWS accounts)
- Experiment scheduler and workflow engine
- AI Assistant (topology analysis, vulnerability detection, experiment suggestion, result summarization)
- Audit log service (append-only, exportable to S3/MinIO and SIEM)
- Identity and access management service (RBAC, ABAC, SSO/SAML, SCIM)
- Data stores: PostgreSQL (primary), Redis (session/pub-sub), S3 (audit exports, backups)

### System Boundaries

In scope:
- ChaosPlane production AWS environment (primary region)
- ChaosPlane web application and API
- ChaosPlane chaos agent (the component deployed in customer environments)
- Supporting services: identity, audit log, workflow engine, AI Assistant
- Internal development and operations processes that affect the security and availability of the system

Out of scope:
- Customer-managed infrastructure where the chaos agent is deployed
- Third-party identity providers (Okta, Azure AD, Google Workspace) — customers are responsible for their IdP security
- AWS infrastructure controls (covered by AWS's own SOC 2 report)

### Principal Service Commitments

ChaosPlane commits to customers via its Terms of Service and Enterprise agreements to:

1. Protect customer data from unauthorized access, disclosure, and loss
2. Maintain platform availability per the SLA defined in the Enterprise plan
3. Process experiments accurately and completely
4. Treat customer data as confidential and not use it for purposes beyond service delivery
5. Handle personal data in accordance with the Privacy Policy and applicable data protection law

---

## Trust Service Criteria Coverage

### CC — Security (Common Criteria)

Security is the foundational criterion and is required for all SOC 2 examinations.

#### CC1 — Control Environment

**CC1.1 — COSO Principle 1: Commitment to integrity and ethical values**
ChaosPlane maintains a Code of Conduct and information security policy, both reviewed annually. All employees acknowledge these policies at onboarding and annually thereafter.

**CC1.2 — COSO Principle 2: Board oversight**
Executive leadership reviews security posture and compliance status quarterly. Material security incidents are escalated to leadership immediately.

**CC1.3 — COSO Principle 3: Organizational structure, reporting lines, and authorities**
Security responsibilities are assigned to a designated Compliance Manager and engineering security lead. Roles and responsibilities are documented.

**CC1.4 — COSO Principle 4: Commitment to competence**
Security awareness training is required at onboarding and annually. Engineering staff receive role-specific secure development training.

**CC1.5 — COSO Principle 5: Accountability**
Performance expectations include security responsibilities. Violations of security policy are subject to the documented disciplinary process.

#### CC2 — Communication and Information

**CC2.1 — Information to support internal control**
Security policies, runbooks, and procedures are maintained in the internal wiki with version history. Changes are communicated to affected staff.

**CC2.2 — Internal communication**
Security incidents, policy changes, and compliance findings are communicated via internal channels. A dedicated security reporting channel is available to all staff.

**CC2.3 — External communication**
Security commitments are published in the Terms of Service and Privacy Policy. Enterprise customers receive security documentation on request. The status page (status.chaosplane.dev) communicates incidents and maintenance in real time.

#### CC3 — Risk Assessment

**CC3.1 — Specify suitable objectives**
Security objectives are defined in the information security policy and aligned with business objectives and customer commitments.

**CC3.2 — Identify and analyze risk**
Formal risk assessment conducted annually and after significant changes. Risks documented with likelihood, impact, and owner.

**CC3.3 — Assess fraud risk**
Fraud risk considered in risk assessment, including insider threat, account takeover, and abuse of the chaos agent.

**CC3.4 — Identify and analyze significant change**
Change management process requires security review for significant changes to the platform architecture, data flows, or access controls.

#### CC4 — Monitoring Activities

**CC4.1 — Conduct ongoing and/or separate evaluations**
Continuous monitoring via infrastructure metrics, application logs, and security alerts. Annual third-party penetration test. Quarterly internal compliance reviews.

**CC4.2 — Evaluate and communicate deficiencies**
Security findings tracked to remediation. Material deficiencies escalated to leadership. Post-incident reviews produce action items with owners and due dates.

#### CC5 — Control Activities

**CC5.1 — Select and develop control activities**
Controls selected based on risk assessment. Defense-in-depth approach: WAF, network segmentation, RBAC/ABAC, encryption, audit logging.

**CC5.2 — Select and develop general controls over technology**
Technology controls include: infrastructure as code (Terraform), CI/CD pipeline with security gates, dependency scanning, SAST, and branch protection.

**CC5.3 — Deploy through policies and procedures**
Controls deployed via documented policies and automated enforcement where possible (e.g., IAM policies, network ACLs, CI checks).

#### CC6 — Logical and Physical Access Controls

**CC6.1 — Logical access security software, infrastructure, and architectures**
Access to the ChaosPlane platform is controlled by the identity and access management service. RBAC enforced on all resources. Enterprise tier adds ABAC. Principle of least privilege applied.

**CC6.2 — Prior to issuing system credentials and granting system access**
User accounts created via self-service registration (email verification required) or SCIM provisioning from customer IdP. Employee access provisioned via access request process with manager approval.

**CC6.3 — Role-based access and least privilege**
Roles defined with minimum necessary permissions. Privileged access (production infrastructure, database) restricted to named individuals with MFA required.

**CC6.4 — Access credentials and authentication**
Passwords hashed with bcrypt (cost factor 12+). MFA available on all tiers, enforceable org-wide on Enterprise. SSO/SAML delegates authentication to customer IdPs. Session tokens rotated on privilege escalation. WebSocket sessions re-authenticated every 5 minutes via server-initiated token revalidation. Token revocation propagated in real time via Redis pub/sub.

**CC6.5 — Discontinue access**
Employee offboarding: all access revoked within 24 hours. Customer user deprovisioning: immediate via SCIM or manual removal. Account deletion: 30-day grace period, then anonymization and cascade soft delete.

**CC6.6 — Logical access security measures against threats from outside system boundaries**
WAF in front of all public endpoints. Rate limiting on authentication endpoints. Bot detection on login and registration. Suspicious login alerts (new device, new location).

**CC6.7 — Restrict transmission, movement, and removal of information**
Customer data not transmitted outside the production environment except via authorized export (audit log export to customer-owned S3, SIEM integration with customer-provided credentials). TLS 1.2+ on all external connections.

**CC6.8 — Prevent or detect unauthorized or malicious software**
Container images scanned for vulnerabilities before deployment (Trivy). Dependency scanning on every build (Dependabot). No user-uploaded executables accepted. Production deployments via CI/CD pipeline only.

#### CC7 — System Operations

**CC7.1 — Detect and monitor for new vulnerabilities**
CVE monitoring via Dependabot and Trivy. Critical vulnerabilities patched within 7 days. High within 30 days. Vulnerability status reviewed in weekly engineering sync.

**CC7.2 — Monitor system components for anomalous behavior**
Infrastructure and application metrics monitored 24/7. Anomaly detection alerts on-call engineer. SIEM integration available for Enterprise customers.

**CC7.3 — Evaluate security events to determine whether they are security incidents**
Security events triaged by on-call engineer using documented severity classification (P0–P3). P0/P1 events escalate to security lead within 1 hour.

**CC7.4 — Respond to identified security incidents**
Incident response runbook covers containment, eradication, recovery, and customer notification. Post-incident review required for P0/P1 events.

**CC7.5 — Identify, develop, and implement activities to recover from identified security incidents**
Recovery procedures documented per incident type. RTO and RPO targets defined. Business continuity plan tested annually.

#### CC8 — Change Management

**CC8.1 — Authorize, design, develop, configure, document, test, approve, and implement changes**
All changes via pull request with peer review. Security review required for significant changes. Production deployments require approval. Automated tests (unit, integration, security) must pass before merge.

#### CC9 — Risk Mitigation

**CC9.1 — Identify, select, and develop risk mitigation activities**
Risk mitigation activities selected based on annual risk assessment. Controls documented and assigned owners.

**CC9.2 — Assess and manage risks from vendors and business partners**
Third-party vendors assessed for security posture before onboarding. DPAs in place with all data processors. Critical vendors reviewed annually.

---

### A — Availability

**A1.1 — Current processing capacity and usage**
Infrastructure auto-scales via AWS EKS. Capacity metrics monitored with alerts for sustained high utilization. Capacity planning reviewed quarterly.

**A1.2 — Environmental, regulatory, and technological changes**
Change management process includes assessment of impact on availability. Infrastructure changes tested in staging before production.

**A1.3 — Recovery from identified availability incidents**
Multi-AZ deployment on AWS. Automatic failover for database and API tiers. Automated daily backups with quarterly restore tests. RTO: 4 hours (Enterprise). RPO: 1 hour.

---

### PI — Processing Integrity

**PI1.1 — Obtain and process inputs completely, accurately, and timely**
Experiment inputs validated at API layer before processing. Invalid inputs rejected with descriptive error responses. Idempotency keys prevent duplicate experiment execution.

**PI1.2 — Processing is complete, accurate, timely, authorized, and error-free**
Experiment execution results recorded with full audit trail. Failed experiments logged with error details. Workflow engine tracks step-level execution state.

**PI1.3 — Outputs are complete, accurate, and timely**
Experiment results and Resilience Score calculations are deterministic and reproducible. Result data retained per plan tier (up to unlimited for Enterprise).

---

### C — Confidentiality

**C1.1 — Identify and maintain confidential information**
Customer data classified as Confidential. Experiment definitions, topology data, audit logs, and user accounts treated as confidential. Data classification policy documented.

**C1.2 — Dispose of confidential information**
Customer data deletion automated via 30-day grace period pipeline. Anonymization applied before cascade soft delete. Hard deletion from backups within 90 days. Audit log exports deleted from ChaosPlane storage after successful transfer to customer destination.

---

### P — Privacy

**P1.1 — Privacy notice**
Privacy Policy published at chaosplane.dev/privacy. Describes what personal data is collected, how it is used, and customer rights.

**P2.1 — Choice and consent**
Users provide consent at registration. Marketing communications require explicit opt-in. Consent records maintained.

**P3.1 — Collection of personal information**
Personal data collected is limited to what is necessary for service delivery: name, email, authentication credentials, usage data, and billing information.

**P4.1 — Use of personal information**
Personal data used only for service delivery, support, and billing. Not sold or shared with third parties for marketing.

**P5.1 — Access**
Users can access their personal data via account settings. Enterprise customers can request a data export.

**P6.1 — Disclosure and notification**
Personal data not disclosed to third parties except as required for service delivery (e.g., payment processor) or by law. Data breach notification per applicable law (72 hours to regulator under GDPR, prompt notification to affected customers).

**P7.1 — Quality of personal information**
Users can update their personal data via account settings. Email changes require verification of the new address and notification to the existing address.

**P8.1 — Monitoring and enforcement**
Privacy compliance reviewed annually. Data subject requests (access, deletion, correction) handled within 30 days. GDPR Right to Erasure automated via the account deletion pipeline.

---

## Complementary User Entity Controls (CUECs)

The following controls are the responsibility of ChaosPlane customers, not ChaosPlane. The effectiveness of ChaosPlane's controls depends on customers implementing these:

1. Customers are responsible for managing their own identity provider (Okta, Azure AD, Google Workspace) security, including MFA enforcement and user lifecycle management.
2. Customers are responsible for controlling who has access to their ChaosPlane organization and assigning appropriate roles.
3. Customers are responsible for securing the environment where the ChaosPlane chaos agent is deployed.
4. Customers are responsible for reviewing and acting on audit log exports and SIEM alerts.
5. Customers are responsible for configuring blast radius policies appropriate for their environment.
6. Customers using audit log export to S3 are responsible for securing their S3 bucket.

---

## Subservice Organizations

ChaosPlane uses the following subservice organizations. Their controls are not covered by this report. Customers should review the subservice organizations' own SOC 2 reports.

| Subservice Organization | Service | SOC 2 Report Available |
|---|---|---|
| Amazon Web Services (AWS) | Cloud infrastructure (compute, storage, networking) | Yes |
| Stripe | Payment processing | Yes |
| OpenAI / Anthropic | LLM API for AI Assistant | Yes (OpenAI); In progress (Anthropic) |
| SendGrid | Transactional email | Yes |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for Type I examination |
| Owner | Compliance Manager |
| Last reviewed | April 2026 |
| Next review | Upon engagement of auditor |
| Standard reference | AICPA Trust Services Criteria (2017, updated 2022) |
