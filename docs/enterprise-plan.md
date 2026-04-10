# ChaosPlane Enterprise Plan

> ChaosPlane v1.0.0 GA — Pricing, features, and SLA guarantees for every tier.

---

## Pricing Tiers

### Free
**$0 / forever**

For individuals and small teams exploring chaos engineering.

- 1 environment
- Up to 3 users
- 10 experiments/month
- Community-supported experiments (CPU, memory, network latency, pod kill)
- Basic Resilience Score (read-only)
- 7-day experiment history
- Community support (GitHub Discussions)

---

### Team
**$49 / user / month** (billed annually) — $59 month-to-month

For growing engineering teams running regular chaos experiments.

Everything in Free, plus:

- Up to 5 environments
- Unlimited users
- Unlimited experiments
- eBPF-based kernel-level chaos (network partition, disk I/O, syscall fault injection)
- AWS chaos actions (EC2 stop/terminate, RDS failover, AZ blackout, EKS node drain)
- Workflow builder (sequential + parallel steps)
- Resilience Score with trend tracking
- 90-day experiment history
- Slack + PagerDuty notifications
- Email support (48h SLA)

---

### Business
**$149 / user / month** (billed annually) — $179 month-to-month

For organizations that need stronger access controls and deeper observability.

Everything in Team, plus:

- Unlimited environments
- RBAC (Role-Based Access Control)
- AI Assistant: topology analysis, vulnerability detection, experiment suggestions, natural language result summaries
- GameDay planning and execution
- Workflow visual builder with DAG editor, conditional branching, loops, and error handling
- Workflow template library
- Audit log (90-day retention)
- Audit log export (CSV, JSON)
- Prometheus + Grafana integration
- Datadog, New Relic, Dynatrace integrations
- 1-year experiment history
- Priority email support (24h SLA)
- Dedicated Slack channel (5+ seats)

---

### Enterprise
**Custom pricing** — contact sales

For large organizations with strict security, compliance, and availability requirements.

Everything in Business, plus:

- SSO / SAML 2.0 (Okta, Azure AD, Google Workspace, generic SAML IdP)
- SCIM provisioning and JIT (Just-In-Time) provisioning
- ABAC (Attribute-Based Access Control) for fine-grained permission policies
- MFA with recovery codes (10 one-time codes, SHA-256 hashed)
- Concurrent session limits (configurable per organization policy)
- WebSocket continuous re-authentication (5-minute token revalidation)
- Automated account deletion pipeline with 30-day grace period (GDPR Right to Erasure)
- Detailed audit log: every API call recorded with actor, timestamp, IP, and resource
- Audit log export to S3 / MinIO
- SIEM integration (Splunk, Datadog)
- Resilience Score with anonymous industry benchmarking
- GameDay with participant role management, real-time event timeline, and auto-generated postmortems
- Extension Kit: custom chaos action development framework and plugin architecture
- Migration tooling: Gremlin → ChaosPlane, LitmusChaos → ChaosPlane
- Unlimited experiment history
- Dedicated Customer Success Manager
- 99.9% uptime SLA (see SLA section below)
- 4-hour critical incident response SLA
- Private Slack channel with engineering escalation path
- Quarterly business reviews
- 30-day free Enterprise trial (PoC program)

---

### Government
**Custom pricing** — contact public sector sales

For government agencies and regulated public sector organizations.

Everything in Enterprise, plus:

- Data residency controls (region-locked storage and processing)
- Enhanced audit trail with tamper-evident log signing
- Dedicated single-tenant deployment option
- Compliance documentation package (ISO 27001 mapping, SOC 2 Type I scope)
- ISMS-P certification support materials
- Named security contact and quarterly security reviews
- Custom contractual terms (DPA, BAA, security addendum)

> Note: FedRAMP authorization, air-gap deployment, and hardware MFA enforcement are planned for Phase 4.

---

## Feature Matrix

| Feature | Free | Team | Business | Enterprise | Government |
|---|:---:|:---:|:---:|:---:|:---:|
| Environments | 1 | 5 | Unlimited | Unlimited | Unlimited |
| Users | 3 | Unlimited | Unlimited | Unlimited | Unlimited |
| Experiments/month | 10 | Unlimited | Unlimited | Unlimited | Unlimited |
| Community chaos actions | ✓ | ✓ | ✓ | ✓ | ✓ |
| eBPF kernel-level chaos | | ✓ | ✓ | ✓ | ✓ |
| AWS chaos actions | | ✓ | ✓ | ✓ | ✓ |
| Workflow builder | | ✓ | ✓ | ✓ | ✓ |
| Visual DAG editor | | | ✓ | ✓ | ✓ |
| Workflow templates | | | ✓ | ✓ | ✓ |
| AI Assistant | | | ✓ | ✓ | ✓ |
| Resilience Score | Read-only | ✓ | ✓ | ✓ | ✓ |
| Industry benchmarking | | | | ✓ | ✓ |
| GameDay | | | ✓ | ✓ | ✓ |
| GameDay postmortems | | | | ✓ | ✓ |
| RBAC | | | ✓ | ✓ | ✓ |
| ABAC | | | | ✓ | ✓ |
| SSO / SAML | | | | ✓ | ✓ |
| SCIM provisioning | | | | ✓ | ✓ |
| MFA recovery codes | | | | ✓ | ✓ |
| Session limits | | | | ✓ | ✓ |
| Audit log | | | ✓ | ✓ | ✓ |
| Audit log export | | | ✓ | ✓ | ✓ |
| SIEM integration | | | | ✓ | ✓ |
| S3 / MinIO export | | | | ✓ | ✓ |
| Extension Kit | | | | ✓ | ✓ |
| Migration tooling | | | | ✓ | ✓ |
| Data residency controls | | | | | ✓ |
| Single-tenant deployment | | | | | ✓ |
| Experiment history | 7 days | 90 days | 1 year | Unlimited | Unlimited |
| Support | Community | Email 48h | Email 24h + Slack | CSM + 4h SLA | Named contact |
| Uptime SLA | None | None | 99.5% | 99.9% | 99.9% |

---

## SLA Guarantees

### Uptime SLA

| Tier | Monthly Uptime Target | Maximum Downtime/Month |
|---|---|---|
| Business | 99.5% | ~3.6 hours |
| Enterprise | 99.9% | ~43 minutes |
| Government | 99.9% | ~43 minutes |

Uptime is measured as the availability of the ChaosPlane API and web application, excluding scheduled maintenance windows (announced 72 hours in advance) and force majeure events.

### Incident Response SLA

| Severity | Definition | Enterprise Response | Government Response |
|---|---|---|---|
| P0 — Critical | Platform unavailable, data loss risk | 1 hour | 1 hour |
| P1 — High | Core feature unavailable, no workaround | 4 hours | 2 hours |
| P2 — Medium | Degraded performance, workaround exists | 1 business day | 1 business day |
| P3 — Low | Minor issue, cosmetic, documentation | 3 business days | 3 business days |

Response time is measured from the moment a ticket is opened via the dedicated support channel.

### SLA Credits

If ChaosPlane fails to meet the monthly uptime SLA, affected Enterprise and Government customers are eligible for service credits:

| Uptime Achieved | Credit |
|---|---|
| 99.0% – 99.9% | 10% of monthly fee |
| 95.0% – 99.0% | 25% of monthly fee |
| Below 95.0% | 50% of monthly fee |

Credits are applied to the next billing cycle. Credits are the sole remedy for SLA breaches and do not entitle customers to refunds.

### Scheduled Maintenance

- Maintenance windows: Sundays 02:00–06:00 UTC
- Minimum 72-hour advance notice via status page and email
- Emergency maintenance may occur with shorter notice; customers are notified immediately

---

## Support Channels

| Tier | Channel | Hours |
|---|---|---|
| Free | GitHub Discussions | Community-driven |
| Team | Email (support@chaosplane.io) | Business hours |
| Business | Email + dedicated Slack channel | Business hours |
| Enterprise | Dedicated Slack + CSM + escalation path | 24/7 for P0/P1 |
| Government | Named security contact + dedicated Slack | 24/7 for P0/P1 |

---

## Add-ons (All Paid Tiers)

| Add-on | Price |
|---|---|
| Additional environments | $200/environment/month |
| Extended audit log retention (beyond plan default) | $50/month per additional year |
| Professional Services (onboarding, GameDay facilitation) | Custom |
| Training and certification program | Custom |

---

## Billing and Terms

- Annual plans are billed upfront. Monthly plans are billed at the start of each billing cycle.
- Seat count changes take effect at the next billing cycle.
- Enterprise and Government contracts require a signed order form and are subject to negotiated terms.
- All prices are in USD and exclude applicable taxes.
- Free tier does not require a credit card.
