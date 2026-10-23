#!/usr/bin/env python3
"""Write a concise Atlas migrate validate summary (CI sticky comment + logs)."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def list_versions(migrations_dir: Path) -> list[str]:
    """Return sorted migration version names from *.up.sql files."""
    if not migrations_dir.is_dir():
        return []
    versions: list[str] = []
    for path in sorted(migrations_dir.glob("*.up.sql")):
        # 20261007235756_baseline.up.sql → 20261007235756_baseline
        versions.append(path.name.removesuffix(".up.sql"))
    return versions


def version_list(versions: list[str]) -> list[str]:
    n = len(versions)
    show = versions if n <= 20 else versions[:10]
    lines = [f"- `{v}`" for v in show]
    if n > 20:
        lines.append(f"- …and {n - 10} more")
    return lines


MAX_ERR_LINES = 20
MAX_ERR_CHARS = 240


def atlas_error_block(log: str) -> list[str]:
    """Atlas CLI text, trimmed for a short sticky-comment code block."""
    lines: list[str] = []
    for raw in log.splitlines():
        line = raw.strip()
        if not line:
            continue
        if len(line) > MAX_ERR_CHARS:
            line = line[: MAX_ERR_CHARS - 1] + "…"
        lines.append(line)
    if not lines:
        return ["No atlas output captured."]
    if len(lines) > MAX_ERR_LINES:
        return lines[-MAX_ERR_LINES:] + ["… (truncated)"]
    return lines


def render(*, outcome: str, versions: list[str], log: str) -> str | None:
    """Return sticky-comment markdown, or None to skip the comment."""
    n = len(versions)

    if outcome == "success" and n == 0:
        return None

    parts = ["## Migration Status\n"]
    label = "version" if n == 1 else "versions"

    if outcome != "success":
        parts.append(f"❌ Failed (**{n}** {label})\n")
        if versions:
            parts.extend(version_list(versions))
            parts.append("")
        parts.append("```")
        parts.extend(atlas_error_block(log))
        parts.append("```\n")
        return "\n".join(parts)

    parts.append(f"✅ Passed (**{n}** {label})\n")
    parts.extend(version_list(versions))
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

    versions = list_versions(args.migrations_dir)
    body = render(outcome=args.outcome, versions=versions, log=log)

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
