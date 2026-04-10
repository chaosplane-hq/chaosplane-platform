# ISO 27001 Control Mapping — ChaosPlane

> ChaosPlane is preparing for ISO 27001 certification. This document maps ChaosPlane platform features and operational controls to the ISO/IEC 27001:2022 Annex A control set. It is intended for use by auditors, enterprise customers conducting vendor assessments, and the internal compliance team.
>
> This document describes controls that are implemented or in active preparation. It does not constitute a certification claim.

---

## Scope

The scope of ChaosPlane's ISO 27001 preparation covers:

- The ChaosPlane SaaS platform (API, web application, agent infrastructure)
- Supporting cloud infrastructure (AWS, primary region)
- Internal development and operations processes
- Customer data processed within the platform (experiment definitions, topology data, audit logs, user accounts)

---

## Annex A Control Mapping

### A.5 — Organizational Controls

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| A.5.1 | Policies for information security | Information security policy documented and reviewed annually. Covers acceptable use, data classification, incident response, and access control. |
| A.5.2 | Information security roles and responsibilities | Security responsibilities assigned to engineering leads, a designated Compliance Manager, and a named security contact for Government tier customers. |
| A.5.3 | Segregation of duties | Production deployments require a second approver. Billing and infrastructure access are separated by role. |
| A.5.4 | Management responsibilities | Engineering management reviews security posture quarterly. Compliance findings are escalated to executive leadership. |
| A.5.5 | Contact with authorities | Incident response runbook includes contact procedures for relevant authorities (data protection regulators, CERT). |
| A.5.6 | Contact with special interest groups | ChaosPlane participates in CNCF security working groups and monitors CVE feeds for dependencies. |
| A.5.7 | Threat intelligence | Dependency vulnerability scanning (Dependabot, Trivy) runs on every build. Threat intelligence feeds inform quarterly risk reviews. |
| A.5.8 | Information security in project management | Security requirements are part of the feature specification template. Security review is a gate in the release process. |
| A.5.9 | Inventory of information and other associated assets | Asset inventory maintained for cloud infrastructure, data stores, and third-party services. Reviewed quarterly. |
| A.5.10 | Acceptable use of information and other associated assets | Acceptable use policy covers employee devices, cloud resources, and customer data handling. |
| A.5.11 | Return of assets | Offboarding checklist requires return or secure wipe of all company devices and revocation of all access within 24 hours. |
| A.5.12 | Classification of information | Data classified as Public, Internal, Confidential, or Restricted. Customer experiment data and audit logs are Confidential. |
| A.5.13 | Labelling of information | Data classification labels applied in internal documentation and data handling procedures. |
| A.5.14 | Information transfer | Customer data transfer uses TLS 1.2+ in transit. Audit log exports to S3/MinIO use server-side encryption. |
| A.5.15 | Access control | RBAC enforced across all platform resources. Enterprise tier adds ABAC for attribute-based fine-grained policies. |
| A.5.16 | Identity management | User identities managed via ChaosPlane's identity service. SSO/SAML integration delegates identity to customer IdPs (Okta, Azure AD, Google Workspace). SCIM handles provisioning and deprovisioning. |
| A.5.17 | Authentication information | Passwords hashed with bcrypt (cost factor 12+). MFA enforced for all staff. Enterprise customers can enforce MFA org-wide. Recovery codes are SHA-256 hashed, 10 per user, with usage tracking. |
| A.5.18 | Access rights | Access rights reviewed quarterly. Principle of least privilege applied. Offboarding triggers immediate access revocation. |
| A.5.19 | Information security in supplier relationships | Third-party vendors assessed for security posture before onboarding. DPAs in place with all data processors. |
| A.5.20 | Addressing information security within supplier agreements | Supplier contracts include security requirements, breach notification obligations, and audit rights. |
| A.5.21 | Managing information security in the ICT supply chain | Software dependencies tracked via SBOM. Critical dependencies reviewed for known vulnerabilities on each release. |
| A.5.22 | Monitoring, review and change management of supplier services | Annual review of critical suppliers (cloud provider, LLM API, payment processor). |
| A.5.23 | Information security for use of cloud services | AWS infrastructure governed by ChaosPlane's cloud security policy. IAM roles follow least privilege. S3 buckets are private by default. |
| A.5.24 | Information security incident management planning | Incident response plan documented with severity classification (P0–P3), escalation paths, and post-incident review process. |
| A.5.25 | Assessment and decision on information security events | Security events triaged by on-call engineer. P0/P1 events escalate to security lead within 1 hour. |
| A.5.26 | Response to information security incidents | Incident response runbook covers containment, eradication, recovery, and customer notification. |
| A.5.27 | Learning from information security incidents | Post-incident reviews produce action items tracked to closure. Trends reviewed quarterly. |
| A.5.28 | Collection of evidence | Audit logs retained for compliance purposes. Log integrity protected via append-only storage and tamper-evident signing (Government tier). |
| A.5.29 | Information security during disruption | Business continuity plan covers platform recovery. RTO and RPO targets defined per tier. |
| A.5.30 | ICT readiness for business continuity | Automated backups, multi-AZ deployment, and runbooks for common failure scenarios. Tested annually. |
| A.5.31 | Legal, statutory, regulatory and contractual requirements | Legal register maintained. GDPR, applicable data protection laws, and contractual obligations tracked. |
| A.5.32 | Intellectual property rights | Open source license compliance tracked via FOSSA. Internal IP policy covers employee contributions. |
| A.5.33 | Protection of records | Audit logs and compliance records retained per policy (minimum 3 years). Deletion follows documented retention schedule. |
| A.5.34 | Privacy and protection of PII | Privacy policy published. GDPR Right to Erasure automated via 30-day grace period pipeline: request → anonymization (email → hash, name → "Deleted User") → cascade soft delete. |
| A.5.35 | Independent review of information security | Annual third-party penetration test. Results reviewed by engineering leadership and tracked to remediation. |
| A.5.36 | Compliance with policies, rules and standards | Internal compliance reviews conducted quarterly. Deviations tracked and risk-accepted or remediated. |
| A.5.37 | Documented operating procedures | Runbooks maintained for deployment, incident response, backup/restore, and key rotation. Stored in internal wiki with version history. |

---

### A.6 — People Controls

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| A.6.1 | Screening | Background checks conducted for all employees prior to start date, consistent with local law. |
| A.6.2 | Terms and conditions of employment | Employment contracts include confidentiality obligations, acceptable use policy acknowledgment, and IP assignment. |
| A.6.3 | Information security awareness, education and training | Security awareness training required at onboarding and annually. Phishing simulation run twice per year. |
| A.6.4 | Disciplinary process | Documented disciplinary process for security policy violations, up to and including termination. |
| A.6.5 | Responsibilities after termination or change of employment | Offboarding checklist: access revoked within 24 hours, devices returned, confidentiality obligations survive termination. |
| A.6.6 | Confidentiality or non-disclosure agreements | NDAs signed by all employees, contractors, and relevant third parties before access to confidential information. |
| A.6.7 | Remote working | Remote work policy covers device security, VPN use, and screen lock requirements. |
| A.6.8 | Information security event reporting | All staff trained to report suspected security events to the security team via a dedicated channel. |

---

### A.7 — Physical Controls

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| A.7.1 | Physical security perimeters | ChaosPlane operates as a cloud-native SaaS. Physical infrastructure is hosted on AWS, which maintains physical security controls documented in AWS's ISO 27001 certification. |
| A.7.2 | Physical entry | Office access controlled via key card. Visitor log maintained. Server rooms not operated by ChaosPlane directly. |
| A.7.3 | Securing offices, rooms and facilities | Office security policy covers clean desk, screen lock, and visitor escort requirements. |
| A.7.4 | Physical security monitoring | Office CCTV in place. AWS data center physical monitoring inherited from AWS controls. |
| A.7.5 | Protecting against physical and environmental threats | AWS data centers provide environmental controls (fire suppression, climate control, power redundancy). |
| A.7.6 | Working in secure areas | Sensitive discussions and work involving customer data conducted in private settings. |
| A.7.7 | Clear desk and clear screen | Clear desk and screen lock policy enforced. Automatic screen lock after 5 minutes of inactivity. |
| A.7.8 | Equipment siting and protection | Employee laptops encrypted (FileVault / BitLocker). MDM enforced for all company devices. |
| A.7.9 | Security of assets off-premises | Remote work policy covers device security for off-premises use. |
| A.7.10 | Storage media | Sensitive data not stored on removable media. Cloud storage used exclusively. |
| A.7.11 | Supporting utilities | AWS provides power redundancy and UPS for hosted infrastructure. |
| A.7.12 | Cabling security | Not applicable (cloud-hosted infrastructure). |
| A.7.13 | Equipment maintenance | Employee device maintenance tracked via MDM. Cloud infrastructure maintenance managed by AWS. |
| A.7.14 | Secure disposal or re-use of equipment | Device wipe procedure documented and executed at offboarding. AWS handles secure disposal of physical media. |

---

### A.8 — Technological Controls

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| A.8.1 | User endpoint devices | MDM (Mobile Device Management) enforced on all company devices. Full-disk encryption required. Jailbroken/rooted devices blocked. |
| A.8.2 | Privileged access rights | Privileged access (AWS root, production database) restricted to named individuals. MFA required. Access reviewed quarterly. |
| A.8.3 | Information access restriction | API access controlled by RBAC/ABAC. Tenant data isolated at the database level. Cross-tenant access is architecturally prevented. |
| A.8.4 | Access to source code | Source code access restricted to engineering staff. Branch protection rules prevent direct pushes to main. All changes require peer review. |
| A.8.5 | Secure authentication | All authentication uses TLS. Passwords hashed with bcrypt. MFA available on all tiers, enforceable on Enterprise. Session tokens rotated on privilege escalation. WebSocket sessions re-authenticated every 5 minutes. |
| A.8.6 | Capacity management | Infrastructure auto-scales via AWS EKS. Capacity metrics monitored with alerts for sustained high utilization. |
| A.8.7 | Protection against malware | Dependency scanning on every build (Trivy, Dependabot). Container images scanned before deployment. No executable uploads accepted from users. |
| A.8.8 | Management of technical vulnerabilities | CVE monitoring via Dependabot and Trivy. Critical vulnerabilities patched within 7 days. High within 30 days. |
| A.8.9 | Configuration management | Infrastructure defined as code (Terraform). Configuration changes go through pull request review and CI validation. Drift detection alerts on manual changes. |
| A.8.10 | Information deletion | Customer data deletion automated via 30-day grace period pipeline. Deleted data anonymized and cascade-soft-deleted. Hard deletion from backups within 90 days. |
| A.8.11 | Data masking | PII masked in logs. Experiment payloads do not log sensitive field values. |
| A.8.12 | Data leakage prevention | Egress controls on production environment. Audit log export requires explicit authorization. SIEM integration uses read-only credentials. |
| A.8.13 | Information backup | Automated daily backups of all databases. Backups encrypted at rest. Restore tested quarterly. |
| A.8.14 | Redundancy of information processing facilities | Multi-AZ deployment on AWS. Automatic failover for database and API tiers. |
| A.8.15 | Logging | All API calls logged with actor, timestamp, IP, resource, and outcome. Audit logs are append-only. Log retention: 90 days hot, 3 years cold (S3). |
| A.8.16 | Monitoring activities | Infrastructure and application metrics monitored 24/7. Anomaly detection alerts on-call engineer. SIEM integration available for Enterprise customers. |
| A.8.17 | Clock synchronization | All systems synchronized to AWS Time Sync Service (NTP). Timestamps in audit logs are UTC. |
| A.8.18 | Use of privileged utility programs | Privileged tooling (kubectl exec, database console) access logged and requires MFA. Usage reviewed monthly. |
| A.8.19 | Installation of software on operational systems | Production deployments via CI/CD pipeline only. No manual software installation on production systems. |
| A.8.20 | Networks security | VPC with private subnets for internal services. Public-facing services behind WAF and load balancer. Network segmentation between tenant workloads. |
| A.8.21 | Security of network services | TLS 1.2+ enforced on all external endpoints. Internal service communication uses mTLS. |
| A.8.22 | Segregation of networks | Production, staging, and development environments in separate VPCs. No direct routing between environments. |
| A.8.23 | Web filtering | Outbound web filtering on corporate devices via DNS-based filtering. |
| A.8.24 | Use of cryptography | Encryption at rest: AES-256 (AWS KMS). Encryption in transit: TLS 1.2+. Key rotation: annual for KMS keys, immediate on suspected compromise. |
| A.8.25 | Secure development lifecycle | Security requirements in feature specs. SAST in CI pipeline. Dependency scanning on every build. Security review gate before GA releases. |
| A.8.26 | Application security requirements | Threat modeling for new features. OWASP Top 10 addressed in secure coding guidelines. |
| A.8.27 | Secure system architecture and engineering principles | Defense in depth: WAF, network segmentation, RBAC/ABAC, audit logging, encryption at rest and in transit. |
| A.8.28 | Secure coding | Secure coding guidelines published internally. Code review required for all changes. SAST (static analysis) runs in CI. |
| A.8.29 | Security testing in development and acceptance | Automated security tests in CI. Annual third-party penetration test. Critical findings block release. |
| A.8.30 | Outsourced development | Third-party development subject to same security requirements as internal development. Code reviewed before merge. |
| A.8.31 | Separation of development, test and production environments | Separate AWS accounts for production, staging, and development. Production credentials not accessible in lower environments. |
| A.8.32 | Change management | All changes via pull request with peer review. Production deployments require approval. Change log maintained. |
| A.8.33 | Test information | Production data not used in test environments. Synthetic data used for testing. |
| A.8.34 | Protection of information systems during audit testing | Penetration tests conducted in a dedicated test environment or with explicit production safeguards. Customer data not accessed during testing. |

---

## Control Coverage Summary

| Annex A Section | Total Controls | Implemented | In Progress | Not Applicable |
|---|:---:|:---:|:---:|:---:|
| A.5 Organizational | 37 | 35 | 2 | 0 |
| A.6 People | 8 | 8 | 0 | 0 |
| A.7 Physical | 14 | 12 | 0 | 2 |
| A.8 Technological | 34 | 32 | 2 | 0 |
| **Total** | **93** | **87** | **4** | **2** |

Not Applicable controls: A.7.12 (cabling security — cloud-hosted), A.7.11 (supporting utilities — inherited from AWS).

---

## Residual Risks and Open Items

| Item | Status | Target Date |
|---|---|---|
| Third-party penetration test (annual) | Scheduled | Q2 2026 |
| Compliance Manager hire | In progress | Q1 2026 |
| Automated compliance tooling (Vanta / Drata) | Evaluating | Q2 2026 |
| ISMS-P certification preparation | Planned | Phase 4 |
| FedRAMP authorization | Not in scope (Phase 4) | TBD |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for certification |
| Owner | Compliance Manager |
| Last reviewed | April 2026 |
| Next review | October 2026 |
| Standard reference | ISO/IEC 27001:2022 |
