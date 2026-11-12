#!/usr/bin/env bash
# Test runner for the YAML scalar round-trip work.
#
# Usage:
#   ./test.sh [--output_path <junit.xml>] <base|new>
#
#   base  run the repository's existing tests (all production packages); these
#         must pass both before and after a change to scalar emission.
#   new   run the scalar round-trip tests; these fail before the fix and pass
#         after it.
set -uo pipefail

cd "$(dirname "$0")"

OUTPUT_PATH=""
if [ "${1:-}" = "--output_path" ]; then
  OUTPUT_PATH="$2"
  shift 2
fi

MODE="${1:-new}"

# Pin test-generated randomness so runs are reproducible.
export YTT_SEED=1
# A fixed level of parallelism keeps runs identical regardless of host CPUs.
export GOMAXPROCS=4
# All dependencies are vendored; never reach for the network.
export GOFLAGS="-mod=vendor"
export CGO_ENABLED=0

TEST_LOG="$(mktemp)"
STATUS=0

run_tests() {
  local packages=("$@")
  go test -count=1 -v "${packages[@]}" 2>&1 | tee -a "$TEST_LOG"
  local status=${PIPESTATUS[0]}
  if [ "$status" -ne 0 ]; then
    STATUS=$status
  fi
}

case "$MODE" in
  base)
    # Regression suite: every production package. Scalar emission sits in the
    # vendored YAML encoder and feeds all of YAML output, formatting, template
    # evaluation and the CLI, so the whole tree is the blast radius.
    run_tests ./pkg/...
    ;;
  new)
    run_tests ./test/scalars/...
    ;;
  *)
    echo "unknown mode: $MODE (expected base or new)" >&2
    exit 2
    ;;
esac

if [ -n "$OUTPUT_PATH" ]; then
  go-junit-report -set-exit-code < "$TEST_LOG" > "$OUTPUT_PATH" || true
fi

exit "$STATUS"
