"""Проверка и публикация подготовленных репозиториев экосистемы Granio."""

import argparse
import json
from pathlib import Path
import subprocess
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parent.parent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--create", action="store_true", help="создать отсутствующие форки")
    parser.add_argument("--push", action="store_true", help="опубликовать подготовленную ветку dev/granio")
    parser.add_argument("--set-default", action="store_true", help="сделать опубликованную ветку основной")
    parser.add_argument("--only", action="append", help="обработать только указанное имя репозитория")
    args = parser.parse_args()
    manifest = json.loads((ROOT / "ecosystem/repositories.json").read_text(encoding="utf-8"))
    credential = subprocess.run(["git", "credential", "fill"], input="protocol=https\nhost=github.com\n\n",
                                capture_output=True, text=True, check=True, timeout=20)
    fields = dict(line.split("=", 1) for line in credential.stdout.splitlines() if "=" in line)
    headers = {"Authorization": "Bearer " + fields["password"], "Accept": "application/vnd.github+json",
               "User-Agent": "Granio-Development", "X-GitHub-Api-Version": "2022-11-28"}

    def api(path, body=None, method=None):
        req = urllib.request.Request("https://api.github.com/" + path,
                                     data=json.dumps(body).encode() if body is not None else None,
                                     headers=headers, method=method)
        with urllib.request.urlopen(req, timeout=40) as response:
            return json.load(response)

    owner = manifest["owner"]
    if api("user")["login"] != owner:
        raise SystemExit("Учётная запись GitHub не совпадает с владельцем из реестра")
    for item in manifest["repositories"]:
        name = item["name"]
        if args.only and name not in args.only:
            continue
        path = (ROOT / item["path"]).resolve()
        if not path.is_relative_to(ROOT) or not path.is_dir():
            raise SystemExit("Рабочая копия отсутствует внутри проекта: " + item["path"])
        upstream = item["upstream"].removeprefix("https://github.com/")
        try:
            repo = api(f"repos/{owner}/{name}")
        except urllib.error.HTTPError as error:
            if error.code != 404:
                raise
            if not args.create:
                print(f"{name}: репозиторий ещё не создан")
                continue
            # После неопределённого результата повторный запуск сначала выполнит GET.
            api(f"repos/{upstream}/forks", {"name": name, "default_branch_only": False}, "POST")
            print(f"{name}: форк создан; повторный запуск проверит его готовность")
            continue
        source = (repo.get("parent") or {}).get("html_url")
        if source != item["upstream"]:
            raise SystemExit(f"Неожиданный исходный репозиторий для {name}: {source}")
        if args.create or args.push:
            api(f"repos/{owner}/{name}", {"description": "Форк " + item["upstream"] + ". Цель: " + item["purpose"],
                                        "homepage": item["upstream"]}, "PATCH")
        if args.push:
            branch = subprocess.check_output(["git", "-C", str(path), "branch", "--show-current"], text=True).strip()
            if branch != "dev/granio":
                raise SystemExit("Ожидается подготовленная ветка dev/granio: " + item["path"])
            dirty = subprocess.check_output(["git", "-C", str(path), "status", "--porcelain", "--ignore-submodules=all"], text=True)
            if dirty.strip():
                raise SystemExit("Сначала сохраните изменения в коммите: " + item["path"])
            subprocess.run(["git", "-C", str(path), "push", f"https://github.com/{owner}/{name}.git",
                            "HEAD:refs/heads/dev/granio"], check=True)
        if args.set_default:
            api(f"repos/{owner}/{name}/git/ref/heads/dev/granio")
            api(f"repos/{owner}/{name}", {"default_branch": "dev/granio"}, "PATCH")
        print(f"{name}: {repo['html_url']}")


if __name__ == "__main__":
    try:
        main()
    except urllib.error.HTTPError as error:
        message = error.read().decode("utf-8", errors="replace")
        raise SystemExit(f"GitHub отклонил запрос ({error.code}). Состояние сохранено; частые повторы не выполняются.\n{message}")
