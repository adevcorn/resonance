package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsRegistry = prometheus.NewRegistry()
	httpHandler     = promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path"},
	)

	orchestrationStartCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ensemble_orchestration_starts_total",
			Help: "Total number of orchestration starts",
		},
	)

	orchestrationCompleteCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_orchestration_completes_total",
			Help: "Total number of orchestration completions",
		},
		[]string{"status"},
	)

	turnCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_turns_total",
			Help: "Total number of agent turns",
		},
		[]string{"agent", "status"},
	)

	turnDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_turn_duration_seconds",
			Help:    "Agent turn duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"agent"},
	)

	tokensTotalPerSession = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_tokens_total_per_session",
			Help: "Total tokens used per session",
		},
		[]string{"session_id"},
	)

	tokensInputPerTurn = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_tokens_input_per_turn",
			Help:    "Tokens input per turn",
			Buckets: []float64{10, 50, 100, 500, 1000},
		},
		[]string{"agent"},
	)

	tokensOutputPerTurn = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_tokens_output_per_turn",
			Help:    "Tokens output per turn",
			Buckets: []float64{10, 50, 100, 500, 1000},
		},
		[]string{"agent"},
	)

	tokensPerMessage = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_tokens_per_message",
			Help:    "Tokens per message",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		},
		[]string{"role", "agent"},
	)

	contextSizePerTurn = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ensemble_context_size_messages",
			Help: "Number of messages in context per turn",
		},
		[]string{"agent"},
	)

	messageCountPerTurn = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ensemble_message_count_per_turn",
			Help: "Total messages in conversation per turn",
		},
		[]string{"agent"},
	)

	contextSizeTokensPerTurn = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ensemble_context_size_tokens",
			Help: "Estimated token count in context per turn",
		},
		[]string{"agent"},
	)

	systemPromptSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_system_prompt_size_tokens",
			Help:    "System prompt size in tokens",
			Buckets: []float64{10, 50, 100, 250, 500, 1000},
		},
		[]string{"agent"},
	)

	teamAssemblyDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_team_assembly_duration_seconds",
			Help:    "Team assembly duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{},
	)

	moderatorDecisionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_moderator_decision_duration_seconds",
			Help:    "Moderator decision duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05},
		},
		[]string{},
	)

	toolExecutionCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_tool_executions_total",
			Help: "Total number of tool executions",
		},
		[]string{"tool", "side", "status"},
	)

	toolExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_tool_execution_duration_seconds",
			Help:    "Tool execution duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"tool", "side"},
	)

	sessionCreationCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ensemble_sessions_created_total",
			Help: "Total number of sessions created",
		},
	)

	sessionCompletionCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_sessions_completed_total",
			Help: "Total number of sessions completed",
		},
		[]string{"status"},
	)

	collaborationMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ensemble_collaboration_messages_total",
			Help: "Total number of collaboration messages",
		},
		[]string{"sender", "receiver"},
	)

	collaborationMessageDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_collaboration_message_duration_seconds",
			Help:    "Collaboration message duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5},
		},
		[]string{"sender"},
	)

	tokenEfficiency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ensemble_token_efficiency_ratio",
			Help:    "Tokens output per input token ratio",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"agent"},
	)
)

func InitializeMetrics(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("Metrics server listening at :9090/metrics")
			httpHandler.ServeHTTP(nil, nil)
		}
	}()
}

func GetMetricsHandler() http.Handler {
	return httpHandler
}

func RecordHTTPRequest(method, path string, status int, duration float64) {
	requestCounter.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
	requestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordOrchestrationStart() {
	orchestrationStartCounter.Inc()
}

func RecordOrchestrationComplete(status string) {
	orchestrationCompleteCounter.WithLabelValues(status).Inc()
}

func RecordTurn(agent, status string, duration float64, inputTokens, outputTokens float64) {
	turnCounter.WithLabelValues(agent, status).Inc()
	turnDuration.WithLabelValues(agent).Observe(duration)
	tokensInputPerTurn.WithLabelValues(agent).Observe(inputTokens)
	tokensOutputPerTurn.WithLabelValues(agent).Observe(outputTokens)
	tokensPerMessage.WithLabelValues("input", agent).Observe(inputTokens)
	tokensPerMessage.WithLabelValues("output", agent).Observe(outputTokens)
}

func RecordTokensPerSession(sessionID string, tokens int) {
	tokensTotalPerSession.WithLabelValues(sessionID).Add(float64(tokens))
}

func RecordContextSize(agent string, messageCount int, estimatedTokens float64) {
	contextSizePerTurn.WithLabelValues(agent).Set(float64(messageCount))
	messageCountPerTurn.WithLabelValues(agent).Set(float64(messageCount))
	contextSizeTokensPerTurn.WithLabelValues(agent).Set(estimatedTokens)
}

func RecordSystemPromptSize(agent string, size int) {
	systemPromptSize.WithLabelValues(agent).Observe(float64(size))
}

func RecordTeamAssemblyDuration(duration float64) {
	teamAssemblyDuration.WithLabelValues().Observe(duration)
}

func RecordModeratorDecisionDuration(duration float64) {
	moderatorDecisionDuration.WithLabelValues().Observe(duration)
}

func RecordToolExecution(tool, side, status string, duration float64) {
	toolExecutionCounter.WithLabelValues(tool, side, status).Inc()
	toolExecutionDuration.WithLabelValues(tool, side).Observe(duration)
}

func RecordSessionCreation(status string) {
	if status == "" {
		status = "unknown"
	}
	sessionCreationCounter.Inc()
	sessionCompletionCounter.WithLabelValues(status).Inc()
}

func RecordCollaborationMessage(sender, receiver string, duration float64) {
	collaborationMessagesTotal.WithLabelValues(sender, receiver).Inc()
	collaborationMessageDuration.WithLabelValues(sender).Observe(duration)
}

func RecordTokenEfficiency(agent string, inputTokens, outputTokens int) {
	if inputTokens > 0 {
		tokens := float64(outputTokens) / float64(inputTokens)
		tokenEfficiency.WithLabelValues(agent).Observe(tokens)
	}
}

func EstimateTokenCount(text string) float64 {
	if text == "" {
		return 0
	}
	charCount := len(text)
	return float64(charCount) / 4.0
}

func TokenCountToEstimate(text string) int {
	return int(EstimateTokenCount(text))
}
