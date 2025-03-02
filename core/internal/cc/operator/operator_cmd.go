package operator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jm33-m0/emp3r0r/core/internal/def"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/posener/h2conn"
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
	req.Header.Add("operator_session", OPERATOR_SESSION)

	resp, err := OperatorHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, status code: %d, url: %s, request body: %v", resp.StatusCode, url, data)
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
	resp, err := sendJSONRequest(url, operation)
	if err != nil {
		logging.Errorf("Failed to set active agent: %v", err)
	}

	err = json.Unmarshal(resp, live.ActiveAgent)
	if err != nil {
		logging.Errorf("Failed to unmarshal active agent: %v", err)
	}
}

func cmdListAgents(_ *cobra.Command, _ []string) {
	err := refreshAgentList()
	if err != nil {
		logging.Errorf("Failed to list agents: %v", err)
		return
	}
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

// connectMsgTun connects to the operator message tunnel
func connectMsgTun() (conn *h2conn.Conn, ctx context.Context, cancel context.CancelFunc, err error) {
	h2 := h2conn.Client{
		Client: OperatorHTTPClient,
		Header: http.Header{
			"operator_session": {OPERATOR_SESSION},
		},
	}
	url := fmt.Sprintf("%s/%s", OperatorRootURL, transport.OperatorMsgTunnel)
	ctx, cancel = context.WithCancel(context.Background())
	conn, resp, err := h2.Connect(ctx, url)
	if err != nil {
		err = fmt.Errorf("connect to message tunnel: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status code: %d", resp.StatusCode)
		return
	}
	logging.Successf("Connected to %s, session ID is %s", url, OPERATOR_SESSION)

	return
}

func msgTunHandler() {
	time.Sleep(3 * time.Second)
	conn, ctx, cancel, err := connectMsgTun()
	if err != nil {
		logging.Errorf("Failed to connect to message tunnel: %v", err)
		return
	}
	defer cancel()

	decoder := json.NewDecoder(bufio.NewReader(conn)) // Buffered reader to prevent partial reads

	// Create a ticker to simulate heartbeat checks every second
	heartbeatTicker := time.NewTicker(1 * time.Second)
	defer heartbeatTicker.Stop()

	// Create a timeout timer for 1 minute (60 seconds)
	timeoutTimer := time.NewTimer(1 * time.Minute)
	defer timeoutTimer.Stop()

	// Channel to track the latest heartbeat
	heartbeatCh := make(chan struct{})

	// Goroutine to monitor the heartbeat and handle the timeout
	go func() {
		for {
			select {
			case <-heartbeatTicker.C:
				// If no heartbeat received in the last minute, close the connection
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
					logging.Warningf("Message tunnel heartbeat timeout, closing connection")
					conn.Close()
					cancel()
					return
				}
				// Reset the timeout timer after receiving a heartbeat
				timeoutTimer.Reset(1 * time.Minute)
			case <-heartbeatCh:
				// Heartbeat received, reset the timeout
				timeoutTimer.Reset(1 * time.Minute)
			}
		}
	}()

	// Keep reading messages from the tunnel
	for ctx.Err() == nil {
		msg := new(def.MsgTunData)
		if err := decoder.Decode(msg); err != nil {
			if errors.Is(err, io.EOF) {
				logging.Warningf("Message tunnel closed")
				return
			}
			logging.Errorf("Failed to decode message: %v", err)
			continue
		}
		logging.Debugf("Message from operator: %v", *msg)

		// Reset the heartbeat timer after receiving a valid message
		heartbeatCh <- struct{}{}

		processAgentData(msg)
	}
}
