"""Build the graphics-stack issue inventory with explicit local assessment notes."""

import html
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
DIRECTORY = ROOT / "docs/graphics-issue-audit"
# Priority concerns this fork's development; it is not a security severity rating.
NOTES = {
    "go-gl/gl": {
        174: ("P1", "Оптимизация", "Нужен прототип с whitelist noescape; nocallback требует отдельного режима без Go debug callbacks. Автоматическая разметка всего API небезопасна."),
        172: ("P1", "Подтверждено по дереву", "В v3.0/gles2/KHR действительно нет dummy.go. Добавить проверку сохранения заголовков в vendor/пакетном архиве."),
        169: ("P2", "Не воспроизведено", "Жалоба на frame pacing Linux/GNOME; примеры upstream работают. Нужны C/Go comparator, latency percentiles, swap interval и upload-stall measurements."),
        168: ("P2", "Окружение/не воспроизведено", "Пакет v2.1/gl существует и собирается на этой Windows-машине; без cgo он исключается. Нужна понятная диагностика toolchain/CGO."),
        167: ("P1", "Кандидат загрузчика", "OSMesa-контекст должен использовать свой resolver через InitWithProcAddrFunc. По умолчанию POSIX использует GLX; контейнеры автора не запускались."),
        155: ("P2", "Сборка/не воспроизведено", "Обсуждение смешивает отсутствующий cross C compiler и runtime Init. Нужны отдельные cross-build и реальный Windows smoke; clang/mingw и контекст фиксировать явно."),
        154: ("P2", "Окружение/не воспроизведено", "Staticcheck зависит от cgo headers/compiler. Само сообщение не доказывает дефект GL; добавить воспроизводимую матрицу анализа."),
        152: ("P1", "Диагностика", "Отсутствующая glActiveTexture связана с resolver/current context/driver. Проверять контекст и capabilities; данная NVIDIA-конфигурация проходит Init."),
        151: ("P1", "Не воспроизведено", "Краш BufferSubData на Windows AMD требует того же GPU и минимального воспроизведения; нельзя объявлять размер/профиль доказанной причиной."),
        150: ("P1", "Диагностика", "Инициализировать каждый импортированный профиль после MakeContextCurrent на закреплённом OS thread. Защита от вызова незагруженной функции нужна в debug API."),
        148: ("P1", "Несоответствие профилей", "gles2 — имя API Khronos, а v3.0/3.1/3.2 — версии GLES. Генерация настоящего GLES 2.0 успешно проверена; текущий Fyne запрашивает 2.0, импортируя 3.1."),
        145: ("P3", "Ограничение окружения", "Статическая линковка требует static GL stack. Наличие динамической libGL не гарантирует -static; документировать поддерживаемые варианты."),
        132: ("P2", "Частично устарело", "Указанного reflect.SliceHeader уже нет в проверенном conversions.go. Str использует указатель на настоящий StringHeader; переход на unsafe.StringData/SliceData остаётся полезен."),
        128: ("P1", "Типизация API", "uint имеет платформенную ширину и не заменяет uint32 индексов GL. Добавить helpers для []uint16/[]uint32 и проверки размера/типа; raw API сохранить."),
        116: ("P2", "Работает в проверенном сценарии", "Синхронный debug callback на NVIDIA действительно вызван. C trampoline игнорирует переданный Go callback pointer; глобальный callback остаётся ограничением нескольких контекстов."),
        109: ("P2", "Архитектурное расширение", "Путь без cgo отсутствует. Генерировать второй backend, сравнивать ABI/callbacks/performance; отдельное решение также требуется для GLFW."),
        104: ("P1", "Диагностика", "Профили имеют отдельные глобальные таблицы функций. all-core с lenientInit не гарантирует поддержку каждой функции; debug guards и capabilities предпочтительнее запрета нескольких профилей."),
        102: ("P2", "Релизный процесс", "GL закрепляется pseudo-version. Собственные semver tags, changelog и generation manifest нужны после выбора публичного module path."),
        80: ("P1", "Частично закрыто API", "PtrOffset deprecated; некоторые WithOffset уже есть. Остаются недостающие overloads. Отключение cgocheck не является исправлением."),
        77: ("P1", "Воспроизведено", "Strs(\"A\", \"B\") выделяет два байта AB без NUL. Контракт с документацией расходится; Fyne уже добавляет терминатор вручную."),
    },
    "go-gl/glfw": {
        426: ("P1", "Фрагмент воспроизведён", "На Linux signed char формат даёт 44 символа/45 байт при guid[33]. Маскирование даёт 32. Тот же код есть в C GLFW 3.3 и 3.4; реальное устройство и эксплуатация не проверялись."),
        425: ("P1", "Подтверждён конфликт условий", "OpenBSD/Wayland native access включается тегом, тогда как C backend OpenBSD остаётся X11. FreeBSD требует отдельной проверки. BSD runtime не запускался."),
        413: ("P1", "Не воспроизведено", "Внешний unknown scancode превращается в panic через panicError. Нужны safe query и тесты fn/media keys на Wayland; соответствующего desktop здесь нет."),
        410: ("P2", "Инфраструктура", "ASan/MSan матрица для C/Go boundary; сообщения из неинструментированных X11/dbus/libdecor нельзя автоматически считать ошибкой GLFW."),
        407: ("P1", "Оптимизация", "Go 1.27.1 снимает барьер минимальной версии. Nocallback учитывать для error callback и событий из native setters; сначала измерить выбранные функции."),
        405: ("P3", "Старый профиль/не воспроизведено", "Отчёт относится к GLFW 3.2 и macOS. Проверить актуальный Cocoa backend; эта Windows/Linux-проверка его не охватывает."),
        402: ("P1", "Сборочная переносимость", "Wayland include paths должны приходить из pkg-config, а не зависеть от расположения headers в дистрибутиве. OpenSUSE отдельно не проверялась."),
        377: ("P2", "Окружение/не воспроизведено", "Требуется macOS SDK/Xcode comparator. Предложение из комментария собирать GLES на macOS не является проверенным исправлением."),
        376: ("P2", "Окружение/не воспроизведено", "Проверить GCC execution charset и Unicode-путь checkout на Windows. Текущий X:/fyne + GCC 13.2 работает."),
        370: ("P1", "Функциональное расширение", "IME требует composition/preedit, selection/candidate rectangle и платформенных callbacks, а не одного CharCallback. Патч только Go-обёрток недостаточен."),
        359: ("P2", "Нативный upstream", "Warning _GNU_SOURCE и прежние missing declarations — разные состояния. Исправление должно переживать обновление C snapshot; смена mkostemp на mkstemp меняет семантику."),
        357: ("P2", "Не воспроизведено", "Gamepad mapping/axis complaint требует устройства автора или записанного backend fixture. Не считать общим дефектом всех контроллеров."),
        353: ("P2", "Есть в GLFW 3.4", "ContextNoError уже экспортирован в v3.4/window.go; старый v3.3 требует отдельного backport. Это не отсутствующая возможность рабочего 3.4 API."),
        345: ("P2", "Не воспроизведено", "NSGL pixel format может зависеть от виртуальной машины/драйвера/context hints. Проверять реальную macOS и параметры выбора контекста."),
        296: ("P2", "Функциональное расширение", "LockKeyMods сообщает modifier события, а не произвольный текущий toggle state. Нужен отдельный контракт CapsLock/NumLock query по платформам."),
        292: ("P1", "Не воспроизведено", "Bluetooth gamepad mapping errors способны перейти в panic при Init. Отделить ошибки внешнего устройства от programmer errors; тесты через fixtures."),
        288: ("P2", "Теги частично/уже добавлены", "В текущих 3.3 и 3.4 есть NetBSD build tags. Это не доказательство успешного NetBSD runtime; дополнить BSD CI."),
        274: ("P1", "Связано с #292", "HORIPAD/8BitDo mapping и глобальный error channel; нужен comparator C vs Go и безопасная обработка внешних устройств."),
        188: ("P2", "Частично выполнено upstream", "Для v3.4/glfw есть v0.1.0-pre.2. Стабильный релиз и собственная версия wrapper должны быть отделены от версии C GLFW."),
    },
    "go-gl/glow": {
        127: ("P1", "Недостающие overloads", "DrawElementsInstanced/ReadPixels PBO требуют offset variants. Вносить в XML overrides и проверять checkptr, zero offset и pointer-width."),
        86: ("P2", "API расширение", "Один GLConst int32 не подходит uint64 TIMEOUT_IGNORED и смешанным enum/integer параметрам. Добавить domain-aware names/string helpers без замены raw constants."),
        63: ("P1", "Частично выполнено", "WithOffset уже решает некоторые вызовы. Дополнить полный перечень, убрать uintptr-to-pointer из безопасного API; legacy raw API не менять молча."),
        51: ("P2", "Архитектура", "Разделить registry/IR и cgo/FFI templates, генерировать static typed dispatch. WebGL/Metal/Vulkan не становятся GL-поддержкой от смены loader."),
        32: ("P2", "Диагностика", "Generate debug/trace build variant с call counters и source labels, выключаемый из release. Сравнить overhead; использовать внешние GPU captures."),
        31: ("P1", "API расширение", "Генерировать typed slices, lengths и offset overloads рядом с raw bindings. XML len metadata требует overrides и тестов retained-pointer функций."),
        23: ("P1", "Оптимизация генерации", "Restrict уже есть. Затем исследовать общие C bridges по ABI signature, object-file splitting, size и cold build time; не обещать исторические проценты экономии."),
        16: ("P1", "Воспроизведено на Linux", "EGL 1.5 генерируется, но Cgo не определяет AHardwareBuffer/EGLClientPixmapHI/wl_display/wl_resource. Нужна корректная обработка struct prefix и native type overrides."),
        2: ("P2", "Частично проверено", "EGL compile failure подтверждён; GLX/WGL generation/native usage требует отдельных fixtures. Наличие ненулевого proc address не доказывает поддержку extension."),
    },
}


def cell(text):
    return html.escape(str(text)).replace("|", "\\|").replace("\n", " ")


def main():
    snapshot = json.loads((DIRECTORY / "open-issues.json").read_text(encoding="utf-8"))
    rows = []
    lines = ["# Открытые issues GL / GLFW / Glow", "", f"Снимок GitHub API: {snapshot['fetched_at']}. Все открытые issues трёх репозиториев; PR исключены. Полные тела и все комментарии сохранены в [open-issues.json](open-issues.json).", "", "Приоритет относится к развитию нашего стека. Статус отделяет локальное воспроизведение, чтение кода и неподтверждённые сообщения пользователей. Закрытые issues и отдельный трекер нативной C GLFW в этот реестр не входят.", ""]
    for repository, issues in snapshot["repositories"].items():
        assert {issue["number"] for issue in issues} == set(NOTES[repository]), f"Refresh assessment notes for {repository}"
        lines += [f"## {repository}: {len(issues)}", "", "| Issue | Приоритет | Проверка | Решение / следующее действие |", "| --- | --- | --- | --- |"]
        for issue in issues:
            priority, status, note = NOTES[repository][issue["number"]]
            rows.append({"repository": repository, "number": issue["number"], "title": issue["title"], "url": issue["html_url"], "priority": priority, "status": status, "note": note, "comments_saved": len(issue["discussion"])})
            lines.append(f"| [#{issue['number']}]({issue['html_url']}) {cell(issue['title'])} | {priority} | {cell(status)} | {cell(note)} |")
        lines.append("")
    (DIRECTORY / "INVENTORY.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    (DIRECTORY / "inventory.json").write_text(json.dumps({"fetched_at": snapshot["fetched_at"], "records": rows}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Assessed {len(rows)} open issues; saved {sum(row['comments_saved'] for row in rows)} comments")


if __name__ == "__main__":
    main()
