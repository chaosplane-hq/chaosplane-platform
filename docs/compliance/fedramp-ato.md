# FedRAMP ATO — Continuous Monitoring and ConMon Reporting — ChaosPlane

> ChaosPlane Government Edition received FedRAMP Moderate Agency Authorization to Operate (ATO) in Phase 5. This document covers the post-ATO continuous monitoring (ConMon) program, monthly ConMon reporting obligations, significant change management, and the path toward JAB Provisional ATO (P-ATO).
>
> This document supplements the FedRAMP SSP (fedramp-ssp-draft.md) and is intended for the ChaosPlane Compliance Manager, the sponsoring agency ISSO, and the FedRAMP PMO.

---

## ATO Status

| Field | Value |
|---|---|
| Authorization type | Agency ATO |
| Sponsoring agency | To be designated |
| Impact level | Moderate |
| Authorization date | Target: Month 27-28 (Phase 5) |
| ATO expiration | 3 years from authorization date |
| 3PAO | To be engaged Month 23 |
| FedRAMP Marketplace status | Listed upon ATO |
| Path to JAB P-ATO | Targeted after 2+ agency ATOs |

---

## Continuous Monitoring Program Overview

FedRAMP requires authorized cloud service providers to maintain a continuous monitoring program that provides ongoing assurance that security controls remain effective. ChaosPlane's ConMon program follows the FedRAMP Continuous Monitoring Strategy Guide and NIST SP 800-137.

The ConMon program has four pillars:

1. Ongoing vulnerability scanning and patch management
2. Security control assessments (annual subset, full triennial)
3. POA&M management and remediation tracking
4. Monthly ConMon reporting to the authorizing agency and FedRAMP PMO

---

## ConMon Activities and Cadence

### Monthly Activities

| Activity | Description | Owner | Deliverable |
|---|---|---|---|
| Vulnerability scan — OS/container | Trivy scan of all running container images and host OS | Security Engineer | Scan report |
| Vulnerability scan — web application | DAST scan of ChaosPlane Government Edition web application | Security Engineer | DAST report |
| POA&M review | Review all open POA&M items, update status, close completed items | Compliance Manager | Updated POA&M |
| ConMon report | Monthly ConMon report to agency AO and FedRAMP PMO | Compliance Manager | ConMon report package |
| Incident review | Review all security incidents from the month, confirm US-CERT reporting | Security Engineer | Incident summary |
| Privileged access review | Review all privileged access grants, revoke unnecessary access | Security Engineer | Access review log |
| Key rotation check | Verify encryption key rotation schedule is on track | Security Engineer | Key rotation status |

### Quarterly Activities

| Activity | Description | Owner | Deliverable |
|---|---|---|---|
| Security control spot-check | Assess 10-15% of Moderate baseline controls | Compliance Manager | Control assessment results |
| Penetration test (lightweight) | Targeted pen test of high-risk areas identified in monthly scans | 3PAO or internal red team | Pen test report |
| DR test | Disaster recovery test for GovCloud deployment | Platform Engineering | DR test report |
| Third-party service review | Review all leveraged FedRAMP authorizations for continued validity | Compliance Manager | Third-party review |
| Backup restore test | Restore from backup to verify integrity | Platform Engineering | Restore test report |

### Annual Activities

| Activity | Description | Owner | Deliverable |
|---|---|---|---|
| Full penetration test | Comprehensive penetration test by DISA-approved 3PAO | 3PAO | Penetration test report |
| Security control assessment | Full assessment of Moderate baseline by 3PAO | 3PAO | Security Assessment Report (SAR) |
| SSP update | Update SSP to reflect any changes to system boundary, components, or controls | Compliance Manager | Updated SSP |
| Privacy impact assessment | Review PII handling, update PIA if system changes affect privacy | Compliance Manager | Updated PIA |
| Contingency plan test | Full DR exercise including failover and recovery | Platform Engineering | CP test report |
| Risk assessment | Annual risk assessment, update risk register | CISO | Risk assessment report |

---

## Monthly ConMon Report Package

Each monthly ConMon report package submitted to the agency AO and FedRAMP PMO contains:

### 1. Executive Summary

One-page summary covering:
- Overall security posture (Green / Yellow / Red)
- New vulnerabilities discovered and remediation status
- Open POA&M count and trend
- Incidents during the month
- Significant changes (if any)

### 2. Vulnerability Scan Results

```
Vulnerability Summary — [Month YYYY]

Container/OS Scans:
  Critical:  [count] — [count] remediated, [count] in POA&M
  High:      [count] — [count] remediated, [count] in POA&M
  Medium:    [count] — tracked, remediation within 90 days
  Low:       [count] — tracked, remediation within 180 days

Web Application Scans:
  Critical:  [count]
  High:      [count]
  Medium:    [count]

New CVEs this month: [list of CVE IDs with CVSS scores]
Remediated this month: [list of CVE IDs]
```

Remediation SLAs (FedRAMP Moderate requirement):
- Critical (CVSS 9.0+): 30 days
- High (CVSS 7.0-8.9): 90 days
- Medium (CVSS 4.0-6.9): 180 days
- Low (CVSS < 4.0): 365 days

### 3. POA&M Status

The Plan of Action and Milestones tracks all open security findings:

| POA&M ID | Control | Finding | Severity | Opened | Target Close | Status |
|---|---|---|---|---|---|---|
| POA-001 | SC-13 | FIPS mode validation pending for chaos-daemon component | High | Month 15 | Month 16 | Closed |
| POA-002 | IA-8 | PIV/CAC integration testing requires agency IdP | Medium | Month 15 | Month 17 | In progress |
| POA-003 | CA-7 | ConMon tooling selection in progress | Low | Month 15 | Month 16 | Closed |

POA&M items are tracked in ChaosPlane's compliance management system. The Compliance Manager reviews all open items monthly and updates the FedRAMP PMO on any items approaching or exceeding their target close date.

### 4. Incident Report Summary

For each security incident during the month:

```
Incident: [INC-YYYY-NNN]
Date/Time: [UTC timestamp]
Severity: [P0 / P1 / P2 / P3]
Description: [Brief description]
US-CERT reported: [Yes/No — timestamp if yes]
Agency AO notified: [Yes/No — timestamp if yes]
Status: [Resolved / Under investigation]
Root cause: [Brief root cause]
Remediation: [Actions taken]
```

P0/P1 incidents are reported to US-CERT within 1 hour of detection. The agency AO is notified within 1 hour. The FedRAMP PMO is notified within 24 hours.

### 5. Significant Change Log

Any changes to the system boundary, components, or security controls during the month:

| Change | Type | Date | Impact | AO Notified |
|---|---|---|---|---|
| Kubernetes version upgrade (1.28 → 1.29) | Minor | [date] | No boundary change | No (minor change) |
| New AI model endpoint added | Significant | [date] | New data flow | Yes — pre-approval obtained |

Significant change categories (require AO pre-approval):
- New external services added to system boundary
- Changes to authentication or authorization mechanisms
- New data types processed
- Changes to encryption algorithms or key management
- New interconnections with external systems

Minor changes (notify AO in monthly report, no pre-approval):
- Software version upgrades within approved component list
- Configuration changes within approved baseline
- Patch application for approved CVEs

---

## Significant Change Process

When a significant change is planned:

1. Compliance Manager drafts a Significant Change Request (SCR) describing the change, security impact, and updated control implementation
2. CISO reviews and approves the SCR
3. SCR submitted to agency AO for review (minimum 30 days before implementation)
4. If AO approves, change is implemented and SSP updated
5. If change affects system boundary, 3PAO assessment of affected controls may be required
6. ConMon report for the month of implementation includes the change details

---

## Incident Response and US-CERT Reporting

### Reporting Thresholds

| Severity | Definition | US-CERT Report | Agency AO | FedRAMP PMO |
|---|---|---|---|---|
| P0 | Active breach, data exfiltration, system compromise | Within 1 hour | Within 1 hour | Within 24 hours |
| P1 | Suspected breach, significant availability impact | Within 1 hour | Within 1 hour | Within 24 hours |
| P2 | Security event with limited impact | Within 24 hours | Within 24 hours | Monthly report |
| P3 | Security event with no customer impact | Monthly report | Monthly report | Monthly report |

### US-CERT Reporting

US-CERT reports are submitted via the US-CERT Incident Reporting System (https://www.cisa.gov/report):

```
Incident Report Fields:
- Organization: ChaosPlane, Inc.
- System: ChaosPlane Government Edition (FedRAMP Moderate)
- Date/Time of detection: [UTC]
- Date/Time of occurrence: [UTC, if known]
- Incident type: [Unauthorized access / Denial of service / Malicious code / etc.]
- Description: [Factual description, no speculation]
- Affected systems: [Component names, not IP addresses in initial report]
- Data involved: [CUI categories if applicable]
- Actions taken: [Containment steps]
- Point of contact: [CISO name, phone, email]
```

---

## JAB P-ATO Path

After achieving Agency ATO and demonstrating a mature ConMon program, ChaosPlane will pursue JAB (Joint Authorization Board) Provisional ATO. JAB P-ATO is issued by DoD, DHS, and GSA jointly and is recognized by all federal agencies without requiring individual agency ATOs.

JAB P-ATO requirements beyond Agency ATO:
- Minimum 2 active agency ATOs
- 12+ months of ConMon history with no critical findings
- FedRAMP PMO sponsorship (requires application and selection)
- Full 3PAO assessment against JAB requirements
- Public listing on FedRAMP Marketplace

JAB P-ATO timeline: targeted 12-18 months after first Agency ATO.

---

## ConMon Tooling

| Tool | Purpose | Notes |
|---|---|---|
| Trivy | Container and OS vulnerability scanning | Runs in CI/CD and as scheduled CronJob in GovCloud |
| AWS Inspector | EC2 and ECR vulnerability scanning | GovCloud-native, continuous |
| AWS Config | Configuration compliance monitoring | Rules mapped to FedRAMP controls |
| AWS CloudTrail | API audit logging | All GovCloud accounts, 7-year retention |
| AWS Security Hub | Aggregated security findings | FedRAMP standard integration |
| OWASP ZAP | DAST web application scanning | Monthly scheduled scan |
| Custom ConMon dashboard | POA&M tracking, report generation | Internal tool, Compliance Manager |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Active — post-ATO |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | Monthly (ConMon report) / Annual (full review) |
| References | FedRAMP Continuous Monitoring Strategy Guide, NIST SP 800-137, FedRAMP Moderate Baseline |
| Prerequisite | FedRAMP SSP (fedramp-ssp-draft.md), Agency ATO |
