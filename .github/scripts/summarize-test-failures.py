#!/usr/bin/env python3
"""Build a concise GitHub Actions step summary from go test -json output."""

from __future__ import annotations

import json
import re
import sys
from collections import defaultdict
from pathlib import Path

DROP = re.compile(
    r"(?:"
    r"^\s*\[GIN\]|"
    r"^\s*=== RUN\b|"
    r"^\s*=== PAUSE\b|"
    r"^\s*=== CONT\b|"
    r"^\s*--- PASS:|"
    r"CREATE TABLE|"
    r"CREATE INDEX|"
    r"INSERT INTO|"
    r"UPDATE |"
    r"DELETE FROM|"
    r"SELECT |"
    r"rows affected|"
    r"gormigrate|"
    r"^\s*\(/|"  # gorm file location lines like (/path/file.go:123)
    r"^\s*$"
    r")",
    re.IGNORECASE,
)

# testify assertion blocks + panics + FAIL headers
SIGNAL = re.compile(
    r"(?:"
    r"Error Trace:|"
    r"^\s*Error:|"
    r"^\s*Test:|"
    r"^\s*Messages?:|"
    r"^\s*expected:|"
    r"^\s*actual\s*:|"
    r"--- FAIL:|"
    r"panic:|"
    r"\.go:\d+:"
    r")"
)


def main() -> int:
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <gotest.json>", file=sys.stderr)
        return 2

    path = Path(sys.argv[1])
    if not path.is_file():
        print(f"no test log at {path}", file=sys.stderr)
        return 1

    outputs: dict[tuple[str, str], list[str]] = defaultdict(list)
    failed: list[tuple[str, str]] = []

    with path.open() as f:
        for line in f:
            line = line.strip()
            if not line.startswith("{"):
                continue
            try:
                ev = json.loads(line)
            except json.JSONDecodeError:
                continue

            action = ev.get("Action")
            pkg = ev.get("Package") or ""
            test = ev.get("Test") or ""
            if not pkg or not test:
                continue

            key = (pkg, test)
            if action == "output":
                out = ev.get("Output") or ""
                for raw in out.splitlines():
                    if DROP.search(raw):
                        continue
                    outputs[key].append(raw.rstrip())
            elif action == "fail":
                failed.append(key)

    # unique while preserving order
    seen: set[tuple[str, str]] = set()
    ordered: list[tuple[str, str]] = []
    for key in failed:
        if key in seen:
            continue
        seen.add(key)
        ordered.append(key)

    print("## Test failures")
    if not ordered:
        print("No failed test cases found in the JSON log.")
        return 0

    print(f"{len(ordered)} failed test(s):\n")
    for pkg, test in ordered:
        short_pkg = pkg.removeprefix("github.com/WilliamsStudentsOnline/wso-go/")
        print(f"### `{short_pkg}` · `{test}`")
        lines = outputs.get((pkg, test), [])
        # Prefer assertion/panic lines; fall back to a short tail of remaining output
        useful = [ln for ln in lines if SIGNAL.search(ln)]
        # Also keep indented continuation lines right after a signal line
        if useful:
            show: list[str] = []
            keep_cont = False
            for ln in lines:
                if SIGNAL.search(ln):
                    show.append(ln)
                    keep_cont = True
                elif keep_cont and (ln.startswith("\t") or ln.startswith(" ")):
                    show.append(ln)
                else:
                    keep_cont = False
        else:
            show = lines[-20:]
        if show:
            print("```")
            for ln in show[-60:]:
                print(ln)
            print("```")
        print()

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
