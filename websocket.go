package dotnet

import (
    "net/http"
    "net/http/httputil"
    "net/url"

)

func (d DotNet) handleWebSocket(w http.ResponseWriter, r *http.Request) error {
    proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "unix", Path: d.Socket})
    
    proxy.Director = func(req *http.Request) {
        req.URL.Scheme = "http"
        req.URL.Host = "unix"
        req.URL.Path = r.URL.Path
        req.Host = r.Host
    }
    
    proxy.ServeHTTP(w, r)
    return nil
}