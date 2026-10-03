#!/usr/bin/env python3
"""Write a concise pass/fail test summary from go test -json output."""

from __future__ import annotations

import argparse
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

PKG_PREFIX = "github.com/WilliamsStudentsOnline/wso-go/"


def parse_failures(path: Path) -> list[tuple[str, str, list[str]]]:
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

    seen: set[tuple[str, str]] = set()
    ordered: list[tuple[str, str, list[str]]] = []
    for key in failed:
        if key in seen:
            continue
        seen.add(key)
        ordered.append((key[0], key[1], outputs.get(key, [])))
    return ordered


def failure_snippet(lines: list[str]) -> list[str]:
    useful = [ln for ln in lines if SIGNAL.search(ln)]
    if not useful:
        return lines[-20:]

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
    return show[-60:]


def render_pass() -> str:
    return "## Tests\n\n✅ All tests green\n"


def render_failures(failures: list[tuple[str, str, list[str]]]) -> str:
    if not failures:
        return (
            "## Tests\n\n"
            "❌ Tests failed, but no failed test cases were found in the JSON log.\n"
        )

    parts = [
        "## Tests\n",
        f"❌ **{len(failures)} failed test(s)**\n",
    ]
    for pkg, test, lines in failures:
        short_pkg = pkg.removeprefix(PKG_PREFIX)
        parts.append(f"### `{short_pkg}` · `{test}`\n")
        show = failure_snippet(lines)
        if show:
            parts.append("```")
            parts.extend(show)
            parts.append("```\n")
        else:
            parts.append("_No assertion output captured._\n")
    return "\n".join(parts)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "jsonfile",
        nargs="?",
        help="go test -json / gotestsum --jsonfile path (omit with --pass)",
    )
    parser.add_argument(
        "--pass",
        dest="passed",
        action="store_true",
        help="emit the all-green summary without reading a log",
    )
    parser.add_argument(
        "-o",
        "--output",
        help="write markdown to this file (also printed to stdout)",
    )
    args = parser.parse_args()

    if args.passed:
        body = render_pass()
    else:
        if not args.jsonfile:
            print("error: jsonfile required unless --pass", file=sys.stderr)
            return 2
        path = Path(args.jsonfile)
        if not path.is_file():
            print(f"error: no test log at {path}", file=sys.stderr)
            return 1
        body = render_failures(parse_failures(path))

    sys.stdout.write(body)
    if args.output:
        Path(args.output).write_text(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
