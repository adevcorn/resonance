package api

import (
	"net/http"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/gorilla/mux"
)

// AgentSummary is a summary of an agent
type AgentSummary struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}

// AgentDetail is detailed information about an agent
type AgentDetail struct {
	AgentSummary
	SystemPrompt string               `json:"system_prompt"`
	Model        protocol.ModelConfig `json:"model"`
	Tools        protocol.ToolsConfig `json:"tools"`
}

// handleListAgents lists all available agents
func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.agentPool.GetAll()

	var summaries []AgentSummary
	for _, agent := range agents {
		def := agent.Definition()
		summaries = append(summaries, AgentSummary{
			Name:         def.Name,
			DisplayName:  def.DisplayName,
			Description:  def.Description,
			Capabilities: def.Capabilities,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"agents": summaries,
	})
}

// handleGetAgent retrieves details for a specific agent
func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	agent, err := s.agentPool.Get(name)
	if err != nil {
		respondError(w, http.StatusNotFound, "Agent not found")
		return
	}

	def := agent.Definition()

	detail := AgentDetail{
		AgentSummary: AgentSummary{
			Name:         def.Name,
			DisplayName:  def.DisplayName,
			Description:  def.Description,
			Capabilities: def.Capabilities,
		},
		SystemPrompt: def.SystemPrompt,
		Model:        def.Model,
		Tools:        def.Tools,
	}

	respondJSON(w, http.StatusOK, detail)
}
