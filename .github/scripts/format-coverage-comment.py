#!/usr/bin/env python3
"""Shrink the Cobertura PR comment: PR-touched packages first, rest collapsed."""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path

PKG_PREFIX = "github.com/WilliamsStudentsOnline/wso-go/"

# Section order for collapsed groups
SECTIONS = ("Jobs", "Services", "Lib", "Other")

SECTION_PREFIXES = {
    "Jobs": "jobs/",
    "Services": "services/",
    "Lib": "lib/",
}

ROW_RE = re.compile(
    r"^\|?\s*(?P<pkg>(?:\*\*)?github\.com/WilliamsStudentsOnline/wso-go/[^\s|*]+(?:\*\*)?)"
    r"\s*\|\s*(?P<rest>.+?)\s*\|?\s*$"
)
SUMMARY_RE = re.compile(r"^\|?\s*\*\*Summary\*\*", re.IGNORECASE)
SEP_RE = re.compile(r"^\|?\s*-{3,}")


def section_for(short_pkg: str) -> str:
    for name, prefix in SECTION_PREFIXES.items():
        if short_pkg == prefix.rstrip("/") or short_pkg.startswith(prefix):
            return name
    return "Other"


def short_name(pkg: str) -> str:
    return pkg.removeprefix(PKG_PREFIX).strip("*")


def packages_from_changed_files(files: list[str]) -> set[str]:
    """Map changed paths to Go import paths that might appear in coverage."""
    pkgs: set[str] = set()
    for path in files:
        if not path.endswith(".go"):
            continue
        # directory of the file is the package
        parent = str(Path(path).parent)
        if parent in (".", ""):
            pkgs.add(PKG_PREFIX.rstrip("/"))
        else:
            pkgs.add(PKG_PREFIX + parent)
    return pkgs


def git_changed_files(diff_range: str) -> list[str] | None:
    result = subprocess.run(
        ["git", "diff", "--name-only", diff_range],
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        print(result.stderr or result.stdout, file=sys.stderr)
        return None
    return [ln.strip() for ln in result.stdout.splitlines() if ln.strip()]


def parse_coverage_md(text: str) -> tuple[list[str], str, list[str], list[tuple[str, str]], str | None]:
    """Return (preamble_lines, header_line, sep_line_parts, rows, summary_line).

    rows are (full_package_cell, rest_of_row including trailing cells).
    """
    lines = text.splitlines()
    preamble: list[str] = []
    header: str | None = None
    sep: list[str] = []
    rows: list[tuple[str, str]] = []
    summary: str | None = None

    i = 0
    while i < len(lines):
        line = lines[i]
        if header is None and re.search(r"Package\s*\|", line, re.IGNORECASE):
            header = line
            i += 1
            # optional separator
            if i < len(lines) and SEP_RE.match(lines[i].replace(" ", "")):
                sep = [lines[i]]
                i += 1
            continue

        if header is not None:
            if SUMMARY_RE.match(line):
                summary = line
                i += 1
                continue
            m = ROW_RE.match(line)
            if m:
                pkg = m.group("pkg").strip().strip("*")
                rest = m.group("rest").rstrip("|").strip()
                rows.append((pkg, rest))
                i += 1
                continue
            if not line.strip():
                i += 1
                continue

        if header is None:
            preamble.append(line)
        i += 1

    if header is None:
        raise ValueError("no coverage table header found")

    return preamble, header, sep, rows, summary


def render_table(header: str, sep: list[str], rows: list[tuple[str, str]]) -> list[str]:
    out = [header]
    out.extend(sep)
    for pkg, rest in rows:
        out.append(f"{pkg} | {rest}")
    return out


def format_comment(
    text: str,
    touched_pkgs: set[str],
    *,
    changed_files_known: bool = False,
) -> str:
    preamble, header, sep, rows, summary = parse_coverage_md(text)

    touched_rows: list[tuple[str, str]] = []
    by_section: dict[str, list[tuple[str, str]]] = defaultdict(list)

    for pkg, rest in rows:
        short = short_name(pkg)
        full = pkg if pkg.startswith("github.com/") else PKG_PREFIX + short
        by_section[section_for(short)].append((pkg, rest))
        if full in touched_pkgs or pkg in touched_pkgs:
            touched_rows.append((pkg, rest))

    # stable sort within each section by package name
    touched_rows.sort(key=lambda r: r[0])
    for name in SECTIONS:
        by_section[name].sort(key=lambda r: r[0])

    parts: list[str] = []
    # keep badge / intro
    while preamble and not preamble[0].strip():
        preamble.pop(0)
    parts.extend(preamble)
    if parts and parts[-1].strip():
        parts.append("")

    if touched_rows:
        parts.append("Packages touched by this PR:")
        parts.append("")
        parts.extend(render_table(header, sep, touched_rows))
        parts.append("")
    elif touched_pkgs:
        parts.append("_No coverage rows for packages touched by this PR._")
        parts.append("")
    elif changed_files_known:
        parts.append("_No Go packages changed in this PR._")
        parts.append("")
    else:
        parts.append("_Could not determine packages touched by this PR._")
        parts.append("")

    for name in SECTIONS:
        section_rows = by_section.get(name, [])
        if not section_rows:
            continue
        parts.append("<details>")
        parts.append(f"<summary><strong>{name} ({len(section_rows)})</strong></summary>")
        parts.append("")
        parts.extend(render_table(header, sep, section_rows))
        parts.append("")
        parts.append("</details>")
        parts.append("")

    if summary:
        parts.append(summary)
        parts.append("")

    return "\n".join(parts).rstrip() + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "-i",
        "--input",
        default="code-coverage-results.md",
        help="markdown from irongut/CodeCoverageSummary",
    )
    parser.add_argument(
        "-o",
        "--output",
        help="write reformatted markdown here (default: overwrite --input)",
    )
    parser.add_argument(
        "--diff-range",
        help="git diff range for touched files (e.g. origin/master...HEAD)",
    )
    parser.add_argument(
        "--changed-file",
        action="append",
        default=[],
        dest="changed_files",
        help="changed file path (repeatable; skips git if set)",
    )
    args = parser.parse_args()

    in_path = Path(args.input)
    if not in_path.is_file():
        print(f"error: no coverage markdown at {in_path}", file=sys.stderr)
        return 1

    changed_files_known = False
    if args.changed_files:
        changed = args.changed_files
        changed_files_known = True
    elif args.diff_range:
        changed_or_none = git_changed_files(args.diff_range)
        if changed_or_none is None:
            changed = []
        else:
            changed = changed_or_none
            changed_files_known = True
    else:
        changed = []

    touched = packages_from_changed_files(changed)
    body = format_comment(
        in_path.read_text(),
        touched,
        changed_files_known=changed_files_known,
    )

    out_path = Path(args.output) if args.output else in_path
    out_path.write_text(body)
    sys.stdout.write(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
