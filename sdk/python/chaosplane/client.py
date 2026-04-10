from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any
from urllib.parse import quote, urlencode

import httpx

DEFAULT_BASE_URL = "http://localhost:8080"


class ChaosPlaneError(Exception):
    def __init__(self, status_code: int, message: str, body: Any = None):
        super().__init__(message)
        self.status_code = status_code
        self.body = body


@dataclass
class AuthUser:
    id: str
    email: str
    name: str
    email_verified: bool
    last_login_at: str | None = None


@dataclass
class AuthTenant:
    id: str
    name: str
    slug: str


@dataclass
class AuthResponse:
    access_token: str
    refresh_token: str
    expires_in: int
    user: AuthUser
    tenant: AuthTenant


@dataclass
class ActionRequest:
    type: str
    parameters: dict[str, Any] | None = None


@dataclass
class TargetRequest:
    kind: str
    namespace: str | None = None
    label_selector: dict[str, str] | None = None
    names: list[str] | None = None


@dataclass
class CreateExperimentRequest:
    name: str
    namespace: str
    action: ActionRequest
    target: TargetRequest
    duration: str


@dataclass
class ExperimentResponse:
    name: str
    namespace: str
    action: str
    phase: str
    start_time: str | None = None
    end_time: str | None = None


@dataclass
class PaginatedExperiments:
    items: list[ExperimentResponse]
    total_count: int
    limit: int
    offset: int


@dataclass
class Organization:
    id: str
    tenant_id: str
    name: str
    slug: str
    description: str | None = None


@dataclass
class Workspace:
    id: str
    tenant_id: str
    organization_id: str
    name: str
    slug: str
    description: str | None = None


@dataclass
class Team:
    id: str
    tenant_id: str
    workspace_id: str
    name: str
    slug: str
    description: str | None = None


@dataclass
class Project:
    id: str
    tenant_id: str
    workspace_id: str
    name: str
    slug: str
    description: str | None = None


@dataclass
class Environment:
    id: str
    tenant_id: str
    project_id: str
    name: str
    slug: str
    type: str
    agent_status: str


@dataclass
class HierarchyResponse:
    organizations: list[Organization]
    workspaces: list[Workspace]
    teams: list[Team]
    projects: list[Project]
    environments: list[Environment]


@dataclass
class Subscription:
    id: str
    tenant_id: str
    plan: str
    status: str
    gateway: str
    created_at: str
    current_period_start: str | None = None
    current_period_end: str | None = None
    trial_ends_at: str | None = None
    cancelled_at: str | None = None


@dataclass
class UsageStats:
    experiments: int
    agents: int
    api_calls: int


@dataclass
class PlanLimits:
    max_experiments: int
    max_agents: int
    max_api_calls: int


@dataclass
class BillingStatusResponse:
    subscription: Subscription | None
    usage: UsageStats | None
    limits: PlanLimits | None


def _to_camel(s: str) -> str:
    parts = s.split("_")
    return parts[0] + "".join(p.capitalize() for p in parts[1:])


def _serialize(obj: Any) -> Any:
    if isinstance(obj, dict):
        return {_to_camel(k): _serialize(v) for k, v in obj.items() if v is not None}
    if isinstance(obj, list):
        return [_serialize(i) for i in obj]
    if hasattr(obj, "__dataclass_fields__"):
        return _serialize({k: getattr(obj, k) for k in obj.__dataclass_fields__})
    return obj


def _parse_auth_response(data: dict[str, Any]) -> AuthResponse:
    return AuthResponse(
        access_token=data["accessToken"],
        refresh_token=data["refreshToken"],
        expires_in=data["expiresIn"],
        user=AuthUser(
            id=data["user"]["id"],
            email=data["user"]["email"],
            name=data["user"]["name"],
            email_verified=data["user"]["emailVerified"],
            last_login_at=data["user"].get("lastLoginAt"),
        ),
        tenant=AuthTenant(
            id=data["tenant"]["id"],
            name=data["tenant"]["name"],
            slug=data["tenant"]["slug"],
        ),
    )


def _parse_experiment(data: dict[str, Any]) -> ExperimentResponse:
    return ExperimentResponse(
        name=data["name"],
        namespace=data["namespace"],
        action=data.get("action", ""),
        phase=data.get("phase", ""),
        start_time=data.get("startTime"),
        end_time=data.get("endTime"),
    )


class ChaosPlaneClient:
    def __init__(
        self,
        base_url: str = DEFAULT_BASE_URL,
        api_key: str | None = None,
        timeout: float = 30.0,
    ):
        self._base_url = base_url.rstrip("/")
        self._api_key = api_key
        self._access_token: str | None = None
        self._refresh_token: str | None = None
        self._client = httpx.Client(timeout=timeout)

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> ChaosPlaneClient:
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def set_tokens(self, access_token: str, refresh_token: str) -> None:
        self._access_token = access_token
        self._refresh_token = refresh_token

    def login(self, email: str, password: str) -> AuthResponse:
        data = self._request(
            "POST", "/auth/login", json={"email": email, "password": password}
        )
        resp = _parse_auth_response(data)
        self._access_token = resp.access_token
        self._refresh_token = resp.refresh_token
        return resp

    def refresh(self) -> AuthResponse:
        if not self._refresh_token:
            raise ChaosPlaneError(0, "No refresh token available")
        data = self._request(
            "POST", "/auth/refresh", json={"refreshToken": self._refresh_token}
        )
        resp = _parse_auth_response(data)
        self._access_token = resp.access_token
        self._refresh_token = resp.refresh_token
        return resp

    def list_experiments(
        self, limit: int = 20, offset: int = 0
    ) -> PaginatedExperiments:
        params = urlencode({"limit": limit, "offset": offset})
        data = self._request("GET", f"/api/v1/experiments?{params}")
        return PaginatedExperiments(
            items=[_parse_experiment(item) for item in data.get("items", [])],
            total_count=data["totalCount"],
            limit=data["limit"],
            offset=data["offset"],
        )

    def create_experiment(self, req: CreateExperimentRequest) -> ExperimentResponse:
        data = self._request("POST", "/api/v1/experiments", json=_serialize(req))
        return _parse_experiment(data)

    def get_experiment(self, name: str) -> ExperimentResponse:
        data = self._request("GET", f"/api/v1/experiments/{quote(name, safe='')}")
        return _parse_experiment(data)

    def delete_experiment(self, name: str) -> None:
        self._request("DELETE", f"/api/v1/experiments/{quote(name, safe='')}")

    def abort_experiment(self, name: str) -> ExperimentResponse:
        data = self._request(
            "POST", f"/api/v1/experiments/{quote(name, safe='')}/abort"
        )
        return _parse_experiment(data)

    def list_hierarchy(self) -> HierarchyResponse:
        data = self._request("GET", "/api/v1/hierarchy")
        return HierarchyResponse(
            organizations=[
                Organization(
                    id=o["id"],
                    tenant_id=o["tenantId"],
                    name=o["name"],
                    slug=o["slug"],
                    description=o.get("description"),
                )
                for o in data.get("organizations", [])
            ],
            workspaces=[
                Workspace(
                    id=w["id"],
                    tenant_id=w["tenantId"],
                    organization_id=w["organizationId"],
                    name=w["name"],
                    slug=w["slug"],
                    description=w.get("description"),
                )
                for w in data.get("workspaces", [])
            ],
            teams=[
                Team(
                    id=t["id"],
                    tenant_id=t["tenantId"],
                    workspace_id=t["workspaceId"],
                    name=t["name"],
                    slug=t["slug"],
                    description=t.get("description"),
                )
                for t in data.get("teams", [])
            ],
            projects=[
                Project(
                    id=p["id"],
                    tenant_id=p["tenantId"],
                    workspace_id=p["workspaceId"],
                    name=p["name"],
                    slug=p["slug"],
                    description=p.get("description"),
                )
                for p in data.get("projects", [])
            ],
            environments=[
                Environment(
                    id=e["id"],
                    tenant_id=e["tenantId"],
                    project_id=e["projectId"],
                    name=e["name"],
                    slug=e["slug"],
                    type=e["type"],
                    agent_status=e["agentStatus"],
                )
                for e in data.get("environments", [])
            ],
        )

    def get_billing_status(self) -> BillingStatusResponse:
        data = self._request("GET", "/api/v1/billing")
        sub_data = data.get("subscription")
        usage_data = data.get("usage")
        limits_data = data.get("limits")
        return BillingStatusResponse(
            subscription=Subscription(
                id=sub_data["id"],
                tenant_id=sub_data["tenantId"],
                plan=sub_data["plan"],
                status=sub_data["status"],
                gateway=sub_data.get("gateway", ""),
                created_at=sub_data["createdAt"],
                current_period_start=sub_data.get("currentPeriodStart"),
                current_period_end=sub_data.get("currentPeriodEnd"),
                trial_ends_at=sub_data.get("trialEndsAt"),
                cancelled_at=sub_data.get("cancelledAt"),
            )
            if sub_data
            else None,
            usage=UsageStats(
                experiments=usage_data["experiments"],
                agents=usage_data["agents"],
                api_calls=usage_data["apiCalls"],
            )
            if usage_data
            else None,
            limits=PlanLimits(
                max_experiments=limits_data["maxExperiments"],
                max_agents=limits_data["maxAgents"],
                max_api_calls=limits_data["maxApiCalls"],
            )
            if limits_data
            else None,
        )

    def _request(self, method: str, path: str, json: Any = None) -> Any:
        headers: dict[str, str] = {"Accept": "application/json"}

        if self._api_key:
            headers["X-API-Key"] = self._api_key
        elif self._access_token:
            headers["Authorization"] = f"Bearer {self._access_token}"

        resp = self._client.request(
            method, f"{self._base_url}{path}", headers=headers, json=json
        )

        if resp.status_code >= 400:
            message = resp.text
            body = None
            try:
                body = resp.json()
                if isinstance(body, dict) and "error" in body:
                    message = body["error"]
            except Exception:
                pass
            raise ChaosPlaneError(resp.status_code, message, body)

        if resp.status_code == 204 or not resp.content:
            return None

        return resp.json()
