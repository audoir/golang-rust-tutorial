#!/usr/bin/env bash
# Start all tutorial servers.
# Each server is launched in the background; Ctrl-C kills them all.
#
# Usage: ./scripts/start-servers.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# ── Cleanup on exit ──────────────────────────────────────────────────────────
cleanup() {
  echo ""
  echo "Stopping all servers..."
  kill "${PIDS[@]}" 2>/dev/null || true
  wait "${PIDS[@]}" 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT INT TERM

PIDS=()

# ── Next.js (port 3000) ───────────────────────────────────────────────────────
echo "▶ Starting Next.js dev server on http://localhost:3000 ..."
(cd "$REPO_ROOT" && npm run dev) &
PIDS+=($!)

# ── Go (port 3005) ─────────────────────────────────────────────────────────────
echo "▶ Starting Go server on http://localhost:3005 ..."
(cd "$REPO_ROOT/go-api" && go run .) &
PIDS+=($!)

echo ""
echo "All servers running. Press Ctrl-C to stop."
wait
