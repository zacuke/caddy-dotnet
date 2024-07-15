package dotnet

import (
    "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    httpcaddyfile.RegisterHandlerDirective("dotnet", parseCaddyfile)
}

// parseCaddyfile sets up the handler from Caddyfile tokens.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
    var d DotNet
    for h.Next() {
        for h.NextBlock(0) {
            switch h.Val() {
            case "exec_path":
                if !h.NextArg() {
                    return nil, h.ArgErr()
                }
                d.ExecPath = h.Val()
            case "args":
                d.Args = h.RemainingArgs()
            case "socket":
                if !h.NextArg() {
                    return nil, h.ArgErr()
                }
                d.Socket = h.Val()
            case "env_vars":
                d.EnvVars = h.RemainingArgs()
            case "working_dir":
                if !h.NextArg() {
                    return nil, h.ArgErr()
                }
                d.WorkingDir = h.Val()
            case "syslog_output":
                d.SyslogOutput = true
            default:
                return nil, h.Errf("unrecognized subdirective %s", h.Val())
            }
        }
    }
    return d, nil
}
