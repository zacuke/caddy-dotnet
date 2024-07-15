package dotnet

import (
    "context"
    "crypto/tls"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "github.com/caddyserver/caddy/v2/modules/caddytls"
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
    res, err := d.RoundTrip(r)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    // Copy headers from the response
    for k, v := range res.Header {
        for _, vv := range v {
            w.Header().Add(k, vv)
        }
    }

    // Write the status code from the response
    w.WriteHeader(res.StatusCode)

    // Copy the response body
    if _, err := io.Copy(w, res.Body); err != nil {
        d.logger.Error("Error copying response body", zap.Error(err))
        return err
    }

    return nil
}

// RoundTrip implements http.RoundTripper.
func (d DotNet) RoundTrip(r *http.Request) (*http.Response, error) {
    env, err := d.buildEnv(r)
    if err != nil {
        d.logger.Error("Error building environment", zap.Error(err))
        return nil, fmt.Errorf("building environment: %v", err)
    }

    ctx := r.Context()

    // create the client that will facilitate the protocol
    transport := &http.Transport{
        DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
            return net.Dial("unix", d.Socket)
        },
    }

    // Construct the URL with the correct scheme
    url := fmt.Sprintf("http://unix%s", r.URL.Path)
    if r.URL.RawQuery != "" {
        url += "?" + r.URL.RawQuery
    }

    d.logger.Info("Constructed URL", zap.String("url", url))

    req, err := http.NewRequestWithContext(ctx, r.Method, url, r.Body)
    if err != nil {
        d.logger.Error("Error creating new request", zap.Error(err))
        return nil, err
    }

    req.Header = r.Header.Clone()
    req.Host = r.Host
    req.RemoteAddr = r.RemoteAddr
    for k, v := range env {
        req.Header.Set(k, v)
    }

    client := &http.Client{Transport: transport}
    resp, err := client.Do(req)
    if err != nil {
        d.logger.Error("Error executing request", zap.Error(err))
        return nil, err
    }

    return resp, nil
}

// buildEnv returns a set of CGI environment variables for the request.
func (d DotNet) buildEnv(r *http.Request) (map[string]string, error) {
    repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

    env := make(map[string]string)

    // Separate remote IP and port; more lenient than net.SplitHostPort
    var ip, port string
    if idx := strings.LastIndex(r.RemoteAddr, ":"); idx > -1 {
        ip = r.RemoteAddr[:idx]
        port = r.RemoteAddr[idx+1:]
    } else {
        ip = r.RemoteAddr
    }

    // Remove [] from IPv6 addresses
    ip = strings.Replace(ip, "[", "", 1)
    ip = strings.Replace(ip, "]", "", 1)

    // make sure file root is absolute
    root, err := filepath.Abs(repl.ReplaceAll(d.WorkingDir, "."))
    if err != nil {
        return nil, err
    }

    docURI := r.URL.Path
    scriptName := docURI

    // SCRIPT_FILENAME is the absolute path of SCRIPT_NAME
    scriptFilename := filepath.Join(root, scriptName)

    requestScheme := "http"
    if r.TLS != nil {
        requestScheme = "https"
    }

    reqHost, reqPort, err := net.SplitHostPort(r.Host)
    if err != nil {
        // whatever, just assume there was no port
        reqHost = r.Host
    }

    authUser, _ := repl.GetString("http.auth.user.id")

    env = map[string]string{
        "AUTH_TYPE":         "", // Not used
        "CONTENT_LENGTH":    r.Header.Get("Content-Length"),
        "CONTENT_TYPE":      r.Header.Get("Content-Type"),
        "GATEWAY_INTERFACE": "CGI/1.1",
        "PATH_INFO":         "",
        "QUERY_STRING":      r.URL.RawQuery,
        "REMOTE_ADDR":       ip,
        "REMOTE_HOST":       ip, // For speed, remote host lookups disabled
        "REMOTE_PORT":       port,
        "REMOTE_IDENT":      "", // Not used
        "REMOTE_USER":       authUser,
        "REQUEST_METHOD":    r.Method,
        "REQUEST_SCHEME":    requestScheme,
        "SERVER_NAME":       reqHost,
        "SERVER_PROTOCOL":   r.Proto,
        "SERVER_SOFTWARE":   d.logger.Name(),

        // Other variables
        "DOCUMENT_ROOT":   root,
        "DOCUMENT_URI":    docURI,
        "HTTP_HOST":       r.Host, // added here, since not always part of headers
        "REQUEST_URI":     r.URL.RequestURI(),
        "SCRIPT_FILENAME": scriptFilename,
        "SCRIPT_NAME":     scriptName,
    }

    if reqPort != "" {
        env["SERVER_PORT"] = reqPort
    } else if requestScheme == "http" {
        env["SERVER_PORT"] = "80"
    } else if requestScheme == "https" {
        env["SERVER_PORT"] = "443"
    }

    if r.TLS != nil {
        env["HTTPS"] = "on"
        v, ok := tlsProtocolStrings[r.TLS.Version]
        if ok {
            env["SSL_PROTOCOL"] = v
        }
        for _, cs := range caddytls.SupportedCipherSuites() {
            if cs.ID == r.TLS.CipherSuite {
                env["SSL_CIPHER"] = cs.Name
                break
            }
        }
    }

    for _, value := range d.EnvVars {
        kv := strings.SplitN(value, "=", 2)
        if len(kv) == 2 {
            env[kv[0]] = repl.ReplaceAll(kv[1], "")
        }
    }

    for field, val := range r.Header {
        header := strings.ToUpper(field)
        header = headerNameReplacer.Replace(header)
        env["HTTP_"+header] = strings.Join(val, ", ")
    }

    return env, nil
}

var headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")

var tlsProtocolStrings = map[uint16]string{
    tls.VersionTLS10: "TLSv1",
    tls.VersionTLS11: "TLSv1.1",
    tls.VersionTLS12: "TLSv1.2",
    tls.VersionTLS13: "TLSv1.3",
}

var _ caddyhttp.MiddlewareHandler = (*DotNet)(nil)

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
