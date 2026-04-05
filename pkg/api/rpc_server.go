package api

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// ShmClaw is the RPC service
type ShmClaw struct {
	// Add orchestrator dependency here later
}

// ExecuteArgs holds the arguments for the Execute RPC method
type ExecuteArgs struct {
	Command string `json:"command"`
}

// ExecuteReply holds the response for the Execute RPC method
type ExecuteReply struct {
	Result string `json:"result"`
}

// Execute is the main entry point for the JSON-RPC interface
func (s *ShmClaw) Execute(args *ExecuteArgs, reply *ExecuteReply) error {
	reply.Result = "Executed: " + args.Command
	return nil
}

// HttpConn wraps http.ResponseWriter and *http.Request into an io.ReadWriteCloser
type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

// NewRPCServer initializes and returns an http.Handler for the JSON-RPC server
func NewRPCServer() (http.Handler, error) {
	server := rpc.NewServer()
	shmClaw := &ShmClaw{}
	err := server.Register(shmClaw)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		conn := &HttpConn{in: r.Body, out: w}
		server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}), nil
}
