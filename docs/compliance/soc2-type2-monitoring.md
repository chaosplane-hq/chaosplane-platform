# SOC 2 Type II Continuous Monitoring Plan — ChaosPlane

> ChaosPlane completed its SOC 2 Type I examination and is now in the SOC 2 Type II observation period. This document defines the continuous monitoring plan for the Type II engagement: the 6-month observation schedule, evidence collection procedures, control testing schedule, and auditor coordination process.
>
> A SOC 2 Type II report provides an auditor's opinion on whether controls are suitably designed AND operating effectively over the observation period. The observation period for ChaosPlane's Type II engagement runs from Month 14 through Month 20 (6 months).
>
> This document is intended for use by the independent service auditor, the internal compliance team, and enterprise customers conducting vendor due diligence.

---

## Observation Period

| Field | Value |
|---|---|
| Observation Start | Month 14 (first day) |
| Observation End | Month 20 (last day) |
| Duration | 6 months |
| Trust Service Criteria | Security (CC), Availability (A), Processing Integrity (PI), Confidentiality (C), Privacy (P) |
| Auditor | To be confirmed — independent CPA firm with SOC 2 practice |
| Report Target | Month 21 (report issuance) |

---

## Control Testing Schedule

Controls are tested across the observation period using a combination of automated evidence collection, periodic manual testing, and auditor walkthroughs. The schedule below defines when each control category is tested and what evidence is collected.

### Month 14 — Observation Period Start

| Activity | Owner | Evidence Produced |
|---|---|---|
| Kick-off meeting with auditor | Compliance Manager | Meeting minutes, agreed testing plan |
| Baseline control inventory review | Compliance Manager | Updated control inventory vs. Type I report |
| Evidence collection tooling configured | Security Engineer | Tooling configuration documentation |
| Automated evidence collection activated | Security Engineer | First automated evidence snapshot |
| Access rights review (quarterly) | Engineering Lead | Access review report, sign-off |
| Backup restore test | SRE | Restore test report with RTO/RPO results |
| Security awareness training records pulled | HR / Compliance | Training completion report |
| Incident response plan review | CISO | Reviewed IR plan, version history |

### Month 15

| Activity | Owner | Evidence Produced |
|---|---|---|
| Vulnerability scan (OS/container) | Security Engineer | Scan report, remediation tickets |
| Dependency scan review | Security Engineer | Dependabot/Trivy report, patch status |
| Change management sample pull | Compliance Manager | 20 pull requests with review/approval records |
| Vendor review (critical subprocessors) | Compliance Manager | Vendor assessment records |
| MFA enforcement verification | Security Engineer | IAM/IdP MFA enforcement screenshot, config export |
| Penetration test scoping (annual) | CISO | Scoping document, 3PAO/pen test firm engagement |

### Month 16

| Activity | Owner | Evidence Produced |
|---|---|---|
| Access rights review (quarterly) | Engineering Lead | Access review report, sign-off |
| Audit log integrity check | Security Engineer | Log integrity verification report |
| Incident log review (Months 14-16) | Compliance Manager | Incident register, post-mortems for P0/P1 events |
| Backup verification | SRE | Backup completion logs, encryption verification |
| Security training completion check | HR / Compliance | Training completion report (% complete) |
| Auditor interim walkthrough | Compliance Manager | Walkthrough notes, open items list |
| SIEM alert review sample | Security Engineer | 30-day alert log, response records |

### Month 17

| Activity | Owner | Evidence Produced |
|---|---|---|
| Annual penetration test execution | 3rd party / Security Engineer | Penetration test report, finding remediation plan |
| Vulnerability scan (OS/container) | Security Engineer | Scan report, remediation tickets |
| Change management sample pull | Compliance Manager | 20 pull requests with review/approval records |
| Offboarding records review | HR / Compliance | Offboarding checklist completions (sample of 5) |
| Encryption configuration verification | Security Engineer | KMS key policy export, TLS config screenshot |
| Capacity management review | SRE | Capacity metrics, auto-scaling event logs |

### Month 18

| Activity | Owner | Evidence Produced |
|---|---|---|
| Access rights review (quarterly) | Engineering Lead | Access review report, sign-off |
| Backup restore test | SRE | Restore test report with RTO/RPO results |
| Incident log review (Months 16-18) | Compliance Manager | Incident register, post-mortems |
| Data subject request log review | DPO / Compliance | DSR log, response time metrics |
| Vendor SOC 2 report collection | Compliance Manager | AWS, Stripe, SendGrid SOC 2 reports |
| Auditor interim walkthrough | Compliance Manager | Walkthrough notes, open items list |
| Penetration test finding remediation check | Security Engineer | Remediation evidence for critical/high findings |

### Month 19

| Activity | Owner | Evidence Produced |
|---|---|---|
| Vulnerability scan (OS/container) | Security Engineer | Scan report, remediation tickets |
| Change management sample pull | Compliance Manager | 20 pull requests with review/approval records |
| Security awareness training records pulled | HR / Compliance | Training completion report |
| Availability metrics review | SRE | Uptime report, SLA compliance metrics |
| Audit log retention verification | Security Engineer | Log retention policy config, S3 lifecycle rules |
| Risk assessment review | CISO | Risk register, treatment plan updates |

### Month 20 — Observation Period End

| Activity | Owner | Evidence Produced |
|---|---|---|
| Access rights review (quarterly) | Engineering Lead | Access review report, sign-off |
| Final evidence package assembly | Compliance Manager | Complete evidence package for auditor |
| Incident log review (Months 18-20) | Compliance Manager | Incident register, post-mortems |
| Backup restore test | SRE | Restore test report |
| Final auditor walkthrough | Compliance Manager + CISO | Walkthrough notes, management representation letter |
| Management representation letter | CISO | Signed management rep letter |
| Outstanding finding remediation | Security Engineer | Remediation evidence for any open items |

---

## Evidence Collection Procedures

### Automated Evidence Collection

The following evidence is collected automatically via compliance management tooling (Vanta or Drata) on a continuous basis:

| Evidence Type | Collection Method | Frequency | Retention |
|---|---|---|---|
| User access list (all systems) | API integration with AWS IAM, GitHub, production DB | Daily | 7 years |
| MFA enforcement status | API integration with IdP (Okta / Azure AD) | Daily | 7 years |
| Vulnerability scan results | Trivy, Dependabot API | Per build + weekly | 7 years |
| Backup completion status | AWS RDS / S3 backup API | Daily | 7 years |
| Encryption configuration | AWS KMS, RDS, S3 API | Weekly | 7 years |
| Uptime / availability metrics | CloudWatch, status page API | Continuous | 7 years |
| Security training completion | HRIS / LMS API | Weekly | 7 years |
| Incident tickets | Ticketing system API | Daily | 7 years |
| Change management records | GitHub PR API | Per PR | 7 years |
| Audit log integrity | Append-only log hash verification | Daily | 7 years |

### Manual Evidence Collection

The following evidence requires manual collection by the compliance team:

| Evidence Type | Collection Method | Frequency | Owner |
|---|---|---|---|
| Access rights review sign-off | Manager sign-off on access review report | Quarterly | Engineering Lead |
| Vendor assessment records | Security questionnaire responses, SOC 2 reports | Annual | Compliance Manager |
| Penetration test report | Engagement with 3rd party pen test firm | Annual | CISO |
| Backup restore test report | SRE executes restore, documents results | Quarterly | SRE |
| Offboarding checklist completions | HR pulls completed checklists | Per event + quarterly sample | HR |
| Incident post-mortems | Engineering writes post-mortem for P0/P1 | Per incident | Engineering Lead |
| Risk assessment | CISO leads annual risk assessment | Annual | CISO |
| Management representation letter | CISO signs at end of observation period | Annual | CISO |

---

## Control Testing Procedures

### CC6.1 — Logical Access Controls

**Test objective:** Verify that access to ChaosPlane systems is restricted to authorized users with appropriate roles.

**Testing procedures:**
1. Pull complete user access list from production environment (automated, monthly)
2. Verify each user has a corresponding active employment record or active customer contract
3. Verify no users have roles exceeding their documented job function
4. Sample 10 users and verify their access was provisioned via the documented access request process
5. Verify MFA is enforced for all staff accounts and all Enterprise customer organizations that have enabled org-wide MFA enforcement

**Evidence:** User access list, employment records (sample), access request tickets (sample), MFA enforcement configuration screenshot.

**Testing frequency:** Quarterly (access review) + continuous automated monitoring.

---

### CC6.2 — New User Provisioning

**Test objective:** Verify that new user accounts are created only after appropriate authorization.

**Testing procedures:**
1. Pull list of all new user accounts created during the observation period
2. For employee accounts: verify each has a corresponding approved access request ticket and manager approval
3. For customer accounts: verify each was created via self-service registration (email verification) or SCIM provisioning from an authorized customer IdP
4. Sample 10 new accounts and trace to authorization record

**Evidence:** New account creation logs, access request tickets (sample), SCIM provisioning logs (sample).

**Testing frequency:** Quarterly sample + continuous automated monitoring.

---

### CC6.6 — Logical Access — External Threats

**Test objective:** Verify that controls are in place to protect against unauthorized access from outside the system boundary.

**Testing procedures:**
1. Verify WAF is active and logging on all public endpoints
2. Review WAF rule set — confirm OWASP Core Rule Set is applied
3. Verify rate limiting is configured on authentication endpoints
4. Review authentication failure logs for evidence of brute force attempts and verify blocking occurred
5. Verify TLS 1.2+ is enforced on all external endpoints (TLS scan)

**Evidence:** WAF configuration export, rate limiting configuration, authentication failure log sample, TLS scan report.

**Testing frequency:** Monthly automated + quarterly manual review.

---

### CC7.2 — System Monitoring

**Test objective:** Verify that the system is monitored for anomalous behavior and that alerts are investigated.

**Testing procedures:**
1. Pull 30-day alert log from monitoring system
2. Sample 10 alerts and verify each was acknowledged and investigated within the documented SLA
3. Verify on-call rotation is documented and staffed
4. Verify SIEM integration is active for at least one Enterprise customer (as evidence of capability)

**Evidence:** 30-day alert log, alert response records (sample), on-call schedule, SIEM integration configuration.

**Testing frequency:** Monthly.

---

### CC8.1 — Change Management

**Test objective:** Verify that all changes to the production environment go through the documented change management process.

**Testing procedures:**
1. Pull list of all production deployments during the sample period
2. For each deployment, verify a corresponding pull request exists with peer review approval
3. Verify CI/CD pipeline ran and all tests passed before deployment
4. Sample 20 pull requests and verify security review was conducted for significant changes
5. Verify no direct pushes to main branch occurred (branch protection rules)

**Evidence:** Deployment logs, pull request records (sample), CI/CD pipeline logs (sample), branch protection configuration.

**Testing frequency:** Quarterly sample (20 PRs per quarter).

---

### A1.2 — Availability — Environmental and Regulatory Changes

**Test objective:** Verify that the system maintains availability per SLA commitments.

**Testing procedures:**
1. Pull uptime metrics for the observation period from CloudWatch and status page
2. Calculate actual uptime percentage vs. SLA commitment (99.9% for Enterprise)
3. Review any downtime incidents — verify each has a post-mortem and root cause analysis
4. Verify multi-AZ configuration is active for all critical components (RDS, EKS)
5. Verify auto-scaling is configured and has triggered at least once during the observation period

**Evidence:** Uptime report (6-month), incident post-mortems, multi-AZ configuration screenshot, auto-scaling event logs.

**Testing frequency:** Monthly metrics pull + quarterly review.

---

### A1.3 — Availability — Recovery

**Test objective:** Verify that backup and recovery procedures are in place and tested.

**Testing procedures:**
1. Verify automated backups are configured and completing successfully (daily)
2. Pull backup completion logs for the observation period — verify no missed backups
3. Review quarterly backup restore test reports — verify RTO and RPO targets were met
4. Verify backups are encrypted at rest (KMS)

**Evidence:** Backup completion logs (6-month), restore test reports (2 per observation period), backup encryption configuration.

**Testing frequency:** Continuous automated monitoring + quarterly restore test.

---

### C1.2 — Confidentiality — Disposal

**Test objective:** Verify that customer data is disposed of per the documented retention and deletion policy.

**Testing procedures:**
1. Pull list of account deletion requests during the observation period
2. For each request, verify the 30-day grace period was applied
3. Verify anonymization pipeline ran at day 30 (email hashed, name replaced)
4. Verify backup purge job ran within 90 days of anonymization
5. Sample 5 deletion requests and trace through the full pipeline

**Evidence:** Erasure request log, anonymization pipeline execution logs (sample), backup purge job logs (sample).

**Testing frequency:** Quarterly sample.

---

### P8.1 — Privacy — Monitoring and Enforcement

**Test objective:** Verify that data subject requests are handled within the documented SLA (30 days).

**Testing procedures:**
1. Pull list of all data subject requests (access, deletion, correction, portability) during the observation period
2. Calculate response time for each request
3. Verify all requests were responded to within 30 days
4. Verify deletion requests triggered the automated erasure pipeline

**Evidence:** Data subject request log, response time metrics, erasure pipeline execution logs (sample).

**Testing frequency:** Quarterly.

---

## Auditor Coordination

### Interim Walkthroughs

Two interim walkthroughs are scheduled during the observation period (Month 16 and Month 18). Each walkthrough covers:

- Review of evidence collected to date
- Walkthrough of any new or changed controls
- Discussion of open items and remediation status
- Preview of any significant changes to the system

### Evidence Submission

Evidence is submitted to the auditor via a secure shared workspace (auditor-provided). Evidence packages are organized by control and include:

- Screenshot or export of the evidence
- Date and time of collection
- Name of the person who collected it
- Brief description of what the evidence demonstrates

### Management Representation Letter

The CISO signs a management representation letter at the end of the observation period confirming:

- The system description is accurate and complete
- Controls were suitably designed and operating effectively throughout the observation period
- All known exceptions have been disclosed to the auditor
- No material changes to the system occurred that are not reflected in the SSP

---

## Metrics and KPIs

The following metrics are tracked throughout the observation period to demonstrate control effectiveness:

| Metric | Target | Measurement |
|---|---|---|
| Platform uptime | ≥ 99.9% (Enterprise SLA) | CloudWatch + status page |
| Critical CVE patch time | ≤ 7 days | Vulnerability management system |
| High CVE patch time | ≤ 30 days | Vulnerability management system |
| Access review completion | 100% quarterly | Access review sign-off records |
| Security training completion | 100% of staff | LMS completion report |
| Backup success rate | 100% | Backup completion logs |
| Restore test RTO | ≤ 4 hours | Restore test reports |
| Restore test RPO | ≤ 1 hour | Restore test reports |
| Data subject request response time | ≤ 30 days | DSR log |
| Incident P0/P1 post-mortem completion | 100% within 5 business days | Incident register |
| MFA enforcement (staff) | 100% | IAM / IdP report |

---

## Exception Handling

If a control exception is identified during the observation period:

1. Exception documented in the exception register with date, description, and affected control
2. Root cause analysis completed within 5 business days
3. Remediation plan developed and approved by CISO
4. Auditor notified of material exceptions within 10 business days
5. Remediation evidence collected and submitted to auditor
6. Exception register reviewed at each interim walkthrough

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Active — observation period in progress |
| Owner | Compliance Manager |
| Last reviewed | April 2026 |
| Next review | Monthly (during observation period) |
| Standard reference | AICPA Trust Services Criteria (2017, updated 2022) |
| Observation period | Month 14 — Month 20 |
