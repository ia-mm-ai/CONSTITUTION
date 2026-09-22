"""Reconstruct a portable edition, not a Body's State or a Git repository."""

import argparse
import subprocess
import sys
from pathlib import Path

from verify import ROOT, binding, check_export, encode, publication, write_fresh


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--revision", help="Full, locally available Git commit ID")
    source.add_argument("--export", type=Path, help="Already exported edition directory")
    parser.add_argument("--repository", type=Path, default=ROOT)
    parser.add_argument("--manifest-sha256", help="Independently trusted digest for --export")
    parser.add_argument("--destination", required=True, type=Path)
    parser.add_argument("--allow-incomplete", action="store_true")
    args = parser.parse_args()
    try:
        if args.export:
            if not args.manifest_sha256:
                raise ValueError("--export requires --manifest-sha256")
            files, _ = check_export(args.export, args.manifest_sha256)
        else:
            files, _ = publication(args.repository, args.revision, args.allow_incomplete)
        write_fresh(args.destination, files)
        digest = binding(files["manifest.json"])["sha256"]
        _, report = check_export(args.destination, digest)
        print(encode(report).decode(), end="")
        return 0
    except (ValueError, KeyError, TypeError, OSError, subprocess.SubprocessError) as error:
        print(encode({"status": "FAILED", "error": str(error)}).decode(), end="")
        return 1


if __name__ == "__main__":
    sys.exit(main())
