# Оценка развития библиотеки Fyne и графических зависимостей

Дата: 12 октября 2026 года. База: `7f99a87b0e1903ebc17f3c8c8f478b840a6eabe5`
с текущими локальными изменениями. Оценка охватывает Fyne, локальные `gl`, `glfw`
и `glow`. Производственный код в ходе оценки не исправлялся.

## Основной вывод

Наибольший практический потенциал сейчас — корректность кэшей и изображений,
обработка событий, перенос длинного текста и геометрия больших коллекций.
Собственные GL/GLFW позволяют дорабатывать нижний слой, но значительная часть
затрат находится выше него, в Fyne. Отдельный backend без cgo полезен для сборки
и распространения; его влияние на скорость выполнения нужно измерить.

Первый пакет работ: три воспроизведённых дефекта, совместимое исправление TIFF
decoder и надёжная база измерений. Затем — цикл событий и текст/коллекции.
После этого — выделение графического backend и путь без cgo. Функциональные
расширения следует строить на этих контрактах.

## Что уже есть и следует сохранить

| Область | Существующие возможности |
| --- | --- |
| Рендеринг | Glyph atlas, batching текста, прямоугольников и линий, повторное использование vertex buffers, кэш uniform/attribute locations и значений uniforms, clipping и проверки видимости. Есть GPU-примитивы, blur и пользовательские shaders. |
| Текст | `go-text/typesetting`, shaping и fallback шрифтов, кэши метрик и shaped runs. Новый shaper с нуля не требуется. |
| Коллекции | List/Table/Tree используют повторно видимые элементы. Есть обновление отдельных элементов, переменная высота строк List, headers, sticky rows/columns и изменение ширины колонок Table. |
| Проверки | Большой набор unit/race/golden-тестов; результаты предыдущей проверки Windows и Linux сохранены в [TEST_RESULTS.md](TEST_RESULTS.md). Они не заменяют измерения настоящего кадра и испытания IME/DPI. |
| Диагностика | `FYNE_GL_DEBUG` уже считает draws, batches, quads и новые textures. Следует расширить эти счётчики, сохранив дешёвый выключенный режим. |

В отчёте различаются: **воспроизведённый дефект**, **измеренная стоимость**,
**наблюдение по исходникам** и **предложение**. Последние два статуса не означают
подтверждённого ускорения. Предыдущий обзор issues находится в
[ISSUES_AUDIT.md](ISSUES_AUDIT.md); текущая оценка добавляет собственные замеры
и минимальные воспроизведения.

## Воспроизведённые дефекты

### R1. Новая запись кэша может сразу истечь после скачка часов

Путь: [internal/cache/base.go](../internal/cache/base.go), `setAlive` и `Clean`.
Срок жизни рассчитывается от `cachedNow`, обновляемого в `Clean`. Если часы уже
продвинулись больше чем на `ValidDuration`, а новый sample ещё не снят, новая
запись получает уже истёкший срок. Следующая очистка её удаляет.

Временный тест продвинул управляемые часы на `3 * ValidDuration`, добавил метрики
нового текста и вызвал `Clean(false)`. Получен размер `{0 0}` вместо `{25 11}`.
Это подтверждает логику при большом разрыве между samples. В обычном работающем
цикле `Clean` вызывается каждый tick; тест не утверждает, что любое бездействие
окна вызывает ошибку. Реальные suspend/resume и блокировка event loop требуют
отдельной проверки.

**Доработка:** актуализировать sample до обслуживания событий после паузы или
защитить регистрацию/продление от слишком старого sample. Сохранить дешёвый
горячий путь. Отдельно проверить восстановление layout после пересоздания
renderer: [cache/widget.go](../internal/cache/widget.go), `Renderer`, самостоятельно
не вызывает `Layout(wid.Size())`; это пока наблюдение по коду, не новое runtime-воспроизведение.

**Приёмка:** fake-clock test, сохранение новой записи после паузы, восстановление
геометрии Label/List/Tree после удаления renderer, benchmarks горячего пути.
Доказательство: [cache-clock-probe.txt](library-assessment/cache-clock-probe.txt).

### R2. Repeat самой клавиши-модификатора очищает её состояние

Путь: [window_desktop.go](../internal/driver/glfw/window_desktop.go),
`desktopModifierCorrected`, строка 738. Любая action, отличная от `Press`,
обрабатывается как отпускание. Для `KeyLeftControl/KeyLeftShift` с `Repeat`
получается нулевая маска вместо удерживаемого Ctrl/Shift.

Воспроизведение синтетическое: вызов существующей функции с соответствующими
key/mod/action. Оно не доказывает, что каждая ОС генерирует Repeat для самих
модификаторов. Повтор обычной буквы при удерживаемом Ctrl — другой случай:
у буквы `glfwKeyToModifier` возвращает ноль.

**Доработка:** разделить `Press`, `Repeat`, `Release`; учитывать левую/правую
клавишу, потерю фокуса и особенности AltGr.
**Приёмка:** последовательности событий и настоящая клавиатура на Windows/X11/Wayland.
Доказательство: [modifier-repeat-probe.txt](library-assessment/modifier-repeat-probe.txt).

### R3. OpenGL upload RGBA.SubImage игнорирует Stride

Путь: [gl/texture.go](../internal/painter/gl/texture.go), `imgToTexture`, строка 112.
В ветке `*image.RGBA` передаётся `i.Pix` и ширина/высота, но не учитывается
`i.Stride`. У `SubImage` шаг между строками остаётся от исходного изображения.

На настоящем скрытом OpenGL-контексте Windows загружено подизображение 2×2
из родительского RGBA 4×2. Его `Stride=16`, плотно упакованная строка — 8 байт.
Чтение GPU texture показало правильную первую строку и чужие пиксели во второй.

**Доработка:** сохранить прямую загрузку для плотно упакованного RGBA, а для
остальных изображений переупаковать строки либо использовать поддерживаемый
row-length/pixel-store путь с восстановлением GL state. GLES требует отдельного
варианта с учётом его capabilities.
**Приёмка:** обычный RGBA, SubImage, ненулевой Rect.Min, нестандартный Stride,
clipping и сравнение с software painter.
Доказательство: [rgba-stride-probe.txt](library-assessment/rgba-stride-probe.txt).

## Измеренная база производительности

Windows/amd64, Go 1.27.1, Intel Core i5-13400, `GOAMD64=v3`, cgo, GOMAXPROCS=16.
Пакеты запускались последовательно (`-p 1`), без `-race`; три измерения по 500 мс.
Ниже медианы обычной сборки без тега `migrated_fynedo`.

| Существующий benchmark | Время / операция | B/op | Аллокации / операция |
| --- | ---: | ---: | ---: |
| FrameEnsureMinSize, дерево 730 узлов | 114,90 мкс | 144 | 2 |
| FrameWalkTrees, то же дерево | 24,05 мкс | 144 | 2 |
| RunOnMain, dispatch пустой функции | 8,84 мкс | 18 | 0¹ |
| GetFontMetrics, 300 горячих lookup | 16,62 мкс | 0 | 0 |
| Text_splitLines | 13,16 мкс | 30 224 | 28 |
| Text_lineBounds_WrapWord, холодный кэш | 1,371 мс | 1 696 388 | 13 948 |
| Text_howManyRunesFit_latin | 1,159 мкс | 2 400 | 3 |
| Text_howManyRunesFit_chinese | 1,598 мкс | 928 | 4 |
| TextGrid_Refresh, fixture 1000×1000 символов | 22,83 мкс | 4 064 | 1001 |
| Opaque rectangle 1280×800, существующий fast path | 0,655 мс | 0 | 0 |
| Тот же fill через generic draw.Draw | 10,641 мс | 0 | 0 |

¹ Go печатает усреднённые allocs/op как целое; `0 allocs/op` вместе с ненулевыми
B/op не доказывает полное отсутствие редких аллокаций.

Важные границы этих результатов:

- Frame benchmarks измеряют CPU-проходы, а не весь GPU-кадр, present или
  задержку ввода. На этой fixture сами обходы не оправдывают переписывание всего дерева.
- WrapWord очищает кэш метрик на каждой операции и использует ширину 10.
  Это стресс холодного переноса, а не типичный кадр или скорость ввода любого Entry.
- TextGrid fixture содержит миллион символов, но benchmark вызывает `Refresh`
  без показанного окна; это не отрисовка миллиона glyphs.
- `BenchmarkContentMinSize` дал около 1,25 нс на cached getter. В нём два
  height overrides, а проверка `assert.Equal(minSize, minSize)` не проверяет результат.
  Он не измеряет scroll или пересчёт геометрии миллиона строк.
- Opaque fast path уже есть и в этом сравнении примерно в 16,2 раза быстрее
  generic fill. Это преимущество существующего кода, не внесённая здесь оптимизация.

CPU/alloc profile холодного WrapWord показывает `wrapWordLines` как главный
источник выделенных байтов: 83,8% flat alloc_space. `SetFontMetrics` — ещё 9,6%.
В CPU profile у `MeasureText` 52,8% cumulative, `GetFontMetrics` — 31,8% cumulative;
эти доли пересекаются и не суммируются. Общие 2,80 GB alloc_space относятся ко
всему профилировочному прогону, не к одной операции. Это делает
[richtext.go](../widget/richtext.go), `wrapWordLines` (строка 1686), первым
конкретным кандидатом для работы с аллокациями.

Отдельный профиль `migrated_fynedo`: EnsureMinSize 100,08 мкс, WalkTrees 22,57 мкс,
300 lookup метрик 13,73 мкс. Выборки небольшие; это ориентир, а не доказательство
универсального процента ускорения. Тег изменяет mutex/map/queue реализации.
Его использование требует соблюдения единой UI-goroutine и `fyne.Do` для
фоновых обновлений, как описывает [документация Fyne](https://docs.fyne.io/started/goroutines/).

`RunOnDraw` один раз остановил общий прогон; процесс был завершён после длительного
ожидания. Изолированные проверки с `1x` и с калибровкой 500 мс × 3 прошли.
Воспроизводимой причины deadlock не установлено. Fixture требует исправления:
`createWindow` уже создаёт окно на main thread, затем benchmark снова вызывает
`w.create()` вне него и не освобождает созданные окна. Отсутствует отдельное
закрепление OS thread для операции с current context. Его числа не включены
в основную baseline; из этого случая не делается вывод о штатном приложении.

Исходные измерения, ranges и source hashes: [library-assessment](library-assessment/).
В частности: [benchmark-summary.json](library-assessment/benchmark-summary.json),
[text-cpu-top.txt](library-assessment/text-cpu-top.txt),
[text-alloc-top.txt](library-assessment/text-alloc-top.txt).

## Приоритетный инженерный backlog

Сложность: **S** — локальная работа; **M** — несколько связанных путей;
**L** — изменение подсистемы; **XL** — платформенная архитектура.
Это оценка объёма и риска, не обещание срока.

| ID / приоритет | Улучшение и основание | Сложность | Критерий готовности |
| --- | --- | --- | --- |
| A1 / до публичного выпуска | Исправить R1/R2/R3 и совместимо обновить TIFF decoder. | S–M | Минимальные воспроизведения проходят, горячие пути не ухудшены, выбран и проверен minimum Go. |
| A2 / P1 | Адаптивный event/frame scheduler: события, wakeup, ближайшая анимация и present deadline. Сейчас `runGL` обслуживает события и кадры таймером 60 Гц. | M–L | Измерены input-to-frame p50/p95/p99, idle CPU, скрытые окна, 60/120/144 Гц; сохранены Wayland frame callbacks. |
| A3 / P1 | Снизить аллокации переноса текста: общий буфер rune/boundaries, устранение отдельных маленьких segment slices, кэш результатов по тексту/стилю/ширине, повторное использование измерений. | M | Снижаются B/op и allocs/op на cold/warm fixtures; одинаковые переносы, truncation, ligatures, Arabic/CJK и смешанные styles. |
| A4 / P1 | Индекс геометрии коллекций. При custom heights List ищет видимую строку от начала; Table аналогично при custom widths. Tree.Layout проходит все открытые узлы. | M–L | Prefix sums/Fenwick или sparse index; flattened open-node view для Tree. Scroll зависит от viewport и изменений, не от позиции в большой коллекции. Проверить 100k/1M элементов и изменение размеров. |
| A5 / P1 | Разделить инвалидацию layout, paint и texture. `EnsureMinSize` ходит по всему видимому дереву; repaint очищает и рисует весь canvas. Дедупликация Refresh уже есть. | L | Изменение одного узла не повторяет весь layout. Damage rendering вводится с retained content/FBO или явным правилом сохранности buffer; blur, shadows, transparency и overlays остаются корректными. |
| A6 / P1 | Ограничения памяти и предсказуемая eviction. В shaped cache на 4096 entries очищается вся map. RGBA atlas 1024/2048 использует только alpha: 4/16 MiB на texture. | M | Byte budget и eviction без массового сброса горячих строк; измерены miss/reset counts и p99. Single-channel atlas — возможные 1/4 MiB, с отдельными GL/GLES format/shader вариантами и color-emoji fallback. |
| A7 / P1 | Асинхронная загрузка/декодирование изображений, отмена устаревших запросов, thumbnail cache и decode budgets. `Image.Refresh` синхронно читает и декодирует источник. | M | UI не ждёт I/O/decode; decode limits, старый результат не заменяет новый, ошибки доступны приложению, смена DPI/resize не декодирует файл заново без причины. |
| A8 / P1 | Контракт renderer: `Init` с результатом, `Capabilities`, ресурсный lifetime/context loss и выбор backend. Сейчас common.Canvas хранит `gl.Painter`, app.NewWithID выбирает GLFW, Init возвращает void, built-in shader errors могут panic. | L | Диагностируемая ошибка и fallback policy; явно определены GL/GLES/window-system версии; тесты отказа и восстановления контекста. |
| A9 / P2 | Уменьшить GL state/native-call overhead. Кэш program/buffer/texture/attrib state по context; большие batches, меньше повторных bind/uniform/setup. | M | Counters до/после, CPU и GPU timings без glFinish в обычном цикле. Сохранены порядок, clipping, внешние shaders и сброс state при смене context. |
| A10 / P2 | Прототип минимального GL binding. gl_core использует 52 Go entry points (включая 4 helpers); package v2.1 содержит 2856 функций и 2848 symbol lookup, около 2,34 MB исходника. | M | Отдельный ограниченный пакет для painter, API manifest и проверяемая генерация. Замер cold build, init time и размера binary; полный совместимый gl остаётся доступен. |
| A11 / P2 | Альтернативный backend без cgo: сначала Windows amd64, затем Linux amd64/arm64. GL loader + GLFW DLL/shared library + callback trampolines. | L–XL | Реальная сборка CGO_ENABLED=0, ABI для float/pointers/callbacks, OS-thread ownership, GC lifetime, shutdown, baseline compatibility и runtime benchmarks. |
| A12 / P1 | Очереди, shutdown и backpressure. `UnboundedChan` имеет дополнительную goroutine и неограниченную очередь; UI updates могут опережать обработку. | M | Coalescing заменяемых updates, сохранение порядка обязательных событий; метрики длины очереди, stress producer/close/DoAndWait, ожидания не зависают, окна и goroutines завершаются. |
| A13 / P1 | Воспроизводимые CI/release/performance проверки. Workspace требует подмодули, checkout workflow их не инициализирует. Нужен отдельный workspace/clean-consumer профиль. | S–M | Recursive checkout, pinned versions, GOWORK=off consumer, minimum Go, GOAMD64=v1 и рабочий профиль, CGO/no-CGO matrix; генерация и benchmark fixtures воспроизводимы. |

Для A2 GLFW уже предоставляет `WaitEventsTimeout` и `PostEmptyEvent`;
это основа для ожидания и wakeup, а не готовая замена Go-select без проекта
взаимодействия очередей и shutdown. Таймер 60 Гц имеет интервал 16,67 мс —
это возможная составляющая ожидания, не измеренная полная latency.
[GLFW 3.4: обработка событий](https://www.glfw.org/docs/3.4/input_guide.html#events).

Для A4 исходники: [list.go](../widget/list.go), `calculateVisibleRowHeights` (498);
[table.go](../widget/table.go), `visibleColumnWidths` (922);
[tree.go](../widget/tree.go), `treeContentRenderer.Layout` (687).
Виртуализация объектов есть, но она не устраняет все линейные проходы геометрии.

Для A9 нужно сравнить существующий BufferData с альтернативами streaming/ring
buffers на конкретных drivers. Текущий код намеренно предпочитает BufferData
BufferSubData на desktop; простая замена вызова не обоснована измерениями.

Для A10 генератор уже поддерживает ограничение списка API через `restrict`:
[Glow](https://github.com/go-gl/glow#custom-packages). GLES binding также широк:
887 функций / 880 lookup. Важный отдельный случай: Fyne просит GLES 2.0 context,
но подключает v3.1/gles2 с обязательными lookup, например `glDispatchCompute`.
Это подтверждённое несовпадение требований в коде; отказ на GLES-2-only hardware
в этой оценке не воспроизводился. Нужны capabilities и загрузка необходимого API.

В cgo-пути можно отдельно оценить `#cgo noescape` для функций с доказанным
временем жизни указателей. `nocallback` нельзя массово назначить GL/GLFW:
возможны error/debug callbacks. Эти директивы доступны с Go 1.24 и требуют
решения о minimum Go либо отдельного совместимого варианта.
[Go 1.24](https://go.dev/doc/go1.24#cgo), [контракты cgo](https://pkg.go.dev/cmd/cgo#hdr-Optimizing_calls_of_C_code).

Для A11 возможен [purego](https://github.com/ebitengine/purego), но следует проверить
его ограничения ABI для выбранных платформ. Он не заменяет библиотеку GLFW и
видеодрайвер. Динамический GLFW оставляет нативную библиотеку при запуске.
Кроме GL/GLFW, остаются платформенные cgo-пути Fyne: native menus, accessibility,
macOS/mobile интеграция и systray. Два форка сами по себе не делают весь framework
независимым от cgo. Обход Go runtime собственным asm без контракта GC/preemption
не является обоснованной первой оптимизацией.

### Зависимость, которую нужно исправить перед выпуском

В текущем go.mod `golang.org/x/image v0.24.0`. Свежий `go list -deps` приложений
подтвердил включение TIFF decoder; `canvas.Image.Refresh` использует `image.Decode`.
Публичный [GO-2026-5032](https://pkg.go.dev/vuln/GO-2026-5032) относится к TIFF
PackBits decoder до v0.41.0 и несоразмерному потреблению ресурсов. Вредоносный
файл не запускался; здесь подтверждены версия и включение decoder.

[go.mod v0.41.0](https://raw.githubusercontent.com/golang/image/v0.41.0/go.mod)
требует Go 1.25, а framework объявляет Go 1.22. Поэтому нужен совместимый backport
либо осознанное повышение minimum Go с clean-consumer проверкой. Одни ограничения
ширины/высоты изображения не заменяют исправление PackBits decoder.

## Функциональные расширения

| Приоритет / возможность | Что добавить | Основание и приёмка |
| --- | --- | --- |
| P1: развитые коллекции | Selection model со stable IDs, Ctrl/Shift/range selection, TreeTable, sorting/filtering adapters, lazy async data provider и transactional updates. | В List/Tree Select заменяет выбранное на один ID; Table выбирает одну cell. Headers/sticky/resize уже есть. Проверить keyboard/mouse, reorder/filter и 100k+ записей; I/O не выполняется внутри UpdateItem. |
| P1: полноценный текстовый ввод | IME composition/preedit/commit/cancel, caret rect для ОС, grapheme-aware cursor/delete/selection, точная bidi navigation. | Есть Unicode char callback и shaping, но нет общего Go-контракта preedit. Entry меняет CursorColumn по rune. Проверка Windows TSF/IMM, Linux IME и мобильных устройств, emoji ZWJ/combining marks/RTL. |
| P1: доступность | Семантическое дерево, роли/значения/состояния/actions/focus/selection; native providers и события изменений. | Сейчас 4 роли, label+role interface; Windows/macOS implementation опциональна через accessibility tag, Linux — no-op. Windows собирает плоские leaf elements. Приёмка — NVDA/VoiceOver/Orca, не PNG. |
| P2: интеграция с ОС | Clipboard MIME/HTML/images/files с plain-text fallback; исходящий и widget-level drag/drop, lazy payload и согласование copy/move. | Clipboard API содержит только строки; Window.SetOnDropped получает URI. Существующие drag/URI drop сохраняются. Проверить обмен с реальными внешними приложениями. |
| P2: SVG/изображения | Определённый SVG compatibility profile, воспроизводимый corpus, mask/use/symbol/styles и отчёт об unsupported features; streamed/thumbnail image API. | Растеризация и SVG уже существуют. Нужны сравнение с независимым renderer, decode limits, cancel, корректные alpha/Stride и bounded caches. |
| P2: темы и наблюдаемость | Locale/theme/font version notifications, независимая инвалидация цветов и метрик, profiler hooks и diagnostics snapshot. | SetLocale уже добавлен в локальной базе. Нужно обновление существующего UI и cache epochs. Hooks: input/queue/layout/paint/upload/present, counters и p95/p99 без постоянных аллокаций. |

Дополнительные готовые composite widgets — sortable TreeTable, grid с редакторами
ячеек, command palette, dock layout и task/progress view — полезно выделить
в расширяемый пакет виджетов после модели selection/data. Это предложения,
не утверждение, что существующие widgets нельзя использовать для этих сценариев.

Vulkan/Metal/D3D, PGO и SIMD/asm — дальнейшие эксперименты с конкретной нагрузкой.
Обязательным основанием должны быть CPU/GPU profile, требование платформы или
контракт качества. Новый GPU API не гарантирует более быстрый обычный интерфейс.

## Матрица совместимости и проверок

| Цель | Сейчас | Что требуется для развития |
| --- | --- | --- |
| Windows amd64 | Свежие CPU benchmarks и реальный GPU SubImage probe; предыдущая проверка обычного GL приложения сохранена в PROJECT.md. | DPI/hotplug/sleep-resume, IME, 60/120/144 Гц, Intel/AMD/NVIDIA; GOAMD64=v1 и v3; cgo/no-cgo и accessibility. |
| Linux amd64, X11/Wayland | Предыдущий CI/race результат есть в TEST_RESULTS.md. Новая SSH preflight на f-test-bench: нет DISPLAY/Wayland session, dev metadata GL/GLES/EGL/GLFW и Xvfb. Native benchmark не запускался. | Отдельный подготовленный графический стенд; event wakeup, frame callbacks, DPI, IME и dynamic loading. |
| Windows/Linux ARM64 и ARM | Выбор GLES частично обусловлен GOARCH; GLES2 hints сочетаются с binding 3.1. | Проверять доступный API/capabilities, EGL/ANGLE и ABI; CPU architecture не должна быть единственным выбором graphics API. |
| macOS amd64/arm64 | В этой оценке не проверен. Имеются native menus/accessibility и другие Objective-C/cgo пути. | Main-thread contracts, IME, VoiceOver, context lifecycle, сборка и runtime отдельного backend. |
| Web/Wasm, Android/iOS | Отдельные gl-js/mobile paths; desktop форки их автоматически не покрывают. | Самостоятельные compatibility/performance profiles и реальные IME/resize/lifecycle испытания. |
| Потребитель Go-модуля | Local workspace использует форки; GOWORK=off выбирает опубликованные зависимости из go.mod. | Адреса собственных modules, доступные версии и CI чистого потребителя; отдельно закреплённый generator/tool workspace. |

## Рекомендуемый порядок внедрения

1. **Корректность и воспроизводимость:** R1–R3, исправление TIFF с выбранным
   minimum Go, надёжные benchmarks и tests recovery/shutdown; источники и
   performance baseline закрепить перед сравнением.
2. **Отзывчивость:** event/frame scheduler, wakeup и queue coalescing с явными
   thread/lifecycle контрактами. Измерять latency и idle CPU вместе.
3. **CPU и память:** WrapWord allocations, geometry indexes коллекций,
   selective invalidation, image workers и bounded atlas/text caches.
4. **Графический слой:** capabilities/Init errors, state counters/cache,
   минимальные bindings, сравнение cgo/FFI. Прототип CGO_ENABLED=0 начать с
   Windows amd64; затем переносить на Linux с общей conformance matrix.
5. **Функциональность:** selection/data model и TreeTable, IME/accessibility,
   расширенный clipboard/drag-drop, SVG/image capabilities и composite widgets.

Для каждого изменения нужны одинаковая fixture, build profile, corpus и
hardware; отдельно cold/warm, CPU/GPU time, allocations, peak memory, idle CPU
и latency percentiles. Ускорение засчитывается вместе с сохранением порядка
рисования, геометрии, Unicode поведения и API compatibility.

## Материалы и воспроизведение

- [Среда Windows](library-assessment/environment-windows.txt),
  [Linux preflight](library-assessment/linux-preflight.txt),
  [source hashes](library-assessment/source-hashes.json).
- [Основной benchmark log](library-assessment/benchmarks-windows.txt),
  [migrated profile](library-assessment/benchmarks-migrated-windows.txt),
  [medians/ranges](library-assessment/benchmark-summary.json).
- [Минимальные тесты](library-assessment/reproduction-tests.txt) и три probe logs.
  Эти временные Go-файлы удалены после запуска; ошибки тестов являются результатом
  воспроизведения, а не оставленными изменениями test suite.
- [Первый exploratory log](library-assessment/benchmarks-exploratory-windows.txt),
  [изолированный RunOnDraw](library-assessment/run-on-draw-calibrated.txt).
- [Профилировочный прогон](library-assessment/text-profile-run.txt), CPU/alloc top.
  Временные test executables и бинарные profiles удалены; текстовые результаты сохранены.

Пример повторения основного CPU-набора из корня проекта:

```powershell
go test -p 1 -timeout=120s -run '^$' -bench '^Benchmark(FrameEnsureMinSize|FrameWalkTrees|RunOnMain|GetFontMetrics|Text_splitLines|Text_lineBounds_WrapWord|Text_howManyRunesFit_latin|Text_howManyRunesFit_chinese|DrawRectangle_Opaque_1280x800|DrawRectangle_Generic_1280x800|ContentMinSize|TextGrid_Refresh)$' -benchmem -benchtime=500ms -count=3 ./internal/driver/glfw ./internal/cache ./widget ./internal/painter/software
```

Для непосредственного повторения probes использовать отдельную рабочую копию
и файлы соответствующих packages из reproduction-tests.txt. В текущем запросе
оценка завершена; исправления, смена backend и новые API ещё не реализованы.
