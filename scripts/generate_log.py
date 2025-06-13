#!/usr/bin/env python3
"""Simple prototype: read JSON file and print a single-line log for each entry.

Usage:
    python scripts/generate_log.py data/input.json

This will evolve as we iterate on log-output formats.
"""
import json
import sys
from datetime import datetime
from pathlib import Path


def main():
    if len(sys.argv) != 2:
        print("Usage: python generate_log.py <input.json>", file=sys.stderr)
        sys.exit(1)

    input_file = Path(sys.argv[1])
    data = json.loads(input_file.read_text())

    now = datetime.utcnow().isoformat()
    print(f"{now} INFO Read {len(data)} records from {input_file.name}")


if __name__ == "__main__":
    main()
