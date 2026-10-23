#!/usr/bin/env python3
"""Write a concise Atlas migrate validate summary (CI sticky comment + logs)."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

SUM_LINE = re.compile(r"^L\d+:\s*(.+)$")


def migration_count(migrations_dir: Path) -> int:
    if not migrations_dir.is_dir():
        return 0
    return len(list(migrations_dir.glob("*.up.sql")))


def clean_error(log: str) -> str:
    """Collapse atlas CLI noise into one short error message."""
    detail = ""
    err = ""
    other: list[str] = []
    for raw in log.splitlines():
        line = raw.strip()
        if not line:
            continue
        if line.startswith("You have a checksum error"):
            continue
        if line.startswith("Please check your migration files"):
            continue
        m = SUM_LINE.match(line)
        if m:
            detail = m.group(1)
            continue
        if line.startswith("Error: "):
            err = line.removeprefix("Error: ")
            continue
        other.append(line)

    if err and detail:
        return f"{err}: {detail}"
    if err:
        return err
    if detail:
        return detail
    if other:
        return "\n".join(other[-40:])
    return "No atlas output captured."


def render(*, outcome: str, count: int, log: str) -> str | None:
    """Return sticky-comment markdown, or None to skip the comment."""
    if outcome == "success" and count == 0:
        return None

    parts = ["## Migration Status\n"]
    if outcome != "success":
        parts.append("❌ Failed\n")
        parts.append("```")
        parts.append(clean_error(log))
        parts.append("```\n")
        return "\n".join(parts)

    parts.append("✅ Passed\n")
    return "\n".join(parts)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--outcome",
        required=True,
        choices=("success", "failure", "cancelled", "skipped"),
        help="GitHub step outcome for atlas migrate validate",
    )
    parser.add_argument(
        "--migrations-dir",
        type=Path,
        default=Path("db/atlas/migrations"),
        help="Path to golang-migrate SQL directory",
    )
    parser.add_argument(
        "--log",
        type=Path,
        help="Captured atlas stdout/stderr (used on failure)",
    )
    parser.add_argument(
        "-o",
        "--output",
        help="Write sticky-comment markdown here (omit file when skipping comment)",
    )
    parser.add_argument(
        "--github-output",
        type=Path,
        help="Append comment=true/false for workflow gating",
    )
    args = parser.parse_args()

    log = ""
    if args.log is not None and args.log.is_file():
        log = args.log.read_text(errors="replace")

    count = migration_count(args.migrations_dir)
    body = render(outcome=args.outcome, count=count, log=log)

    if body is None:
        sys.stdout.write("No migrations detected\n")
        if args.output:
            Path(args.output).unlink(missing_ok=True)
        if args.github_output:
            with args.github_output.open("a") as f:
                f.write("comment=false\n")
        return 0

    sys.stdout.write(body)
    if args.output:
        Path(args.output).write_text(body)
    if args.github_output:
        with args.github_output.open("a") as f:
            f.write("comment=true\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
