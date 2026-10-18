# Библиотеки и функции для визуального ТЗ: реализация и проверки

12 октября 2026 года, рабочее дерево `X:\fyne`, Go 1.27.1.
Работа продолжает предыдущие итерации, описанные в
[LIBRARY_DEVELOPMENT_VALIDATION.md](LIBRARY_DEVELOPMENT_VALIDATION.md).
Входное ТЗ остаётся источником требований; этот отчёт не является приёмкой всего ТЗ.

## Реализовано

| Компонент | Новые функции и поведение |
| --- | --- |
| `deps/gl` | Immutable `Functions`/`Load` для compact/material/GLES 2; отдельный `material` package: проверенные core/ARB/EXT RGBA8 FBO, GPU snapshot нижних слоёв без CPU readback, framebuffer/viewport scope, два ping-pong targets на слой, cache и бюджет |
| `deps/gl/material` | GPU Gaussian в два прохода, rounded mask, tint, статичный noise; premultiplied alpha; обычная и linear-light обработка sRGB с decode **до** bilinear interpolation; обе оси масштаба учитываются отдельно |
| `deps/glfw`, 3.3/3.4/3.5 | Optional `QueryWindowVisuals`/`ApplyWindowVisuals`: документированные Windows DWM backdrop/caption/corners/redirection alpha; support OS отдельно от actual WGL surface; preflight, HRESULT/Accepted, owner-thread check, безопасное состояние после Destroy |
| `deps/glow` | `-dispatch`, методы и offset overloads для отдельной immutable таблицы; уникальные имена receiver; отказ от unsafe callback profiles до публикации; `ShaderSourceString` через C-stack pointer array без C allocation; manifest сохраняет `context_dispatch` |
| `deps/dbus/portal` | Retry временных ошибок чтения с backoff 250 ms → максимум 5 s; отсутствующий owner не опрашивается; `WaitReady(ctx)` для нескольких ожидающих без потребления UI Updates; closed watcher/connection публикует unavailable |
| Fyne | Desktop/GLES painter действительно использует таблицу своего контекста и новый upload shader source; optional `SystemAppearanceProvider` без расширения обязательного Settings; один XDG provider/connection на приложение; автоматическая тема сохраняет явные overrides/custom Theme; reduced motion подавляет анимации; один ожидающий UI callback применяет latest snapshot; shutdown отменяет pending work |

Новые публичные настройки сохраняют availability каждого ключа. Стандартный XDG
portal не предоставляет text scale и permission для прозрачности; эти поля остаются
недоступными. Windows/macOS пока не заполняют новый Fyne snapshot: их существующие
theme paths сохраняются. Источник: [XDG Settings](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.Settings.html).

## Ресурсы и контракт renderer

`material.NewPool(device, 0)` задаёт 64 MiB **запрошенного RGBA8 texel storage**.
Максимум два слоя, у каждого два targets. Resize учитывает старые и новые textures
одновременно и проверяет бюджет до allocation. При ошибке второго allocation
первый новый target освобождается, старый cache сохраняется. Ошибка render callback
инвалидирует результат; renderer должен выбрать fallback.

Cache key включает surface/content/style generations, ScaleX/Y, radius/corners,
tint/noise и linear policy. Статичная сцена повторно не копируется и не размывается.
Нулевой framebuffer даёт `ErrInvalid`: rendering следует пропустить до восстановления.
Source generation должен описывать **нижние слои**, а не прошлый результат материала.

`CopySnapshot` копирует текущий read framebuffer на GPU, сохраняет texture binding
и запрещает feedback. Координаты физические, с началом снизу слева; source bounds и
read buffer выбирает compositor. Операция не ресайзит изображение. Downsampling
в отдельный уменьшенный target делается GPU draw pass приложения; `Size.Downsample`
выделяет размеры с округлением вверх.

`RenderTarget` восстанавливает read/draw FBO и viewport при returned error и panic.
`Blur.Apply` владеет program/vertex input/array buffer/texture unit 0/blend/scissor/
cull/color mask; после него compositor восстанавливает свой draw state. Это явный
контракт, а не автоматическое сохранение всего GL state. Core-profile VAO и
desktop alpha-test/rasterizer-discard state задаёт compositor. Базовые runtime
проверки выполнены для GL 2.1 compatibility и GLES 2.

`Pool.Close`/`Blur.Close` выполняются с matching context current до его уничтожения.
После context loss используется `Abandon`, затем новые table/device/pool/program.
Таблица функций не владеет контекстом и не делает GL calls thread-safe.

64 MiB здесь не являются измеренным полным GPU allocation процесса. Program/VBO,
window framebuffer, другие caches и driver bookkeeping учитываются отдельно.
Linear-light intermediate RGBA8 имеет конечную точность; float/HDR pipeline не добавлялся.

## Проверки текущего кода

Raw logs: [material-runtime](graphics-assessment/material-runtime).
Сводка с SHA-256 источников и logs: [material-runtime.json](graphics-assessment/material-runtime.json).

| Проверка | Результат |
| --- | --- |
| Windows, Go race/cgocheck2 | Glow, material/capabilities, compact/material/GLES units; все три GLFW native suites; daemon-independent DBus units normal/no-TCP и portal |
| Windows native | NVIDIA RTX 4080 SUPER / GL 4.6, driver 616.92; GPU snapshot, blur/sRGB/mask/tint/noise, cache, resize peak, separate FBO/PBO/viewport restore, error/panic restore, два независимых контекста |
| Windows DWM | Build 26300; composition=true, remote=false, high contrast=false (known=true); actual WGL pixel format 12 / alpha 8; accepted/read-back backdrop roles; caption/corners/alpha accepted; opaque-alpha preflight ничего не меняет; wrong thread и после Destroy дают typed errors |
| Linux native | Go 1.27.1, Mesa 25.0.5 llvmpipe; X11 GL 4.5, X11 strict GLES 2.0, X11 GL 2.1 **EXT FBO** с выключенным ARB FBO, Wayland/EGL GL и strict GLES 2.0; каждый probe создаёт два отдельных контекста |
| Linux D-Bus | Реальный isolated daemon, полные normal/no-TCP suites с race; owner/restart/spoof/overflow/canceled-old-read и recovery без нового signal повторены 10 раз; app-owned provider lifecycle повторён пять раз |
| Fyne | app/painter/GLFW driver race suites на Windows и Linux; two-window capture после failed global reload; native test выполняет Show/capture на GUI owner thread и ждёт согласованный canvas size |
| Pure material | JS/WASM compilation, `CGO_ENABLED=0`; это compile, не WebGL runtime |
| Generation | Повторный compact/material generation не меняет bytes или timestamps; manifest и overloads сохраняются |
| Patches | Пять patches воспроизводятся на чистых pinned upstream archives; bytes сравниваются с рабочими форками, reverse check проходит |

GPU CPU-reference сравнивает весь 32×24 RGBA8 output для обычного и linear-light
проходов, включая fractional sample offsets и rounded mask. Максимальная ошибка
на проверенных драйверах — 2 уровня байта; threshold — 3. Все pixels соблюдают
RGB ≤ A и RGB=0 при A=0. Noise меняет pixels, соблюдает alpha contract и остаётся
побайтово тем же после cache hit. Snapshot сверяется с независимым input pattern.
Для каждого контекста: 2001 cache hits, 6 pool target allocations; `UsedBytes=23040`,
`PeakBytes=29184` при budget 32768. Отдельные targets test fixture не входят в эти pool stats.

Microbenchmark cache-hit path (три samples, median): Windows **35.83 ns/op**,
Linux **38.74 ns/op**, оба **0 B/op / 0 allocs/op**. Это lookup/validation в тёплом
pool на fake device; показатель не является FPS или latency общего renderer.
Native `RenderTarget` и `ShaderSourceString`: 0 Go allocations в probes.

Новый handwritten Go код проходит `go vet`. Полный generated `gl` сохраняет
предыдущие предупреждения `Ptr(uintptr)`/deprecated `PtrOffset`; их нельзя считать
устранёнными только потому, что новый слой использует безопасные overloads.

Полная native Windows DBus suite без daemon завершается ошибками `dbus-launch`
и отказом default system TCP endpoint. Её прохождение на Windows не заявляется;
полный protocol/IPC путь проверен на `f-test-bench`, как разрешено инструкциями.
Go units на Windows вынесены в повторяемый `scripts/test-dbus-windows-unit.py`.
Первый Linux запуск пришлось перенести в короткий `/tmp/fmrt-...`: длинный
temporary path превышал лимит адреса Unix socket. Это исправление harness.

## Воспроизведение и границы ТЗ

Основные scripts: `material-runtime-probe/main.go`,
`validate-material-runtime-linux.sh` (полный прогон; `--gpu-only` для изменения
только GPU layer; `--resume-renderers` после сохранённых предыдущих checks),
`verify-unchanged-gl-generation.py`, `verify-fork-patches.py`.
Для Linux export используйте короткий изолированный каталог, отдельные Go caches,
Xvfb, `dbus-run-session`, собственный Weston runtime; не меняйте system/session buses.

DWM support: backdrop с build 22621, caption/corners с 22000, redirection alpha
с 26100 и только при verified transparent WGL surface. Accepted HRESULT не
доказывает видимый материал; API не переопределяет native frame/buttons.
Источники: [DWM attributes](https://learn.microsoft.com/en-us/windows/win32/api/dwmapi/ne-dwmapi-dwmwindowattribute),
[backdrop roles](https://learn.microsoft.com/en-us/windows/win32/api/dwmapi/ne-dwmapi-dwm_systembackdrop_type).

Остаётся продуктовая интеграция material graph в canvas/Fyne popup, optional
WindowVisuals state machine/request IDs/observers и provider остальных ОС.
Не реализованы AppKit material views и compositor blur protocols Wayland/X11;
на этих путях native API возвращает `ErrVisualUnsupported`, in-app primitives доступны.
Не выполнены visual acceptance DWM на 22621/22631, AMD/Intel/ARM64, RDP/VM,
macOS Intel/Apple Silicon, physical fractional-DPI matrix, WebGL/mobile/BSD runtime,
performance gate 5×60 seconds, input-to-present и измеренный полный GPU memory.
Эти проверки нельзя заменять зелёными microbenchmarks или всегда opaque fallback.

Изменения оставлены в рабочем дереве; releases, tags, commits и push не выполнялись.
Временные validation каталоги этой работы удаляются после сохранения evidence.
