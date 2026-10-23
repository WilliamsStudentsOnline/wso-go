#!/usr/bin/env python3
"""Write a concise Atlas migrate validate summary (CI sticky comment + logs)."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def list_versions(migrations_dir: Path) -> list[str]:
    """Return sorted golang-migrate version names from *.up.sql files."""
    if not migrations_dir.is_dir():
        return []
    versions: list[str] = []
    for path in sorted(migrations_dir.glob("*.up.sql")):
        # 20261007235756_baseline.up.sql → 20261007235756_baseline
        versions.append(path.name.removesuffix(".up.sql"))
    return versions


def render(*, outcome: str, versions: list[str], log: str) -> str:
    n = len(versions)
    parts = ["## Atlas\n"]

    if outcome != "success":
        parts.append("❌ **Migration validation failed**\n")
        if n == 0:
            parts.append("_No `*.up.sql` migrations found under `db/atlas/migrations/`._\n")
        else:
            parts.append(f"Checked **{n}** migration(s).\n")
        trimmed = log.strip()
        if trimmed:
            parts.append("```")
            parts.extend(trimmed.splitlines()[-80:])
            parts.append("```\n")
        else:
            parts.append("_No atlas output captured._\n")
        return "\n".join(parts)

    if n == 0:
        parts.append("ℹ️ No migrations detected\n")
        parts.append(
            "_`db/atlas/migrations/` has no `*.up.sql` files; nothing to validate._\n"
        )
        return "\n".join(parts)

    parts.append(f"✅ Migrations passed (**{n}** version{'s' if n != 1 else ''})\n")
    show = versions if n <= 20 else versions[:10]
    for v in show:
        parts.append(f"- `{v}`")
    if n > 20:
        parts.append(f"- …and {n - 10} more")
    parts.append("")
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
        help="Write markdown here (also printed to stdout)",
    )
    args = parser.parse_args()

    log = ""
    if args.log is not None and args.log.is_file():
        log = args.log.read_text(errors="replace")

    body = render(
        outcome=args.outcome,
        versions=list_versions(args.migrations_dir),
        log=log,
    )
    sys.stdout.write(body)
    if args.output:
        Path(args.output).write_text(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
