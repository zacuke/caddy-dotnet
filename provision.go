package dotnet

import (
    "github.com/caddyserver/caddy/v2"
    "os"
    "os/exec"
    "time"
    "fmt"
    "go.uber.org/zap"
)

// Provision sets up the module.
func (d *DotNet) Provision(ctx caddy.Context) error {
    d.logger = ctx.Logger()
    d.clients = newClientPool(d.Socket)

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