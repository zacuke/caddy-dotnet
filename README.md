# Caddy .NET Plugin

This plugin allows Caddy to serve .NET applications directly.

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
 
This command will build a Caddy binary that includes the .NET plugin. The binary will be named `caddy` and will be in your current directory. Replace the binary on server to enable plugin.

## Usage

To use the plugin, you need to configure it in your Caddyfile. Here's a sample configuration:

```
:80 {
    route {
        dotnet {
            exec_path /usr/bin/dotnet
            working_dir /var/www/example
            args /var/www/example/example.dll arg2
            env_vars ASPNETCORE_ENVIRONMENT=Test KEY2=VALUE2
            syslog_output
        }
    }
}
```

### Configuration Options

- `exec_path`: Path to your .NET application executable or /usr/bin/dotnet (required)
- `working_dir`: Working directory for your .NET application
- `args`: Additional arguments to pass to your .NET application or specify .NET application dll when exec_path=/usr/bin/dotnet
- `env_vars`: Environment variables to set for your .NET application
- `syslog_output`: If present, redirects the .NET application's output to syslog
 
## Notes

- This plugin uses Unix sockets for communication between Caddy and your .NET application.
- The --urls parameter is added to your .NET application to listen to unix socket.
- The plugin will start your .NET application automatically when Caddy starts.
- WebSocket support is included.
- The process currently runs under caddy user. 

 
For more detailed information about Caddy and its configuration, please refer to the [official Caddy documentation](https://caddyserver.com/docs/).
