package dotnet

import (
    "context"
    "fmt"
    "net"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "time"
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(Dotnet{})
}

type Dotnet struct {
    ExecPath string `json:"exec_path"`
    Args     string `json:"args"`
    Socket   string `json:"socket"`
}

func (Dotnet) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.dotnet",
        New: func() caddy.Module { return new(Dotnet) },
    }
}

func (d Dotnet) Provision(ctx caddy.Context) error {
    return nil
}

func (d Dotnet) Validate() error {
    if d.ExecPath == "" {
        return fmt.Errorf("exec_path is required")
    }
    if d.Socket == "" {
        d.Socket = "/tmp/kestrel.sock"
    }
    return nil
}

func (d Dotnet) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    if _, err := os.Stat(d.Socket); err == nil {
        os.Remove(d.Socket)
    }

    args := append(strings.Split(d.Args, " "), "--urls", "http://unix:"+d.Socket)
    cmd := exec.Command(d.ExecPath, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    go func() {
        if err := cmd.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to start .NET application: %v\n", err)
        }
    }()

    for i := 0; i < 10; i++ {
        if _, err := os.Stat(d.Socket); err == nil {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }

    conn, err := net.Dial("unix", d.Socket)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }
    defer conn.Close()

    proxy := &http.Transport{
        DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
            return conn, nil
        },
    }
    proxyClient := &http.Client{Transport: proxy}
    proxyReq, err := http.NewRequest(r.Method, "http://unix"+r.URL.RequestURI(), r.Body)
    if err != nil {
        return err
    }
    proxyReq.Header = r.Header

    proxyResp, err := proxyClient.Do(proxyReq)
    if err != nil {
        return err
    }
    defer proxyResp.Body.Close()

    for key, values := range proxyResp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    w.WriteHeader(proxyResp.StatusCode)
    _, err = w.Write([]byte(proxyResp.Status))
    return err
}

func (d *Dotnet) UnmarshalCaddyfile(h *caddyfile.Dispenser) error {
    for h.Next() {
        for h.NextBlock(0) {
            switch h.Val() {
            case "exec_path":
                if !h.Args(&d.ExecPath) {
                    return h.ArgErr()
                }
            case "args":
                if !h.Args(&d.Args) {
                    return h.ArgErr()
                }
            case "socket":
                if !h.Args(&d.Socket) {
                    return h.ArgErr()
                }
            }
        }
    }
    return nil
}

var (
    _ caddyhttp.MiddlewareHandler = (*Dotnet)(nil)
    _ caddyfile.Unmarshaler       = (*Dotnet)(nil)
)
