# GPU material pipeline, surface generations и portal deltas

Проверено 12 октября 2026 года на Go 1.27.1. Это следующий реализованный этап
библиотечной части `Fyne_Visual_Modernization_TZ.md`; документ требований не
является заявлением о полной реализации или аппаратной приёмке ТЗ.

## Что реализовано

| Библиотека | Изменение | Практический результат |
| --- | --- | --- |
| `gl` | `material.Renderer`, `Blur.Capture`, `Pool.UpdateAdaptive`, source/crop identity в cache key | Один API соединяет GPU crop, downsampling, два blur-прохода и общий бюджет двух слоёв. Выбирается лучшее доступное 1/2/4/8 разрешение; resize учитывает старые и новые texels. Статичный результат возвращается без GL calls и Go allocations |
| `glow` | `GetFramebufferViewport` и `GetTargetBindings` для dispatch-профилей | Фиксированные запросы состояния объединены в один cgo transition с C-stack outputs. Compact/material/strict GLES2 регенерированы; минимальная версия и число обязательных GL entry points не увеличены |
| `glfw` 3.3/3.4/3.5 | `SurfaceTracker`, `UpdateSurface`, clipped `FramebufferRegion`, GL-origin conversion | Geometry/DPI/context дают отдельное resource generation. Focus не сбрасывает cache. Hidden/iconified/zero framebuffer дают `Drawable=false`; destroyed window возвращает typed error. Callbacks не заменяются |
| `dbus` | `ApplyAppearanceChange` и ordered delta path в appearance watcher | Обычные доверенные изменения обновляют snapshot без ReadAll/owner round-trips. Startup/restart, malformed payload, неизвестный порядок и overflow сохраняют authoritative recovery |

`SettingChanged` передаёт новое значение ключа, поэтому watcher может применять
валидированную typed delta. Он проверяет unique sender, namespace/path/interface
и incoming sequence. Начальный GetNameOwner reply служит нижней границей для
queued settings; более новые события, пришедшие до старого ReadAll reply, не
отбрасываются. Final owner check отдельно защищает смену владельца.
Monotonic dropped counter проверяется перед принятием ответа/дельты, поскольку
выбор из готовых Go channels не гарантирует обработку Overflow token первым.
Основание протокола: [XDG Settings](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.Settings.html).

## Контракты для compositor

Один Renderer обслуживает панели одного окна/context generation. Source — уже
нарисованные нижние слои в premultiplied RGBA8 texture. Source lifetime, scene
ordering, content/style/surface generations и приведение UI units к физическим
пикселям принадлежат compositor. Crop — physical pixels с GL bottom-left origin.
Texture identity, source dimensions и crop movement автоматически участвуют в key.
Собственный старый результат слота запрещён как его новый источник до resize/delete.

Два слоя используют один pool budget, default 64 MiB. Старые targets сохраняются
при невозможности выделить замену; implicit eviction отсутствует. Ошибки draw не
маскируются повторным rendering с другим quality. Render error инвалидирует cache;
при `ErrBudget` приложение выбирает явный fallback. После context loss вызывается
`Abandon`, затем создаются новые table/device/renderer. До Destroy выполняется Close.

GPU capture использует четыре bilinear samples с clamp к crop. Factor 1 сохраняет
texel sampling, factor 2 — 2x2 box. Factors 4/8 являются дешёвой аппроксимацией,
могут терять мелкие детали и не называются точным area filter. Linear policy
декодирует texels до interpolation/filtering и re-encodes captured RGBA8;
blur/mask/tint/noise сохраняют premultiplied output и RGB=0 при A=0.

FBO/viewport восстанавливаются, в том числе при error/panic; остальные draw states
принадлежат compositor. GL 2.1 compatibility / GLES2 используются как baseline.
GL core-profile VAO и дополнительные desktop rasterization states задаёт caller.
Пример и API: [material README](../deps/gl/material/README.md),
[surface tracker contract](../deps/glfw/WINDOW_VISUALS.md).

## Проверки

Raw logs: [material-pipeline](graphics-assessment/material-pipeline).
Сводка: [material-pipeline.json](graphics-assessment/material-pipeline.json) —
31 validation/benchmark logs, 37 source SHA-256 и 12 native contexts.
[Source verification](graphics-assessment/material-pipeline/source-verification.txt)
подтверждает совпадение всех 37 источников с Linux snapshot.

| Стенд / проверка | Результат |
| --- | --- |
| Windows x64, NVIDIA RTX 4080 SUPER / GL 4.6 | Два независимых context tables; global failed reload не ломает rendering; crop 1/2/4/8 normal/linear, odd dimensions, crop edge clamp, alpha, shared adaptive renderer budget/cache/release/Close |
| Linux X11 | Mesa 25.0.5 llvmpipe: GL 4.5, strict GLES 2.0 и GL 2.1 с EXT FBO при отключённом ARB FBO; по два независимых context |
| Linux Wayland/EGL | Собственный headless Weston runtime, scale=2; GL и strict GLES 2.0, по два context; это smoke test, не physical mixed-DPI/compositor acceptance |
| Native GLFW 3.3/3.4/3.5 | Windows/Linux suites с race; tracker сверяется с actual hidden surface, stable generation и zero allocations; после Destroy — typed error. Fractional conversion/zero framebuffer/generation проверены units |
| D-Bus | Полные Linux normal/no-TCP suites с реальным isolated daemon и race. Appearance watcher tests повторены 10 раз; 20 последовательных обычных deltas не увеличивают ReadAll count. Malformed delta запускает recovery, concurrent change до старого ReadAll response не теряется |
| Windows D-Bus | Daemon-independent units normal/no-TCP и полная platform-eligible portal unit suite; native daemon interoperability здесь не заявляется |
| Fyne | app/painter/GLFW driver race suites Windows/Linux; Linux app-owned provider lifecycle повторён пять раз |
| Generation / patches | Compact/material повторная генерация сохраняет bytes/mtime; пять fork patches применяются к чистым pinned archives, файлы совпадают, reverse check проходит |
| WASM | Pure material package компилируется с `CGO_ENABLED=0`; browser runtime не проверен |

Максимальная ошибка crop против CPU reference — **1/255** на всех 12 contexts;
blur reference — **2/255**, threshold 3/255. RGB≤A и RGB=0 при A=0 проверены.
GPU Capture, fixed-size queries, RenderTarget и ShaderSourceString дают **0 Go
allocations**. State-query snapshot — **один cgo transition**. Статичный Renderer
даёт **0 GL calls / 0 allocations**, даже с race/cgocheck2/checkptr=2.
Проверка выявила девять временных float allocations в widened finite validation
при checkptr; float32 comparisons сохранили NaN/Inf rejection и устранили их.

Микроизмерения: три samples, median; fixed fake-device/cache scenarios, не FPS.

| Операция | Windows ns/op | Linux ns/op | B/op / allocs/op |
| --- | ---: | ---: | ---: |
| Pool cache hit | 28.31 | 34.58 | 0 / 0 |
| Adaptive pool cache hit | 65.67 | 79.87 | 0 / 0 |
| Renderer cache hit | 48.44 | 66.08 | 0 / 0 |
| Typed appearance delta | 20.55 | 24.44 | 0 / 0 |
| Static SurfaceTracker | 9.00 | 9.92 | 0 / 0 |

## Воспроизведение и оставшиеся границы

Для Linux: `pack-graphics-validation.py`, короткий собственный temporary path,
`validate-material-runtime-linux.sh` с отдельными caches, Xvfb/isolated D-Bus и
собственным Weston. Для Windows: `test-dbus-windows-unit.py`, native GLFW suites,
`material-runtime-probe/main.go` с `graphics_probe`, generation/patch verification.
`record-material-pipeline.py --work <validation-directory>` отвергает неполные,
failed либо allocating результаты и записывает summary/source/log hashes.

Pool budget покрывает принадлежащие ему RGBA8 texels, включая resize peak.
Compositor обязан отдельно учитывать внешнюю source texture, shaders, window
framebuffer, другие caches и driver bookkeeping; полный GPU memory процесса не
измерялся. На первоначальном Windows запуске compiler встретил нехватку virtual
memory; проверки завершены с `GOMAXPROCS=2`, `GOFLAGS=-p=1`.

Интеграция Renderer в Fyne canvas/popups и optional WindowVisuals state machine
ещё требуется. AppKit native materials, Wayland/X11 compositor blur protocols,
providers остальных ОС и native visual acceptance здесь не добавлены. Не выполнены
macOS/BSD/ARM/mobile/WebGL runtime, physical fractional/mixed DPI, Windows
22621/22631/RDP/VM, GPU vendor matrix и acceptance 5×60 seconds с present/input
latency. Microbenchmarks и accepted DWM attributes этих проверок не заменяют.

Изменения находятся в рабочем дереве. Releases, tags, commits и push не выполнялись.
Собственные temporary directories удалены после сохранения validation evidence.
