#!/usr/bin/env python3
"""Render the sticky OpenAPI PR comment from check artifacts."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def load_json(path: Path):
    if not path.is_file():
        return None
    return json.loads(path.read_text())


def spectral_line(spectral_json: Path) -> str:
    data = load_json(spectral_json)
    if data is None:
        return "❓ Spectral: not run"
    errors = sum(1 for x in data if x.get("severity") == 0)
    warnings = sum(1 for x in data if x.get("severity") == 1)
    if errors:
        return f"❌ Spectral: **{errors} error(s)**, {warnings} warning(s)"
    if warnings:
        return f"✅ Spectral: 0 errors ({warnings} warning(s))"
    return "✅ Spectral: clean"


def freshness_line(status_path: Path) -> str:
    if not status_path.is_file():
        return "❓ Docs freshness: not run"
    status = status_path.read_text().strip()
    if status == "pass":
        return "✅ API documentation is in sync!"
    return "❌ API documentation is out of date: run `make api`, commit, and re-push"


def breaking_section(status_path: Path, report_path: Path, base_ref: str) -> list[str]:
    lines: list[str] = []
    if not status_path.is_file():
        return ["❓ Breaking-change check: not run"]
    status = status_path.read_text().strip()
    report = report_path.read_text().strip() if report_path.is_file() else ""
    if status == "pass":
        lines.append(f"✅ No breaking changes vs `{base_ref}`")
    else:
        lines.append(f"❌ Breaking changes vs `{base_ref}`")
        if report:
            lines.append("")
            lines.append("```")
            lines.append(report[:6000])
            lines.append("```")
    return lines


def changelog_section(changelog_json: Path, *, breaking_ok: bool) -> tuple[list[str], bool]:
    data = load_json(changelog_json)
    if data is None:
        return [], False
    if not data:
        if breaking_ok:
            return ["", "_No API changelog vs base._"], False
        return [], False

    lines = [
        "",
        "<details>",
        f"<summary><strong>Changelog ({len(data)})</strong></summary>",
        "",
    ]
    for item in data[:40]:
        level = {0: "error", 1: "warn", 2: "info", 3: "info"}.get(
            item.get("level"), "info"
        )
        op = item.get("operation") or ""
        path = item.get("path") or ""
        text = item.get("text") or item.get("id") or "change"
        where = f" — `{op} {path}`".rstrip() if path else ""
        lines.append(f"- **{level}**: {text}{where}")
    if len(data) > 40:
        lines.append(f"- …and {len(data) - 40} more")
    lines.extend(["", "</details>"])
    return lines, True


def publish_nudge(*, has_api_delta: bool) -> list[str]:
    if not has_api_delta:
        return []
    return [
        "",
        "> **NOTE:** Downstream clients will not see these API changes yet! "
        "See README for next steps.",
    ]


def render(results_dir: Path, base_ref: str) -> str:
    breaking_status = ""
    status_path = results_dir / "breaking.status"
    if status_path.is_file():
        breaking_status = status_path.read_text().strip()

    changelog_lines, has_changelog = changelog_section(
        results_dir / "changelog.json",
        breaking_ok=breaking_status == "pass",
    )
    has_api_delta = has_changelog or (breaking_status == "fail")

    parts = [
        "## OpenAPI",
        "",
        freshness_line(results_dir / "freshness.status"),
        spectral_line(results_dir / "spectral.json"),
    ]
    parts.extend(
        breaking_section(
            results_dir / "breaking.status",
            results_dir / "breaking.txt",
            base_ref,
        )
    )
    parts.extend(changelog_lines)
    parts.extend(publish_nudge(has_api_delta=has_api_delta))
    parts.append("")
    return "\n".join(parts)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--results-dir",
        type=Path,
        default=Path("openapi-results"),
    )
    parser.add_argument(
        "--base-ref",
        default="origin/master",
        help="Base ref shown in the comment",
    )
    parser.add_argument("-o", "--output", type=Path, help="Write markdown here")
    args = parser.parse_args()

    body = render(args.results_dir, args.base_ref)
    sys.stdout.write(body)
    if args.output:
        args.output.write_text(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
