# cofiswarm-adapter-agentic

Cofiswarm component: `adapter-agentic`.

- Layout: [REPO-STANDARD-LAYOUT](https://github.com/keepdevops/cofiswarm-docs/blob/main/REPO-STANDARD-LAYOUT.md)
- Migration: [MIGRATION-SPRINTS](https://github.com/keepdevops/cofiswarm-docs/blob/main/MIGRATION-SPRINTS.md)

## FHS paths

| Path | Purpose |
|------|---------|
| `/etc/cofiswarm/adapter-agentic/` | config |
| `/var/lib/cofiswarm/adapter-agentic/` | state |
| `/var/log/cofiswarm/adapter-agentic/` | logs |

## Test

```bash
./test/scripts/assert-layout.sh adapter-agentic
```
