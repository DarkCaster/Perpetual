#!/usr/bin/env bash
#
# Perpetual.sh - launcher script that detects current OS architecture
# and executes the matching Perpetual binary located alongside this script.
#
# Fallback: if a matching binary cannot be found next to this script,
# a generic "Perpetual" binary in the current working directory will be used instead.
#
# All command line arguments and stdin are passed through to the selected binary,
# and its exit code becomes the exit code of this script.

set -u

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

# Detect CPU architecture and map it to the naming scheme used by release binaries.
detect_arch() {
    local machine
    machine="$(uname -m 2>/dev/null || echo unknown)"
    case "$machine" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        i386|i486|i586|i686|x86)
            echo "x86"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

ARCH="$(detect_arch)"

TARGET_BIN=""

if [ "$ARCH" != "unknown" ]; then
    CANDIDATE="$SCRIPT_DIR/Perpetual_${ARCH}.bin"
    if [ -f "$CANDIDATE" ]; then
        TARGET_BIN="$CANDIDATE"
    fi
fi

# Fallback to a generic "Perpetual" binary
if [ -z "$TARGET_BIN" ]; then
    FALLBACK="$SCRIPT_DIR/Perpetual"
    if [ -f "$FALLBACK" ]; then
        TARGET_BIN="$FALLBACK"
    fi
fi
if [ -z "$TARGET_BIN" ]; then
    FALLBACK="./Perpetual"
    if [ -f "$FALLBACK" ]; then
        TARGET_BIN="$FALLBACK"
    fi
fi

if [ -z "$TARGET_BIN" ]; then
    echo "Error: could not find a suitable Perpetual binary to run." >&2
    echo "Looked for: $SCRIPT_DIR/Perpetual_${ARCH}.bin and ./Perpetual" >&2
    exit 1
fi

# Make sure the resolved binary is executable (archive extraction may strip permissions).
if [ ! -x "$TARGET_BIN" ]; then
    chmod +x "$TARGET_BIN" 2>/dev/null
fi

if [ ! -x "$TARGET_BIN" ]; then
    echo "Error: found Perpetual binary at '$TARGET_BIN' but it is not executable, and permissions could not be changed." >&2
    exit 1
fi

exec "$TARGET_BIN" "$@"
