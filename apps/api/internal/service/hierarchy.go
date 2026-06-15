package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var ErrHierarchyNotFound = errors.New("resource not found")

type HierarchyService struct {
	pool *database.Pool
}

func NewHierarchyService(pool *database.Pool) *HierarchyService {
	return &HierarchyService{pool: pool}
}

type ActorContext struct {
	UserID   string
	TenantID string
}

type Organization struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Workspace struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description,omitempty"`
}

type Team struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Project struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Environment struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	AgentStatus string `json:"agentStatus"`
}

type HierarchyResponse struct {
	Organizations []Organization `json:"organizations"`
	Workspaces    []Workspace    `json:"workspaces"`
	Teams         []Team         `json:"teams"`
	Projects      []Project      `json:"projects"`
	Environments  []Environment  `json:"environments"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateWorkspaceRequest struct {
	OrganizationID string `json:"organizationId" binding:"required,uuid"`
	Name           string `json:"name" binding:"required,min=2"`
	Slug           string `json:"slug,omitempty"`
	Description    string `json:"description,omitempty"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateTeamRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required,uuid"`
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateProjectRequest struct {
	WorkspaceID string `json:"workspaceId" binding:"required,uuid"`
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreateEnvironmentRequest struct {
	ProjectID string `json:"projectId" binding:"required,uuid"`
	Name      string `json:"name" binding:"required,min=2"`
	Slug      string `json:"slug,omitempty"`
	Type      string `json:"type,omitempty"`
}

type UpdateEnvironmentRequest struct {
	Name string `json:"name" binding:"required,min=2"`
	Slug string `json:"slug,omitempty"`
	Type string `json:"type,omitempty"`
}

func (s *HierarchyService) List(ctx context.Context, actor ActorContext) (*HierarchyResponse, error) {
	if err := s.ensureMembership(ctx, actor); err != nil {
		return nil, err
	}

	orgs, err := s.listOrganizations(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}
	workspaces, err := s.listWorkspaces(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}
	teams, err := s.listTeams(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}
	projects, err := s.listProjects(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}
	envs, err := s.listEnvironments(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}

	return &HierarchyResponse{Organizations: orgs, Workspaces: workspaces, Teams: teams, Projects: projects, Environments: envs}, nil
}

func (s *HierarchyService) CreateOrganization(ctx context.Context, actor ActorContext, req *CreateOrganizationRequest) (*Organization, error) {
	if err := s.ensureMembership(ctx, actor); err != nil {
		return nil, err
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO organizations (tenant_id, name, slug, description)
		VALUES ($1::uuid, $2, $3, NULLIF($4, ''))
		RETURNING id::text, tenant_id::text, name, slug, COALESCE(description, '')
	`, actor.TenantID, req.Name, slug, req.Description)
	return scanOrganization(row)
}

func (s *HierarchyService) UpdateOrganization(ctx context.Context, actor ActorContext, id string, req *UpdateOrganizationRequest) (*Organization, error) {
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE organizations
		SET name = $3, slug = $4, description = NULLIF($5, ''), updated_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		RETURNING id::text, tenant_id::text, name, slug, COALESCE(description, '')
	`, id, actor.TenantID, req.Name, slug, req.Description)
	return scanOrganization(row)
}

func (s *HierarchyService) CreateWorkspace(ctx context.Context, actor ActorContext, req *CreateWorkspaceRequest) (*Workspace, error) {
	if err := s.ensureOrgTenant(ctx, actor.TenantID, req.OrganizationID); err != nil {
		return nil, err
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO workspaces (tenant_id, organization_id, name, slug, description)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''))
		RETURNING id::text, tenant_id::text, organization_id::text, name, slug, COALESCE(description, '')
	`, actor.TenantID, req.OrganizationID, req.Name, slug, req.Description)
	return scanWorkspace(row)
}

func (s *HierarchyService) UpdateWorkspace(ctx context.Context, actor ActorContext, id string, req *UpdateWorkspaceRequest) (*Workspace, error) {
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE workspaces
		SET name = $3, slug = $4, description = NULLIF($5, ''), updated_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		RETURNING id::text, tenant_id::text, organization_id::text, name, slug, COALESCE(description, '')
	`, id, actor.TenantID, req.Name, slug, req.Description)
	return scanWorkspace(row)
}

func (s *HierarchyService) CreateTeam(ctx context.Context, actor ActorContext, req *CreateTeamRequest) (*Team, error) {
	if err := s.ensureWorkspaceTenant(ctx, actor.TenantID, req.WorkspaceID); err != nil {
		return nil, err
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO teams (tenant_id, workspace_id, name, slug, description)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''))
		RETURNING id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
	`, actor.TenantID, req.WorkspaceID, req.Name, slug, req.Description)
	return scanTeam(row)
}

func (s *HierarchyService) UpdateTeam(ctx context.Context, actor ActorContext, id string, req *UpdateTeamRequest) (*Team, error) {
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE teams
		SET name = $3, slug = $4, description = NULLIF($5, ''), updated_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		RETURNING id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
	`, id, actor.TenantID, req.Name, slug, req.Description)
	return scanTeam(row)
}

func (s *HierarchyService) CreateProject(ctx context.Context, actor ActorContext, req *CreateProjectRequest) (*Project, error) {
	if err := s.ensureWorkspaceTenant(ctx, actor.TenantID, req.WorkspaceID); err != nil {
		return nil, err
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO projects (tenant_id, workspace_id, name, slug, description)
		VALUES ($1::uuid, $2::uuid, $3, $4, NULLIF($5, ''))
		RETURNING id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
	`, actor.TenantID, req.WorkspaceID, req.Name, slug, req.Description)
	return scanProject(row)
}

func (s *HierarchyService) UpdateProject(ctx context.Context, actor ActorContext, id string, req *UpdateProjectRequest) (*Project, error) {
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE projects
		SET name = $3, slug = $4, description = NULLIF($5, ''), updated_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		RETURNING id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
	`, id, actor.TenantID, req.Name, slug, req.Description)
	return scanProject(row)
}

func (s *HierarchyService) CreateEnvironment(ctx context.Context, actor ActorContext, req *CreateEnvironmentRequest) (*Environment, error) {
	if err := s.ensureProjectTenant(ctx, actor.TenantID, req.ProjectID); err != nil {
		return nil, err
	}
	typeValue := req.Type
	if typeValue == "" {
		typeValue = "staging"
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO environments (tenant_id, project_id, name, slug, type)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING id::text, tenant_id::text, project_id::text, name, slug, type, agent_status
	`, actor.TenantID, req.ProjectID, req.Name, slug, typeValue)
	return scanEnvironment(row)
}

func (s *HierarchyService) UpdateEnvironment(ctx context.Context, actor ActorContext, id string, req *UpdateEnvironmentRequest) (*Environment, error) {
	typeValue := req.Type
	if typeValue == "" {
		typeValue = "staging"
	}
	slug := defaultSlug(req.Slug, req.Name)
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE environments
		SET name = $3, slug = $4, type = $5, updated_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		RETURNING id::text, tenant_id::text, project_id::text, name, slug, type, agent_status
	`, id, actor.TenantID, req.Name, slug, typeValue)
	return scanEnvironment(row)
}

func (s *HierarchyService) ensureMembership(ctx context.Context, actor ActorContext) error {
	var exists bool
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_tenants WHERE user_id = $1::uuid AND tenant_id = $2::uuid
		)
	`, actor.UserID, actor.TenantID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check tenant membership: %w", err)
	}
	if !exists {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *HierarchyService) ensureOrgTenant(ctx context.Context, tenantID, orgID string) error {
	return ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM organizations WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, orgID, tenantID)
}

func (s *HierarchyService) ensureWorkspaceTenant(ctx context.Context, tenantID, workspaceID string) error {
	return ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM workspaces WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, workspaceID, tenantID)
}

func (s *HierarchyService) ensureProjectTenant(ctx context.Context, tenantID, projectID string) error {
	return ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM projects WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, projectID, tenantID)
}

func ensureTenantScopedRow(ctx context.Context, pool *database.Pool, query, id, tenantID string) error {
	var marker int
	err := pool.Conn(ctx).QueryRow(ctx, query, id, tenantID).Scan(&marker)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrHierarchyNotFound
		}
		return err
	}
	return nil
}

func (s *HierarchyService) listOrganizations(ctx context.Context, tenantID string) ([]Organization, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, name, slug, COALESCE(description, '')
		FROM organizations
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var items []Organization
	for rows.Next() {
		var item Organization
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Name, &item.Slug, &item.Description); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		items = append(items, item)
	}
	return nonNilOrganizations(items), rows.Err()
}

func (s *HierarchyService) listWorkspaces(ctx context.Context, tenantID string) ([]Workspace, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, organization_id::text, name, slug, COALESCE(description, '')
		FROM workspaces
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()
	var items []Workspace
	for rows.Next() {
		var item Workspace
		if err := rows.Scan(&item.ID, &item.TenantID, &item.OrganizationID, &item.Name, &item.Slug, &item.Description); err != nil {
			return nil, fmt.Errorf("scan workspace: %w", err)
		}
		items = append(items, item)
	}
	return nonNilWorkspaces(items), rows.Err()
}

func (s *HierarchyService) listTeams(ctx context.Context, tenantID string) ([]Team, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
		FROM teams
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()
	var items []Team
	for rows.Next() {
		var item Team
		if err := rows.Scan(&item.ID, &item.TenantID, &item.WorkspaceID, &item.Name, &item.Slug, &item.Description); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		items = append(items, item)
	}
	return nonNilTeams(items), rows.Err()
}

func (s *HierarchyService) listProjects(ctx context.Context, tenantID string) ([]Project, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, workspace_id::text, name, slug, COALESCE(description, '')
		FROM projects
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var items []Project
	for rows.Next() {
		var item Project
		if err := rows.Scan(&item.ID, &item.TenantID, &item.WorkspaceID, &item.Name, &item.Slug, &item.Description); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		items = append(items, item)
	}
	return nonNilProjects(items), rows.Err()
}

func (s *HierarchyService) listEnvironments(ctx context.Context, tenantID string) ([]Environment, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, project_id::text, name, slug, type, agent_status
		FROM environments
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	defer rows.Close()
	var items []Environment
	for rows.Next() {
		var item Environment
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ProjectID, &item.Name, &item.Slug, &item.Type, &item.AgentStatus); err != nil {
			return nil, fmt.Errorf("scan environment: %w", err)
		}
		items = append(items, item)
	}
	return nonNilEnvironments(items), rows.Err()
}

func scanOrganization(row pgx.Row) (*Organization, error) {
	var item Organization
	if err := row.Scan(&item.ID, &item.TenantID, &item.Name, &item.Slug, &item.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanWorkspace(row pgx.Row) (*Workspace, error) {
	var item Workspace
	if err := row.Scan(&item.ID, &item.TenantID, &item.OrganizationID, &item.Name, &item.Slug, &item.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanTeam(row pgx.Row) (*Team, error) {
	var item Team
	if err := row.Scan(&item.ID, &item.TenantID, &item.WorkspaceID, &item.Name, &item.Slug, &item.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanProject(row pgx.Row) (*Project, error) {
	var item Project
	if err := row.Scan(&item.ID, &item.TenantID, &item.WorkspaceID, &item.Name, &item.Slug, &item.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanEnvironment(row pgx.Row) (*Environment, error) {
	var item Environment
	if err := row.Scan(&item.ID, &item.TenantID, &item.ProjectID, &item.Name, &item.Slug, &item.Type, &item.AgentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func defaultSlug(slug, name string) string {
	if trimmed := strings.TrimSpace(slug); trimmed != "" {
		return slugify(trimmed)
	}
	return slugify(name)
}

func nonNilOrganizations(items []Organization) []Organization {
	if items == nil {
		return []Organization{}
	}
	return items
}

func nonNilWorkspaces(items []Workspace) []Workspace {
	if items == nil {
		return []Workspace{}
	}
	return items
}

func nonNilTeams(items []Team) []Team {
	if items == nil {
		return []Team{}
	}
	return items
}

func nonNilProjects(items []Project) []Project {
	if items == nil {
		return []Project{}
	}
	return items
}

func nonNilEnvironments(items []Environment) []Environment {
	if items == nil {
		return []Environment{}
	}
	return items
}
