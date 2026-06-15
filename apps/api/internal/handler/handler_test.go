package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"", 20, 0},
		{"limit=50&offset=10", 50, 10},
		{"limit=200", 100, 0},
		{"limit=-1", 20, 0},
		{"offset=-5", 20, 0},
		{"limit=abc", 20, 0},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)

		limit, offset := parsePagination(c)
		if limit != tt.wantLimit {
			t.Errorf("query=%q: expected limit %d, got %d", tt.query, tt.wantLimit, limit)
		}
		if offset != tt.wantOffset {
			t.Errorf("query=%q: expected offset %d, got %d", tt.query, tt.wantOffset, offset)
		}
	}
}
