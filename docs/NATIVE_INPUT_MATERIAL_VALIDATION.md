# IME, логическая доступность и материалы Fyne

13 октября 2026 года. Рабочий форк `X:\fyne`, Go 1.27.1, Windows build 26300
и Linux `f-test-bench`. Исходное визуальное ТЗ использовано как предложение;
кроссплатформенные решения сверены с [переработанной архитектурой](VISUAL_MODERNIZATION_REVIEW.md).
Этот протокол фиксирует выполненные сценарии и оставшиеся ограничения. Приёмка
всего ТЗ и всех вариантов IME/screen-reader здесь не объявляется.

## Доступность за пределами viewport

`AccessibleVirtualCollection` описывает полную логическую коллекцию, включая
невидимые строки и потомков закрытых ветвей. List, Table и Tree реализуют этот
контракт. Table использует устойчивые IDs `row:column`; Tree сохраняет родителя
и индекс среди его детей, включая пользовательский непустой Root.

Чтение числа детей, metadata, выделения или соседнего элемента не прокручивает
коллекцию и не раскрывает ветви. Реализация UIA `Realize`/`ScrollIntoView` выполняет
явное действие. Для имени невидимого элемента запрашивается один временный
шаблон; миллион элементов не создаётся при чтении Count. Renderer этого шаблона
освобождается после чтения. UI callbacks источника данных должны оставаться быстрыми.

Windows получил ItemContainer, VirtualizedItem и ScrollItem, логическую fragment
navigation, невидимые Grid cells и Selection.GetSelection с выделением вне
viewport. `IsOffscreen` сообщается также для обычных прокрученных контролов.
Реальный UIA-клиент выявил потерю соседних контролов после виртуальной коллекции:
collection navigation ошибочно перехватывала Next/PreviousSibling. Контейнер
теперь использует логическую navigation только для своих детей; переходы к его
соседям/родителю сохраняют обычное дерево. COM regression проверяет контролы
до и после List вместе с последней offscreen строкой и выделением.
ItemContainer сейчас поддерживает итерацию `propertyId=0`; фильтры Name,
AutomationId и IsSelected ещё не реализованы. Основание:
[UIA virtualization](https://learn.microsoft.com/en-us/windows/win32/winauto/uiauto-implementingvirtualizeditem),
[ItemContainer](https://learn.microsoft.com/en-us/windows/win32/winauto/uiauto-implementingitemcontainer).

Linux получил логические ChildCount/GetChildAtIndex/GetIndexInParent, полное
выделение, невидимые Table.GetAccessibleAt и `org.a11y.atspi.Cache` с протокольной
структурой после 2015 года. Cache содержит видимые и запрошенные элементы;
закрытые ветви раскрываются в логическом дереве по запросу клиента. GetChildren
для коллекции больше 4096 возвращает отказ от bulk enumeration: клиент должен
использовать ChildCount/GetChildAtIndex. Это предусмотрено
[Accessible protocol](https://gnome.pages.gitlab.gnome.org/at-spi2-core/devel-docs/doc-org.a11y.atspi.Accessible.html).
Формат Cache сверялся с
[AT-SPI Cache](https://gnome.pages.gitlab.gnome.org/at-spi2-core/devel-docs/doc-org.a11y.atspi.Cache.html).

Исправлены события активации окна, пустое text selection и TextChanged:
insert/delete теперь описывают только изменённый фрагмент в rune offsets.
Эти исправления проверены реальным Orca, а не только внутренним snapshot.

Нативные UIA и AT-SPI клиенты проверяют List из 100000 строк: строку 99999,
выделенную строку 90000, невидимость, логический индекс и устойчивость ранее
запрошенного provider после прокрутки. На Linux этот контракт прошёл и X11,
и Wayland. Cache в этом сценарии содержит меньше 200 элементов.

Границы: спрос клиента определяет объём retained providers; отдельный абсолютный
лимит памяти этого cache пока не введён. Имена cached offscreen providers могут
оставаться прежними после изменения данных до повторного логического запроса.
Невидимая строка представлена логическим элементом; произвольные вложенные
interactive controls шаблона не разворачиваются в отдельное виртуальное поддерево.
Полные Table/TableItem headers, relations, font attributes и сценарии всех
экранных читателей остаются следующей работой. Это полная адресация логических
List/Table/Tree, а не заявление о завершении всех accessibility patterns.
Обнаружена также практическая граница: Windows UIA FindFirst по всем descendants
окна с List из 100000 строк может завершиться timeout. Для проверки соседних
контролов использован bounded TreeWalker, который не раскрывает List/Tree/Table.
Ускорение массовых predicate queries и offscreen cache refresh остаётся задачей;
обычный полный обход большого виртуального дерева не объявляется быстрым.

## Реальные IME и экранные читатели

| Сценарий | Результат и доказательство |
| --- | --- |
| Linux IBus + штатный Anthy 1.5.17, X11 | `日本語`: отдельный inline preedit, commit, один undo и Escape cancellation; [log](native-input-validation/real-anthy.txt), [events](native-input-validation/ime-anthy.jsonl) |
| Linux IBus password/lifetime | В реальном daemon: surrounding text пустой, offsets нулевые, purpose password/private; закрытый context и window reference освобождены; [race test](native-input-validation/ibus-privacy-lifetime.txt) |
| Linux Fcitx Pinyin 4.2.9.9, X11 | Реальные commits `你好` и `中国`; [log](native-input-validation/real-pinyin-orca.txt), [events](native-input-validation/ime-pinyin.jsonl) |
| Orca 43.1 | Реальный процесс формирует речь `CJK input entry`; независимый libatspi client проверяет password redaction и offscreen selection; [speech](native-input-validation/orca-speech.txt), [client](native-input-validation/reader-client.json) |
| NVDA 2026.2, Windows | Реальный portable NVDA получает окно Fyne, обычный edit и protected edit с непустым password; внешний UIA-клиент проверяет redaction; [filtered speech](native-input-validation/nvda-speech.txt), [client](native-input-validation/windows-reader-client.json). Использован silence synth: последовательности речи подтверждены, акустическая/брайль-приёмка не выполнялась |
| Microsoft Japanese IME, Windows | **Не принят.** Активация stage 4 завершилась `0x80004005`; другой пробный запуск с успешной активацией ввёл ASCII вместо `日本語`. [failed test](native-input-validation/windows-real-ime-failed.txt), [environment](native-input-validation/windows-ime-environment.txt) |

Сценарии IME запускают обычное Fyne приложение и штатные процессы input method.
Клавиши поступают через XTest/SendInput. В успешном Anthy сценарии engine не
заменялся тестовым обработчиком. Engine собран из
[официального ibus-anthy](https://github.com/ibus/ibus-anthy).
Screen-reader логи от других окон Windows не сохранялись в evidence.
Внешний password fixture содержит непустое значение: AT-SPI проверяет отсутствие
текста и character count; UIA — protected state и отсутствие Value/Text contents.
Воспроизводимый Windows harness:
[validate-native-reader-windows.ps1](../scripts/validate-native-reader-windows.ps1)
с `-WorkDirectory`, содержащим race-built `input-probe.exe`, standalone portable
NVDA и собственный `nvda-profile` с silence synth. Существующий NVDA не завершается:
для его присутствия harness требует отдельную сессию.

На стенде IBus не зарегистрировал XIM server. Поэтому добавлен прямой локальный
IBus клиент для X11: ProcessKeyEvent, CommitText, preedit, surrounding text,
DeleteSurroundingText, caret, focus/reset и завершение context. Нет изменения
пользовательского input-method профиля. Поздние сигналы старого focus отбрасываются
по поколению focus и D-Bus response sequence. Password текст не передаётся engine.
Контракт сверялся с
[IBus InputContext](https://ibus.github.io/docs/ibus-1.5/IBusInputContext.html).

На Windows TSF key filtering перенесён перед TranslateMessage/DispatchMessage,
чтобы обработанная engine клавиша не порождала второй WM_CHAR. Исправлена также
headless-сборка: scalar VARIANT helper не зависит от VariantInit/OleAut32,
которую ранее косвенно подключал accessibility build.

Windows имеет файлы/регистрацию японского IME, но состояние Basic Typing не
подтверждено: текущая Windows-сессия не elevated, а `Get-WindowsCapability -Online`
возвращает требование повышения. Нельзя выводить отсутствие language capability
из одного отказа активации. Языковой список ru/en и исходный enable flag профиля
после временных диагностических проб восстановлены. Следующий шаг — запросить
состояние `Language.Basic~~~ja-JP~0.0.1.0` из elevated сессии, при необходимости
подготовить Basic Typing и повторить `TestRealMicrosoftJapaneseIME`.
[Microsoft IME troubleshooting](https://learn.microsoft.com/en-us/troubleshoot/windows-client/shell-experience/troubleshoot-ime-common-issues).
Для read-only проверки подготовлен
[windows-ime-preflight.ps1](../scripts/windows-ime-preflight.ps1): запускать
из PowerShell с повышением. Он сохраняет только состояния CJK Basic Typing и
языковые профили; компоненты не устанавливает, настройки не изменяет.

Пока не приняты: Windows реальные CJK/reconversion; Linux Korean и другие
engines; Wayland CJK text-input/input-method связка; attributes всех preedit
segments; IBus ForwardKeyEvent. Прямой IBus transport также требует последующей
проверки sender identity и поведения при переполнении очереди сигналов.
Синхронный ProcessKeyEvent ограничен 100 ms; это предел ожидания ошибки,
а не доказательство задержки ввода в 100 ms или аппаратной частоты кадров.

## Внешний Wayland drag

[Harness](../scripts/validate-wayland-drag-linux.sh) запускает отдельный
`wl_data_device` source и настоящее Wayland окно Fyne в Weston 10.0.1. Gesture
поступает в seat через тестовый модуль libweston; enter/button serial, hit testing,
grab, offer, pipe transfer и finish формирует compositor. Source не вызывает
приватные Fyne callbacks. Это нативный протокольный сценарий, а не физическая мышь.

| Offer | Проверка |
| --- | --- |
| text/plain | Unicode `Привет 日本 😀`, 24 bytes |
| text/html | HTML с CJK/emoji, 18 bytes |
| text/uri-list | URI с percent-encoded пробелом, 32 bytes |
| application/x-fyne-drag-test | 524288 bytes с нулевыми и non-UTF8 bytes |
| application/x-not-accepted | Cancelled, ни одного дополнительного drop callback |

Все положительные передачи сравниваются побайтно; source должен получить
`dnd_finished`. Source отказывается передавать MIME, которого не объявлял.
[Результат](native-input-validation/wayland-drag.txt),
[source/gesture evidence](native-input-validation/wayland-protocol-results.json).

Исправлены согласование версии data-device v3, COPY actions/finish, сохранение
непринятого offer до Leave и выбор только объявленного text/plain/UTF8 alias.
Основой служат pinned Wayland/xdg-shell protocol definitions в локальном GLFW.
X11 external Xdnd и Windows external OLE также входят в полные native suites.

Следующая приёмка: GTK/Qt/file-manager источники, MOVE/link, несколько seats,
outgoing drag и виртуальные файлы. Получение native pipe пока выполняется
синхронно с лимитом 64 MiB и deadline 3 s; асинхронный receive не добавлялся.

## Материалы подключены к canvas и popup

Новый `canvas.Material` фильтрует уже нарисованные нижние слои этого canvas:
GPU snapshot, blur, corners, tint/noise, downsample и linear policy. Он не
захватывает чужие окна/desktop и не добавляет периодическую animation.
У `widget.PopUp` появился optional `Material`. По умолчанию существующий popup
сохраняет прежнее оформление.

```go
popup := widget.NewPopUp(content, window.Canvas())
popup.Material = canvas.NewMaterial(theme.Color(theme.ColorNameOverlayBackground))
popup.ShowAtPosition(fyne.NewPos(30, 20))
```

Материал можно разместить также между другими объектами canvas. Объекты поверх
него сохраняют обычное hit testing. Unsupported backend, invalid parameters,
high contrast, запрещённая transparency, третий material layer или превышение
бюджета выбирают opaque fallback. Software/Web/mobile профиль использует fallback;
браузерный gallery с этим API собран, native GPU material на Web не заявляется.

Один painter владеет двумя слоями, 48 MiB pool и резервом 16 MiB для GPU capture,
включая старую и новую capture texture при resize. Общий лимит — 64 MiB
**запрошенного RGBA texel storage**, не полный GPU allocation приложения.
Capture растёт до максимальных нужных размеров и переиспользуется. Cache key
учитывает содержимое, поверхность, physical geometry, обе оси scale и стиль.
Статичная сцена использует cache; ресурсы освобождаются с matching context до
Destroy. `Canvas.Capture()` теперь применяет dirty texture invalidation перед
новой отрисовкой, включая нижние слои материала.

Полноэкранная 5K capture превышает резерв 16 MiB и выбирает fallback; небольшой
popup на 5K может использовать GPU pipeline. Автоматический downsample большого
framebuffer capture, освобождение hidden layers в конце кадра, HDR/float targets
ещё не реализованы. Native window backdrop state machine добавлен позднее:
см. [WINDOW_VISUALS_VALIDATION.md](WINDOW_VISUALS_VALIDATION.md), включая
самостоятельные границы приёмки WGL/DWM.

Нативный интеграционный тест сравнивает пиксели gradient/snapshot, orientation,
cache hit, бюджет, opaque fallback, Show/Capture/Hide реального popup. Он прошёл
Windows GL, Linux Mesa GL/X11, strict GLES 2.0 и Wayland. Это интеграция Fyne;
нижний renderer и его performance probes описаны в
[MATERIAL_PIPELINE_VALIDATION.md](MATERIAL_PIPELINE_VALIDATION.md).
Проверка не заменяет длительные input-to-present измерения на 120–360 Hz,
mixed-DPI monitor migration, physical 2K/4K/5K, macOS/mobile и HDR приёмку.
Отдельно остаётся визуальная проверка контраста tint/foreground при live theme
switch: пользовательские material colors не заменяются автоматически темой.

## Воспроизводимость

| Полный набор Fyne + apps, race detector | Пакетов PASS | Test/subtest PASS events | FAIL |
| --- | ---: | ---: | ---: |
| Windows native, accessibility, clipboard/drag/material/display gates | 47 | 4941 | 0 |
| Windows headless (`no_glfw`) | 47 | 4753 | 0 |
| Linux X11, AT-SPI, clipboard/drag/material/display gates | 49 | 5036 | 0 |

Нативная матрица framebuffer: 2048×1080, 2560×1440, 3840×2160, 5120×2880,
каждая при scales 1, 1.25 и 2. Physical monitor/refresh-rate приёмкой это не является.
Все 3252 fingerprinted source inputs совпали между workspace и Linux export.
Windows-only sibling fix проверен повторным полным native suite и реальным NVDA;
его файлы не входят в Linux build. После усиления только `input_probe` fixture
реальные CJK/Orca/NVDA/Wayland harnesses повторены, обычные suite inputs не менялись.

Полные suites включают Fyne и `extras/apps`, `-count=1`, race detector и timeout.
Точные counts/skips, обязательные native checks, SHA-256 источников и артефактов:
[summary.json](native-input-validation/summary.json),
[evidence.json](native-input-validation/evidence.json).
Отдельный отказ real Windows Japanese IME сохранён, не превращён в PASS.

```powershell
$env:FYNE_TEST_MATERIAL='1'
$env:FYNE_TEST_CLIPBOARD='1'
$env:FYNE_TEST_DRAGDROP='1'
$env:FYNE_TEST_DISPLAY_MATRIX='1'
go test -p 1 -race -timeout 5m -tags accessibility,migrated_fynedo -count=1 ./... ./extras/apps/...
go test -p 1 -race -timeout 5m -tags no_glfw,migrated_fynedo -count=1 ./... ./extras/apps/...
```

Linux harnesses ожидают собственный абсолютный work directory, `source`,
race-built `input-probe`/`input-probe-wayland` и временный Anthy prefix.
Запускать внутри отдельных `dbus-run-session -- xvfb-run`:

```sh
FYNE_TEST_REAL_IBUS=1 sh scripts/validate-anthy-linux.sh "$work"
sh scripts/validate-native-input-linux.sh "$work"
FYNE_TEST_MATERIAL=1 sh scripts/validate-wayland-drag-linux.sh "$work"
```

Только явный `FYNE_TEST_REAL_IME=1` включает реальный Microsoft IME test;
он требует корректно настроенную Windows IME-сессию. Обычный native suite не
считает отсутствие языкового компонента успешной IME приёмкой.

Пять dependency patches воспроизведены на чистых pinned upstream sources:
[fork-patch-replay.json](native-input-validation/fork-patch-replay.json).
Временные standalone NVDA, engine prefix, compositors, profiles и caches этого
запуска очищены после сохранения evidence:
[cleanup.txt](native-input-validation/cleanup.txt). Оставшиеся две Windows NVDA
proxy DLL и каталог `X:\fyne\.build-accessibility-c28a1842` удалены 2026-10-13
после разрешённого пользователем закрытия Яндекс Браузера и перезапуска оболочки
Explorer. Отсутствие каталога и восстановление оболочки проверены.
Восстановление Windows языковых профилей проверено:
[windows-language-restoration.json](native-input-validation/windows-language-restoration.json).
Установленные на стенде test tools
сохраняются; пользовательские и чужие временные каталоги не очищаются.
