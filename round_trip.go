package dotnet

import (
    "net/http"
    "go.uber.org/zap"
    "fmt"
    "strings"
)

// RoundTrip implements http.RoundTripper.
func (d DotNet) RoundTrip(r *http.Request) (*http.Response, error) {
    env, err := d.buildEnv(r)
    if err != nil {
        d.logger.Error("Error building environment", zap.Error(err))
        return nil, fmt.Errorf("building environment: %v", err)
    }

    url := fmt.Sprintf("http://unix%s", r.URL.Path)
    if r.URL.RawQuery != "" {
        url += "?" + r.URL.RawQuery
    }

    d.logger.Info("Constructed URL", zap.String("url", url))

    req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
    if err != nil {
        d.logger.Error("Error creating new request", zap.Error(err))
        return nil, err
    }

    // Copy original headers
    req.Header = r.Header.Clone()

    // Set environment variables as headers
    for k, v := range env {
        if !strings.HasPrefix(k, "HTTP_") {
            // Convert environment variable names to header names
            headerName := "X-Dotnet-Env-" + strings.ReplaceAll(k, "_", "-")
            req.Header.Set(headerName, v)
        }
    }

    // Special handling for CONTENT_LENGTH and CONTENT_TYPE
    if contentLength, ok := env["CONTENT_LENGTH"]; ok {
        req.Header.Set("Content-Length", contentLength)
    }
    if contentType, ok := env["CONTENT_TYPE"]; ok {
        req.Header.Set("Content-Type", contentType)
    }

    req.Host = r.Host

    client := d.clients.get()
    defer d.clients.put(client)

    resp, err := client.Do(req)
    if err != nil {
        d.logger.Error("Error executing request", zap.Error(err))
        return nil, err
    }

    return resp, nil
}