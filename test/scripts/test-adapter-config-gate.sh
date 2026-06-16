#!/usr/bin/env bash
set -euo pipefail
ROLE="${1:?}"
YAML="$(cd "$(dirname "$0")/../standalone/etc/cofiswarm/${ROLE}" && pwd)/${ROLE}.yaml"
grep -qE 'dispatch_url:' "$YAML" && grep -qE 'slot_manager_url:' "$YAML"
echo "ok: ${ROLE} → dispatch + slot-manager"
