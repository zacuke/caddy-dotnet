package dotnet

import (
    "crypto/tls"
    "net/http"
    "github.com/caddyserver/caddy/v2/modules/caddytls"
)

var tlsProtocolStrings = map[uint16]string{
    tls.VersionTLS10: "TLSv1",
    tls.VersionTLS11: "TLSv1.1",
    tls.VersionTLS12: "TLSv1.2",
    tls.VersionTLS13: "TLSv1.3",
}

func addTLSEnv(r *http.Request, env map[string]string) {
    if r.TLS != nil {
        env["HTTPS"] = "on"
        if v, ok := tlsProtocolStrings[r.TLS.Version]; ok {
            env["SSL_PROTOCOL"] = v
        }
        for _, cs := range caddytls.SupportedCipherSuites() {
            if cs.ID == r.TLS.CipherSuite {
                env["SSL_CIPHER"] = cs.Name
                break
            }
        }
    }
}