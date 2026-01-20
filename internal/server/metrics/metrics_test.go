package metrics

import (
	"testing"
)

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 1.25, // 5 chars / 4
		},
		{
			name:     "longer text",
			text:     "This is a test message",
			expected: 5.5, // 22 chars / 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokenCount(tt.text)
			if result != tt.expected {
				t.Errorf("EstimateTokenCount(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestTokenCountToEstimate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 1, // 5 chars / 4 = 1.25 -> 1
		},
		{
			name:     "longer text",
			text:     "This is a test message",
			expected: 5, // 22 chars / 4 = 5.5 -> 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokenCountToEstimate(tt.text)
			if result != tt.expected {
				t.Errorf("TokenCountToEstimate(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestRecordMetrics(t *testing.T) {
	// Test that metrics recording doesn't panic
	t.Run("RecordHTTPRequest", func(t *testing.T) {
		RecordHTTPRequest("GET", "/api/health", 200, 0.1)
	})

	t.Run("RecordOrchestrationStart", func(t *testing.T) {
		RecordOrchestrationStart()
	})

	t.Run("RecordOrchestrationComplete", func(t *testing.T) {
		RecordOrchestrationComplete("success")
	})

	t.Run("RecordTurn", func(t *testing.T) {
		RecordTurn("developer", "success", 1.5, 100.0, 50.0)
	})

	t.Run("RecordTokensPerSession", func(t *testing.T) {
		RecordTokensPerSession("session-123", 500)
	})

	t.Run("RecordContextSize", func(t *testing.T) {
		RecordContextSize("developer", 10, 500.0)
	})

	t.Run("RecordSystemPromptSize", func(t *testing.T) {
		RecordSystemPromptSize("developer", 200)
	})

	t.Run("RecordTeamAssemblyDuration", func(t *testing.T) {
		RecordTeamAssemblyDuration(0.5)
	})

	t.Run("RecordModeratorDecisionDuration", func(t *testing.T) {
		RecordModeratorDecisionDuration(0.05)
	})

	t.Run("RecordToolExecution", func(t *testing.T) {
		RecordToolExecution("read_file", "client", "success", 0.2)
	})

	t.Run("RecordSessionCreation", func(t *testing.T) {
		RecordSessionCreation("active")
	})

	t.Run("RecordCollaborationMessage", func(t *testing.T) {
		RecordCollaborationMessage("developer", "architect", 0.1)
	})

	t.Run("RecordTokenEfficiency", func(t *testing.T) {
		RecordTokenEfficiency("developer", 100, 50)
	})
}
