package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type GameDayService struct {
	pool *database.Pool
}

func NewGameDayService(pool *database.Pool) *GameDayService {
	return &GameDayService{pool: pool}
}

type GameDay struct {
	ID            string     `json:"id"`
	EnvironmentID string     `json:"environmentId"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	Status        string     `json:"status"`
	ScheduledAt   *time.Time `json:"scheduledAt,omitempty"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
	CreatedBy     string     `json:"createdBy"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type GameDayEvent struct {
	ID          string          `json:"id"`
	EventType   string          `json:"eventType"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	UserID      *string         `json:"userId,omitempty"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type GameDayPostmortem struct {
	ID            string          `json:"id"`
	Summary       string          `json:"summary"`
	WhatWentWell  *string         `json:"whatWentWell,omitempty"`
	WhatWentWrong *string         `json:"whatWentWrong,omitempty"`
	ActionItems   json.RawMessage `json:"actionItems"`
	CreatedBy     string          `json:"createdBy"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type GameDayListResponse struct {
	Items []GameDay `json:"items"`
}

type GameDayDetailResponse struct {
	GameDay    GameDay            `json:"gameday"`
	Events     []GameDayEvent     `json:"events"`
	Postmortem *GameDayPostmortem `json:"postmortem,omitempty"`
}

type CreateGameDayRequest struct {
	EnvironmentID string     `json:"environmentId" binding:"required,uuid"`
	Title         string     `json:"title" binding:"required"`
	Description   *string    `json:"description,omitempty"`
	ScheduledAt   *time.Time `json:"scheduledAt,omitempty"`
}

type UpdateGameDayStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AddGameDayEventRequest struct {
	EventType   string          `json:"eventType" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Description *string         `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type CreatePostmortemRequest struct {
	Summary       string          `json:"summary" binding:"required"`
	WhatWentWell  *string         `json:"whatWentWell,omitempty"`
	WhatWentWrong *string         `json:"whatWentWrong,omitempty"`
	ActionItems   json.RawMessage `json:"actionItems,omitempty"`
}

func (s *GameDayService) List(ctx context.Context, actor ActorContext) (*GameDayListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, environment_id::text, title, description, status, scheduled_at, started_at, ended_at, created_by::text, created_at
		FROM gamedays WHERE tenant_id = $1::uuid ORDER BY created_at DESC LIMIT 50
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list gamedays: %w", err)
	}
	defer rows.Close()
	items := []GameDay{}
	for rows.Next() {
		var g GameDay
		if err := rows.Scan(&g.ID, &g.EnvironmentID, &g.Title, &g.Description, &g.Status, &g.ScheduledAt, &g.StartedAt, &g.EndedAt, &g.CreatedBy, &g.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return &GameDayListResponse{Items: items}, rows.Err()
}

func (s *GameDayService) Get(ctx context.Context, actor ActorContext, gamedayID string) (*GameDayDetailResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var g GameDay
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, environment_id::text, title, description, status, scheduled_at, started_at, ended_at, created_by::text, created_at
		FROM gamedays WHERE id = $1::uuid AND tenant_id = $2::uuid
	`, gamedayID, actor.TenantID).Scan(&g.ID, &g.EnvironmentID, &g.Title, &g.Description, &g.Status, &g.ScheduledAt, &g.StartedAt, &g.EndedAt, &g.CreatedBy, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}

	eventRows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, event_type, title, description, user_id::text, metadata, created_at
		FROM gameday_events WHERE gameday_id = $1::uuid ORDER BY created_at ASC
	`, gamedayID)
	if err != nil {
		return nil, err
	}
	defer eventRows.Close()
	events := []GameDayEvent{}
	for eventRows.Next() {
		var e GameDayEvent
		if eventRows.Scan(&e.ID, &e.EventType, &e.Title, &e.Description, &e.UserID, &e.Metadata, &e.CreatedAt) == nil {
			events = append(events, e)
		}
	}

	var postmortem *GameDayPostmortem
	var pm GameDayPostmortem
	err = s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, summary, what_went_well, what_went_wrong, action_items, created_by::text, created_at
		FROM gameday_postmortems WHERE gameday_id = $1::uuid
	`, gamedayID).Scan(&pm.ID, &pm.Summary, &pm.WhatWentWell, &pm.WhatWentWrong, &pm.ActionItems, &pm.CreatedBy, &pm.CreatedAt)
	if err == nil {
		postmortem = &pm
	}

	return &GameDayDetailResponse{GameDay: g, Events: events, Postmortem: postmortem}, nil
}

func (s *GameDayService) Create(ctx context.Context, actor ActorContext, req *CreateGameDayRequest) (*GameDay, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var g GameDay
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO gamedays (tenant_id, environment_id, title, description, scheduled_at, created_by)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::uuid)
		RETURNING id::text, environment_id::text, title, description, status, scheduled_at, started_at, ended_at, created_by::text, created_at
	`, actor.TenantID, req.EnvironmentID, req.Title, req.Description, req.ScheduledAt, actor.UserID).Scan(
		&g.ID, &g.EnvironmentID, &g.Title, &g.Description, &g.Status, &g.ScheduledAt, &g.StartedAt, &g.EndedAt, &g.CreatedBy, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create gameday: %w", err)
	}
	return &g, nil
}

func (s *GameDayService) UpdateStatus(ctx context.Context, actor ActorContext, gamedayID string, req *UpdateGameDayStatusRequest) (*GameDay, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var startedAt, endedAt any
	if req.Status == "running" {
		now := time.Now()
		startedAt = now
	}
	if req.Status == "completed" || req.Status == "cancelled" {
		now := time.Now()
		endedAt = now
	}

	var g GameDay
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE gamedays SET status = $3, started_at = COALESCE($4, started_at), ended_at = COALESCE($5, ended_at)
		WHERE id = $1::uuid AND tenant_id = $2::uuid
		RETURNING id::text, environment_id::text, title, description, status, scheduled_at, started_at, ended_at, created_by::text, created_at
	`, gamedayID, actor.TenantID, req.Status, startedAt, endedAt).Scan(
		&g.ID, &g.EnvironmentID, &g.Title, &g.Description, &g.Status, &g.ScheduledAt, &g.StartedAt, &g.EndedAt, &g.CreatedBy, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (s *GameDayService) AddEvent(ctx context.Context, actor ActorContext, gamedayID string, req *AddGameDayEventRequest) (*GameDayEvent, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	metadata := "{}"
	if len(req.Metadata) > 0 {
		metadata = string(req.Metadata)
	}
	var e GameDayEvent
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO gameday_events (gameday_id, event_type, title, description, user_id, metadata)
		VALUES ($1::uuid, $2, $3, $4, $5::uuid, $6::jsonb)
		RETURNING id::text, event_type, title, description, user_id::text, metadata, created_at
	`, gamedayID, req.EventType, req.Title, req.Description, actor.UserID, metadata).Scan(
		&e.ID, &e.EventType, &e.Title, &e.Description, &e.UserID, &e.Metadata, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add gameday event: %w", err)
	}
	return &e, nil
}

func (s *GameDayService) CreatePostmortem(ctx context.Context, actor ActorContext, gamedayID string, req *CreatePostmortemRequest) (*GameDayPostmortem, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	actionItems := "[]"
	if len(req.ActionItems) > 0 {
		actionItems = string(req.ActionItems)
	}
	var pm GameDayPostmortem
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO gameday_postmortems (gameday_id, summary, what_went_well, what_went_wrong, action_items, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5::jsonb, $6::uuid)
		RETURNING id::text, summary, what_went_well, what_went_wrong, action_items, created_by::text, created_at
	`, gamedayID, req.Summary, req.WhatWentWell, req.WhatWentWrong, actionItems, actor.UserID).Scan(
		&pm.ID, &pm.Summary, &pm.WhatWentWell, &pm.WhatWentWrong, &pm.ActionItems, &pm.CreatedBy, &pm.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create postmortem: %w", err)
	}
	return &pm, nil
}
