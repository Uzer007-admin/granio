# Высокие частоты, разрешения и native input/accessibility

Продолжение форка, 12 октября 2026: Go 1.27.1, Windows amd64 и Linux f-test-bench.
Полные результаты: [suite-summary.json](display-validation/suite-summary.json).
SHA-256 журналов и source inputs: [evidence.json](display-validation/evidence.json).
`tests_pass` включает subtests; это не число независимых test functions.

| Полный набор | Пакеты PASS | Passed test actions | FAIL |
| --- | ---: | ---: | ---: |
| Windows native + UIA, race | 47 | 4937 | 0 |
| Windows headless, race | 47 | 4751 | 0 |
| Linux native + AT-SPI, race | 49 | 5032 | 0 |

Последний повтор Windows native driver после Close: 211 actions, 0 FAIL.
Проверенный Linux снимок содержит 3217 source inputs. При первой сверке все
совпали по точным байтам с Windows workspace. Последующие изменения общего дерева
в GL material и GLFW surface tracker отражены в `linux_source_comparison` и не
включены в эту приёмку автоматически; проверенные hash сохраняются отдельно.
Все пять patches зависимостей воспроизводятся на чистых upstream snapshots.
fyne_settings.exe, apps.exe и fyne_gallery.exe пересобраны; gallery smoke прошёл.

## Реализованные изменения

- Desktop scheduler принимает 120/144/160/180/240/280/300/360 Гц, сохраняет
  привязку к периоду и пропускает устаревшие кадры после дорогого render.
  Browser использует requestAnimationFrame и прекращает запросы кадров в простое.
- `desktop.ScalableWindow`: opt-in ScaleExact, отдельные system/user/effective/
  framebuffer scales и live DPI callback. Legacy сохранён; маленький scale
  больше не округляется в нулевой делитель.
- Capture рисует canvas в back buffer перед чтением: GLX front buffer возвращал
  чёрное изображение без GL error. Вложенные context calls/panic восстанавливают
  исходный context. Close удаляет текстуры при действующем context своего окна.
- Windows TSF ITextStoreACP, native edit sessions/composition sink, UTF-16 ACP,
  selection/text locks, IMM reconversion/document-feed и password privacy.
  Linux XIM: перемещение preedit caret по символам, словам и строкам.
- Общая accessibility model: реальная text geometry, прокрутка без изменения
  selection/undo, numeric values, logical selection, table cells и стабильные IDs.
  UIA: Text/TextRange, RangeValue, Selection, Grid/GridItem и события.
  AT-SPI: Text/Value/Selection/Table и события через accessibility bus.
  Layout-контейнеры больше не скрывают строки от selection provider; native child
  index правильно сопоставляется логическому индексу после прокрутки.
- External drop: независимые OLE/Xlib Xdnd source processes. CF_HDROP достигает
  старого файлового callback при включённом MIME target. Неподдерживаемые desktop
  backend возвращают ErrDataDropUnsupported.
- `theme.NewModern`: общий opaque light/dark профиль, density, независимый text
  scale, accent и явный high contrast. Gallery содержит Unicode Entry/password,
  live-переключатели и виртуализированную Table 10000×3.
- Browser runtime выявил gl-js дефекты: named Enum в TexSubImage2D и raw Go slices
  в ReadPixels. Исправления сохранены в [локальном публичном snapshot](../deps/gl-js/UPSTREAM.md),
  выбранном через go.work. Добавлен русский Select placeholder.

## Выполненная проверка

| Область | Проверка | Граница результата |
| --- | --- | --- |
| Fyne/apps | Полные Windows native/headless и Linux native наборы с race | Результаты/skips в suite-summary; Windows driver повторён после изменения Close |
| Частоты | 8 частот 120–360 Гц; 400 ms native animation, idle wake | Scheduler callbacks; не измерение present/input-to-photon |
| Разрешения | 2048×1080, 2560×1440, 3840×2160, 5120×2880 × 100/125/200% | 12 native framebuffer/capture сценариев Windows + 12 Xvfb/GLX; размеры и пиксели |
| DPI | 8 system scales × 4 user scales и native ScaleExact | Геометрия/преобразования; physical multi-monitor DPI ещё не принят |
| TSF/UIA | Реальный ITfContext.RequestEditSession; reconversion, COM TextRange/selection/events/lifetime | Не заменяет CJK IME, Narrator/NVDA |
| AT-SPI | Отдельное D-Bus соединение: Unicode edit, privacy/detach, Table/Selection/Value и signal | Не является приёмкой Orca |
| Drop | Unicode, HTML, binary 512 KiB; Windows CF_HDROP; X11 URI/INCR | Реальный native IPC с самостоятельными source processes |
| Gallery | Windows, X11, Weston/Wayland: Capture и Close/Quit; wasm в headless Chrome | Общий renderer/widget путь; headless не заменяет physical display |
| Browser idle | За 750 ms delivered RAF: 4→4; после resize 4→7 | Статичная сцена не перерисовывается постоянно, resize пробуждает loop |

Под 2K явно проверены DCI 2048×1080 и QHD 2560×1440. RGBA8 4K — 31.64 MiB,
5K — 56.25 MiB на поверхность. Матрица не доказывает, что сложная сцена укладывается
в 2.778 ms на физическом 360 Гц дисплее: нужны GPU duration, present intervals,
latency и учёт временных GPU surfaces.

## Повторение

```powershell
$env:GOMAXPROCS='4'
$env:FYNE_TEST_DISPLAY_MATRIX='1'
$env:FYNE_TEST_DRAGDROP='1'
$env:FYNE_TEST_CLIPBOARD='1'
go test -timeout 5m -p 1 -race '-tags=accessibility,migrated_fynedo' -count=1 ./... ./extras/apps/...
go test -timeout 5m -p 1 -race '-tags=no_glfw,migrated_fynedo' -count=1 ./... ./extras/apps/...
go run -tags accessibility,migrated_fynedo ./cmd/fyne_gallery -fps 360 -smoke
```

Linux: X11/GL/Wayland development packages, Xvfb, D-Bus/at-spi2-core,
fontconfig/CJK fonts. `scripts/native-xdnd-source.c` собирается с `-lX11`.

```sh
xvfb-run -a -s '-screen 0 6400x3600x24' dbus-run-session -- env \
  FYNE_PLATFORM=x11 FYNE_TEST_ATSPI=1 FYNE_TEST_CLIPBOARD=1 \
  FYNE_TEST_DISPLAY_MATRIX=1 FYNE_TEST_DRAGDROP=1 FYNE_XDND_SOURCE=/path/to/native-xdnd-source \
  go test -timeout 5m -p 4 -race -tags=accessibility,migrated_fynedo -count=1 ./... ./extras/apps/...
```

Промежуточные Windows runs исчерпывали system commit: ThreadSanitizer error 1455
и compiler out of memory. Итоговые native/headless runs выполнены последовательно
с GOMAXPROCS=4; системные настройки памяти не менялись. Android binary-table
fixture skips — upstream; subprocess-only clipboard/OLE gates вызываются parent
тестом и не означают пропуск проверки IPC.

## Экспериментальные границы

Реальные CJK engines/reconversion UX; Narrator/NVDA/Orca; полная иерархия offscreen/
закрытых tree nodes и selection proxies, Table headers/relations, TextEdit/active
composition и детальные font attributes. Selection provider экспортирует видимые
children и не обещает complete offscreen accessible tree.

Не приняты external Wayland drag, virtual file streams, исходящий drag,
macOS native IME/AX/drag и mobile hardware matrix. Browser smoke фиксирует старое
предупреждение о недоступном filesystem cache на js; IndexedDB cache parity требует
отдельного изменения. Browser CJK font fallback не принят: в gallery часть глифов
заменяется missing-glyph markers. Physical 160–360 Гц, multi-monitor DPI, 4K/5K GPU performance,
HDR и native materials требуют аппаратной приёмки.

Переработка пожеланий приложенного документа:
[VISUAL_MODERNIZATION_REVIEW.md](VISUAL_MODERNIZATION_REVIEW.md). Windows-only
критерии не приняты автоматически. Изменения сохранены локально, без публикации.
