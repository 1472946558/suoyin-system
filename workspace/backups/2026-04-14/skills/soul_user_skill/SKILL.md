# SKILL.md - Soul / User Skill

## Purpose

Maintain deep understanding of the user (廖总) and ensure all agents operate with consistent user context.

## What This Skill Provides

- User identity and preferences
- Communication style rules
- Project priority hierarchy
- Decision-making patterns
- Work preference constraints

## Key Data

### User Profile
- Name: Xiao Liao (廖总)
- Timezone: UTC+8
- Language: Chinese (primary), English (technical)
- Style: Direct, no fluff, results-first

### Priority Hierarchy
1. Golf iOS delivery (deadline-driven)
2. GEO product (strategic)
3. OpenClaw maintenance (infrastructure)
4. Business operations (revenue)

### Communication Rules
- Never use "Sure, I can help!" or "Great question!"
- Never explain what you're about to do — just do it
- Output: Conclusion -> Cause -> Actions -> Priority -> Risk
- Chinese default, English for technical terms
- Give options with trade-offs, not open questions

### Decision Patterns
- Fast decisions on delivery tasks
- Research before product decisions
- Zero tolerance for security risks
- Prefers SOPs over one-off answers

## When to Use

Every agent must read this skill at session start. It provides the user context layer that all task routing and output formatting depends on.

## Update Triggers

- User explicitly states a preference change
- Repeated corrections on same topic (3+ times)
- New project added or removed
- Priority shift detected

## Related Files

- `workspace/USER.md` — full user profile
- `workspace/SOUL.md` — agent personality
- `workspace/IDENTITY.md` — agent identity
- `workspace/memory/` — session context

---

_Skill file for LiaoZong framework. Updates should be reflected in USER.md._
