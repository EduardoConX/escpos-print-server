# ESC/POS Print Server

Local HTTP server for sending ESC/POS tickets to a thermal printer installed on Windows. It is designed to receive POS requests over the network and send the raw print data directly to the Windows print queue.

Based on the work of [Parzibyte](https://github.com/parzibyte/ejemplos-plugin-impresoras-termicas-v2).

## Requirements

- Windows on the computer where the printer is installed.
- A printer compatible with ESC/POS installed and visible in Windows.
- Go installed on the computer where the server is compiled. Go is not required on the printer machine if you copy and run the compiled `.exe`.

The local print queue in this setup is named `XP-58`. `TicketsPrinter` is the shared Windows name; POS devices send requests to the server, which then prints to the local queue `XP-58`.

## Build

From PowerShell in the project folder:

```powershell
go build -o escpos-print-server.exe .
```

You can copy `escpos-print-server.exe` to the printer computer. The `XP-58` printer must be installed there with that name.

## Configuration and startup

The server uses these environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port for the server. |
| `HOST` | `0.0.0.0` | Network interface to bind to. |
| `PRINTER_NAME` | `XP-58` | Name of the Windows print queue. |
| `LOG_LEVEL` | `info` | Logging level. |
| `CONFIG_FILE` | `config.json` | Optional path to a JSON configuration file. |
| `API_KEY` | unset | Optional Bearer token required for printing and status endpoints. |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins, or `*`. |
| `READ_TIMEOUT_SECONDS` | `15` | Maximum time to read an HTTP request. |
| `WRITE_TIMEOUT_SECONDS` | `60` | Maximum time to write an HTTP response. |

The server loads `config.json` from its working directory if the file exists. Environment variables override values from the file. See [`config.example.json`](config.example.json) for the available JSON fields.

Example startup in PowerShell:

```powershell
$env:PORT="8080"
$env:PRINTER_NAME="XP-58"
$env:LOG_LEVEL="info"
.\escpos-print-server.exe
```

To run it from source, use `go run .` instead of the `.exe`. The terminal must remain open while the server is running. To stop it, press `Ctrl+C`.

## Test a local print

With the server running, open another PowerShell window and send a sample ticket:

```powershell
$ticket = @{
    operations = @(
        @{ action = "alignment"; data = "C" }
        @{ action = "boldText"; data = "1" }
        @{ action = "text"; data = "ESC/POS TEST" }
        @{ action = "enter"; data = "" }
        @{ action = "boldText"; data = "0" }
        @{ action = "text"; data = "Connection successful" }
        @{ action = "feed"; data = "2" }
    )
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
    -Uri "http://localhost:8080/" `
    -Method Post `
    -ContentType "application/json" `
    -Body $ticket
```

The expected response is `status: ok`. Also check that the ticket appears in the printer and in the Windows print queue. If there is an error, review the console output from the server process.

The service also provides `GET /health` for a basic health check, `GET /status` for the configured printer name, and `POST /test` to print a test ticket. `/status` does not verify whether the printer is online; use `/test` to check the end-to-end print path. When `API_KEY` is configured, send it as `Authorization: Bearer <key>` to `/`, `/status`, and `/test`. The health endpoint does not require authentication.

## Connect the POS from another computer

Get the local IP address of the machine running the server. From the POS, send the same `POST` request to:

```text
http://IP-OF-THE-PC:8080/
```

For example, if the server PC has IP `192.168.1.25`, the URL will be `http://192.168.1.25:8080/`. If the connection fails, allow the configured port through the Windows Firewall for the private network.

The `printer` field in the JSON is optional. If omitted, the server uses `PRINTER_NAME`. If provided, it must match the name of a valid print queue available to the server process.

## Request format

```json
{
    "operations": [
        { "action": "alignment", "data": "C" },
        { "action": "boldText", "data": "1" },
        { "action": "text", "data": "My store" },
        { "action": "enter", "data": "" },
        { "action": "boldText", "data": "0" },
        { "action": "separator", "data": "-" },
        { "action": "text", "data": "Thanks for your purchase" },
        { "action": "feed", "data": "2" },
        { "action": "cut", "data": "" }
    ]
}
```

Supported actions: `text`, `enter`, `line`, `separator`, `alignment` (`L`, `C`, `R`), `boldText` (`1` to enable, `0` to disable), `fontSize` (`width,height`), `feed` (line count), `cut`, and `openCashDrawer`.

Successful responses are HTTP `200` with `{"status":"ok"}`. Invalid requests respond with `400`; printer send failures respond with `500`.

## Security and operation

- The server listens on the configured port on the machine's network interfaces. Do not expose it directly to the internet.
- Restrict access to the port in the Windows Firewall to the computer's private network. You can also configure `API_KEY` to require a Bearer token for printing and status endpoints.
- For continuous operation, configure the `.exe` to start with Windows using Task Scheduler or a Windows service. The project does not install that service automatically.
- Cut and cash drawer actions depend on the printer model and its ESC/POS configuration supporting them.
