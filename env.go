package dotnet

import (
    "net"
    "net/http"
    "path/filepath"
    "strings"
    "github.com/caddyserver/caddy/v2"
)

var headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")

// buildEnv returns a set of environment variables for the request.
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

    env["AUTH_TYPE"] = ""  // Not used
    env["CONTENT_LENGTH"] = r.Header.Get("Content-Length")
    env["CONTENT_TYPE"] = r.Header.Get("Content-Type")
    env["GATEWAY_INTERFACE"] = "CGI/1.1"
    env["PATH_INFO"] = ""
    env["QUERY_STRING"] = r.URL.RawQuery
    env["REMOTE_ADDR"] = ip
    env["REMOTE_HOST"] = ip  // For speed, remote host lookups disabled
    env["REMOTE_PORT"] = port
    env["REMOTE_IDENT"] = ""  // Not used
    env["REMOTE_USER"] = authUser
    env["REQUEST_METHOD"] = r.Method
    env["REQUEST_SCHEME"] = requestScheme
    env["SERVER_NAME"] = reqHost
    env["SERVER_PROTOCOL"] = r.Proto
    env["SERVER_SOFTWARE"] = d.logger.Name()

    // Other variables
    env["DOCUMENT_ROOT"] = root
    env["DOCUMENT_URI"] = docURI
    env["HTTP_HOST"] = r.Host  // added here, since not always part of headers
    env["REQUEST_URI"] = r.URL.RequestURI()
    env["SCRIPT_FILENAME"] = scriptFilename
    env["SCRIPT_NAME"] = scriptName

    if reqPort != "" {
        env["SERVER_PORT"] = reqPort
    } else if requestScheme == "http" {
        env["SERVER_PORT"] = "80"
    } else if requestScheme == "https" {
        env["SERVER_PORT"] = "443"
    }

    addTLSEnv(r, env)
    addUserDefinedEnv(d, repl, env)
    addHTTPHeadersEnv(r, env)

    return env, nil
}