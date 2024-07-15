package dotnet

import (
    "net/http"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "io"
    "go.uber.org/zap"
    "github.com/gorilla/websocket"
)

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (d DotNet) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    if websocket.IsWebSocketUpgrade(r) {
        return d.handleWebSocket(w, r)
    }

    res, err := d.RoundTrip(r)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    for k, v := range res.Header {
        for _, vv := range v {
            w.Header().Add(k, vv)
        }
    }

    w.WriteHeader(res.StatusCode)

    if _, err := io.Copy(w, res.Body); err != nil {
        d.logger.Error("Error copying response body", zap.Error(err))
        return err
    }

    return nil
}