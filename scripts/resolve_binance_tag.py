#!/usr/bin/env python3
"""Resolve a git tag (or the latest tag) to a commit hash for binance-go."""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from typing import Dict, Tuple


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Resolve a binance-go tag or commit reference"
    )
    parser.add_argument(
        "--tag",
        dest="tag",
        default="",
        help="Specific tag or commit hash to resolve (leave empty for latest tag)",
    )
    parser.add_argument(
        "--repo",
        dest="repo",
        default="https://github.com/openxapi/binance-go.git",
        help="Repository URL to query",
    )
    return parser.parse_args()


def load_tags(repo: str) -> Tuple[Dict[str, str], Dict[str, str]]:
    try:
        data = subprocess.check_output(
            ["git", "ls-remote", "--tags", repo], text=True
        )
    except subprocess.CalledProcessError as exc:  # pragma: no cover - external command
        raise RuntimeError(
            f"git ls-remote failed for {repo} (exit {exc.returncode})"
        ) from exc
    if not data.strip():
        raise RuntimeError("no tags discovered in repository")

    direct: Dict[str, str] = {}
    deref: Dict[str, str] = {}
    for raw_line in data.strip().splitlines():
        parts = raw_line.strip().split()
        if len(parts) != 2:
            continue
        sha, ref = parts
        if not ref.startswith("refs/tags/"):
            continue
        name = ref[len("refs/tags/") :]
        if name.endswith("^{}"):
            deref[name[:-3]] = sha.lower()
        else:
            direct[name] = sha.lower()
    return direct, deref


def resolve_tag(tag: str, repo: str) -> Tuple[str, str]:
    direct, deref = load_tags(repo)

    def resolver(name: str) -> str | None:
        return deref.get(name) or direct.get(name)

    if tag:
        commit = resolver(tag)
        if not commit and re.fullmatch(r"[0-9a-fA-F]{7,40}", tag):
            return tag, tag.lower()
        if not commit:
            raise RuntimeError(f"unable to resolve tag '{tag}' in {repo}")
        return tag, commit

    all_tags = sorted(set(direct) | set(deref))
    if not all_tags:
        raise RuntimeError("no tags available to resolve")

    semver: list[Tuple[int, int, int, str]] = []
    pattern = re.compile(r"v(\d+)\.(\d+)\.(\d+)$")
    for name in all_tags:
        match = pattern.match(name)
        if match:
            semver.append(
                (int(match.group(1)), int(match.group(2)), int(match.group(3)), name)
            )

    if semver:
        semver.sort()
        candidate = semver[-1][-1]
    else:
        candidate = all_tags[-1]

    commit = resolver(candidate)
    if not commit:
        raise RuntimeError(f"unable to resolve commit for tag '{candidate}'")
    return candidate, commit


def main() -> int:
    args = parse_args()
    try:
        resolved_tag, commit = resolve_tag(args.tag.strip(), args.repo.strip())
    except RuntimeError as exc:
        print(str(exc), file=sys.stderr)
        return 1

    print(f"RESOLVED_TAG={resolved_tag}")
    print(f"RESOLVED_COMMIT={commit}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
