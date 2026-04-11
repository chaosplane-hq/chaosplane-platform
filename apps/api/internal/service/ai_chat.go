package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type AIChatService struct {
	pool *database.Pool
	llm  *LLMClient
}

func NewAIChatService(pool *database.Pool, llm *LLMClient) *AIChatService {
	return &AIChatService{pool: pool, llm: llm}
}

type ChatSession struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChatSessionListResponse struct {
	Items []ChatSession `json:"items"`
}

type ChatMessageListResponse struct {
	Items []ChatMessage `json:"items"`
}

type CreateChatSessionRequest struct {
	Title string `json:"title,omitempty"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type SendMessageResponse struct {
	UserMessage      ChatMessage `json:"userMessage"`
	AssistantMessage ChatMessage `json:"assistantMessage"`
}

func (s *AIChatService) ListSessions(ctx context.Context, actor ActorContext) (*ChatSessionListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, title, created_at, updated_at
		FROM ai_chat_sessions
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
		ORDER BY updated_at DESC
		LIMIT 50
	`, actor.TenantID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	defer rows.Close()

	items := []ChatSession{}
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat session: %w", err)
		}
		items = append(items, s)
	}
	return &ChatSessionListResponse{Items: items}, rows.Err()
}

func (s *AIChatService) CreateSession(ctx context.Context, actor ActorContext, req *CreateChatSessionRequest) (*ChatSession, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	title := req.Title
	if title == "" {
		title = "New Chat"
	}

	var session ChatSession
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO ai_chat_sessions (tenant_id, user_id, title)
		VALUES ($1::uuid, $2::uuid, $3)
		RETURNING id::text, title, created_at, updated_at
	`, actor.TenantID, actor.UserID, title).Scan(&session.ID, &session.Title, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create chat session: %w", err)
	}
	return &session, nil
}

func (s *AIChatService) GetMessages(ctx context.Context, actor ActorContext, sessionID string) (*ChatMessageListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := s.ensureSessionOwner(ctx, actor, sessionID); err != nil {
		return nil, err
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, role, content, created_at
		FROM ai_chat_messages
		WHERE session_id = $1::uuid
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}
	defer rows.Close()

	items := []ChatMessage{}
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		items = append(items, m)
	}
	return &ChatMessageListResponse{Items: items}, rows.Err()
}

func (s *AIChatService) SendMessage(ctx context.Context, actor ActorContext, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := s.ensureSessionOwner(ctx, actor, sessionID); err != nil {
		return nil, err
	}

	var userMsg ChatMessage
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO ai_chat_messages (session_id, role, content)
		VALUES ($1::uuid, 'user', $2)
		RETURNING id::text, role, content, created_at
	`, sessionID, req.Content).Scan(&userMsg.ID, &userMsg.Role, &userMsg.Content, &userMsg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("store user message: %w", err)
	}

	assistantContent, err := s.generateResponse(ctx, sessionID, req.Content)
	if err != nil {
		assistantContent = fmt.Sprintf("Error generating response: %v", err)
	}

	var assistantMsg ChatMessage
	err = s.pool.App.QueryRow(ctx, `
		INSERT INTO ai_chat_messages (session_id, role, content)
		VALUES ($1::uuid, 'assistant', $2)
		RETURNING id::text, role, content, created_at
	`, sessionID, assistantContent).Scan(&assistantMsg.ID, &assistantMsg.Role, &assistantMsg.Content, &assistantMsg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("store assistant message: %w", err)
	}

	_, _ = s.pool.App.Exec(ctx, `UPDATE ai_chat_sessions SET updated_at = now() WHERE id = $1::uuid`, sessionID)

	return &SendMessageResponse{UserMessage: userMsg, AssistantMessage: assistantMsg}, nil
}

func (s *AIChatService) DeleteSession(ctx context.Context, actor ActorContext, sessionID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		DELETE FROM ai_chat_sessions
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND user_id = $3::uuid
	`, sessionID, actor.TenantID, actor.UserID)
	if err != nil {
		return fmt.Errorf("delete chat session: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *AIChatService) ensureSessionOwner(ctx context.Context, actor ActorContext, sessionID string) error {
	var exists bool
	err := s.pool.App.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM ai_chat_sessions
			WHERE id = $1::uuid AND tenant_id = $2::uuid AND user_id = $3::uuid
		)
	`, sessionID, actor.TenantID, actor.UserID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check session ownership: %w", err)
	}
	if !exists {
		return ErrHierarchyNotFound
	}
	return nil
}

func generateAssistantResponse(userMessage string) string {
	return "AI assistant is not configured. Set LLM_API_KEY and LLM_PROVIDER environment variables to enable AI-powered responses. Supported providers: openai (default), anthropic."
}

func (s *AIChatService) generateResponse(ctx context.Context, sessionID, userMessage string) (string, error) {
	if s.llm == nil || !s.llm.IsConfigured() {
		return generateAssistantResponse(userMessage), nil
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT role, content FROM ai_chat_messages
		WHERE session_id = $1::uuid
		ORDER BY created_at ASC
		LIMIT 20
	`, sessionID)
	if err != nil {
		return generateAssistantResponse(userMessage), nil
	}
	defer rows.Close()

	var history []LLMMessage
	for rows.Next() {
		var role, content string
		if rows.Scan(&role, &content) == nil {
			history = append(history, LLMMessage{Role: role, Content: content})
		}
	}

	systemPrompt := `You are ChaosPlane AI Assistant, an expert in chaos engineering, resilience testing, and Kubernetes operations.
You help users:
- Analyze their cluster topology and identify vulnerabilities
- Suggest chaos experiments to test resilience
- Explain experiment results and recommend improvements
- Answer questions about chaos engineering best practices
Be concise, technical, and actionable.`

	return s.llm.Complete(ctx, systemPrompt, history)
}

var _ = errors.Is
var _ = pgx.ErrNoRows
