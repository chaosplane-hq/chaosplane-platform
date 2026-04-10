const DEFAULT_BASE_URL = "http://localhost:8080";

export class ChaosPlaneError extends Error {
  constructor(
    public readonly statusCode: number,
    message: string,
    public readonly body?: unknown,
  ) {
    super(message);
    this.name = "ChaosPlaneError";
  }
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshRequest {
  refreshToken: string;
}

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  emailVerified: boolean;
  lastLoginAt?: string;
}

export interface AuthTenant {
  id: string;
  name: string;
  slug: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: AuthUser;
  tenant: AuthTenant;
}

export interface ActionRequest {
  type: string;
  parameters?: Record<string, unknown>;
}

export interface TargetRequest {
  kind: string;
  namespace?: string;
  labelSelector?: Record<string, string>;
  names?: string[];
}

export interface CreateExperimentRequest {
  name: string;
  namespace: string;
  action: ActionRequest;
  target: TargetRequest;
  duration: string;
}

export interface ExperimentResponse {
  name: string;
  namespace: string;
  action: string;
  phase: string;
  startTime?: string;
  endTime?: string;
}

export interface PaginatedExperiments {
  items: ExperimentResponse[];
  totalCount: number;
  limit: number;
  offset: number;
}

export interface Organization {
  id: string;
  tenantId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Workspace {
  id: string;
  tenantId: string;
  organizationId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Team {
  id: string;
  tenantId: string;
  workspaceId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Project {
  id: string;
  tenantId: string;
  workspaceId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Environment {
  id: string;
  tenantId: string;
  projectId: string;
  name: string;
  slug: string;
  type: string;
  agentStatus: string;
}

export interface HierarchyResponse {
  organizations: Organization[];
  workspaces: Workspace[];
  teams: Team[];
  projects: Project[];
  environments: Environment[];
}

export interface Subscription {
  id: string;
  tenantId: string;
  plan: string;
  status: string;
  gateway: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  trialEndsAt?: string;
  cancelledAt?: string;
  createdAt: string;
}

export interface UsageStats {
  experiments: number;
  agents: number;
  apiCalls: number;
}

export interface PlanLimits {
  maxExperiments: number;
  maxAgents: number;
  maxApiCalls: number;
}

export interface BillingStatusResponse {
  subscription: Subscription | null;
  usage: UsageStats | null;
  limits: PlanLimits | null;
}

export interface ChaosPlaneClientOptions {
  /** API base URL. Defaults to http://localhost:8080 */
  baseUrl?: string;
  /** API key for X-API-Key authentication. */
  apiKey?: string;
  /** Custom fetch implementation (defaults to global fetch). */
  fetch?: typeof globalThis.fetch;
}

export class ChaosPlaneClient {
  private baseUrl: string;
  private apiKey?: string;
  private accessToken?: string;
  private refreshTokenValue?: string;
  private fetchFn: typeof globalThis.fetch;

  constructor(options: ChaosPlaneClientOptions = {}) {
    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
    this.apiKey = options.apiKey;
    this.fetchFn = options.fetch ?? globalThis.fetch.bind(globalThis);
  }

  setTokens(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken;
    this.refreshTokenValue = refreshToken;
  }

  async login(req: LoginRequest): Promise<AuthResponse> {
    const resp = await this.request<AuthResponse>("POST", "/auth/login", req);
    this.accessToken = resp.accessToken;
    this.refreshTokenValue = resp.refreshToken;
    return resp;
  }

  async refresh(): Promise<AuthResponse> {
    if (!this.refreshTokenValue) {
      throw new ChaosPlaneError(0, "No refresh token available");
    }
    const resp = await this.request<AuthResponse>("POST", "/auth/refresh", {
      refreshToken: this.refreshTokenValue,
    });
    this.accessToken = resp.accessToken;
    this.refreshTokenValue = resp.refreshToken;
    return resp;
  }

  async listExperiments(
    limit = 20,
    offset = 0,
  ): Promise<PaginatedExperiments> {
    const params = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    return this.request<PaginatedExperiments>(
      "GET",
      `/api/v1/experiments?${params}`,
    );
  }

  async createExperiment(
    req: CreateExperimentRequest,
  ): Promise<ExperimentResponse> {
    return this.request<ExperimentResponse>(
      "POST",
      "/api/v1/experiments",
      req,
    );
  }

  async getExperiment(name: string): Promise<ExperimentResponse> {
    return this.request<ExperimentResponse>(
      "GET",
      `/api/v1/experiments/${encodeURIComponent(name)}`,
    );
  }

  async deleteExperiment(name: string): Promise<void> {
    await this.request(
      "DELETE",
      `/api/v1/experiments/${encodeURIComponent(name)}`,
    );
  }

  async abortExperiment(name: string): Promise<ExperimentResponse> {
    return this.request<ExperimentResponse>(
      "POST",
      `/api/v1/experiments/${encodeURIComponent(name)}/abort`,
    );
  }

  async listHierarchy(): Promise<HierarchyResponse> {
    return this.request<HierarchyResponse>("GET", "/api/v1/hierarchy");
  }

  async getBillingStatus(): Promise<BillingStatusResponse> {
    return this.request<BillingStatusResponse>("GET", "/api/v1/billing");
  }

  private async request<T = void>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const headers: Record<string, string> = {
      Accept: "application/json",
    };

    if (this.apiKey) {
      headers["X-API-Key"] = this.apiKey;
    } else if (this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    const init: RequestInit = { method, headers };

    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
      init.body = JSON.stringify(body);
    }

    const resp = await this.fetchFn(`${this.baseUrl}${path}`, init);

    if (!resp.ok) {
      let message = resp.statusText;
      let parsed: unknown;
      try {
        parsed = await resp.json();
        if (
          typeof parsed === "object" &&
          parsed !== null &&
          "error" in parsed
        ) {
          message = (parsed as { error: string }).error;
        }
      } catch {
        /* non-JSON error body */
      }
      throw new ChaosPlaneError(resp.status, message, parsed);
    }

    if (resp.status === 204 || resp.headers.get("content-length") === "0") {
      return undefined as T;
    }

    return (await resp.json()) as T;
  }
}

export default ChaosPlaneClient;
