# Оценка GL / GLFW / Glow

Дата: 12 октября 2026 года. Область: `deps/gl`, `deps/glfw`, `deps/glow`, включая открытые issues именно этих репозиториев. Fyne здесь выступает потребителем API и интеграционным тестом. Дефекты его layout/cache не отнесены к дефектам GL/GLFW.

**Реализация после оценки:** подтверждённые defects strings/GUID/lifecycle/GLES/packaging/generator/EGL исправлены, добавлены D-Bus 5.3.0 и GLFW 3.5.1. Итоги, benchmarks и точные ограничения: [NATIVE_STACK_IMPLEMENTATION.md](NATIVE_STACK_IMPLEMENTATION.md). Далее сохранён исходный анализ и roadmap.

## Вывод и порядок работы

Форк `gl` имеет смысл поддерживать вместе с `glow`: исправления сгенерированных файлов должны воспроизводиться из шаблонов и XML overrides. Форк `glfw` полезен для исправления жизненного цикла, ошибок ввода, IME и будущей сборки без cgo. Первые изменения стоит направить на корректность указателей/строк и профилей, затем на стоимость обёрток и загрузки символов. Переписывание всех трёх библиотек на ассемблер не оправдано имеющимися измерениями.

| Очередь | Изменение | Причина и критерий готовности |
| --- | --- | --- |
| 1 | GUID форматирование в C GLFW 3.3/3.4 | Подтверждён опасный signed-char фрагмент; два UTF-8 байта дают 45 байт при ёмкости 33. Маскировать байты и ограничить запись; проверить ASCII/UTF-8/длинные имена, ASan и стабильность GUID. |
| 1 | Контракт `Strs`, типизированные uploads, offset overloads | `Strs` не добавляет NUL, документация обещает обратное. Исключить чтение за буфером и неверную ширину индексов; пройти checkptr/cgocheck2 и shader/PBO/VBO cases. |
| 1 | Очистка Go-реестра окон в `Terminate` | После уничтожения нативного окна остаётся запись с Go-объектом и callbacks. Проверены скрытое окно и оставшаяся запись. Нужны reset/reinit tests, включая окна без явного `Destroy`. |
| 1 | Соответствие GLES-контекста и bindings | В интеграции запрашивается GLES 2.0, импортируется strict GLES 3.1. Настоящий GLES 2.0 успешно генерируется; выбрать минимальный API и явное повышение версии при необходимости функций 3.x. |
| 1 | Packaging и воспроизводимость генерации | В `v3.0/gles2/KHR` отсутствует `dummy.go`; свежий vendor consumer теряет header. На проверенной Windows сборка всё же проходит за счёт внешних headers. GL 2.1/GLES 3.1 повторная генерация совпала после одинаковых LF/gofmt. |
| 2 | Меньший профиль и более дешёвый loader | Fyne использует 48 native entrypoints, GL 2.1 загружает 2848 символов. Проверить `restrict`, required/optional functions и выигрыш в cold build/startup/binary size. |
| 2 | Выборочные `noescape`, отдельный callback-safe режим | Доступны при Go 1.27.1. Разметка требует знания retained pointers и debug/error callbacks; выигрыш подтвердить сравнением с обычным cgo. |
| 2 | Ошибки, threading, context-specific dispatch | Ненулевой pointer не подтверждает поддержку функции. Контекст/версия/extension/error должны быть явными; без locks/checks на каждом release-вызове. |
| 3 | Второй backend без cgo, EGL/headless, IME | Полезные расширения, требующие отдельной ABI/platform матрицы. Сохранить рабочий cgo backend для сравнения и совместимости. |

На этапе этой оценки C snapshot сохранялся. По последующему запросу подготовлен отдельный GLFW 3.5.1 модуль и подключён к Fyne; модули 3.3/3.4 сохранены.

## Проверенная база

| Репозиторий | Upstream commit | Рабочий API |
| --- | --- | --- |
| go-gl/gl | `4566fea9a2762279aa48096fa3a0339db7406044` | GL 2.1 и GLES 3.1; остальные профили сохранены |
| go-gl/glfw | `d41da22a9587f777098f96d37014f6cdd35d1afb` | Вложенные Go-модули 3.3 и 3.4; Fyne использует 3.4 |
| go-gl/glow | `394d535962b324f7dd84b9244afbd952a4fe717e` | Khronos XML, templates, overrides |

Удалённые `master` этих трёх репозиториев проверены: новых commits относительно локальной базы не было. Нативный snapshot GLFW 3.4: `7b6aead9fb88b3623e3b3725ebb42670cbe4c579`. Go GLFW 3.4 имеет tag `v0.1.0-pre.2`; версия Go-wrapper и версия C-библиотеки — разные величины.

Рабочий Windows smoke: Go 1.27.1, GCC/MSYS2, NVIDIA RTX 4080 SUPER, OpenGL 4.6.0 / driver 616.92. Создано скрытое окно, выполнены Clear/ReadPixels, прочитан `[0, 255, 0, 255]`; синхронный debug callback действительно вызван. Источник проверки: [graphics-probe/main.go](../scripts/graphics-probe/main.go), результаты: [windows-gl-probe.log](graphics-assessment/windows-gl-probe.log).

## Что исправлять в GL

### Строки и память

`conversions.go:90` выделяет сумму длин строк, копирует их подряд и не добавляет терминаторы. Безопасная проверка `Strs("A", "B")` прочитала ровно два выделенных байта `AB`, без чтения за границей. Это подтверждает [gl#77](https://github.com/go-gl/gl/issues/77). При `ShaderSource(..., lengths=nil)` такой результат не соответствует требованию NUL. Fyne уже добавляет `\x00` в `internal/painter/gl/gl_core.go:275`, поэтому данный smoke не означает краша Fyne.

Нужны явно разделённые операции: null-terminated strings для имён/исходников без lengths и строки с точными lengths для случаев с embedded NUL. При исправлении проверить пустые строки, несколько строк, уже добавленный NUL, overflow общей длины и вызов `free`. Ошибочный контракт надо исправить в `glow/tmpl/conversions.tmpl` и перегенерировать профили.

`Ptr(interface{})` принимает `uint`/`int` платформенной ширины, использует reflection и паникует на пустом slice. В проверенном паттерне `Ptr([]float32)` даёт одну Go-аллокацию на вызов. Предпочтительны добавочные typed helpers/overloads с `[]uint16`, `[]uint32`, `[]float32`, вычислением byte length и обработкой пустого slice; данные из Go не должны сохраняться нативной функцией без отдельного lifetime-контракта. `unsafe.SliceData`/`unsafe.StringData` позволяют убрать устаревшую механическую работу с headers, но не отменяют требования GC/cgo.

Offset и memory pointer нельзя смешивать. `PtrOffset` уже deprecated, `VertexAttribPointerWithOffset` существует. Остались пробелы вроде DrawElementsInstanced и ReadPixels при PBO, о которых сообщает [glow#127](https://github.com/go-gl/glow/issues/127). Использовать integer offset до C-преобразования; не предлагать отключать `cgocheck` как решение.

### Загрузчик, capabilities и несколько контекстов

`InitWithProcAddrFunc` заполняет глобальные pointers пакета. Strict init требует core-функции выбранного профиля; `all-core` с lenient init может оставить nil. Нужна диагностика, содержащая backend, фактический current context, GL/GLES version и missing symbol. Наличие адреса функции не доказывает поддержку extension — она определяется версией и extension set.

Для OSMesa нужен `OSMesaGetProcAddress`/соответствующий resolver через `InitWithProcAddrFunc`; POSIX default сейчас GLX. Для EGL выбирать EGL loader, а не рассчитывать на GLX/WGL случайно. Это объясняет направление проверки [gl#167](https://github.com/go-gl/gl/issues/167), но контейнеры автора не воспроизводились.

Предлагается объект `Dispatch/Context` с immutable capability set и отдельным callback registration. Старые package functions можно сохранить как совместимый facade. На Windows адреса extensions зависят от pixel format/контекста, что описано в [документации Microsoft](https://learn.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-wglgetprocaddress). Глобальная таблица и один `userDebugCallback` ограничивают независимую работу разных контекстов/GPU. Проверять два окна, shared/unshared contexts, teardown/reinit и callback routing. OS thread закрепляет владелец контекста; release-вызов не должен добавлять отдельный syscall для проверки thread id.

## Где есть потенциал оптимизации

| Измерение | Результат на текущей Windows-машине | Интерпретация |
| --- | --- | --- |
| Init через GLFW resolver | 2848 lookups, около 8.4 ms в исходном прогоне | Время загрузки, а не кадра; повторная инициализация с уже созданным контекстом |
| ClearColor | Около 40 ns/call, 0 Go allocations | Этот driver/state call уже дешёвый; не переносить результат на uploads/draw/swap |
| GetError | Около 570–620 ns/call, 0 Go allocations | Стоимость включает драйвер; не проверять после каждой операции в release без необходимости |
| Ptr(slice) | 1 allocation/call; время заметно меняется между прогонами | Надёжнее сначала устранить allocation, затем измерить стабильный throughput |
| GL 2.1 generated package | 2,342,812 bytes; 2856 Go functions; 2848 lookups | Обёртки и lookup table заметно шире потребности Fyne |
| GLES 3.1 generated package | 846,138 bytes; 887 Go functions; 880 lookups | GLES 3.1 strict loader не является GLES 2.0 loader |
| Fyne GL facade | 52 именных вызова, включая 4 helpers; 48 native entrypoints | Основа для маленького manifest профиля; иные потребители требуют своих manifests |

Измерения — короткие CPU samples на одной NVIDIA/Windows конфигурации, без изоляции всей фоновой нагрузки. Они не измеряют GPU frame time и не доказывают улучшение FPS. Первоначальные samples сохранены в [baseline JSON](graphics-assessment/windows-gl-probe-baseline.json), повторный прогон — в журнале; будущие изменения сравнивать с одинаковым драйвером, context hints, power settings и benchmark body.

`glow -restrict` уже умеет ограничивать функции/enums; `-remext`/`-addext` ограничивают extensions. Добавить проверяемый manifest минимального Fyne API и extension packs для отдельных потребителей. Сначала измерить cold compile, peak RAM, Go/C object size, binary size и Init; не обещать одинаковое ускорение для всех приложений.

Следующий эксперимент — группировать C bridges с одинаковой ABI signature и разделить generated C на файлы, которые linker может удалить. Одинаковые Go-параметры ещё не гарантируют одинаковую C ABI: нужны pointer qualifiers, calling convention, width, floating registers, return types и arch tests. Приоритет [glow#23](https://github.com/go-gl/glow/issues/23) выше ручной оптимизации каждого из тысяч wrappers.

Для `#cgo noescape` нужен whitelist функций, не сохраняющих Go pointer и не возвращающих его в Go callback. Для `nocallback` требуется доказательство отсутствия перехода C→Go. GL synchronous debug output может сделать callback из обычного вызова драйвера, а GLFW вызывает Go error callback даже из некоторых getters/setters. Нужен отдельный backend/build variant или C error/debug buffering с последующим Go drain. Ошибочная разметка приводит к panic/повреждению памяти по [контракту cgo](https://pkg.go.dev/cmd/cgo#hdr-Optimizing_calls_of_C_code). Go 1.27.1 даёт механизм, но не гарантирует безопасность разметки.

## Что дорабатывать в GLFW

### Подтверждённые проблемы

1. **C GUID buffer.** В Linux joystick backend обеих веток 3.3/3.4 используется `sprintf(guid[33], "%02x...", name[i])` с `char`. Safe standalone test на `f-test-bench` с `-fsigned-char` дал `original_length=44`, `bytes_with_nul=45`, `masked_length=32`. Проверка использовала буфер 128 и не выполняла переполнение. Исходный источник — [glfw#426](https://github.com/go-gl/glfw/issues/426). Это подтверждение форматирующего фрагмента и применимости исходного кода, а не проверка эксплуатации или реального EVIOCGNAME устройства. Исправление должно быть сохранено в процессе обновления C snapshot.
2. **Registry lifetime.** `Terminate()` вызывает C cleanup, но не очищает `windows.m`. Временный native race-тест создал одно скрытое окно, вызвал Terminate и обнаружил одну оставшуюся Go-запись. Файл теста удалён; [журнал](graphics-assessment/glfw-terminate-probe.log) и рецепт ниже сохранены. При явном `Destroy` запись удаляется, поэтому это отдельный сценарий библиотеки, не доказательство утечки при каждом закрытии Fyne.
3. **Условия BSD/Wayland.** `native_linbsd_wayland.go` допускает OpenBSD с `wayland`, а `build.go` выбирает `_GLFW_X11` для OpenBSD. Это подтверждённое несоответствие conditions, связанное с [glfw#425](https://github.com/go-gl/glfw/issues/425). FreeBSD и фактические BSD link/runtime отдельно не проверялись.

Для повторения registry-проверки поместить временный тест в package `glfw`: LockOSThread → Init → WindowHint(Visible,False) → CreateWindow → Terminate → проверить `len(windows.m)` под mutex. После нативного Terminate не вызывать `w.Destroy()` на освобождённом handle. Тестовый Go-entry удалить отдельно. Production fix должен очищать реестры и учитывать повторную инициализацию.

### Ошибки и ввод

Глобальный `lastError` channel имеет capacity 1, тогда как нативный GLFW error state относится к thread. Очередь безопасна как Go channel, но это не доказывает правильную принадлежность ошибок вызовам/потокам. Добавить тесты параллельных допустимых queries и последовательных ошибок, отделить programmer errors от неизвестных scancodes, gamepad mapping и environment failures. `GetKeyName` для внешнего fn/media scancode не должен обрушать приложение через panic без документированной альтернативы.

Расширения с практической ценностью: IME composition/preedit/candidate rectangle, стабильные DPI/framebuffer/content-scale transitions, корректный focus/lost-input cleanup, raw mouse/gamepad fixtures, lock-state query. Часть возможностей уже есть: `WaitEventsTimeout`, `PostEmptyEvents`, `SwapInterval`, runtime platform selection, `ContextNoError` в 3.4. Их не надо реализовывать повторно. Частоту обработки событий и batching определяет также приложение/Fyne; её нельзя исправить одним форком wrappers.

IME затрагивает нативные Windows TSF/IMM, macOS text input и Wayland protocols/X11 input methods. Нужны composition lifecycle, Unicode caret/selection и координаты кандидатов. CharCallback передаёт завершённый текст, но не заменяет эти состояния. [glfw#370](https://github.com/go-gl/glfw/issues/370) связан с upstream C PR; переносить такой patch следует с отдельными platform tests и процедурой rebase.

## Что улучшать в Glow

Генератор — место для правил memory lifetime, typed length/offset helpers, backend templates, profiles и diagnostics. Текущая структура уже разделяет спецификации, Package IR и templates, но часть семантики приходится добавлять через overrides: XML не объясняет все случаи retained pointers и integer-as-pointer offsets.

Нужны fixture tests для каждой новой override, determinism/manifest с SHA XML и templates, обязательная compile-проверка generated output и fresh consumer/vendor build. `GeneratePackage` должен сохранять stderr gofmt, проверять ошибки шаблонов и оставлять диагностируемые частичные результаты. `importPathToDir` следует проверять `len(pkgs)` и `packages.Errors`; сейчас код обращается к `pkgs[0]` и анализирует только GoFiles.

**EGL 1.5:** генерация на Go 1.27.1 завершается успешно, но последующая Linux-компиляция с SDK не проходит: Cgo не определяет `C.AHardwareBuffer`, `C.EGLClientPixmapHI`, `C.wl_display`, `C.wl_resource`. Также `DEFAULT_DISPLAY` остаётся выражением `EGL_CAST(...)` в Go source. [glow#16](https://github.com/go-gl/glow/issues/16) остаётся актуален; successful gofmt недостаточен. Нужно сохранить `struct` prefix, разбирать native platform types и приводить macro constants через Go/C-safe overrides. Журнал: [egl-compile.log](graphics-assessment/egl-compile.log).

Полезен отдельный debug/trace variant: имена enums по domain, counters по функциям, upload bytes, source labels и timings без release overhead. Не заменять все constants одним `GLConst int32`: `TIMEOUT_IGNORED` — uint64, многие аргументы смешивают enum и обычные integers. Безопасный слой дополняет raw API и сохраняет путь для PBO/client arrays/persistent mappings.

## Расширение без cgo и переход C GLFW

Путь без cgo стоит делать вторым generated backend. Начать с Windows amd64 и минимального GL manifest, затем Linux/macOS amd64/arm64. [purego](https://github.com/ebitengine/purego) позволяет вызывать native shared libraries, но сам проект отмечает beta status; его поддержку ABI/callbacks надо проверять для используемых signatures. Стандартного syscall вызова недостаточно для произвольных floating-point GL signatures. Сравнить static typed stubs, выбранный FFI adapter и cgo на одном workload.

Для GLFW без cgo потребуется поставка DLL/shared library, её загрузка/поиск symbols, callback ABI и управление lifetime. Отсутствие GCC при сборке не отменяет native runtime и видеодрайвер. Не выдавать возможность собрать маленький GL test за готовность всего desktop framework.

Нативная [GLFW 3.5.1](https://github.com/glfw/glfw/releases/tag/3.5.1) доступна. Перед переходом сопоставить public header с Go wrappers, regenerated Wayland protocol sources, изменённые platform flags, retained local patches, hints/behaviour и SDK minimums. Лучше сохранить 3.4 API и добавить отдельный модуль/profile 3.5, чем молча менять значение существующего пути. Матрица: Windows amd64/arm64, Linux X11/Wayland amd64/arm64, macOS arm64/amd64; resize/focus/clipboard/IME/gamepads, multi-window/context, sleep/resume, ASan и race.

## Issues и обновление Go

GitHub API snapshot содержит **48 открытых issues: GL 20, GLFW 19, Glow 9**, а также все **227 комментариев**. Полный разбор каждого issue с приоритетом, состоянием проверки и следующим действием: [INVENTORY.md](graphics-issue-audit/INVENTORY.md). Полные тела: [open-issues.json](graphics-issue-audit/open-issues.json). PR и закрытые issues не включены; C GLFW имеет отдельный upstream tracker.

Не все старые жалобы актуальны: reflect.SliceHeader уже отсутствует в проверенном файле; debug callback в проверенном сценарии работает; ContextNoError есть в GLFW 3.4; некоторые offset overloads уже реализованы; NetBSD tags уже появились. Для crashes AMD/macOS/gamepad/BSD нет достаточного runtime evidence этой проверки. Открытый issue не равен подтверждённому дефекту текущего форка.

Go 1.27.1 установлен в `go.work` и всех девяти `go.mod`, включая два CLI fixtures. Обновлены Go-зависимости основного модуля, apps и Glow, их полный транзитивный/инструментальный граф и GitHub Actions. C GLFW сохранена по выбору пользователя. Из-за Windows compile regression v5.2.3 и upstream master оставлена **godbus/dbus v5.2.2**. Старые пути major-версий не мигрировались автоматически. Изменения модулей сохраняются в воспроизводимых patches; детали и ограничения проверки — [GO_DEPENDENCY_UPDATE.md](GO_DEPENDENCY_UPDATE.md).

## План следующих итераций и приёмка

1. **Корректность:** strings/offsets/GUID/Terminate/packaging; минимальные regression fixtures, zero-length/retained-pointer cases, fresh module/vendor consumers, race/cgocheck2/ASan. Изменения GL выводить из Glow, нативные patches применять воспроизводимо.
2. **Производительность:** минимальный профиль; cold/warm build, binary size, init lookups/allocations, upload throughput; selective noescape; frame-time p50/p95/p99, input-to-render latency, idle CPU. Comparator C обязателен для жалобы на frame pacing; swap и GPU stalls учитывать отдельно.
3. **Функции:** capabilities/context object, callback routing, safe input/error queries, EGL/headless и IME; затем второй no-cgo backend и C GLFW 3.5.1 как отдельные изменения. Не объединять смену ABI, renderer и window backend в один неразличимый эксперимент.

Подтверждённые runtime исправления реализованы в последующем изменении; архитектурные предложения остаются roadmap. Итоговый статус приведён в NATIVE_STACK_IMPLEMENTATION.md. Оценки срока/процента ускорения без прототипов не даны.
