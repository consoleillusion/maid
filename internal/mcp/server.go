// Package mcp is maid's MCP server: JSON-RPC 2.0 over newline-delimited
// stdio, with tools only. See design/implementation/interfaces.md.
package mcp

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// ProtocolVersions the server accepts, newest first.
var ProtocolVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

// Version of maid reported to clients.
var Version = "0.1.0-dev"

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Server serves MCP over a reader and a writer.
type Server struct {
	Tools       *Tools
	mu          sync.Mutex
	out         *bufio.Writer
	clientName  string
	session     string
	lastPhase   string
	listChanged bool
}

// NewServer returns a server for the given tools.
func NewServer(t *Tools) *Server {
	// Purpose: create a server with a random session ID for agent identity.
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return &Server{Tools: t, session: hex.EncodeToString(b)}
}

// Serve reads requests until in is closed.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	// Purpose: run the JSON-RPC loop on stdio.
	defer s.Tools.Close()
	s.out = bufio.NewWriter(out)
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			s.send(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{-32700, "parse error"}})
			continue
		}
		s.handle(ctx, req)
	}
	return sc.Err()
}

func (s *Server) handle(ctx context.Context, req request) {
	// Purpose: dispatch one request or notification.
	isNotification := len(req.ID) == 0
	var result any
	var rerr *rpcError
	switch req.Method {
	case "initialize":
		result = s.initialize(req.Params)
	case "notifications/initialized", "notifications/cancelled":
		return
	case "ping":
		result = map[string]any{}
	case "tools/list":
		result = map[string]any{"tools": s.Tools.List(ctx)}
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			rerr = &rpcError{-32602, "invalid params"}
			break
		}
		result = s.Tools.Call(ctx, p.Name, p.Arguments, s.defaultAgent())
		defer s.maybeNotifyListChanged(ctx)
	default:
		rerr = &rpcError{-32601, "method not found: " + req.Method}
	}
	if isNotification {
		return
	}
	s.send(response{JSONRPC: "2.0", ID: req.ID, Result: result, Error: rerr})
}

func (s *Server) initialize(params json.RawMessage) any {
	// Purpose: negotiate the protocol version and send the server instructions.
	var p struct {
		ProtocolVersion string `json:"protocolVersion"`
		ClientInfo      struct {
			Name string `json:"name"`
		} `json:"clientInfo"`
		Capabilities map[string]any `json:"capabilities"`
	}
	_ = json.Unmarshal(params, &p)
	s.clientName = p.ClientInfo.Name
	version := ProtocolVersions[0]
	for _, v := range ProtocolVersions {
		if v == p.ProtocolVersion {
			version = v
		}
	}
	phase := s.Tools.Phase(context.Background())
	s.lastPhase = phase
	return map[string]any{
		"protocolVersion": version,
		"capabilities":    map[string]any{"tools": map[string]any{"listChanged": true}},
		"serverInfo":      map[string]any{"name": "maid", "version": Version},
		"instructions": fmt.Sprintf("maid (always written in lowercase) guides this project from design to accepted code. Current phase: %s. "+
			"Call maid_next before anything else. It returns the step, the questions to ask the user, and ready-to-run tool calls in \"calls\". "+
			"Whenever the user asks you to work on this project, call maid_next and work on the tasks it lists. "+
			"Use maid's tools to read and change the design; you do not need file access for the design phases. "+
			"Ask the user everything in ask_user in ONE message, never one question at a time, and act on all the answers together. "+
			"design/ is the source of truth; read only the sections you need. Never guess missing information: ask the user instead.", phase),
	}
}

func (s *Server) maybeNotifyListChanged(ctx context.Context) {
	// Purpose: tell the client when the phase changes the tool list.
	phase := s.Tools.Phase(ctx)
	if phase != s.lastPhase {
		s.lastPhase = phase
		s.send(notification{JSONRPC: "2.0", Method: "notifications/tools/list_changed"})
	}
}

func (s *Server) defaultAgent() string {
	// Purpose: derive an agent ID from the client name and session.
	name := s.clientName
	if name == "" {
		name = "agent"
	}
	return name + "-" + s.session
}

func (s *Server) send(v any) {
	// Purpose: write one JSON-RPC message followed by a newline.
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		b, _ = json.Marshal(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{-32603, err.Error()}})
	}
	_, _ = s.out.Write(append(b, '\n'))
	_ = s.out.Flush()
}
