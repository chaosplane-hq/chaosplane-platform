# CSAP 상등급 Upgrade Preparation — ChaosPlane

> ChaosPlane holds CSAP 표준등급 (Standard Grade) certification for its Korea-region SaaS deployment. This document covers the preparation for upgrading to CSAP 상등급 (Advanced Grade), which is required for cloud services handling sensitive public institution data and national security-adjacent workloads. 상등급 is administered by KISA (한국인터넷진흥원) and involves a more rigorous control set and on-site audit process.
>
> Target: CSAP 상등급 certification by Month 32-34 (Phase 5).

---

## 표준등급 vs 상등급 Comparison

| Dimension | 표준등급 (Standard) | 상등급 (Advanced) |
|---|---|---|
| Control items | 117 | 177 |
| Additional controls | — | +60 controls |
| Audit process | Document review + on-site | Document review + extended on-site + technical verification |
| Target customers | General public institutions | Sensitive public institutions, national security-adjacent |
| Infrastructure requirement | CSAP-certified cloud (NHN Cloud / NCP) | Same, with stricter isolation requirements |
| Encryption | AES-256, KCMVP-approved | AES-256, KCMVP-approved + additional key management controls |
| Audit log retention | 1 year | 3 years |
| Penetration test | Annual (third-party) | Semi-annual (KISA-approved tester) |
| Incident response | 24-hour KISA notification | 2-hour KISA notification for P0/P1 |
| Personnel security | Background check | Enhanced background check + periodic reinvestigation |

---

## Additional Control Domains for 상등급

상등급 adds 60 control items beyond 표준등급 across the following areas:

### Enhanced Access Control (접근 통제 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| AC-E-01 | 특권 계정 분리 (Privileged account separation) | Production privileged access uses dedicated accounts separate from developer accounts. No shared privileged accounts. |
| AC-E-02 | 특권 세션 녹화 (Privileged session recording) | All privileged SSH and kubectl sessions recorded via session recording proxy. Recordings retained 3 years. |
| AC-E-03 | 접근 권한 자동 만료 (Automatic access expiry) | Privileged access grants expire after 8 hours. Re-approval required for extended access. |
| AC-E-04 | 다중 인증 강화 (Enhanced MFA) | Hardware security key (FIDO2) required for all privileged access to Korea-region production. Software MFA not accepted for privileged roles. |
| AC-E-05 | 접근 이상 탐지 (Access anomaly detection) | ML-based anomaly detection on access patterns. Alerts on unusual login times, locations, or access volumes. |
| AC-E-06 | 제로 트러스트 접근 (Zero-trust access) | All internal service-to-service calls authenticated via mTLS with short-lived certificates (24-hour TTL). No implicit trust based on network location. |

### Enhanced Cryptography (암호화 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| CR-E-01 | KCMVP 검증 모듈 (KCMVP validated modules) | All cryptographic operations use KCMVP-validated modules. ARIA-256 available as alternative to AES-256 for Korean public sector customers who require it. |
| CR-E-02 | 암호키 이중 관리 (Dual-control key management) | Encryption key operations (creation, rotation, deletion) require approval from two authorized personnel. Enforced via NHN Cloud / NCP KMS policy. |
| CR-E-03 | 키 에스크로 금지 (No key escrow) | Encryption keys are never escrowed to third parties. Customer data keys are accessible only to the customer's tenant context. |
| CR-E-04 | 암호화 감사 (Cryptography audit) | All KMS key usage logged. Key access anomalies trigger alerts. Annual cryptography review by CISO. |
| CR-E-05 | 전송 암호화 강화 (Enhanced transit encryption) | TLS 1.3 required for all new connections (TLS 1.2 permitted only for legacy compatibility with explicit approval). Perfect forward secrecy enforced. |

### Enhanced Network Security (네트워크 보안 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| NS-E-01 | 망 분리 강화 (Enhanced network segmentation) | Korea-region production network segmented into 5 zones: DMZ, application, data, management, monitoring. No direct routing between non-adjacent zones. |
| NS-E-02 | 마이크로 세그멘테이션 (Micro-segmentation) | Kubernetes NetworkPolicy enforces pod-level network isolation. Each service allows only necessary inbound connections from specific source pods. |
| NS-E-03 | 동서 트래픽 검사 (East-west traffic inspection) | Internal service traffic inspected via service mesh (Istio). Anomalous internal traffic patterns trigger alerts. |
| NS-E-04 | DNS 보안 (DNS security) | DNSSEC enabled for chaosplane.dev Korea subdomain. DNS queries from Korea-region pods resolved via internal resolver only. |
| NS-E-05 | 네트워크 포렌식 (Network forensics) | Full packet capture capability available for incident investigation. Captures stored encrypted, accessible only to CISO and authorized security engineers. |

### Enhanced Data Security (데이터 보안 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| DS-E-01 | 데이터 분류 자동화 (Automated data classification) | ML-based data classification scans experiment definitions and results for sensitive data patterns (PII, CUI-equivalent Korean data). Auto-tags and applies handling controls. |
| DS-E-02 | 데이터 손실 방지 (DLP) | DLP controls on all data egress paths. Bulk exports of sensitive data require supervisor approval and are logged. |
| DS-E-03 | 데이터 마스킹 (Data masking) | Sensitive fields (personal information, system credentials) masked in logs and non-production environments. |
| DS-E-04 | 데이터 접근 감사 강화 (Enhanced data access audit) | Row-level access logging for all sensitive data. Access patterns analyzed for anomalies. Reports available to agency CISO on request. |
| DS-E-05 | 개인정보 비식별화 (PII de-identification) | Personal information de-identified before use in analytics or AI training. Re-identification controls prevent reverse engineering. |
| DS-E-06 | 데이터 주권 (Data sovereignty) | All Korea-region data remains in Korea. Technical controls prevent data egress to non-Korea regions. Verified via network egress monitoring and quarterly audit. |

### Enhanced Incident Response (침해사고 대응 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| IR-E-01 | 2시간 KISA 신고 (2-hour KISA notification) | P0/P1 incidents involving Korea-region data reported to KISA within 2 hours. Automated notification pipeline with KISA contact integration. |
| IR-E-02 | 사이버 위기 대응 (Cyber crisis response) | Cyber crisis response plan documented for nation-state level attacks. Coordination procedures with KISA and NIS (국가정보원) defined. |
| IR-E-03 | 포렌식 역량 (Forensic capability) | Digital forensics capability maintained. Evidence preservation procedures documented. Chain of custody maintained for all forensic artifacts. |
| IR-E-04 | 침해사고 모의훈련 (IR tabletop exercises) | Semi-annual IR tabletop exercises. Korea-region scenarios included. Results documented and gaps remediated. |

### Enhanced Continuity (업무 연속성 강화)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| BC-E-01 | RTO/RPO 강화 (Enhanced RTO/RPO) | Korea-region 상등급: RTO 2 hours, RPO 30 minutes. Continuous WAL archiving for PostgreSQL. |
| BC-E-02 | 이중화 강화 (Enhanced redundancy) | All Korea-region critical components deployed in active-active configuration across two availability zones. Single AZ failure does not impact service. |
| BC-E-03 | DR 훈련 강화 (Enhanced DR testing) | Semi-annual DR test for Korea-region. Full failover test, not just backup restore. Results reported to KISA. |

### Supply Chain Security (공급망 보안)

| Item | Control | ChaosPlane Implementation |
|---|---|---|
| SC-E-01 | 소프트웨어 공급망 (Software supply chain) | SBOM generated for every release. All dependencies reviewed for country-of-origin risk. No components from sanctioned countries. |
| SC-E-02 | 오픈소스 보안 (Open source security) | Open source components reviewed against KISA vulnerability database and NVD. Critical CVEs patched within 7 days. |
| SC-E-03 | 클라우드 공급자 감사 (Cloud provider audit) | NHN Cloud / NCP CSAP certification verified annually. Provider security posture reviewed quarterly. |

---

## Infrastructure Changes for 상등급

### Enhanced Korea-Region Architecture

상등급 requires stricter network segmentation and dedicated management infrastructure:

```
[Public Internet]
       |
  [WAF / DDoS — NHN Cloud / NCP]
       |
  [Load Balancer]
       |
  [Korea-Region VPC — 상등급]
  ┌────────────────────────────────────────────────────────┐
  │  Zone 1: DMZ                                           │
  │    - Ingress controller                                │
  │    - WAF (application-level, OWASP CRS + custom rules) │
  │    - Rate limiting                                     │
  │                                                        │
  │  Zone 2: Application (CUI processing)                  │
  │    - ChaosPlane API (Go, mTLS)                         │
  │    - Workflow engine                                   │
  │    - AI Assistant (Korea-region model endpoint)        │
  │    - CUI Marking Service                               │
  │                                                        │
  │  Zone 3: Data (encrypted at rest, KCMVP)               │
  │    - PostgreSQL (primary + synchronous standby)        │
  │    - Redis (session / pub-sub)                         │
  │    - Object storage (audit exports, backups)           │
  │                                                        │
  │  Zone 4: Management (restricted, session recorded)     │
  │    - Bastion (FIDO2 hardware key required)             │
  │    - Session recording proxy                           │
  │    - Privileged access workstation (PAW)               │
  │                                                        │
  │  Zone 5: Monitoring / SIEM                             │
  │    - Log aggregation (append-only)                     │
  │    - SIEM (anomaly detection)                          │
  │    - Network forensics capture                         │
  └────────────────────────────────────────────────────────┘
       |
  [NHN Cloud / NCP KMS — KCMVP validated]
```

---

## 상등급 Certification Timeline

| Phase | Activity | Target |
|---|---|---|
| Gap analysis | Identify delta from 표준등급 to 상등급 | Month 28 |
| Control implementation | Implement 60 additional controls | Month 28-30 |
| Internal audit | Self-assessment against 177 controls | Month 30 |
| Documentation | Update SSP, control evidence, architecture docs | Month 30-31 |
| KISA pre-consultation | Pre-audit consultation with KISA | Month 31 |
| Document review (서류 심사) | KISA document review | Month 31-32 |
| On-site audit (현장 심사) | KISA extended on-site assessment | Month 32-33 |
| Technical verification | KISA technical verification testing | Month 33 |
| Certification | CSAP 상등급 certificate issued | Month 33-34 |

---

## Control Coverage Summary

| Domain | 표준등급 Items | 상등급 Additional | Total | Implemented | In Progress |
|---|:---:|:---:|:---:|:---:|:---:|
| Management | 20 | 8 | 28 | 20 | 8 |
| Physical | 12 | 4 | 16 | 12 | 4 |
| Technical | 85 | 48 | 133 | 85 | 48 |
| **Total** | **117** | **60** | **177** | **117** | **60** |

All 117 표준등급 controls are implemented (CSAP 표준등급 certified). The 60 additional 상등급 controls are in active implementation for Phase 5.

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | In preparation |
| Owner | CISO / Compliance Manager |
| Last reviewed | April 2026 |
| Next review | Upon KISA pre-consultation |
| Standard reference | CSAP 평가기준 (KISA, 상등급) |
| Prerequisite | CSAP 표준등급 (certified) |
