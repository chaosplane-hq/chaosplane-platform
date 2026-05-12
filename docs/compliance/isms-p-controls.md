# ISMS-P Control Mapping — ChaosPlane

> ChaosPlane is preparing for ISMS-P (정보보호 및 개인정보보호 관리체계) certification. This document maps ChaosPlane platform features and operational controls to the ISMS-P control framework administered by KISA (한국인터넷진흥원). It is intended for use by the KISA audit team, Korean enterprise customers conducting vendor assessments, and the internal compliance team.
>
> ISMS-P certification covers both information security management (ISMS) and personal information protection management (PIMS). This document addresses both domains.
>
> This document describes controls that are implemented or in active preparation. It does not constitute a certification claim.

---

## Scope

The scope of ChaosPlane's ISMS-P preparation covers:

- The ChaosPlane SaaS platform (API, web application, chaos agent infrastructure)
- Supporting cloud infrastructure (AWS primary region; NHN Cloud / NCP for Korea-region CSAP deployment)
- Internal development and operations processes
- Personal information processed within the platform (customer accounts, usage data, audit logs)
- Korean customer data subject to the Personal Information Protection Act (PIPA, 개인정보보호법)

---

## ISMS-P Framework Overview

The ISMS-P framework consists of three domains:

| Domain | Korean | Controls |
|---|---|---|
| 1. Management System | 관리체계 수립 및 운영 | 16 controls |
| 2. Protection Measures | 보호대책 요구사항 | 64 controls |
| 3. Personal Information Protection | 개인정보 처리 단계별 요구사항 | 22 controls |

**Total: 102 controls**

---

## Domain 1 — Management System (관리체계 수립 및 운영)

### 1.1 Management System Establishment (관리체계 기반 마련)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 1.1.1 | 경영진의 참여 (Executive involvement) | Executive leadership reviews security posture and compliance status quarterly. CISO role established in Phase 4 (Month 14). Material security incidents escalated to executive leadership immediately. |
| 1.1.2 | 최고책임자의 지정 (CISO designation) | CISO designated as of Phase 4 Month 14. Responsible for overall information security strategy, ISMS-P certification, and regulatory compliance. |
| 1.1.3 | 조직 구성 (Organization structure) | Security team structure: CISO, Security Engineers (1-2), Compliance Manager, DPO (Month 18). Roles and responsibilities documented. |
| 1.1.4 | 범위 설정 (Scope definition) | ISMS-P scope formally defined to cover ChaosPlane SaaS platform, supporting infrastructure, and all personal information processing activities. Scope document reviewed annually. |
| 1.1.5 | 정책 수립 (Policy establishment) | Information security policy and personal information protection policy documented. Reviewed annually. Covers acceptable use, data classification, incident response, access control, and personal information handling. |
| 1.1.6 | 자원 할당 (Resource allocation) | Security budget allocated annually. Headcount plan includes dedicated security and compliance roles. Tool budget covers compliance automation, penetration testing, and audit tooling. |

### 1.2 Risk Management (위험 관리)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 1.2.1 | 정보자산 식별 (Asset identification) | Asset inventory maintained for cloud infrastructure, data stores, third-party services, and personal information processing systems. Reviewed quarterly. |
| 1.2.2 | 현황 및 흐름 분석 (Current state and flow analysis) | Data flow diagrams maintained for all personal information processing activities. Updated when new features are introduced. |
| 1.2.3 | 위험 평가 (Risk assessment) | Formal risk assessment conducted annually and after significant changes. Risks documented with likelihood, impact, owner, and treatment plan. |
| 1.2.4 | 보호대책 선정 (Control selection) | Controls selected based on risk assessment results. Defense-in-depth approach applied. Control gaps tracked to remediation. |
| 1.2.5 | 정보보호 계획 수립 (Security plan) | Annual information security plan documents objectives, controls, resource requirements, and timelines. Reviewed by CISO and executive leadership. |

### 1.3 Management System Operation (관리체계 운영)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 1.3.1 | 보호대책 구현 (Control implementation) | Controls implemented per the annual security plan. Implementation status tracked in compliance management tooling. |
| 1.3.2 | 보호대책 공유 (Control communication) | Security policies and procedures published in internal wiki. Changes communicated to affected staff. New employee security orientation at onboarding. |
| 1.3.3 | 운영현황 관리 (Operational status management) | Monthly security metrics reviewed by CISO. Quarterly compliance review by executive leadership. |

### 1.4 Management System Improvement (관리체계 개선)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 1.4.1 | 법적 요구사항 준수 검토 (Legal compliance review) | Legal register maintained covering PIPA, GDPR, and applicable sector regulations. Reviewed quarterly. Legal counsel engaged for material changes. |
| 1.4.2 | 관리체계 점검 (Management system review) | Annual internal audit of ISMS-P controls. Findings tracked to remediation. Results reported to executive leadership. |
| 1.4.3 | 관리체계 개선 (Management system improvement) | Audit findings, incident post-mortems, and risk assessment results feed into the annual security plan update. Continuous improvement process documented. |

---

## Domain 2 — Protection Measures (보호대책 요구사항)

### 2.1 Human Security (인적 보안)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.1.1 | 주요 직무자 지정 및 관리 (Key role designation) | Key roles with elevated access (CISO, Security Engineers, DBA, SRE) formally designated. Access reviewed quarterly. |
| 2.1.2 | 직무 분리 (Segregation of duties) | Production deployments require a second approver. Billing and infrastructure access separated by role. No single person can approve and deploy their own changes. |
| 2.1.3 | 보안 서약 (Security pledge) | All employees sign confidentiality and security obligations at onboarding. Contractors sign NDA before access to any systems. |
| 2.1.4 | 인식 제고 및 교육 훈련 (Awareness and training) | Security awareness training required at onboarding and annually. Role-specific training for engineers (secure coding), operations (incident response), and compliance staff. Phishing simulation twice per year. |
| 2.1.5 | 퇴직 및 직무 변경 관리 (Termination and role change) | Offboarding checklist: all access revoked within 24 hours, devices returned, confidentiality obligations survive termination. Role changes trigger access review within 5 business days. |
| 2.1.6 | 보안 위반 시 조치 (Security violation response) | Documented disciplinary process for security policy violations, up to and including termination. Violations reported to CISO. |

### 2.2 Physical and Environmental Security (물리 보안)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.2.1 | 물리 보안 구역 설정 (Physical security zones) | ChaosPlane operates as cloud-native SaaS. Physical infrastructure hosted on AWS (and NHN Cloud / NCP for Korea region), which maintain physical security controls documented in their own certifications. |
| 2.2.2 | 출입통제 (Access control) | Office access controlled via key card. Visitor log maintained. AWS/NHN Cloud data center physical access inherited from provider controls. |
| 2.2.3 | 정보시스템 보호 (Information system protection) | Employee laptops encrypted (FileVault / BitLocker). MDM enforced for all company devices. Automatic screen lock after 5 minutes. |
| 2.2.4 | 보호구역 내 작업 (Work in secure areas) | Sensitive work involving customer data conducted in private settings. Clean desk policy enforced. |
| 2.2.5 | 보호구역 내 반출입 통제 (Asset movement control) | Sensitive data not stored on removable media. Cloud storage used exclusively. Device removal tracked via MDM. |
| 2.2.6 | 업무환경 보안 (Work environment security) | Remote work policy covers device security, VPN use, and screen lock requirements. |

### 2.3 External Party Security (외부자 보안)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.3.1 | 외부자 현황 관리 (External party management) | Register of third-party vendors and subprocessors maintained. Reviewed quarterly. |
| 2.3.2 | 외부자 계약 시 보안 (Security in external contracts) | Supplier contracts include security requirements, breach notification obligations, and audit rights. DPAs in place with all data processors. |
| 2.3.3 | 외부자 보안 이행 관리 (External party security monitoring) | Annual review of critical suppliers. Security questionnaires sent to key vendors. |
| 2.3.4 | 외부자 계약 변경 및 만료 시 보안 (Contract change/expiry) | Contract termination triggers data return or deletion per DPA terms. Access revoked immediately upon contract end. |

### 2.4 Physical and Logical Asset Management (물리 및 논리적 자산 관리)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.4.1 | 정보자산 분류 (Asset classification) | Data classified as Public, Internal, Confidential, or Restricted. Customer experiment data and audit logs are Confidential. Personal information is Restricted. |
| 2.4.2 | 정보자산 취급 및 운반 (Asset handling) | Data handling procedures documented per classification level. Confidential and Restricted data requires encryption in transit and at rest. |
| 2.4.3 | 정보자산 반납 및 폐기 (Asset return and disposal) | Device wipe procedure executed at offboarding. AWS/NHN Cloud handle secure disposal of physical media. Data deletion follows documented retention schedule. |

### 2.5 Authentication and Authorization (인증 및 권한 관리)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.5.1 | 사용자 계정 관리 (User account management) | User accounts created via self-service registration (email verification required) or SCIM provisioning. Employee access provisioned via access request with manager approval. |
| 2.5.2 | 사용자 식별 (User identification) | Each user has a unique identifier (UUID). Shared accounts not permitted. Service accounts have dedicated identities. |
| 2.5.3 | 사용자 인증 (User authentication) | Passwords hashed with bcrypt (cost factor 12+). MFA available on all tiers, enforceable org-wide on Enterprise. SSO/SAML delegates authentication to customer IdPs. |
| 2.5.4 | 비밀번호 관리 (Password management) | Password policy: minimum 12 characters, complexity requirements. Passwords hashed with bcrypt. Breach password detection via HaveIBeenPwned API at registration and login. |
| 2.5.5 | 특수 계정 및 권한 관리 (Privileged account management) | Privileged access (AWS root, production database, kubectl) restricted to named individuals. MFA required. Access reviewed quarterly. Usage logged. |
| 2.5.6 | 접근권한 검토 (Access rights review) | Access rights reviewed quarterly. Principle of least privilege applied. Offboarding triggers immediate access revocation. |

### 2.6 Access Control (접근 통제)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.6.1 | 네트워크 접근 (Network access) | VPC with private subnets for internal services. Public-facing services behind WAF and load balancer. Network segmentation between tenant workloads. |
| 2.6.2 | 정보시스템 접근 (Information system access) | RBAC enforced across all platform resources. Enterprise tier adds ABAC for fine-grained attribute-based policies. Tenant data isolated at database level. |
| 2.6.3 | 응용프로그램 접근 (Application access) | API access controlled by RBAC/ABAC. Rate limiting on all endpoints. Bot detection on authentication endpoints. |
| 2.6.4 | 데이터베이스 접근 (Database access) | Database access restricted to application service accounts and named DBAs. Direct database access requires MFA and is logged. No direct database access from application servers in production. |
| 2.6.5 | 무선 네트워크 접근 (Wireless network access) | Office wireless network uses WPA3. Guest network isolated from corporate network. Remote work uses VPN. |
| 2.6.6 | 원격 접근 (Remote access) | Remote access to production infrastructure via VPN with MFA. SSH access to production servers requires key-based authentication and is logged. |

### 2.7 Cryptography Application (암호화 적용)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.7.1 | 암호정책 적용 (Cryptography policy) | Cryptography policy documents approved algorithms, key lengths, and key management procedures. AES-256 for data at rest, TLS 1.2+ for data in transit. |
| 2.7.2 | 암호키 관리 (Key management) | Encryption keys managed via AWS KMS. Key rotation: annual for KMS keys, immediate on suspected compromise. Key access restricted to authorized services. |

### 2.8 Information System Security (정보시스템 도입 및 개발 보안)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.8.1 | 보안 요구사항 정의 (Security requirements) | Security requirements included in feature specification template. Threat modeling for new features. OWASP Top 10 addressed in secure coding guidelines. |
| 2.8.2 | 보안 요구사항 검토 및 시험 (Security testing) | SAST in CI pipeline. Dependency scanning on every build (Trivy, Dependabot). Annual third-party penetration test. Critical findings block release. |
| 2.8.3 | 시험과 운영 환경 분리 (Environment separation) | Separate AWS accounts for production, staging, and development. Production credentials not accessible in lower environments. Production data not used in test environments. |
| 2.8.4 | 시험 데이터 보안 (Test data security) | Synthetic data used for testing. Production data not copied to test environments. |
| 2.8.5 | 소스 프로그램 관리 (Source code management) | Source code access restricted to engineering staff. Branch protection rules prevent direct pushes to main. All changes require peer review. |
| 2.8.6 | 운영환경 이관 (Production deployment) | Production deployments via CI/CD pipeline only. Deployment requires approval. Change log maintained. |

### 2.9 System and Service Operation Security (시스템 및 서비스 운영 관리)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.9.1 | 변경관리 (Change management) | All changes via pull request with peer review. Security review required for significant changes. Production deployments require approval. |
| 2.9.2 | 성능 및 장애관리 (Performance and incident management) | Infrastructure and application metrics monitored 24/7. Anomaly detection alerts on-call engineer. Incident response runbook covers P0-P3 severity levels. |
| 2.9.3 | 백업 및 복구관리 (Backup and recovery) | Automated daily backups of all databases. Backups encrypted at rest. Restore tested quarterly. RTO: 4 hours (Enterprise). RPO: 1 hour. |
| 2.9.4 | 로그 및 접속기록 관리 (Log and access record management) | All API calls logged with actor, timestamp, IP, resource, and outcome. Audit logs are append-only. Log retention: 90 days hot, 3 years cold (S3). |
| 2.9.5 | 시간 동기화 (Time synchronization) | All systems synchronized to AWS Time Sync Service (NTP). Timestamps in audit logs are UTC. |
| 2.9.6 | 정보시스템 도입 및 유지보수 보안 (System maintenance security) | Infrastructure defined as code (Terraform). Configuration changes go through pull request review and CI validation. Drift detection alerts on manual changes. |
| 2.9.7 | 공개서버 보안 (Public server security) | WAF in front of all public endpoints. DDoS protection via AWS Shield. Security headers enforced (HSTS, CSP, X-Frame-Options). |

### 2.10 Incident Management (침해사고 관리)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.10.1 | 보안 침해사고 대응 절차 수립 (Incident response procedures) | Incident response plan documented with severity classification (P0-P3), escalation paths, containment, eradication, recovery, and post-incident review. |
| 2.10.2 | 취약점 점검 및 조치 (Vulnerability management) | CVE monitoring via Dependabot and Trivy. Critical vulnerabilities patched within 7 days. High within 30 days. Annual third-party penetration test. |
| 2.10.3 | 이상행위 분석 및 모니터링 (Anomaly detection) | SIEM integration available for Enterprise customers. Anomaly detection alerts on-call engineer. Suspicious login alerts (new device, new location). |
| 2.10.4 | 사고 대응 및 복구 (Incident response and recovery) | Incident response runbook covers containment, eradication, recovery, and customer notification. Post-incident review required for P0/P1 events. |
| 2.10.5 | 사고 대응 결과 보고 (Incident reporting) | Post-incident reports produced for P0/P1 events. Trends reviewed quarterly. Material incidents reported to KISA within 24 hours per PIPA requirements. |
| 2.10.6 | 재해 복구 (Disaster recovery) | Business continuity plan covers platform recovery. Multi-AZ deployment. Automated failover. DR tested annually. |

### 2.11 Business Continuity (IT 재해 복구)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.11.1 | 연속성 계획 수립 (Continuity planning) | Business continuity plan documented. RTO and RPO targets defined per tier. Plan reviewed annually. |
| 2.11.2 | 연속성 계획 시험 및 유지관리 (Continuity testing) | DR test conducted annually. Results documented. Gaps remediated before next test cycle. |

### 2.12 Compliance (준거성)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 2.12.1 | 개인정보 및 저작권 보호 (Personal information and copyright) | PIPA compliance framework implemented. Open source license compliance tracked via FOSSA. |
| 2.12.2 | 정보보호 관련 법규 준수 검토 (Legal compliance review) | Legal register maintained. Quarterly review of applicable laws and regulations. Legal counsel engaged for material changes. |

---

## Domain 3 — Personal Information Protection (개인정보 처리 단계별 요구사항)

### 3.1 Personal Information Collection (개인정보 수집 시 보호조치)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 3.1.1 | 개인정보 수집 제한 (Collection limitation) | Personal data collected is limited to what is necessary for service delivery: name, email, authentication credentials, usage data, and billing information. Minimization principle applied to all new features. |
| 3.1.2 | 개인정보의 수집 동의 (Consent for collection) | Users provide explicit consent at registration. Consent records maintained with timestamp, version, and channel. Marketing consent is separate and opt-in only. |
| 3.1.3 | 주민등록번호 처리 제한 (Resident registration number restriction) | ChaosPlane does not collect Korean resident registration numbers (주민등록번호). Identity verification uses email and optional phone number only. |
| 3.1.4 | 민감정보 처리 제한 (Sensitive information restriction) | ChaosPlane does not process sensitive personal information (민감정보) as defined under PIPA. If sensitive information is incidentally included in experiment topology data, it is not stored. |
| 3.1.5 | 간접수집 보호조치 (Indirect collection protection) | Personal information collected indirectly (e.g., via SCIM from customer IdP) is handled under the same protection measures as directly collected data. |
| 3.1.6 | 영상정보처리기기 설치 및 운영 (CCTV operation) | Office CCTV operated per applicable law. Signage posted. Footage retained for 30 days. Access restricted to authorized personnel. |

### 3.2 Personal Information Use and Provision (개인정보 보유 및 이용 시 보호조치)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 3.2.1 | 개인정보 현황 관리 (Personal information status management) | Personal information inventory maintained as part of the RoPA. Updated when new processing activities are introduced. |
| 3.2.2 | 개인정보 품질 보장 (Data quality) | Users can update their personal data via account settings. Email changes require re-verification. |
| 3.2.3 | 개인정보 표시 제한 및 이용 시 보호조치 (Display restriction) | PII masked in logs and internal tooling. Full personal data accessible only to authorized roles. |
| 3.2.4 | 이용자 단말기에 대한 접근 보호 (User device access protection) | ChaosPlane does not access user devices beyond what is necessary for service delivery. Cookie policy documented. |
| 3.2.5 | 개인정보 목적 외 이용 및 제공 제한 (Purpose limitation) | Personal data used only for service delivery, support, and billing. Not sold or shared with third parties for marketing. |
| 3.2.6 | 개인정보 제3자 제공 시 보호조치 (Third-party provision) | Personal data shared with third parties only as required for service delivery (payment processor, email provider) or by law. DPAs in place with all recipients. |

### 3.3 Personal Information Provision and Entrustment (개인정보 제공 시 보호조치)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 3.3.1 | 개인정보 처리 업무 위탁 (Processing entrustment) | Personal information processing entrusted to subprocessors (AWS, Stripe, SendGrid) under written agreements. Subprocessor list published and updated with 30 days' notice. |
| 3.3.2 | 영업의 양도 등에 따른 개인정보 이전 (Transfer on business transfer) | In the event of a business transfer, personal information transfer will be handled per PIPA requirements, including notification to data subjects. |
| 3.3.3 | 개인정보의 국외 이전 (Cross-border transfer) | Personal data transferred outside Korea (to AWS us-east-1) under standard contractual clauses and with data subject consent where required. Korean customers can request Korea-region deployment (NHN Cloud / NCP) to keep data in-country. |

### 3.4 Personal Information Destruction (개인정보 파기 시 보호조치)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 3.4.1 | 개인정보의 파기 (Personal information destruction) | Account deletion pipeline: 30-day grace period, then anonymization (email hashed, name replaced with "Deleted User"), cascade soft delete, hard deletion from backups within 90 days. |
| 3.4.2 | 처리목적 달성 후 보유 시 조치 (Retention after purpose) | Data retained beyond its primary purpose only where required by law (billing records 7 years, consent records 3 years). Retained data access restricted. |

### 3.5 Data Subject Rights (정보주체 권리 보호)

| Control | Title | ChaosPlane Implementation |
|---|---|---|
| 3.5.1 | 개인정보 처리방침 공개 (Privacy policy publication) | Privacy policy published at chaosplane.dev/privacy. Korean-language version available. Reviewed annually. |
| 3.5.2 | 정보주체 권리보장 (Data subject rights) | Rights supported: access, correction, deletion, processing suspension. Requests handled within 10 business days per PIPA. Automated via account settings where possible. |
| 3.5.3 | 이용자의 권리 보장 (User rights) | Users can access, export, correct, and delete their personal data via account settings. Written requests handled by DPO / Compliance Manager. |
| 3.5.4 | 개인정보 유출 통지 (Breach notification) | Personal information breach notification to KISA within 24 hours of discovery (per PIPA Article 34). Notification to affected individuals without undue delay. |

---

## KISA Audit Preparation

### Required Documentation for KISA Submission

| Document | Status | Owner |
|---|---|---|
| ISMS-P 인증 신청서 (Application) | To prepare | Compliance Manager |
| 관리체계 범위 정의서 (Scope definition) | In progress | CISO |
| 위험 평가 보고서 (Risk assessment report) | In progress | Security Engineer |
| 정보보호 정책서 (Security policy) | Drafted | CISO |
| 개인정보 처리방침 (Privacy policy) | Published | DPO |
| 개인정보 처리 현황 (Personal information inventory) | In progress | DPO |
| 내부 감사 결과 (Internal audit results) | Scheduled | Compliance Manager |
| 보안 교육 실시 기록 (Training records) | Maintained | HR / Compliance |

### KISA Audit Timeline (Planned)

| Phase | Activity | Target |
|---|---|---|
| 서류 심사 (Document review) | Submit documentation to KISA | Month 16 |
| 현장 심사 준비 (On-site audit preparation) | Internal mock audit | Month 17 |
| 현장 심사 (On-site audit) | KISA on-site assessment | Month 18 |
| 인증 획득 (Certification) | ISMS-P certificate issued | Month 18-19 |

---

## Control Coverage Summary

| Domain | Total Controls | Implemented | In Progress | Not Applicable |
|---|:---:|:---:|:---:|:---:|
| 1. Management System | 16 | 12 | 4 | 0 |
| 2. Protection Measures | 64 | 55 | 9 | 0 |
| 3. Personal Information Protection | 22 | 18 | 4 | 0 |
| **Total** | **102** | **85** | **17** | **0** |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for KISA audit |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | October 2026 |
| Standard reference | ISMS-P 인증기준 (KISA, 2023) |
| Language | English (Korean translation in progress) |
