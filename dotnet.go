package dotnet

import (
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "time"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
    "go.uber.org/zap"
)

func init() {
    caddy.RegisterModule(DotNet{})
}

// DotNet represents the configuration for running .NET applications
type DotNet struct {
    ExecPath     string   `json:"exec_path"`
    Args         []string `json:"args"`
    Socket       string   `json:"socket"`
    EnvVars      []string `json:"env_vars"`
    WorkingDir   string   `json:"working_dir"`
    SyslogOutput bool     `json:"syslog_output"`
    logger       *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (DotNet) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.dotnet",
        New: func() caddy.Module { return new(DotNet) },
    }
}

// Provision sets up the module.
func (d *DotNet) Provision(ctx caddy.Context) error {
    d.logger = ctx.Logger()

    args := append(d.Args, fmt.Sprintf("--urls=http://unix:%s", d.Socket))
    cmd := exec.Command(d.ExecPath, args...)
    cmd.Env = append(os.Environ(), d.EnvVars...)
    cmd.Dir = d.WorkingDir

    if d.SyslogOutput {
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    }

    d.logger.Info("Starting .NET application", zap.String("exec_path", d.ExecPath), zap.Strings("args", args), zap.Strings("env_vars", d.EnvVars))

    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("failed to start the .NET application: %w", err)
    }

    // Give the application some time to start
    time.Sleep(5 * time.Second)

    return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (d DotNet) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    d.logger.Info("Handling request", zap.String("method", r.Method), zap.String("url", r.URL.String()))

    proxy := reverseproxy.Handler{
        Upstreams: reverseproxy.UpstreamPool{
            &reverseproxy.Upstream{
                Dial: "unix:" + d.Socket,
            },
        },
    }

    return proxy.ServeHTTP(w, r, next)
}

// Interface guards
var (
    _ caddy.Module                = (*DotNet)(nil)
    _ caddyhttp.MiddlewareHandler = (*DotNet)(nil)
)
