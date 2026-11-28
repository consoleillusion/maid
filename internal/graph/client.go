package graph

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// Client is a bridge to the upstream code-review-graph MCP server.
type Client struct {
	cmd    *exec.Cmd
	in     io.WriteCloser
	out    *bufio.Reader
	stderr *bytes.Buffer
	mu     sync.Mutex
	next   int64
	closed bool
}

type rpcResponse struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Start launches the upstream MCP server.
func Start(ctx context.Context, opts Options) (*Client, error) {
	// Purpose: start the upstream MCP server in the project root.
	if ctx == nil {
		ctx = context.Background()
	}
	if len(opts.Command) == 0 {
		return nil, errors.New("missing code-review-graph command")
	}
	args := append([]string{}, opts.Args...)
	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}
	args = append(append([]string{}, opts.Command[1:]...), args...)
	cmd := exec.CommandContext(ctx, opts.Command[0], args...)
	cmd.Dir = opts.Repo
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := &Client{cmd: cmd, in: in, out: bufio.NewReader(out), stderr: stderr}
	if err := c.initialize(ctx); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) initialize(ctx context.Context) error {
	// Purpose: complete the MCP handshake with the upstream server.
	params := map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "maid", "version": "0"},
	}
	if err := c.request(ctx, "initialize", params, nil); err != nil {
		return err
	}
	return c.notify(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
}

func (c *Client) request(ctx context.Context, method string, params any, out any) error {
	// Purpose: send a JSON-RPC request and wait for its response.
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("code-review-graph client is closed")
	}
	c.next++
	id := c.next
	message := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		message["params"] = params
	}
	if err := c.write(message); err != nil {
		return err
	}
	for {
		line, err := c.out.ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("read code-review-graph response: %w: %s", err, c.stderr.String())
		}
		var resp rpcResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		if string(resp.ID) != strconv.FormatInt(id, 10) {
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("code-review-graph: %s", resp.Error.Message)
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(resp.Result, out)
	}
}

func (c *Client) notify(message map[string]any) error {
	// Purpose: write one JSON-RPC notification.
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.write(message)
}

func (c *Client) write(message map[string]any) error {
	// Purpose: write one newline-delimited JSON-RPC message.
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = c.in.Write(append(b, '\n'))
	return err
}

// Tools lists the upstream tools.
func (c *Client) Tools(ctx context.Context) ([]Tool, error) {
	// Purpose: fetch the upstream MCP tool list.
	var result struct {
		Tools []Tool `json:"tools"`
	}
	if err := c.request(ctx, "tools/list", nil, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// Call forwards one tool call to the upstream server.
func (c *Client) Call(ctx context.Context, name string, args json.RawMessage) (map[string]any, error) {
	// Purpose: forward one MCP tool call and return the upstream result.
	if len(args) == 0 {
		args = json.RawMessage("{}")
	}
	var result map[string]any
	err := c.request(ctx, "tools/call", map[string]any{"name": name, "arguments": args}, &result)
	return result, err
}

// Close stops the upstream server.
func (c *Client) Close() error {
	// Purpose: stop the upstream child process.
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	_ = c.in.Close()
	if c.cmd.Process == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		_ = c.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
		_ = c.cmd.Process.Kill()
		<-done
		return nil
	}
}
