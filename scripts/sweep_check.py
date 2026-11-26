#!/usr/bin/env python3
"""Checks one captured screen for make sweep: the expected title is on its
first line and no line is cut at the window's edge. Prints a result line
and exits 1 on a failure."""
import sys
import unicodedata


def width(s):
    return sum(2 if unicodedata.east_asian_width(c) in "WF" else 1 for c in s)


def main():
    w, name, title, full_ok = int(sys.argv[1]), sys.argv[2], sys.argv[3], sys.argv[4] == "1"
    lines = sys.stdin.read().split("\n")
    top = [l.strip() for l in lines if l.strip()][:3]  # a boxed title starts with its border
    first = " | ".join(top)
    cut = [l for l in lines if width(l.rstrip()) >= w]
    problems = []
    if title not in first:
        problems.append(f"top is {top[0][:40] if top else ''!r}, not the title {title!r}")
    if cut and not full_ok:
        problems.append(f"{len(cut)} line(s) reach the edge: {cut[0].strip()[:50]!r}")
    print(f"  {'FAIL' if problems else 'ok  '} {name:12} {'; '.join(problems)}")
    sys.exit(1 if problems else 0)


main()
