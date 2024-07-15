package dotnet

import (
    "github.com/caddyserver/caddy/v2"
    "os"
    "os/exec"
    "time"
    "fmt"
    "go.uber.org/zap"
    "path/filepath"
    "crypto/rand"
    "encoding/hex"
)

// Provision sets up the module.
func (d *DotNet) Provision(ctx caddy.Context) error {
    d.logger = ctx.Logger()

    if d.Socket == "" {
        // Generate a random socket name in Caddy's app data directory
        randomBytes := make([]byte, 16)
        if _, err := rand.Read(randomBytes); err != nil {
            return fmt.Errorf("failed to generate random bytes: %w", err)
        }
        randomHex := hex.EncodeToString(randomBytes)
        
        // Use Caddy's app data directory
        appDataDir := caddy.AppDataDir()
        socketDir := filepath.Join(appDataDir, "dotnet", "sockets")
        
        err := os.MkdirAll(socketDir, 0755)
        if err != nil {
            return fmt.Errorf("failed to create socket directory: %w", err)
        }
        
        d.generatedSocket = filepath.Join(socketDir, fmt.Sprintf("dotnet-%s.sock", randomHex))
    } else {
        d.generatedSocket = d.Socket
    }

    d.clients = newClientPool(d.generatedSocket)

    args := append(d.Args, fmt.Sprintf("--urls=http://unix:%s", d.generatedSocket))
    cmd := exec.Command(d.ExecPath, args...)
    cmd.Env = append(os.Environ(), d.EnvVars...)
    cmd.Dir = d.WorkingDir

    if d.SyslogOutput {
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    }

    d.logger.Info("Starting .NET application", 
        zap.String("exec_path", d.ExecPath), 
        zap.Strings("args", args), 
        zap.Strings("env_vars", d.EnvVars),
        zap.String("socket", d.generatedSocket))

    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("failed to start the .NET application: %w", err)
    }

    // Give the application some time to start
    time.Sleep(5 * time.Second)

    return nil
}