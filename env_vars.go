package dotnet

import (
    "strings"
    "github.com/caddyserver/caddy/v2"
    "net/http"
)

func addUserDefinedEnv(d DotNet, repl *caddy.Replacer, env map[string]string) {
    for _, value := range d.EnvVars {
        kv := strings.SplitN(value, "=", 2)
        if len(kv) == 2 {
            env[kv[0]] = repl.ReplaceAll(kv[1], "")
        }
    }
}

func addHTTPHeadersEnv(r *http.Request, env map[string]string) {
    for field, val := range r.Header {
        header := "HTTP_" + strings.ToUpper(headerNameReplacer.Replace(field))
        env[header] = strings.Join(val, ", ")
    }
}