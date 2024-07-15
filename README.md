# Caddy .NET Plugin

This plugin allows Caddy to serve .NET applications using a Unix socket for communication.

## Installation

To use this plugin, you need to compile Caddy with the plugin included. Follow these steps:

1. Ensure you have Go installed on your system (version 1.16 or later).

2. Clone this repository:
   ```
   git clone https://github.com/yourusername/caddy-dotnet-plugin.git
   ```

3. Navigate to the cloned directory:
   ```
   cd caddy-dotnet-plugin
   ```

4. Create a new directory for your custom Caddy build:
   ```
   mkdir custom-caddy
   cd custom-caddy
   ```

5. Initialize a new Go module:
   ```
   go mod init github.com/yourusername/custom-caddy
   ```

6. Create a `main.go` file with the following content:
   ```go
   package main

   import (
       caddycmd "github.com/caddyserver/caddy/v2/cmd"

       // plug in Caddy modules here
       _ "github.com/caddyserver/caddy/v2/modules/standard"
       _ "github.com/yourusername/caddy-dotnet-plugin"
   )

   func main() {
       caddycmd.Main()
   }
   ```

7. Build Caddy with the plugin:
   ```
   go build
   ```

You should now have a `custom-caddy` binary in your current directory that includes the .NET plugin.

## Usage

To use the plugin, you need to configure it in your Caddyfile. Here's a sample configuration:

```
:80 {
    dotnet {
        exec_path /path/to/your/dotnet/app
        working_dir /path/to/your/working/directory
        args arg1 arg2
        env_vars KEY1=VALUE1 KEY2=VALUE2
        socket /path/to/socket/file.sock
        syslog_output
    }
}
```

### Configuration Options

- `exec_path`: Path to your .NET application executable (required)
- `working_dir`: Working directory for your .NET application
- `args`: Additional arguments to pass to your .NET application
- `env_vars`: Environment variables to set for your .NET application
- `socket`: Path to the Unix socket file (if not specified, a random socket will be generated)
- `syslog_output`: If present, redirects the .NET application's output to syslog

Make sure your .NET application is configured to listen on the Unix socket specified in the Caddyfile (or the generated one if not specified).

## Notes

- This plugin uses Unix sockets for communication between Caddy and your .NET application. Ensure your .NET application is configured to use the same socket.
- The plugin will start your .NET application automatically when Caddy starts.
- WebSocket support is included, but make sure your .NET application is properly configured to handle WebSocket connections over the Unix socket.

## Troubleshooting

If you encounter any issues:

1. Check Caddy's error logs for any error messages.
2. Ensure your .NET application is properly configured to listen on the Unix socket.
3. Verify that the paths in your Caddyfile are correct and accessible.
4. If using WebSockets, ensure your .NET application is properly handling WebSocket upgrades.

For more detailed information about Caddy and its configuration, please refer to the [official Caddy documentation](https://caddyserver.com/docs/).