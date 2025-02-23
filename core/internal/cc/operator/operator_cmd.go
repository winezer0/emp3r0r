package operator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/spf13/cobra"
)

func sendJSONRequest(url string, data any) ([]byte, error) {
	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := OperatorHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// operatorSendCommand2Agent sends a command to an agent through the mTLS C2 operator server
func operatorSendCommand2Agent(cmd, cmdID, agentTag string) error {
	operation := def.Operation{
		AgentTag:  agentTag,
		Action:    "command",
		Command:   &cmd,
		CommandID: &cmdID,
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSendCommand)
	_, err := sendJSONRequest(url, operation)
	return err
}

func cmdSetActiveAgent(cmd *cobra.Command, args []string) {
	operation := def.Operation{
		AgentTag: args[0],
		Action:   "command",
	}

	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorSetActiveAgent)
	if _, err := sendJSONRequest(url, operation); err != nil {
		logging.Errorf("Failed to set active agent: %v", err)
	}
}

func cmdListAgents(_ *cobra.Command, _ []string) {
	err := getAgentListFromServer()
	if err != nil {
		logging.Errorf("Failed to list agents: %v", err)
		return
	}

	RenderAgentTable(live.AgentList)
	cli.TmuxSwitchWindow(cli.AgentListPane.WindowID)
}

func getAgentListFromServer() error {
	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorListConnectedAgents)
	body, err := sendJSONRequest(url, nil)
	if err != nil {
		return fmt.Errorf("failed to list agents: %v", err)
	}

	var agents []*def.Emp3r0rAgent
	if err := json.Unmarshal(body, &agents); err != nil {
		return fmt.Errorf("failed to unmarshal agents: %v", err)
	}
	live.AgentList = agents

	return nil
}
