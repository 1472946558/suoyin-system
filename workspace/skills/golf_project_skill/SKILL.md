# SKILL.md - Golf iOS Project

## Purpose

All knowledge, rules, and context for the Golf iOS app project. Any agent working on Golf iOS must read this first.

## Project Overview

- **Type:** Outsourced iOS app development
- **Platform:** iOS (iPhone)
- **Reference:** Android version is source of truth
- **Owner:** 廖总 (product owner)

## Absolute Constraints

1. **No third-party libraries** — Apple frameworks only (UIKit, SwiftUI, Foundation, etc.)
2. **No unnecessary refactoring** — only change what's requested
3. **Acceptance-first** — match acceptance criteria exactly
4. **Android parity** — every feature must match Android behavior
5. **Minimal patches** — smallest possible change

## Technical Stack

- **Language:** Swift
- **UI Framework:** UIKit / SwiftUI (match existing project patterns)
- **Architecture:** MVVM or MVC (match existing)
- **Dependency Management:** None — no SPM, CocoaPods, or Carthage
- **Testing:** XCTest if required by acceptance criteria

## Development Rules

### DO
- Match existing code style in the project
- Use Apple-native APIs for all features
- Test against acceptance criteria before submitting
- Document every file change
- Keep patches minimal

### DON'T
- Add any third-party dependency
- Restructure project folders
- Change naming conventions
- Modify unrelated screens
- "Improve" working code
- Add abstractions that aren't needed

## Delivery Checklist

```
- [ ] Acceptance criteria fully met
- [ ] Android parity verified (if applicable)
- [ ] No third-party libraries introduced
- [ ] Modified files documented
- [ ] Minimal patch confirmed
- [ ] Log keywords provided
- [ ] Risk boundary identified
- [ ] What NOT to change listed
```

## Common Patterns

### When receiving a Golf task:
1. Read acceptance criteria carefully
2. Find the relevant existing code
3. Make the minimum change
4. Verify against Android reference
5. Output the delivery format

### Delivery Output Format:
```markdown
## Golf iOS Delivery
- Task: [description]
- Acceptance: [criteria]
- Modified files: [list]
- Patch: [diff]
- Log keywords: [grep patterns]
- Acceptance steps: [verification]
- Risk: [what could break]
- NOT changed: [explicit list]
```

## Related Agents

- `golf-ios-agent.md` — primary agent for this project
- `delivery-agent.md` — general delivery support
- `strategist-agent.md` — priority and scheduling

## Related Files

- `agents/golf-ios-agent.md`
- `agents/delivery-agent.md`
- `prompts/` — generated prompts for Golf tasks

---

_Skill file for LiaoZong framework. Update when project constraints change._
