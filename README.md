# Caddy .NET Plugin

This plugin allows Caddy to serve .NET applications directly in Caddy without running seperate services.

## Installation

To use this plugin, you need to build Caddy with the plugin included. The easiest way to do this is using `xcaddy`. Follow these steps:

1. Ensure you have Go installed on your system (version 1.22.5 or later).

2. Install `xcaddy` if you haven't already:
   ```
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

3. Build Caddy with the .NET plugin:
   ```
   xcaddy build --with github.com/zacuke/caddy-dotnet
   ```
 
This command will build a Caddy binary that includes the .NET plugin. The binary will be named `caddy` and will be in your current directory.

Use this caddy executable instead of the  default to utilize dotnet 

## Usage

To use the plugin, you need to configure it in your Caddyfile. Here's a sample configuration:

```
:80 {
    route {
        dotnet {
            exec_path /path/to/your/dotnet/app
            working_dir /path/to/your/working/directory
            args arg1 arg2
            env_vars KEY1=VALUE1 KEY2=VALUE2
            socket /path/to/socket/file.sock
            syslog_output
        }
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

 
For more detailed information about Caddy and its configuration, please refer to the [official Caddy documentation](https://caddyserver.com/docs/).