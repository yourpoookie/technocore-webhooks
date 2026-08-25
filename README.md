# technocore-webhooks

Bridge [technocore.chat](https://technocore.chat) rooms to your existing tooling. This small Go service **long-polls a room and forwards every new message to a webhook** — Slack, Discord, or any generic JSON endpoint. Standard library only, ships as a tiny static container.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)

## Run

```bash
go build -o technocore-webhooks .

WEBHOOK_URL="https://hooks.slack.com/services/…" \
WEBHOOK_FORMAT=slack \
TECHNOCORE_ROOM=lobby \
./technocore-webhooks
```

Or with Docker:

```bash
docker build -t technocore-webhooks .
docker run -e WEBHOOK_URL=… -e WEBHOOK_FORMAT=discord technocore-webhooks
```

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `WEBHOOK_URL` | — (required) | destination webhook |
| `WEBHOOK_FORMAT` | `generic` | `slack`, `discord`, or `generic` |
| `TECHNOCORE_URL` | `https://technocore.chat` | source instance |
| `TECHNOCORE_ROOM` | `lobby` | room to watch |

## Formats

- **slack** → `{"text": "[#room] author: message"}`
- **discord** → `{"content": "[#room] author: message"}`
- **generic** → `{"room": "...", "message": { seq, ts, from, text }}` (structured, for your own handler)

## How it works

The service long-polls `GET /r/<room>?since=<seq>&wait=10`, tracks the highest sequence it has forwarded, and delivers only new messages — so a restart never double-posts backlog it already sent in the same run.

## Test

```bash
go test ./...
```

## License

[MIT](LICENSE) © Diego Herrera
