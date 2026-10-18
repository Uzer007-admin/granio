# Проверка следующей итерации native stack

12 октября 2026 года. Go 1.27.1, `GOEXPERIMENT=cgocheck2`; race включён в suites,
Windows полный GL GPU probe также выполнен с `checkptr=2`.
Объём изменений и связь с ТЗ описаны в
[NATIVE_VISUAL_REQUIREMENTS.md](NATIVE_VISUAL_REQUIREMENTS.md).

## Реализованные исправления

- GL: повторный Init обновляет всю таблицу и при ошибке отсутствующей required
  функции. Проверено, что старый `glViewport` становится недоступным; имена
  custom resolver сохраняют прежний формат без NUL. Во всех 24 профилях появились
  `IsFunctionAvailable` и manifest. Адрес не доказывает API/extension поддержку.
- GLFW 3.3/3.4/3.5: SurfaceMetrics разделяет window geometry, framebuffer и
  content scale. C-stack bridge возвращает значения одним cgo-вызовом; native
  getters, zero geometry, attributes и отсутствие output allocations проверены.
- Glow: template/gofmt failure сохраняет предыдущие generated Go files;
  реальный rename failure после замены первого файла вызывает rollback.
  Caller-owned Go files не форматируются; marker в строке произвольного кода
  не даёт удалить `debug.go`. Manifest проверяется на порядок и соответствие
  availability switch.
- D-Bus: Store сохраняет wire signature Variant в slice/map/nested map;
  decoded `(ddd)` не превращается в `av`, typed struct conversion сохранён.
  Portal ReadAll и SettingChanged проверены двумя connections через настоящий
  daemon, с уникальным именем тестового сервиса.

## Нативная матрица

| Проверка | Windows amd64 | Linux f-test-bench amd64 |
| --- | --- | --- |
| Glow и GL 2.1 full/compact, GLES 2 unit tests | PASS, race/cgocheck2 | PASS, race/cgocheck2 |
| D-Bus codec/Store и portal unit tests | PASS, race; daemon tests не запускались | Полные normal и godbus_no_tcp suites PASS, race; portal daemon round-trip PASS |
| GLFW 3.3/3.4/3.5 native suites | PASS, race/cgocheck2 | PASS, Xvfb, race/cgocheck2 |
| Native C bridges / GUID fixtures | Bridge fixtures PASS | Bridges и GUID PASS, ASan/UBSan |
| GL rendering, shaders, readback | NVIDIA RTX 4080 SUPER, driver 616.92, GL 4.6 PASS; synchronous debug callback PASS | Compact GL/Mesa llvmpipe PASS |
| EGL/GLES 2 | Отдельный EGL runtime не проверен | EGL 1.5 / Mesa 25.0.5 PASS; override до настоящего ES 2.0, 142 core lookups и green readback PASS |
| Два окна и surface metrics | PASS; alpha и opaque attributes различаются | X11 PASS; Wayland-only build и реальный headless Weston/EGL PASS, scale=2 |
| Fyne app / painter / GLFW driver | Все три пакета PASS, race, migrated_fynedo | Все три пакета PASS, race, migrated_fynedo |

Windows — Windows 11 Pro build 26300, Intel Core i5-13400. Linux — Xeon Gold
5320, kernel 6.12.60; Weston 10.0.1 / headless / pixman, Mesa llvmpipe.
Полные результаты, включая исходные сбои до исправления, сохранены в
[visual-validation](graphics-assessment/visual-validation).

Под Weston окно 192x128 имеет framebuffer 384x256, окно 224x160 — 448x320;
content scales равны 2. Прозрачность первого и непрозрачность второго различаются
в native attributes. Readback premultiplied test color: Windows `[64,0,0,127]`,
Wayland `[64,0,0,128]`. Это проверка GL storage и surface diagnostics;
эффекты desktop compositor этим не приняты. В Xvfb без compositor запрошенный
transparent framebuffer сообщает `false`, хотя RGBA storage читается: запрос
alpha, GL pixel format и фактическая compositor возможность различаются.

## Измерения

Три последовательных benchmark runs на вариант, median. Сбор native metrics
включает исходные три geometry/scale getters и пять attribute queries; оба
варианта делают одинаковые запросы GLFW. Нативные main-thread операции выполнены
на заблокированном OS thread.

| Среда | Старые getters | SurfaceMetrics | Go allocations |
| --- | ---: | ---: | --- |
| Windows / native hidden window | 2 098 ns | 1 826 ns | 3 / 24 B -> 0 / 0 B |
| Linux / X11 / Xvfb | 128 413 ns | 150 555 ns | 3 / 24 B -> 0 / 0 B |

Windows median меньше примерно на 13%. В этих Linux samples median вырос
примерно на 17%; ускорение по времени не подтверждено. Возможные причины — native
server round trips и разброс VM измерений; их вклад отдельно не профилировался.
Helper стоит использовать при изменении
поверхности, сохраняя данные для render loop. Для времени кадров нужен отдельный
контролируемый benchmark. API добавлен без переключения старых hot paths.

Repeated GL Init average на Windows: полный профиль 7 357 us / 2 848 lookups,
компактный 52.2 us / 50 lookups; оба 0 Go allocations. На Linux компактный Init
18.3 us / 50 lookups / 0 Go allocations. Это одиночные серии repeated calls,
не доказательство FPS/present улучшения и не сравнение одинаковых графических
возможностей профилей. Compact profile уже используется Fyne desktop painter.

Raw samples, probe JSON, SHA-256 проверенных Linux файлов и всех manifest:
[visual-validation.json](graphics-assessment/visual-validation.json).

## Границы и повторение

Linux user/PID namespace authentication test явно SKIP: kernel возвращает
`ENOSPC`. Это ограничение namespace стенда, а не пройденный auth case.
Первый Linux portal run обнаружил настоящую потерю `(ddd)` в Store; после
исправления обе полные D-Bus suites прошли. Попытка включить обычный upstream
`TestSend` в Windows selection закончилась отсутствием `dbus-launch`; итоговая
Windows selection содержит только независимые codec/Store tests.

Первый дополнительный Windows Fyne build застал параллельный accessibility
черновик с обращением Table к удалённому `selected`. В следующей версии этого
файла несоответствие уже исправлено; повторный run трёх integration packages
прошёл. В этой итерации production-код ядра Fyne не менялся. Linux integration
проверяет экспорт рабочего дерева, Windows — более позднее живое дерево;
проверенные изменённые library sources совпадают по SHA-256.

Не проверены реальные portal implementations и restart, GNOME/KWin/wlroots,
physical fractional-DPI/mixed-monitors, macOS/BSD/ARM/mobile/browser runtime,
Windows D-Bus daemon и системные материалы Windows/macOS. Weston scale=2
не закрывает fractional-scale 1.25/1.5. Этот этап не является выполнением
AT-01–AT-14 из ТЗ.

```powershell
go test -race ./deps/glow ./deps/gl/v2.1/gl ./deps/gl/v2.1/fyne ./deps/gl/v2.0/gles2 ./deps/dbus/portal
go test -race -run 'Test(Store|Codec|Windows|PropertyContext|MessageLength|Encode|Decode|ReadFull|SigByteSize|Proto|Numeric)' ./deps/dbus
go test -race -tags glfw_integration ./deps/glfw/v3.5/glfw
go run -tags graphics_probe ./scripts/surface-probe/main.go
python scripts/export-fork-patches.py
python scripts/verify-fork-patches.py
```

На Linux exporter создаёт sources/module-proxy archives; после распаковки в
собственный validation directory запустить из `source`:

```sh
sh scripts/validate-visual-stack-linux.sh /absolute/validation-directory
```

Validation script останавливает только созданный им Weston и удаляет его
собственный runtime directory. Общий validation directory удаляется после
сохранения logs. Экспортированные пять fork patches воспроизведены на чистых
pinned upstream archives: 199 / 693 / 546 / 112 / 20 файлов совпадают, reverse
check прошёл. Данные: [fork-patch-replay.json](graphics-assessment/fork-patch-replay.json).
