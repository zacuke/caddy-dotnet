package dotnet

import (
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "go.uber.org/zap"
    "context"
)

func (d DotNet) handleWebSocket(w http.ResponseWriter, r *http.Request) error {
    socketPath := d.generatedSocket // Use the generated socket path
    
    proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "localhost"})
    
    proxy.Transport = &http.Transport{
        DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
            return net.Dial("unix", socketPath)
        },
    }
    
    proxy.Director = func(req *http.Request) {
        req.URL.Scheme = "http"
        req.URL.Host = "localhost"
        req.Host = r.Host
    }
    
    d.logger.Info("Proxying WebSocket request",
        zap.String("path", r.URL.Path),
        zap.String("socket", socketPath))
    
    proxy.ServeHTTP(w, r)
    
    return nil
}