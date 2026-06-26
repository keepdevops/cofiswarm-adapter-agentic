# cofiswarm-adapter-agentic

Cofiswarm component: `adapter-agentic`.

**Latest release: [v1.3.0](https://github.com/keepdevops/cofiswarm-adapter-agentic/releases/tag/v1.3.0)** — chat-completions input validation, HTTP server timeouts, no-silent-failure handlers, and a curated `/v1/info`. See the [release notes](https://github.com/keepdevops/cofiswarm-adapter-agentic/releases/tag/v1.3.0) for the full changelog.

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
