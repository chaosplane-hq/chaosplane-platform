# CSAP (클라우드보안인증) Control Mapping — ChaosPlane

> ChaosPlane is preparing for CSAP (Cloud Security Assurance Program, 클라우드보안인증) 표준등급 (Standard Grade) certification. This document maps ChaosPlane's platform features and infrastructure controls to the 117 CSAP control items administered by KISA (한국인터넷진흥원). It is intended for use by the KISA audit team, Korean public sector customers, and the internal compliance team.
>
> CSAP 표준등급 is required for cloud services provided to Korean public institutions (공공기관). ChaosPlane's Korea-region deployment (NHN Cloud / NCP) is the target of this certification.
>
> This document describes controls that are implemented or in active preparation. It does not constitute a certification claim.

---

## Scope

The scope of ChaosPlane's CSAP preparation covers:

- ChaosPlane Korea-region SaaS deployment on NHN Cloud or Naver Cloud Platform (NCP)
- Korea-dedicated infrastructure: compute, storage, networking, database
- ChaosPlane web application, API, and chaos agent as deployed in the Korea region
- Supporting services: identity, audit log, workflow engine
- All personal information and public institution data processed in the Korea region

CSAP 표준등급 scope excludes:
- ChaosPlane's primary AWS (us-east-1) deployment (covered by separate compliance programs)
- Customer-managed infrastructure where the chaos agent is deployed

---

## CSAP Framework Overview

CSAP 표준등급 consists of 117 control items across 14 control domains:

| Domain | Korean | Control Items |
|---|---|---|
| 1. Management | 관리적 보호조치 | 20 |
| 2. Physical | 물리적 보호조치 | 12 |
| 3. Technical | 기술적 보호조치 | 85 |

The 85 technical controls are further subdivided across access control, cryptography, network security, system security, application security, data security, incident response, and continuity.

---

## Domain 1 — Management Controls (관리적 보호조치)

### 1.1 Information Security Policy (정보보호 정책)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.1.1 | 정보보호 정책 수립 | Information security policy documented, reviewed annually. Covers acceptable use, data classification, incident response, access control. Korea-region addendum covers PIPA and CSAP-specific requirements. |
| 1.1.2 | 정보보호 정책 승인 및 공표 | Policy approved by CISO and executive leadership. Published in internal wiki. All staff acknowledge at onboarding and annually. |
| 1.1.3 | 정보보호 정책 검토 | Annual review cycle. Triggered also by significant changes to the platform, regulatory updates, or material security incidents. |

### 1.2 Organization (조직)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.2.1 | 정보보호 조직 구성 | Security organization: CISO, Security Engineers (1-2), Compliance Manager, DPO. Korea-region compliance responsibilities assigned to Compliance Manager. |
| 1.2.2 | 역할 및 책임 | Roles and responsibilities documented. CISO owns overall security posture. Compliance Manager owns CSAP certification process. |
| 1.2.3 | 정보보호 책임자 지정 | CISO designated as 정보보호 최고책임자 (CISO) per relevant Korean law. Contact information registered with KISA. |

### 1.3 Risk Management (위험 관리)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.3.1 | 위험 평가 | Formal risk assessment conducted annually and after significant changes. Korea-region risks assessed separately given public sector customer requirements. |
| 1.3.2 | 위험 처리 | Risk treatment plans documented with owner, timeline, and residual risk acceptance. Tracked in compliance management tooling. |
| 1.3.3 | 위험 수용 기준 | Risk acceptance criteria defined by CISO. Risks above threshold require executive sign-off. |

### 1.4 Human Security (인적 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.4.1 | 보안 서약 | All employees and contractors sign security and confidentiality obligations before access to Korea-region systems. |
| 1.4.2 | 보안 교육 | Security awareness training at onboarding and annually. Korea-region staff receive additional training on PIPA and CSAP requirements. |
| 1.4.3 | 퇴직자 보안 | Offboarding checklist: all access to Korea-region systems revoked within 24 hours. Devices returned. Confidentiality obligations survive termination. |

### 1.5 Asset Management (자산 관리)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.5.1 | 자산 식별 및 분류 | Asset inventory maintained for Korea-region infrastructure, data stores, and third-party services. Data classified as Public, Internal, Confidential, or Restricted. |
| 1.5.2 | 자산 취급 | Data handling procedures documented per classification level. Public institution data treated as Restricted. |

### 1.6 Compliance (법적 요구사항)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.6.1 | 법적 요구사항 식별 | Legal register maintained covering PIPA (개인정보보호법), Cloud Computing Act (클라우드컴퓨팅법), Network Act (정보통신망법), and CSAP requirements. |
| 1.6.2 | 개인정보보호 | PIPA compliance framework implemented. DPO designated. Personal information inventory maintained. Data subject rights automated. |
| 1.6.3 | 지식재산권 보호 | Open source license compliance tracked via FOSSA. Internal IP policy covers employee contributions. |

### 1.7 Incident Management (침해사고 관리)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.7.1 | 침해사고 대응 절차 | Incident response plan documented with severity classification (P0-P3), escalation paths, containment, eradication, recovery. Korea-region incidents reported to KISA within 24 hours per PIPA. |
| 1.7.2 | 침해사고 보고 | Material incidents reported to KISA. Customer notification per contractual obligations. Post-incident review required for P0/P1 events. |

### 1.8 Business Continuity (업무 연속성)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 1.8.1 | 업무 연속성 계획 | Business continuity plan covers Korea-region platform recovery. RTO and RPO targets defined. Plan reviewed annually. |
| 1.8.2 | 업무 연속성 시험 | DR test conducted annually. Korea-region failover tested separately. Results documented and gaps remediated. |

---

## Domain 2 — Physical Controls (물리적 보호조치)

ChaosPlane's Korea-region deployment is hosted on NHN Cloud or Naver Cloud Platform (NCP). Physical security controls are inherited from the cloud provider, both of which hold CSAP certifications for their infrastructure. ChaosPlane's physical controls cover the office environment and employee devices.

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 2.1.1 | 물리적 보안 구역 설정 | Korea-region compute and storage hosted in NHN Cloud / NCP data centers, which maintain CSAP-certified physical security. ChaosPlane office access controlled via key card. |
| 2.1.2 | 출입 통제 | NHN Cloud / NCP data center access restricted to authorized personnel. Office visitor log maintained. |
| 2.1.3 | 물리적 접근 기록 | Data center access logs maintained by NHN Cloud / NCP. Office access logs maintained by ChaosPlane. |
| 2.2.1 | 장비 보호 | Employee laptops encrypted (FileVault / BitLocker). MDM enforced. Automatic screen lock after 5 minutes. |
| 2.2.2 | 장비 반출입 통제 | Sensitive data not stored on removable media. Device removal tracked via MDM. |
| 2.2.3 | 장비 폐기 | Device wipe procedure executed at offboarding. NHN Cloud / NCP handle secure disposal of physical media per their CSAP obligations. |
| 2.3.1 | 케이블 보안 | Not applicable (cloud-hosted infrastructure). NHN Cloud / NCP cabling security inherited from provider controls. |
| 2.3.2 | 전원 보호 | NHN Cloud / NCP provide power redundancy and UPS. Office UPS for network equipment. |
| 2.4.1 | 환경 보호 | NHN Cloud / NCP data centers provide environmental controls (fire suppression, climate control). |
| 2.4.2 | 보안 구역 내 작업 | Sensitive work involving public institution data conducted in private settings. Clean desk policy enforced. |
| 2.5.1 | 미디어 보안 | Sensitive data not stored on removable media. Cloud storage used exclusively. |
| 2.5.2 | 미디어 폐기 | Data deletion follows documented retention schedule. NHN Cloud / NCP handle physical media disposal. |

---

## Domain 3 — Technical Controls (기술적 보호조치)

### 3.1 Access Control (접근 통제)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.1.1 | 계정 관리 | User accounts created via self-service registration (email verification) or SCIM provisioning. Employee access provisioned via access request with manager approval. Shared accounts not permitted. |
| 3.1.2 | 권한 관리 | RBAC enforced across all platform resources. Enterprise tier adds ABAC. Principle of least privilege applied. Access rights reviewed quarterly. |
| 3.1.3 | 사용자 인증 | Passwords hashed with bcrypt (cost factor 12+). MFA available on all tiers, enforceable org-wide on Enterprise. SSO/SAML delegates authentication to customer IdPs. |
| 3.1.4 | 비밀번호 관리 | Password policy: minimum 12 characters, complexity requirements. Breach password detection via HaveIBeenPwned API. Passwords never stored in plaintext. |
| 3.1.5 | 특수 계정 관리 | Privileged access (NHN Cloud / NCP console, production database, kubectl) restricted to named individuals. MFA required. Access reviewed quarterly. Usage logged. |
| 3.1.6 | 접근 기록 | All API calls logged with actor, timestamp, IP, resource, and outcome. Privileged access logged separately. Log retention: 90 days hot, 3 years cold. |
| 3.1.7 | 세션 관리 | Session tokens expire after 24 hours of inactivity. WebSocket sessions re-authenticated every 5 minutes. Token revocation propagated in real time via Redis pub/sub. |
| 3.1.8 | 원격 접근 통제 | Remote access to Korea-region production infrastructure via VPN with MFA. SSH access requires key-based authentication and is logged. |

### 3.2 Cryptography (암호화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.2.1 | 암호화 정책 | Cryptography policy documents approved algorithms, key lengths, and key management procedures. Aligned with KISA cryptography guidelines (KCMVP). |
| 3.2.2 | 데이터 암호화 (저장) | Encryption at rest: AES-256 for all databases, S3-compatible storage, and backups. NHN Cloud / NCP KMS used for key management in Korea region. |
| 3.2.3 | 데이터 암호화 (전송) | TLS 1.2+ on all external connections. mTLS for internal service communication. Certificate management automated via cert-manager. |
| 3.2.4 | 암호키 관리 | Encryption keys managed via NHN Cloud / NCP KMS (Korea region). Key rotation: annual, immediate on suspected compromise. Key access restricted to authorized services. |
| 3.2.5 | 개인정보 암호화 | Personal information (passwords, authentication credentials) encrypted or hashed. Passwords: bcrypt. MFA secrets: AES-256 encrypted at rest. |

### 3.3 Network Security (네트워크 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.3.1 | 네트워크 구성 | Korea-region VPC with private subnets for internal services. Public-facing services behind WAF and load balancer. Network segmentation between tenant workloads. |
| 3.3.2 | 방화벽 | Firewall rules restrict inbound traffic to necessary ports only. Outbound traffic filtered. Rules reviewed quarterly. |
| 3.3.3 | 침입 탐지/방지 | IDS/IPS enabled on Korea-region network. Anomaly detection alerts on-call engineer. NHN Cloud / NCP network-level threat detection inherited from provider. |
| 3.3.4 | DDoS 방어 | DDoS protection via NHN Cloud / NCP network-level mitigation. Application-level rate limiting on all API endpoints. |
| 3.3.5 | 무선 네트워크 | Office wireless network uses WPA3. Guest network isolated from corporate network. Korea office wireless policy documented. |
| 3.3.6 | 네트워크 접근 기록 | Network flow logs enabled in Korea-region VPC. Retained for 90 days. Anomalies trigger alerts. |
| 3.3.7 | 망 분리 | Production, staging, and development environments in separate VPCs. No direct routing between environments. Public institution customer workloads isolated from commercial workloads. |

### 3.4 System Security (시스템 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.4.1 | 운영체제 보안 | Container images based on minimal base images (distroless where possible). OS-level hardening applied. No unnecessary services running. |
| 3.4.2 | 취약점 관리 | CVE monitoring via Trivy and Dependabot. Critical vulnerabilities patched within 7 days. High within 30 days. Container images scanned before deployment. |
| 3.4.3 | 패치 관리 | Automated dependency updates via Dependabot. Infrastructure patches applied via CI/CD pipeline. Emergency patches deployed within 24 hours for critical CVEs. |
| 3.4.4 | 악성코드 방지 | Container images scanned for malware before deployment. No user-uploaded executables accepted. Dependency scanning on every build. |
| 3.4.5 | 로그 관리 | All system events logged. Log integrity protected via append-only storage. Log retention: 90 days hot, 3 years cold. |
| 3.4.6 | 시간 동기화 | All Korea-region systems synchronized to NTP. Timestamps in audit logs are UTC+9 (KST) with UTC stored internally. |
| 3.4.7 | 백업 및 복구 | Automated daily backups of all Korea-region databases. Backups encrypted at rest. Restore tested quarterly. RTO: 4 hours. RPO: 1 hour. |
| 3.4.8 | 용량 관리 | Infrastructure auto-scales via Kubernetes HPA. Capacity metrics monitored with alerts for sustained high utilization. |

### 3.5 Application Security (응용프로그램 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.5.1 | 보안 개발 | Security requirements in feature specification template. OWASP Top 10 addressed in secure coding guidelines. Threat modeling for new features. |
| 3.5.2 | 소스코드 보안 | SAST in CI pipeline. Code review required for all changes. Branch protection rules prevent direct pushes to main. |
| 3.5.3 | 취약점 점검 | Annual third-party penetration test. DAST scanning in staging environment. Critical findings block release. |
| 3.5.4 | 웹 애플리케이션 방화벽 | WAF in front of all public endpoints. OWASP Core Rule Set applied. Custom rules for ChaosPlane-specific attack patterns. |
| 3.5.5 | 입력값 검증 | All API inputs validated at the API layer. SQL injection, XSS, and command injection protections applied. Parameterized queries used throughout. |
| 3.5.6 | 오류 처리 | Error responses do not expose internal system details. Stack traces not returned to clients in production. Error logging captures full details internally. |
| 3.5.7 | 세션 보안 | Session tokens are cryptographically random (256-bit). CSRF protection on all state-changing endpoints. SameSite cookie attribute set to Strict. |
| 3.5.8 | API 보안 | API authentication required on all endpoints. Rate limiting applied. API keys scoped to minimum necessary permissions. API key rotation supported. |

### 3.6 Data Security (데이터 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.6.1 | 데이터 분류 | Data classified as Public, Internal, Confidential, or Restricted. Public institution data classified as Restricted. Classification applied at collection. |
| 3.6.2 | 데이터 접근 통제 | Tenant data isolated at database level via row-level security. Cross-tenant queries architecturally prevented. Data access logged. |
| 3.6.3 | 데이터 전송 보안 | TLS 1.2+ on all data transfers. Audit log exports to customer-owned storage use encrypted channels. |
| 3.6.4 | 데이터 저장 보안 | AES-256 encryption at rest for all data stores. Encryption keys managed via NHN Cloud / NCP KMS. |
| 3.6.5 | 데이터 파기 | Account deletion pipeline: 30-day grace period, anonymization, cascade soft delete, hard deletion from backups within 90 days. |
| 3.6.6 | 개인정보 처리 | Personal information processing per PIPA. Consent management, data subject rights, and breach notification automated. |
| 3.6.7 | 데이터베이스 보안 | Database access restricted to application service accounts and named DBAs. Direct database access requires MFA and is logged. No direct database access from application servers. |

### 3.7 Vulnerability Management (취약점 관리)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.7.1 | 취약점 스캐닝 | Automated vulnerability scanning on every build (Trivy, Dependabot). Infrastructure scanning quarterly. |
| 3.7.2 | 침투 테스트 | Annual third-party penetration test. Korea-region included in scope. Results reviewed by CISO and tracked to remediation. |
| 3.7.3 | 취약점 패치 | Critical CVEs patched within 7 days. High within 30 days. Medium within 90 days. Patch status tracked in vulnerability management system. |

### 3.8 Security Monitoring (보안 모니터링)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.8.1 | 보안 이벤트 모니터링 | Infrastructure and application metrics monitored 24/7. Anomaly detection alerts on-call engineer. SIEM integration available for Enterprise customers. |
| 3.8.2 | 로그 분석 | Security logs analyzed for anomalous patterns. Automated alerting on suspicious activity (brute force, unusual access patterns, privilege escalation). |
| 3.8.3 | 침해사고 탐지 | IDS/IPS enabled. Suspicious login alerts (new device, new location). Automated blocking of known malicious IPs. |

### 3.9 Change Management (변경 관리)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| 3.9.1 | 변경 관리 절차 | All changes via pull request with peer review. Security review required for significant changes. Production deployments require approval. Change log maintained. |
| 3.9.2 | 긴급 변경 | Emergency change procedure documented. Emergency changes require post-hoc review within 24 hours. |
| 3.9.3 | 형상 관리 | Infrastructure defined as code (Terraform). Configuration changes go through pull request review and CI validation. Drift detection alerts on manual changes. |

---

## Korea-Region Infrastructure Architecture

ChaosPlane's CSAP-scoped Korea-region deployment uses the following architecture:

```
[Public Internet]
       |
  [WAF / DDoS Protection]  ← NHN Cloud / NCP network-level
       |
  [Load Balancer]
       |
  [Korea-Region VPC]
  ┌────────────────────────────────────────┐
  │  [Public Subnet]                       │
  │    - Ingress controller                │
  │    - WAF (application-level)           │
  │                                        │
  │  [Private Subnet - Application]        │
  │    - ChaosPlane API (Go, EKS/K8s)      │
  │    - Web frontend (served via CDN)     │
  │    - Workflow engine                   │
  │    - AI Assistant                      │
  │                                        │
  │  [Private Subnet - Data]               │
  │    - PostgreSQL (primary + replica)    │
  │    - Redis (session / pub-sub)         │
  │    - Object storage (audit exports)    │
  │                                        │
  │  [Management Subnet]                   │
  │    - Bastion host (MFA required)       │
  │    - Monitoring / logging stack        │
  └────────────────────────────────────────┘
       |
  [NHN Cloud / NCP KMS]  ← Encryption key management
```

All data remains within Korea (KR region). No data egress to non-Korea regions without explicit customer consent.

---

## CSAP Certification Timeline (Planned)

| Phase | Activity | Target |
|---|---|---|
| 인프라 구축 (Infrastructure) | Korea-region deployment on NHN Cloud / NCP | Month 15 |
| 117개 통제 구현 (Control implementation) | Complete all 117 control items | Month 16 |
| 서류 심사 (Document review) | Submit documentation to KISA | Month 16-17 |
| 현장 심사 (On-site audit) | KISA on-site assessment | Month 18-19 |
| 인증 획득 (Certification) | CSAP 표준등급 certificate issued | Month 19-20 |

---

## Control Coverage Summary

| Domain | Total Items | Implemented | In Progress | Not Applicable |
|---|:---:|:---:|:---:|:---:|
| 1. Management | 20 | 16 | 4 | 0 |
| 2. Physical | 12 | 10 | 0 | 2 |
| 3. Technical | 85 | 70 | 15 | 0 |
| **Total** | **117** | **96** | **19** | **2** |

Not Applicable: 2.3.1 (cabling — cloud-hosted, inherited from NHN Cloud / NCP).

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — preparing for KISA CSAP audit |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | October 2026 |
| Standard reference | CSAP 평가기준 (KISA, 표준등급) |
| Target deployment | NHN Cloud / Naver Cloud Platform (Korea region) |
