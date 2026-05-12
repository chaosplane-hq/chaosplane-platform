# GDPR Compliance Framework — ChaosPlane

> ChaosPlane is preparing for full GDPR compliance. This document describes the data protection framework, consent management approach, data subject rights automation, and organizational measures in place. It is intended for use by the designated Data Protection Officer (DPO), enterprise customers conducting vendor assessments, and the internal compliance team.
>
> This document does not constitute a legal opinion. ChaosPlane is preparing for compliance; it does not claim that all obligations are fully satisfied at the time of writing.

---

## Scope

GDPR applies to ChaosPlane's processing of personal data belonging to individuals in the European Economic Area (EEA). This includes:

- Personal data of ChaosPlane customers (name, email, billing information, usage data)
- Personal data of end users within customer organizations who access the ChaosPlane platform
- Personal data processed incidentally through experiment topology data (e.g., hostnames, IP addresses that may identify individuals)

ChaosPlane acts as a **data controller** for data collected directly from customers and as a **data processor** for personal data processed on behalf of enterprise customers within their ChaosPlane organization.

---

## Article 13/14 — Transparency and Privacy Notice

ChaosPlane's Privacy Policy (chaosplane.dev/privacy) discloses:

- Categories of personal data collected
- Purposes and legal bases for processing
- Data retention periods
- Third-party recipients and subprocessors
- Data subject rights and how to exercise them
- Contact details for the DPO

The Privacy Policy is reviewed annually and updated whenever processing activities change materially. Version history is maintained.

**Legal bases used:**

| Processing Activity | Legal Basis |
|---|---|
| Account registration and authentication | Contract (Art. 6(1)(b)) |
| Service delivery (experiment execution, audit logging) | Contract (Art. 6(1)(b)) |
| Billing and invoicing | Contract + Legal obligation (Art. 6(1)(b)(c)) |
| Marketing communications | Consent (Art. 6(1)(a)) |
| Security monitoring and fraud prevention | Legitimate interests (Art. 6(1)(f)) |
| Compliance with legal obligations | Legal obligation (Art. 6(1)(c)) |
| Product analytics (aggregated, anonymized) | Legitimate interests (Art. 6(1)(f)) |

---

## Article 7 — Consent Management

### Consent at Registration

Users provide consent for marketing communications at registration via an explicit opt-in checkbox. The checkbox is unchecked by default. Consent is not bundled with acceptance of Terms of Service.

Consent records capture:
- User ID
- Timestamp (UTC)
- Consent version (tied to Privacy Policy version)
- Channel (web registration, API, SCIM provisioning)
- IP address at time of consent

### Consent Withdrawal

Users can withdraw marketing consent at any time via:
- Account settings (Privacy tab)
- Unsubscribe link in every marketing email
- Written request to privacy@chaosplane.dev

Withdrawal is processed within 24 hours. Withdrawal does not affect the lawfulness of processing prior to withdrawal.

### Consent Records Storage

Consent records are stored in a dedicated `consent_records` table, separate from user profile data, to ensure they survive account deletion for the legally required retention period (3 years after withdrawal or account deletion).

---

## Article 30 — Records of Processing Activities (RoPA)

ChaosPlane maintains a RoPA covering all processing activities. Key entries:

| Processing Activity | Controller/Processor | Data Categories | Retention | Recipients |
|---|---|---|---|---|
| User account management | Controller | Name, email, password hash, MFA secrets | Duration of account + 30-day grace | Internal only |
| Authentication and session management | Controller | Email, IP address, device fingerprint, session tokens | 90 days (logs) | Internal only |
| Experiment execution | Processor (on behalf of customer) | Topology data, experiment definitions, results | Per customer plan tier | Customer-controlled |
| Audit logging | Processor | Actor ID, IP, resource, action, timestamp | 90 days hot / 3 years cold | Customer (on export) |
| Billing | Controller | Name, email, billing address, payment method (tokenized) | 7 years (legal obligation) | Stripe |
| Marketing communications | Controller | Name, email, consent record | Until withdrawal + 3 years | SendGrid |
| Security monitoring | Controller | IP address, user agent, event type | 90 days | Internal only |
| Support tickets | Controller | Name, email, ticket content | 3 years after closure | Zendesk |

The RoPA is reviewed quarterly and updated when new processing activities are introduced.

---

## Article 25 — Data Protection by Design and by Default

### By Design

- Tenant data is isolated at the database level. Cross-tenant queries are architecturally prevented via row-level security and tenant-scoped API middleware.
- PII is masked in application logs. Experiment payloads do not log sensitive field values.
- Encryption at rest (AES-256 via AWS KMS) and in transit (TLS 1.2+) applied to all personal data.
- Pseudonymization applied where possible: internal user IDs (UUIDs) used in experiment data rather than email addresses.

### By Default

- Marketing opt-in is unchecked by default.
- Data export requires explicit user action; no automatic sharing with third parties.
- New features default to minimum data collection. Product analytics use aggregated, anonymized data.
- Session tokens expire after 24 hours of inactivity. WebSocket sessions re-authenticated every 5 minutes.

---

## Article 17 — Right to Erasure (Right to be Forgotten)

### Automation Pipeline

ChaosPlane has implemented an automated Right to Erasure pipeline. The process is:

1. **Request intake** — User submits deletion request via account settings ("Delete my account") or via written request to privacy@chaosplane.dev. Enterprise customers can also trigger deletion via the SCIM deprovisioning API.

2. **30-day grace period** — Account enters a soft-deleted state. The user can cancel deletion during this period. The account is inaccessible but data is retained for recovery. A confirmation email is sent immediately, and a reminder is sent at day 25.

3. **Anonymization** — At day 30, the pipeline executes:
   - Email address replaced with a SHA-256 hash (irreversible, preserves referential integrity in audit logs)
   - Display name replaced with "Deleted User"
   - Profile fields (phone, avatar, bio) set to null
   - API keys and OAuth tokens revoked and deleted
   - MFA secrets deleted
   - Session tokens invalidated via Redis pub/sub

4. **Cascade soft delete** — All user-owned resources (experiment definitions, schedules, personal API keys) are soft-deleted. Organization-owned resources (shared experiments, audit logs) are retained per the organization's data retention policy, with the actor field anonymized.

5. **Hard deletion from backups** — Backup purge job runs within 90 days of anonymization. Backup snapshots older than 90 days are deleted. This is the outer bound for complete erasure from all systems.

6. **Consent record retention** — Consent records are retained for 3 years after deletion (legal obligation to demonstrate lawful processing). They contain only the hashed user ID, not the original email.

### Exceptions to Erasure

Erasure may be refused or deferred where:
- Data is required for compliance with a legal obligation (e.g., billing records retained 7 years)
- Data is required for the establishment, exercise, or defense of legal claims
- The request is manifestly unfounded or excessive

Exceptions are documented and communicated to the data subject within 30 days.

### Erasure Request Tracking

All erasure requests are logged in the `erasure_requests` table with:
- Request ID
- Hashed user identifier (post-anonymization)
- Request date
- Grace period expiry
- Anonymization completion date
- Backup purge completion date
- Any exceptions applied

---

## Article 20 — Right to Data Portability

### Data Export API

Users can export their personal data in machine-readable format (JSON) via:
- Account settings (Export my data)
- `GET /api/v1/users/me/export` (authenticated)

The export includes:
- Account profile (name, email, registration date, preferences)
- Experiment definitions authored by the user
- Experiment run history (last 12 months, or full history on Enterprise)
- Audit log entries where the user is the actor
- Consent records

Export is generated asynchronously. The user receives an email with a time-limited download link (24-hour expiry). The export file is deleted from ChaosPlane storage after download or expiry.

Enterprise customers can request bulk exports for all users in their organization via the admin API.

---

## Article 15/16 — Right of Access and Rectification

Users can view and update their personal data via account settings at any time. Updatable fields include name, email (requires re-verification), and notification preferences.

For data held in audit logs (which are append-only and cannot be modified), ChaosPlane provides read access via the audit log export feature. Corrections to audit log entries are not possible by design (tamper-evident logs), but supplementary notes can be added.

Data subject access requests submitted via privacy@chaosplane.dev are fulfilled within 30 days.

---

## Article 21 — Right to Object

Users can object to processing based on legitimate interests (e.g., product analytics) via account settings or by contacting privacy@chaosplane.dev. Objections are reviewed within 30 days. Where the objection is upheld, processing ceases and data is deleted or anonymized.

---

## Article 28 — Data Processing Agreements (DPAs)

ChaosPlane enters into DPAs with:
- All enterprise customers where ChaosPlane processes personal data on their behalf (ChaosPlane as processor)
- All subprocessors that handle personal data on ChaosPlane's behalf

ChaosPlane's standard DPA is available at chaosplane.dev/dpa. It covers:
- Subject matter and duration of processing
- Nature and purpose of processing
- Type of personal data and categories of data subjects
- Obligations and rights of the controller
- Subprocessor list and change notification process
- Data subject rights assistance obligations
- Security measures (referencing this document and the ISO 27001 control mapping)
- Breach notification (72-hour obligation)
- Return or deletion of data on contract termination
- Audit rights

### Subprocessor List

| Subprocessor | Service | Location | Transfer Mechanism |
|---|---|---|---|
| Amazon Web Services | Cloud infrastructure | US (primary), EU (optional) | SCCs / AWS DPA |
| Stripe | Payment processing | US | SCCs / Stripe DPA |
| SendGrid (Twilio) | Transactional email | US | SCCs / Twilio DPA |
| OpenAI | AI Assistant (LLM API) | US | SCCs / OpenAI DPA |
| Anthropic | AI Assistant (LLM API) | US | SCCs / Anthropic DPA |
| Zendesk | Customer support | US | SCCs / Zendesk DPA |

Customers are notified of subprocessor changes with 30 days' notice via email and the changelog at chaosplane.dev/subprocessors.

---

## Article 32 — Security of Processing

Technical and organizational measures applied to personal data:

| Measure | Implementation |
|---|---|
| Encryption at rest | AES-256 via AWS KMS. All databases, S3 buckets, and backups encrypted. |
| Encryption in transit | TLS 1.2+ on all external connections. mTLS for internal service communication. |
| Access control | RBAC/ABAC. Principle of least privilege. MFA required for all staff. |
| Pseudonymization | Internal UUIDs used in experiment data. Email addresses not logged in experiment payloads. |
| Resilience | Multi-AZ deployment. Automated backups. RTO 4 hours, RPO 1 hour (Enterprise). |
| Testing | Annual third-party penetration test. Quarterly internal security reviews. |
| Incident response | Documented IR plan. 72-hour breach notification to supervisory authority. |

---

## Article 33/34 — Breach Notification

ChaosPlane's incident response plan includes a GDPR breach notification track:

1. Security event detected and classified as potential personal data breach
2. Breach assessment completed within 24 hours of detection: scope, categories of data, number of individuals affected, likely consequences
3. If breach meets notification threshold (risk to individuals' rights and freedoms): notification to lead supervisory authority within 72 hours of becoming aware
4. If breach is likely to result in high risk to individuals: direct notification to affected individuals without undue delay
5. Breach documented in the breach register regardless of notification outcome

The breach register records: date of breach, date of discovery, nature of breach, data categories affected, approximate number of records, consequences, remediation actions, and notification decisions.

---

## Article 37 — Data Protection Officer (DPO)

ChaosPlane is designating a DPO as part of Phase 4 (Month 18). The DPO will:

- Monitor compliance with GDPR and this framework
- Advise on data protection impact assessments (DPIAs)
- Act as the contact point for supervisory authorities
- Handle data subject requests escalated from the support team
- Conduct annual reviews of the RoPA and consent management processes

DPO contact: dpo@chaosplane.dev (to be activated upon appointment)

Until DPO appointment, data protection queries are handled by the Compliance Manager at compliance@chaosplane.dev.

---

## Article 35 — Data Protection Impact Assessments (DPIAs)

DPIAs are conducted for processing activities that are likely to result in high risk to individuals. Triggers include:

- Systematic and extensive profiling
- Large-scale processing of special category data
- Systematic monitoring of publicly accessible areas
- New features involving significant new data collection or processing

ChaosPlane's AI Assistant feature (topology analysis, vulnerability detection) was assessed via DPIA. Key findings: topology data may incidentally include hostnames or IP addresses that could identify individuals in some environments. Mitigation: topology data is processed in-memory and not persisted beyond the experiment session unless explicitly saved by the user.

---

## International Data Transfers

ChaosPlane's primary infrastructure is in AWS us-east-1 (US). For EEA customers, data transfer to the US is covered by Standard Contractual Clauses (SCCs, 2021 version) incorporated into the DPA.

Enterprise customers requiring EU data residency can request deployment to AWS eu-west-1 (Ireland) or eu-central-1 (Frankfurt). This is available as a contractual option on the Enterprise plan.

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for compliance |
| Owner | DPO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | October 2026 |
| Standard reference | Regulation (EU) 2016/679 (GDPR) |
