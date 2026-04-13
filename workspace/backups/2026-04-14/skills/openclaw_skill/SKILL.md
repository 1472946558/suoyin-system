# SKILL.md - OpenClaw Platform

## Purpose

All knowledge, configuration, and maintenance procedures for the OpenClaw multi-agent platform. Any agent working on system tasks must read this first.

## Platform Overview

OpenClaw is a multi-agent orchestration system running locally, integrated with Feishu for messaging.

## Current Configuration

### Gateway
- Port: 18789
- Mode: local
- Bind: loopback
- Auth: token-based

### Agent Profiles
| Agent | Model | Provider |
|-------|-------|----------|
| default | deepseek-chat | DeepSeek |
| coder | deepseek-chat | DeepSeek |
| analysis | deepseek-chat | DeepSeek |
| general | gpt-5.2 | OpenAI |

### Channels
- **Feishu** (websocket mode, domain: feishu)
  - App ID: configured
  - Group policy: open

### Skills
- openai-image-gen (OpenAI DALL-E)
- openai-whisper-api (speech-to-text)

### Extensions
- Feishu plugin (enabled)

## Directory Structure

```
~/.openclaw/
├── agents/          → Agent runtime data
│   └── main/        → Primary agent sessions
├── extensions/      → Plugins (feishu)
│   └── feishu/      → Feishu integration
├── workspace/       → THIS workspace (knowledge base)
│   ├── agents/      → Agent definitions
│   ├── skills/      → Skill definitions
│   ├── docs/        → SOPs and templates
│   ├── prompts/     → Generated prompts
│   ├── backups/     → Backup snapshots
│   └── memory/      → Daily logs
├── devices/         → Device management
├── identity/        → Identity data
├── logs/            → Runtime logs
├── canvas/          → Canvas data
├── cron/            → Scheduled tasks
└── openclaw.json    → Main configuration
```

## Maintenance Procedures

### Daily
- Check gateway is running (port 18789)
- Verify Feishu websocket connection
- Check memory/log disk usage

### Weekly
- Clean old session data
- Review and rotate API keys if needed
- Verify backups are current
- Update agent definitions if needed

### Before Changes
1. Backup current state
2. Document what's changing
3. Apply change
4. Verify system health
5. Record in memory

## Common Operations

### Add a new agent model
1. Edit `openclaw.json` → `agents` section
2. Add auth profile if new provider
3. Restart gateway
4. Test with simple query

### Update Feishu integration
1. Check `extensions/feishu/` for config
2. Update `openclaw.json` → `channels.feishu`
3. Verify websocket reconnection

### Backup/Restore
```bash
# Backup
cp -r ~/.openclaw/workspace/ ~/.openclaw/workspace/backups/YYYY-MM-DD/

# Restore
cp -r ~/.openclaw/workspace/backups/YYYY-MM-DD/* ~/.openclaw/workspace/
```

## Troubleshooting

| Symptom | Check | Fix |
|---------|-------|-----|
| Gateway down | `lsof -i :18789` | Restart openclaw |
| Agent timeout | API key validity | Update in openclaw.json |
| Feishu disconnect | Websocket status | Check app credentials |
| Memory full | `du -sh ~/.openclaw/` | Clean old sessions/logs |
| Config corrupt | Compare with .bak | Restore from backup |

## Security Notes

- `openclaw.json` contains API keys — never commit to public repo
- `.gitignore` must exclude sensitive files
- Gateway binds to loopback only (no external access)
- Token auth enabled for gateway access

## Related Agents

- `openclaw-agent.md` — primary maintenance agent
- `security-agent.md` — security verification
- `backup-agent.md` — backup procedures

## Related Files

- `agents/openclaw-agent.md`
- `agents/security-agent.md`
- `agents/backup-agent.md`
- `openclaw.json` (root config)

---

_Skill file for LiaoZong framework. Update when platform configuration changes._
