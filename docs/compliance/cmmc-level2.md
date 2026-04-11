# CMMC Level 2 Preparation — ChaosPlane

> ChaosPlane is preparing for Cybersecurity Maturity Model Certification (CMMC) Level 2. This document maps ChaosPlane's platform features and operational controls to the 110 practices required for CMMC Level 2, which aligns with NIST SP 800-171 Rev 2. It also describes ChaosPlane's procedures for handling Controlled Unclassified Information (CUI) within the platform.
>
> CMMC Level 2 is required for defense contractors and subcontractors that process, store, or transmit CUI. ChaosPlane Government Edition may be used by DoD contractors to test the resilience of systems that handle CUI, making CMMC Level 2 preparation necessary for that market segment.
>
> This document does not constitute a CMMC certification claim. ChaosPlane is preparing for a third-party CMMC assessment (C3PAO assessment) as part of Phase 4.

---

## Scope

The CMMC Level 2 scope covers:

- ChaosPlane Government Edition deployed in AWS GovCloud (us-gov-west-1)
- ChaosPlane systems that may process, store, or transmit CUI on behalf of DoD contractor customers
- ChaosPlane staff with access to CUI or systems that handle CUI
- The CUI marking system within the ChaosPlane platform

CUI in the ChaosPlane context includes:
- Experiment definitions that reference CUI-bearing system names, IP addresses, or topology data
- Audit logs that capture CUI-related experiment activity
- Customer-provided topology data that may contain CUI

---

## CUI Handling Procedures

### CUI Definition and Categories

CUI is information the Government creates or possesses, or that an entity creates or possesses on behalf of the Government, that a law, regulation, or Government-wide policy requires or permits an agency to handle using safeguarding or dissemination controls.

CUI categories relevant to ChaosPlane customers include:
- Critical Infrastructure (CRIT)
- Defense (DEF)
- Intelligence (INTEL)
- Law Enforcement (LAW)
- Privacy (PRVCY)

### CUI Marking in ChaosPlane

ChaosPlane Government Edition includes a CUI marking system that allows customers to:

1. Mark experiment definitions as containing CUI, with the applicable CUI category
2. Mark topology data imports as CUI
3. Apply CUI handling restrictions to audit log exports (CUI-marked exports require additional authorization)
4. Display CUI banner on all pages when a CUI-marked experiment is active

CUI markings follow the format: `CUI//[Category]` (e.g., `CUI//DEF`, `CUI//CRIT`).

### CUI Access Controls

- CUI-marked resources are accessible only to users with the `cui-access` attribute in the ABAC policy
- CUI access requires explicit grant by the organization administrator
- CUI access is logged separately in the audit log with the `CUI_ACCESS` event type
- CUI-marked experiment results are not included in standard data exports — they require a separate authorized export

### CUI Storage and Transmission

- CUI stored in ChaosPlane is encrypted at rest using AES-256 (AWS KMS GovCloud, FIPS 140-2 validated)
- CUI transmitted between components uses TLS 1.2+ with FIPS-approved cipher suites
- CUI is not stored in logs beyond what is necessary for audit purposes
- CUI is not transmitted to non-GovCloud systems

### CUI Destruction

- CUI is destroyed per the account deletion pipeline (30-day grace period, anonymization, hard deletion within 90 days)
- CUI destruction is logged in the audit log with the `CUI_DESTRUCTION` event type
- Customers can request immediate CUI destruction via the admin API

---

## CMMC Level 2 — 110 Practice Mapping

CMMC Level 2 maps directly to the 110 security requirements in NIST SP 800-171 Rev 2, organized across 14 domains.

### AC — Access Control (22 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| AC.L2-3.1.1 | Limit system access to authorized users | RBAC/ABAC enforced on all platform resources. User accounts require email verification or SCIM provisioning. Unauthorized access attempts logged and alerted. |
| AC.L2-3.1.2 | Limit system access to types of transactions and functions authorized users are permitted to execute | RBAC roles define permitted actions per resource type. ABAC adds attribute-based restrictions (e.g., CUI access requires `cui-access` attribute). |
| AC.L2-3.1.3 | Control the flow of CUI in accordance with approved authorizations | CUI-marked resources restricted to users with `cui-access` attribute. CUI export requires additional authorization. CUI not transmitted outside GovCloud boundary. |
| AC.L2-3.1.4 | Separate the duties of individuals to reduce the risk of malevolent activity | Production deployments require second approver. Billing and infrastructure access separated by role. No single person can approve and deploy their own changes. |
| AC.L2-3.1.5 | Employ the principle of least privilege | All IAM roles and application roles follow least privilege. Privileged access restricted to named individuals. Access reviewed quarterly. |
| AC.L2-3.1.6 | Use non-privileged accounts when accessing non-security functions | Staff use standard accounts for day-to-day work. Privileged accounts used only for privileged operations. Separate accounts for privileged and non-privileged access. |
| AC.L2-3.1.7 | Prevent non-privileged users from executing privileged functions | Privileged functions (production deployment, database access, key management) restricted to named privileged accounts. Enforcement via IAM policies and RBAC. |
| AC.L2-3.1.8 | Limit unsuccessful logon attempts | Account lockout after 10 failed login attempts. Lockout duration: 15 minutes. Brute force detection triggers CAPTCHA and alerts. |
| AC.L2-3.1.9 | Provide privacy and security notices consistent with CUI rules | Privacy and security notices displayed at login. CUI handling notice displayed when CUI-marked resources are accessed. |
| AC.L2-3.1.10 | Use session lock with pattern-hiding displays after period of inactivity | Session timeout after 30 minutes of inactivity (Government Edition default, configurable). Screen lock policy enforced on employee devices (5 minutes). |
| AC.L2-3.1.11 | Terminate sessions after defined conditions | Sessions terminated after 30 minutes of inactivity. Sessions terminated on logout. Sessions terminated on password change or MFA reset. |
| AC.L2-3.1.12 | Monitor and control remote access sessions | Remote access to production infrastructure via VPN with MFA. SSH sessions logged and monitored. Session recording enabled for privileged remote sessions. |
| AC.L2-3.1.13 | Employ cryptographic mechanisms to protect the confidentiality of remote access sessions | All remote access uses TLS 1.2+ with FIPS-approved cipher suites. VPN uses IKEv2 with AES-256. |
| AC.L2-3.1.14 | Route remote access via managed access control points | All remote access to production infrastructure routes through the bastion host (MFA required). No direct access to production systems from the internet. |
| AC.L2-3.1.15 | Authorize remote execution of privileged commands via remote access only for documented operational needs | Privileged remote commands (kubectl exec, database console) require documented operational justification. Usage logged and reviewed monthly. |
| AC.L2-3.1.16 | Authorize wireless access prior to allowing connections | Office wireless access requires corporate credentials. Guest network isolated. Remote work uses VPN. |
| AC.L2-3.1.17 | Protect wireless access using authentication and encryption | Office wireless uses WPA3. VPN uses AES-256. |
| AC.L2-3.1.18 | Control connection of mobile devices | MDM enforced on all company mobile devices. Jailbroken/rooted devices blocked. |
| AC.L2-3.1.19 | Encrypt CUI on mobile devices and mobile computing platforms | Employee devices with access to CUI use full-disk encryption (FileVault / BitLocker). MDM enforces encryption policy. |
| AC.L2-3.1.20 | Verify and control all connections to external systems | External connections documented and approved. Third-party integrations use least-privilege API keys. External connections reviewed quarterly. |
| AC.L2-3.1.21 | Limit use of portable storage devices on external systems | Sensitive data not stored on removable media. Removable media blocked on production systems via MDM policy. |
| AC.L2-3.1.22 | Control CUI posted or processed on publicly accessible systems | CUI not posted on publicly accessible systems. Public-facing ChaosPlane documentation does not contain CUI. |

### AT — Awareness and Training (3 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| AT.L2-3.2.1 | Ensure personnel are aware of security risks associated with their activities | Security awareness training required at onboarding and annually. Covers phishing, social engineering, CUI handling, and incident reporting. |
| AT.L2-3.2.2 | Ensure personnel are trained to carry out assigned security responsibilities | Role-specific training: engineers (secure coding, FIPS crypto), operations (incident response, CUI handling), compliance (CMMC requirements). |
| AT.L2-3.2.3 | Provide security awareness training on recognizing and reporting threats | Phishing simulation twice per year. Threat recognition training included in annual security awareness program. Dedicated channel for reporting suspicious activity. |

### AU — Audit and Accountability (9 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| AU.L2-3.3.1 | Create and retain system audit logs to enable monitoring, analysis, investigation, and reporting | All API calls logged with actor, timestamp, IP, resource, and outcome. Audit logs are append-only. Retention: 90 days hot, 3 years cold (S3 GovCloud). |
| AU.L2-3.3.2 | Ensure the actions of individual users can be traced to those users | Each audit log entry includes the authenticated user ID, session ID, and IP address. No shared accounts permitted. |
| AU.L2-3.3.3 | Review and update logged events | Audit log event types reviewed annually and updated when new features are introduced. Current event types cover authentication, authorization, data access, configuration changes, and CUI access. |
| AU.L2-3.3.4 | Alert in the event of audit logging process failure | Monitoring alerts on audit log service failures. Log pipeline health checked every 5 minutes. On-call engineer alerted within 5 minutes of failure. |
| AU.L2-3.3.5 | Correlate audit record review, analysis, and reporting processes | SIEM integration available for Enterprise/Government customers. Audit logs exportable to customer SIEM in real time. |
| AU.L2-3.3.6 | Provide audit record reduction and report generation to support on-demand analysis | Audit log search and filtering available via API and web UI. Export to JSON/CSV for offline analysis. |
| AU.L2-3.3.7 | Provide a system capability that compares and synchronizes internal clocks | All systems synchronized to AWS Time Sync Service (NTP). Timestamps in audit logs are UTC. Clock drift monitored. |
| AU.L2-3.3.8 | Protect audit information and tools from unauthorized access, modification, and deletion | Audit logs are append-only. No delete or modify API exists. Log integrity protected via tamper-evident signing (Government Edition). Log access restricted to authorized roles. |
| AU.L2-3.3.9 | Limit management of audit logging to a subset of privileged users | Audit log configuration changes restricted to CISO and Security Engineers. Changes logged. |

### CM — Configuration Management (9 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| CM.L2-3.4.1 | Establish and maintain baseline configurations | Infrastructure defined as code (Terraform). Baseline configurations version-controlled in Git. Drift detection alerts on manual changes. |
| CM.L2-3.4.2 | Establish and enforce security configuration settings | Security configuration settings enforced via IaC and CI/CD pipeline. CIS Benchmarks applied to container images. |
| CM.L2-3.4.3 | Track, review, approve, and log changes to organizational systems | All changes via pull request with peer review. Production deployments require approval. Change log maintained. CloudTrail logs all AWS configuration changes. |
| CM.L2-3.4.4 | Analyze the security impact of changes prior to implementation | Security review required for significant changes. Threat modeling for new features. Security review gate before GA releases. |
| CM.L2-3.4.5 | Define, document, approve, and enforce physical and logical access restrictions | Access restrictions documented in IAM policies and RBAC configuration. Changes to access restrictions require CISO approval. |
| CM.L2-3.4.6 | Employ the principle of least functionality | Minimal base images (distroless where possible). No unnecessary services running. Unused ports closed. Unnecessary software not installed on production systems. |
| CM.L2-3.4.7 | Restrict, disable, or prevent the use of nonessential programs, functions, ports, protocols, and services | Port allowlist enforced via security groups. Unnecessary protocols disabled. Application functionality restricted to documented use cases. |
| CM.L2-3.4.8 | Apply deny-by-default and allow-by-exception policy | Security groups default-deny. IAM policies default-deny. Application RBAC default-deny. Explicit allow required for all access. |
| CM.L2-3.4.9 | Control and monitor user-installed software | Production deployments via CI/CD pipeline only. No manual software installation on production systems. Employee devices managed via MDM. |

### IA — Identification and Authentication (11 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| IA.L2-3.5.1 | Identify system users, processes, and devices | Each user has a unique UUID. Service accounts have dedicated identities. Devices identified via MDM. |
| IA.L2-3.5.2 | Authenticate the identities of users, processes, or devices | Passwords hashed with bcrypt (cost factor 12+). MFA available on all tiers, enforceable org-wide. SSO/SAML delegates to agency IdP (PIV/CAC). |
| IA.L2-3.5.3 | Use multifactor authentication for local and network access to privileged accounts | MFA required for all staff accounts. MFA required for all privileged access (AWS console, production database, kubectl). |
| IA.L2-3.5.4 | Employ replay-resistant authentication mechanisms | Session tokens are single-use and time-limited. SAML assertions include NotOnOrAfter timestamp. WebSocket sessions re-authenticated every 5 minutes. |
| IA.L2-3.5.5 | Employ identifier management | User identifiers (UUIDs) are unique and never reused. Deprovisioned accounts are disabled, not deleted, to preserve audit trail. |
| IA.L2-3.5.6 | Disable identifiers after defined inactivity period | Accounts inactive for 90 days are automatically suspended. Suspended accounts require re-authentication and manager approval to reactivate. |
| IA.L2-3.5.7 | Enforce a minimum password complexity and change requirements | Password policy: minimum 12 characters, uppercase, lowercase, number, special character. Passwords expire after 365 days. Breach password detection at registration and login. |
| IA.L2-3.5.8 | Prohibit password reuse | Last 12 passwords remembered. Password reuse prohibited. |
| IA.L2-3.5.9 | Allow temporary password use with immediate change requirement | Temporary passwords (password reset) expire after 24 hours. User must change on first login. |
| IA.L2-3.5.10 | Store and transmit only cryptographically protected passwords | Passwords hashed with bcrypt (cost factor 12+). Never stored in plaintext. Transmitted only over TLS 1.2+. |
| IA.L2-3.5.11 | Obscure feedback of authentication information | Password fields mask input. Authentication error messages do not reveal whether the username or password was incorrect. |

### IR — Incident Response (3 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| IR.L2-3.6.1 | Establish an operational incident-handling capability | Incident response plan documented with severity classification (P0-P3), escalation paths, containment, eradication, recovery. On-call rotation staffed 24/7. |
| IR.L2-3.6.2 | Track, document, and report incidents to designated officials | Incidents tracked in ticketing system. P0/P1 incidents reported to US-CERT within 1 hour. CUI-related incidents reported to the contracting agency within 72 hours per DFARS 252.204-7012. |
| IR.L2-3.6.3 | Test the organizational incident response capability | Tabletop exercise conducted annually. IR plan reviewed and updated after each P0/P1 incident. |

### MA — Maintenance (6 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| MA.L2-3.7.1 | Perform maintenance on organizational systems | Infrastructure maintenance performed via CI/CD pipeline. Emergency maintenance follows documented emergency change procedure. |
| MA.L2-3.7.2 | Provide controls on the tools, techniques, mechanisms, and personnel for system maintenance | Maintenance tools access restricted to authorized personnel. Maintenance activities logged. |
| MA.L2-3.7.3 | Ensure equipment removed for maintenance is sanitized | Not applicable (cloud-hosted). AWS handles physical equipment maintenance and sanitization. |
| MA.L2-3.7.4 | Check media containing diagnostic and test programs for malicious code | Container images scanned for malware before deployment (Trivy). Diagnostic tools reviewed before use in production. |
| MA.L2-3.7.5 | Require MFA for remote maintenance sessions | Remote maintenance sessions require MFA via VPN. Session recording enabled for privileged maintenance sessions. |
| MA.L2-3.7.6 | Supervise the maintenance activities of personnel without required access authorization | Third-party maintenance activities supervised by authorized ChaosPlane staff. Access granted only for the duration of the maintenance window. |

### MP — Media Protection (9 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| MP.L2-3.8.1 | Protect system media containing CUI | CUI stored only in encrypted cloud storage (AWS S3 GovCloud, AES-256). No CUI on removable media. |
| MP.L2-3.8.2 | Limit access to CUI on system media to authorized users | CUI-marked resources accessible only to users with `cui-access` attribute. Access logged. |
| MP.L2-3.8.3 | Sanitize or destroy system media before disposal | AWS handles physical media disposal per their FedRAMP High controls. ChaosPlane data deleted per documented retention schedule before account closure. |
| MP.L2-3.8.4 | Mark media with necessary CUI markings and distribution limitations | CUI markings applied to experiment definitions, topology data, and audit log exports that contain CUI. Marking format: `CUI//[Category]`. |
| MP.L2-3.8.5 | Control access to media containing CUI | CUI-marked exports require additional authorization. CUI export access logged. |
| MP.L2-3.8.6 | Implement cryptographic mechanisms to protect CUI during transport | CUI transmitted only over TLS 1.2+ with FIPS-approved cipher suites. CUI exports encrypted in transit and at rest. |
| MP.L2-3.8.7 | Control the use of removable media on system components | Removable media blocked on production systems. Employee devices managed via MDM with removable media policy. |
| MP.L2-3.8.8 | Prohibit the use of portable storage devices when such devices have no identifiable owner | Unidentified removable media not permitted on company systems. MDM enforces this policy. |
| MP.L2-3.8.9 | Protect the backup and restore of CUI | Backups encrypted at rest (AES-256, AWS KMS GovCloud). Backup access restricted to authorized personnel. Restore operations logged. |

### PE — Physical Protection (6 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| PE.L2-3.10.1 | Limit physical access to organizational systems to authorized individuals | ChaosPlane production systems hosted in AWS GovCloud data centers. Physical access controls inherited from AWS GovCloud FedRAMP High P-ATO. Office access controlled via key card. |
| PE.L2-3.10.2 | Protect and monitor the physical facility and support infrastructure | AWS GovCloud data center physical monitoring inherited from provider. Office CCTV in place. |
| PE.L2-3.10.3 | Escort visitors and monitor visitor activity | Office visitor log maintained. Visitors escorted at all times. AWS GovCloud data center visitor controls inherited from provider. |
| PE.L2-3.10.4 | Maintain audit logs of physical access | AWS GovCloud data center access logs maintained by AWS. Office access logs maintained by ChaosPlane. |
| PE.L2-3.10.5 | Control and manage physical access devices | Office key cards managed by facilities. Lost/stolen cards deactivated immediately. |
| PE.L2-3.10.6 | Enforce safeguarding measures for CUI at alternate work sites | Remote work policy covers CUI handling at alternate work sites. VPN required. Screen lock enforced. CUI not printed or stored locally. |

### PS — Personnel Security (2 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| PS.L2-3.9.1 | Screen individuals prior to authorizing access to organizational systems | Background checks conducted for all employees prior to start date. Enhanced background checks for staff with access to CUI or GovCloud systems. |
| PS.L2-3.9.2 | Ensure CUI is protected during and after personnel actions | Offboarding checklist: all access revoked within 24 hours. CUI access revoked immediately on termination notice. Confidentiality obligations survive termination. |

### RA — Risk Assessment (3 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| RA.L2-3.11.1 | Periodically assess the risk to organizational operations | Formal risk assessment conducted annually and after significant changes. CUI-specific risks assessed separately. |
| RA.L2-3.11.2 | Scan for vulnerabilities in organizational systems periodically and when new vulnerabilities are identified | Automated vulnerability scanning on every build (Trivy, Dependabot). Infrastructure scanning quarterly. Annual penetration test. |
| RA.L2-3.11.3 | Remediate vulnerabilities in accordance with risk assessments | Critical CVEs patched within 7 days. High within 30 days. Medium within 90 days. Patch status tracked in vulnerability management system. |

### CA — Security Assessment (4 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| CA.L2-3.12.1 | Periodically assess the security controls to determine if they are effective | Annual internal security control assessment. Third-party penetration test annually. CMMC C3PAO assessment planned for Phase 4. |
| CA.L2-3.12.2 | Develop and implement plans of action to correct deficiencies | POA&M maintained for all identified deficiencies. Reviewed monthly by Compliance Manager. Material deficiencies escalated to CISO. |
| CA.L2-3.12.3 | Monitor security controls on an ongoing basis | Continuous monitoring via compliance management tooling (Vanta / Drata). Automated evidence collection. Monthly ConMon report. |
| CA.L2-3.12.4 | Develop, document, and periodically update system security plans | System Security Plan (SSP) maintained and updated annually or upon significant change. FedRAMP SSP draft serves as the primary SSP for Government Edition. |

### SC — System and Communications Protection (16 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| SC.L2-3.13.1 | Monitor, control, and protect communications at external boundaries | WAF, ALB, and VPC security groups control external communications. VPC Flow Logs monitor all traffic. |
| SC.L2-3.13.2 | Employ architectural designs, software development techniques, and systems engineering principles | Defense in depth: WAF, network segmentation, RBAC/ABAC, encryption, audit logging. Secure SDLC with security gates. |
| SC.L2-3.13.3 | Separate user functionality from system management functionality | Admin functions accessible only via dedicated admin API endpoints with elevated RBAC roles. Admin UI separated from user UI. |
| SC.L2-3.13.4 | Prevent unauthorized and unintended information transfer | Tenant data isolated at database level. Cross-tenant queries architecturally prevented. Egress controls on production environment. |
| SC.L2-3.13.5 | Implement subnetworks for publicly accessible system components | Public subnet contains only ALB and WAF. Application and data tiers in private subnets. No direct internet access from private subnets. |
| SC.L2-3.13.6 | Deny network communications traffic by default | Security groups default-deny. Explicit allow rules required for all traffic. |
| SC.L2-3.13.7 | Prevent remote devices from simultaneously connecting to the system and other resources (split tunneling) | VPN policy prohibits split tunneling. All traffic routes through VPN when connected. |
| SC.L2-3.13.8 | Implement cryptographic mechanisms to prevent unauthorized disclosure of CUI during transmission | TLS 1.2+ with FIPS-approved cipher suites on all connections. mTLS for internal service communication. |
| SC.L2-3.13.9 | Terminate network connections after defined period of inactivity | Network connections terminated after 30 minutes of inactivity. TCP keepalive configured. |
| SC.L2-3.13.10 | Establish and manage cryptographic keys | Keys managed via AWS KMS GovCloud (FIPS 140-2 validated). Annual rotation. Immediate rotation on suspected compromise. |
| SC.L2-3.13.11 | Employ FIPS-validated cryptography | BoringCrypto (FIPS 140-2 cert #3678) used in all Go services. AWS KMS GovCloud (FIPS 140-2 validated) for key management. FIPS-only cipher suites enforced. |
| SC.L2-3.13.12 | Prohibit remote activation of collaborative computing devices | Collaborative computing devices (cameras, microphones) not present in production systems. Remote activation controls not applicable to server infrastructure. |
| SC.L2-3.13.13 | Control and monitor the use of mobile code | No mobile code (JavaScript) executed in server-side components. Client-side JavaScript served from CDN with Content Security Policy (CSP) headers. |
| SC.L2-3.13.14 | Control and monitor the use of VoIP technologies | VoIP not used in production systems. Corporate VoIP (if used) on isolated network segment. |
| SC.L2-3.13.15 | Protect the authenticity of communications sessions | Session tokens are cryptographically random (256-bit). CSRF protection on all state-changing endpoints. mTLS for service-to-service communication. |
| SC.L2-3.13.16 | Protect CUI at rest | CUI encrypted at rest using AES-256 (AWS KMS GovCloud, FIPS 140-2 validated). |

### SI — System and Information Integrity (7 practices)

| Practice | Requirement | ChaosPlane Implementation |
|---|---|---|
| SI.L2-3.14.1 | Identify, report, and correct information and information system flaws | CVE monitoring via Dependabot and Trivy. Critical CVEs patched within 7 days. Flaw remediation tracked in vulnerability management system. |
| SI.L2-3.14.2 | Provide protection from malicious code at appropriate locations | Container images scanned for malware before deployment (Trivy). Dependency scanning on every build. No user-uploaded executables accepted. |
| SI.L2-3.14.3 | Monitor system security alerts and advisories | CVE feeds monitored. AWS Security Hub alerts reviewed. Vendor security advisories tracked. |
| SI.L2-3.14.4 | Update malicious code protection mechanisms | Container image scanning updated continuously. Dependency scanning uses latest vulnerability databases. |
| SI.L2-3.14.5 | Perform periodic scans and real-time scans of files from external sources | Container images scanned before deployment. Dependencies scanned on every build. External data imports validated and sanitized. |
| SI.L2-3.14.6 | Monitor organizational systems to detect attacks and indicators of compromise | Infrastructure and application metrics monitored 24/7. Anomaly detection alerts on-call engineer. IDS/IPS enabled on GovCloud network. |
| SI.L2-3.14.7 | Identify unauthorized use of organizational systems | Audit logs reviewed for anomalous access patterns. Automated alerting on suspicious activity. SIEM integration for Government customers. |

---

## Practice Coverage Summary

| Domain | Practices | Implemented | In Progress | Not Applicable |
|---|:---:|:---:|:---:|:---:|
| AC — Access Control | 22 | 20 | 2 | 0 |
| AT — Awareness and Training | 3 | 3 | 0 | 0 |
| AU — Audit and Accountability | 9 | 9 | 0 | 0 |
| CM — Configuration Management | 9 | 9 | 0 | 0 |
| IA — Identification and Authentication | 11 | 11 | 0 | 0 |
| IR — Incident Response | 3 | 3 | 0 | 0 |
| MA — Maintenance | 6 | 5 | 0 | 1 |
| MP — Media Protection | 9 | 9 | 0 | 0 |
| PE — Physical Protection | 6 | 6 | 0 | 0 |
| PS — Personnel Security | 2 | 2 | 0 | 0 |
| RA — Risk Assessment | 3 | 3 | 0 | 0 |
| CA — Security Assessment | 4 | 3 | 1 | 0 |
| SC — System and Communications Protection | 16 | 15 | 1 | 0 |
| SI — System and Information Integrity | 7 | 7 | 0 | 0 |
| **Total** | **110** | **105** | **4** | **1** |

Not Applicable: MA.L2-3.7.3 (equipment sanitization — cloud-hosted, inherited from AWS GovCloud).

---

## C3PAO Assessment Timeline (Planned)

| Phase | Activity | Target |
|---|---|---|
| Self-assessment (SPRS) | Complete NIST SP 800-171 self-assessment, submit score to SPRS | Month 19 |
| C3PAO selection | Engage a CMMC Third Party Assessment Organization | Month 20 |
| Assessment preparation | Evidence package assembly, gap remediation | Month 21-22 |
| C3PAO assessment | On-site / remote assessment | Month 22-23 |
| CMMC Level 2 certification | Certificate issued | Month 23-24 |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for C3PAO assessment |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | October 2026 |
| Standard reference | CMMC Level 2, NIST SP 800-171 Rev 2, DFARS 252.204-7012 |
