"""Save all currently open upstream issues using the public, read-only GitHub API."""

import argparse
import datetime
import json
from pathlib import Path
import urllib.request


def read_page(url):
    request = urllib.request.Request(url, headers={
        "Accept": "application/vnd.github+json",
        "User-Agent": "fyne-fork-issue-audit",
        "X-GitHub-Api-Version": "2022-11-28",
    })
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response), response.headers.get("X-RateLimit-Remaining")


def fetch(repository, include_comments=False):
    issues = []
    page = 1
    while True:
        url = f"https://api.github.com/repos/{repository}/issues?state=open&per_page=100&page={page}&sort=created&direction=desc"
        entries, remaining = read_page(url)
        issues.extend(entry for entry in entries if "pull_request" not in entry)
        print(f"{repository}: page={page}, issues={len(issues)}, API remaining={remaining}", flush=True)
        if len(entries) < 100:
            break
        page += 1
    if include_comments:
        for issue in issues:
            discussion = []
            page = 1
            if issue["comments"]:
                while True:
                    entries, remaining = read_page(f"{issue['comments_url']}?per_page=100&page={page}")
                    discussion.extend(entries)
                    if len(entries) < 100:
                        break
                    page += 1
            issue["discussion"] = discussion
            print(f"{repository}#{issue['number']}: {len(discussion)} comments", flush=True)
    return issues


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=Path("docs/issue-audit/open-issues.json"))
    parser.add_argument("--repository", action="append", help="Repository owner/name; repeat for multiple repositories.")
    parser.add_argument("--include-comments", action="store_true", help="Also save all comments on each open issue.")
    arguments = parser.parse_args()
    snapshot = {
        "fetched_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "scope": "All open issues; pull requests excluded; " + ("all issue comments included." if arguments.include_comments else "issue comments not included."),
        "repositories": {name: fetch(name, arguments.include_comments) for name in (arguments.repository or ("fyne-io/fyne", "fyne-io/apps"))},
    }
    arguments.output.parent.mkdir(parents=True, exist_ok=True)
    arguments.output.write_text(json.dumps(snapshot, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Saved snapshot: {arguments.output}", flush=True)


if __name__ == "__main__":
    main()
