"""Construct a portable edition from one explicit, locally available commit."""

import argparse
import subprocess
import sys
from pathlib import Path

from verify import ROOT, binding, check_export, encode, publication, write_fresh


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--revision", required=True)
    parser.add_argument("--destination", required=True, type=Path)
    parser.add_argument("--repository", type=Path, default=ROOT)
    parser.add_argument("--allow-incomplete", action="store_true",
                        help="Emit an explicitly incomplete preview when the core checker is absent")
    args = parser.parse_args()
    try:
        files, report = publication(args.repository, args.revision, args.allow_incomplete)
        write_fresh(args.destination, files)
        digest = binding(files["manifest.json"])["sha256"]
        check_export(args.destination, digest)
        print(encode({"revision": args.revision, "verification": report,
                      "destination": str(args.destination), "manifest_sha256": digest}).decode(), end="")
        return 0
    except (ValueError, KeyError, TypeError, OSError, subprocess.SubprocessError) as error:
        print(encode({"status": "FAILED", "error": str(error)}).decode(), end="")
        return 1


if __name__ == "__main__":
    sys.exit(main())
