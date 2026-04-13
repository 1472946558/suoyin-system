# TOOLS.md - LiaoZong Environment

## OpenClaw Platform

### Gateway
- Host: localhost
- Port: 18789
- Auth: token (stored in openclaw.json)

### Agent Models
| Profile | Provider | Model | Use For |
|---------|----------|-------|---------|
| deepseek | DeepSeek API | deepseek-chat | Default, coding, analysis |
| openai | OpenAI API | gpt-5.2 | General tasks |
| anthropic | Anthropic API | (backup) | Fallback |

### Channels
- **Feishu:** websocket mode, app configured, group policy open

### Skills
- **openai-image-gen:** DALL-E image generation
- **openai-whisper-api:** Speech to text

### Extensions
- **feishu:** Full Feishu integration (wiki, doc, drive, chat, permissions)

## Project Locations

| Project | Location | Status |
|---------|----------|--------|
| OpenClaw | /Users/xiaoliao/Desktop/openclaw | Active |
| Golf iOS | ( outsourced, location TBD ) | Active |
| GEO Product | ( in planning, location TBD ) | Planning |
| Candy | ( in planning, location TBD ) | Planning |

## API Keys (Reference Only — Actual Keys in openclaw.json)

- DeepSeek API: configured
- OpenAI API: configured
- Anthropic API: configured
- Feishu App: configured

## Backup Location

- Workspace backups: `workspace/backups/YYYY-MM-DD/`
- Config backups: `openclaw.json.bak`, `openclaw.json.bak.1`, `openclaw.json.save`

## Notes

- Gateway binds to loopback only (no external access)
- Max concurrent agents: 4
- Max concurrent subagents: 8
- Compaction mode: safeguard
- Session scope: per-channel-peer

---

_Last updated: 2026-04-13_
