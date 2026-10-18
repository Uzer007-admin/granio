"""Generate a searchable inventory from the complete open-issue snapshot."""

from collections import Counter
import datetime
import html
import json
from pathlib import Path
import re


ROOT = Path(__file__).resolve().parent.parent
DIRECTORY = ROOT / "docs" / "issue-audit"


def topic(issue):
    title = issue["title"].lower()
    labels = {label["name"] for label in issue["labels"]}
    if "testing" in labels or re.search(r"\btests?\b|assertimage|assert.*markup|fuzz", title):
        return "Тесты и диагностика"
    if "optimization" in labels or re.search(r"memory|cache|\bcpu\b|slow|performance|\blag|\bfps\b|alloc|leak", title):
        return "Производительность и память"
    if re.search(r"panic|crash|segfault|sigsegv|race|hang|freez|shutdown|not responsive|unresponsive", title):
        return "Стабильность и потоки"
    if re.search(r"vulnerab|dependencies|\bcgo\b|install|packag|\bbuild|compil|manifest|keystore|cross.compile|version|metadata", title):
        return "Сборка зависимости и установка"
    if "Accessibility" in labels or re.search(r"accessibility|\ba11y\b|high contrast", title):
        return "Доступность"
    if re.search(r"keyboard|keybind|shortcut|\bkeys?\b|modifier|mouse|cursor|input|touch|tap|drag|focus|selection|selectable|clipboard|paste|\bime\b", title):
        return "Ввод фокус и навигация"
    if "Storage" in labels or "FileDialog" in labels or re.search(r"filedialog|file.dialog|folder|directory|directories|\buri\b|storage|filesystem|webdav|files? open|fileopen|filesave", title):
        return "Файлы диалоги и хранилище"
    if "Internationalisation" in labels or re.search(r"locale|language|translat|\brtl\b|unicode|chinese|greek|cjk|glyph", title):
        return "Локализация и языки"
    if re.search(r"font|render|paint|\bsvg\b|image|texture|pixel|\bdpi\b|scal|blur|flick|graphic|opengl|\bgles\b|\begl\b|transparen|color|colour", title):
        return "Графика шрифты и масштабирование"
    if re.search(r"window|wayland|\bx11\b|systray|system.tray|notification|\bdock\b|portal|android|\bios\b|wasm|web.driver|browser|mobile", title):
        return "Платформы окна и интеграция с ОС"
    if re.search(r"widget|entry|richtext|textgrid|\blist\b|\btable\b|\btree\b|menu|scroll|tab|layout|container|button|form|label|dialog|binding|accordion|toolbar|slider|popup|pop.up|check|radio|validation", title):
        return "Виджеты компоновка и привязки данных"
    return "API документация и прочие запросы"


def kind(issue):
    labels = {label["name"] for label in issue["labels"]}
    for key, label in (
        ("bug", "bug"), ("unverified", "unverified"),
        ("enhancement", "enhancement"), ("information-needed", "information-needed"),
        ("question", "question"), ("meta", "meta"),
    ):
        if key in labels:
            return label
    return "прочие"


def cell(value):
    return html.escape(str(value)).replace("|", "\\|").replace("\n", " ").replace("\r", " ")


def main():
    data = json.loads((DIRECTORY / "open-issues.json").read_text(encoding="utf-8"))
    snapshot_time = datetime.datetime.fromisoformat(data["fetched_at"]).astimezone(datetime.timezone(datetime.timedelta(hours=3)))
    records = []
    for repository, issues in data["repositories"].items():
        records.extend({
            "repository": repository, "number": issue["number"],
            "title": issue["title"], "url": issue["html_url"],
            "topic": topic(issue), "kind": kind(issue),
            "labels": [label["name"] for label in issue["labels"]],
            "created_at": issue["created_at"], "updated_at": issue["updated_at"],
            "comments": issue["comments"],
        } for issue in issues)
    summary = {
        "fetched_at": data["fetched_at"],
        "repositories": {name: len(issues) for name, issues in data["repositories"].items()},
        "fyne_topics": dict(Counter(record["topic"] for record in records if record["repository"] == "fyne-io/fyne")),
        "fyne_kinds": dict(Counter(record["kind"] for record in records if record["repository"] == "fyne-io/fyne")),
        "classification": "Topics are assigned by title keywords and upstream labels; kinds use the first matching label, not local runtime verification.",
        "records": records,
    }
    (DIRECTORY / "inventory.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    lines = [
        "# Реестр открытых issues Fyne и apps", "",
        f"Снимок от {snapshot_time:%Y-%m-%d %H:%M} МСК содержит все {len(data['repositories']['fyne-io/fyne'])} открытых issues Fyne и {len(data['repositories']['fyne-io/apps'])} issues apps.", "",
        "Тема назначена по заголовку и меткам. Метки передают состояние upstream; подтверждение конкретного дефекта в нашем форке требует воспроизведения. Полные описания сохранены в [open-issues.json](open-issues.json). Анализ приоритетов находится в [ISSUES_AUDIT.md](../ISSUES_AUDIT.md).", "",
    ]
    for repository in data["repositories"]:
        lines.extend([f"## {repository}", "", "| Issue | Заголовок | Тема | Метки upstream |", "| --- | --- | --- | --- |"])
        for record in records:
            if record["repository"] == repository:
                lines.append(f"| [#{record['number']}]({record['url']}) | {cell(record['title'])} | {cell(record['topic'])} | {cell(', '.join(record['labels']) or 'нет')} |")
        lines.append("")
    (DIRECTORY / "INVENTORY.md").write_text("\n".join(lines), encoding="utf-8")
    print(json.dumps({key: value for key, value in summary.items() if key != "records"}, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
