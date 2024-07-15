package dotnet

import (
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "go.uber.org/zap"
    "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
    caddy.RegisterModule(DotNet{})
    httpcaddyfile.RegisterHandlerDirective("dotnet", parseCaddyfile)
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
    clients      *clientPool
}

// CaddyModule returns the Caddy module information.
func (DotNet) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.dotnet",
        New: func() caddy.Module { return new(DotNet) },
    }
}

var _ caddyhttp.MiddlewareHandler = (*DotNet)(nil)