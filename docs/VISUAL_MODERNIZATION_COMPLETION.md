# Завершение программной части визуального ТЗ

Дата: 14 октября 2026. Основание: `Fyne_Visual_Modernization_TZ.md`, v1.0,
и текущая рабочая база форка Fyne. Последнее указание пользователя исключает
длительные performance-проверки из этой итерации. Дополнительные Windows/GPU,
macOS, GNOME/Fly стенды пользователь подтвердил как недоступные. Позже предложены
KDE-стенды f-test-osnovacustom (X11, 172.21.9.27) и f-autotest-altkde
(172.21.9.21). Подключение через devproxy и f-test-bench вернуло No route to host/
SSH timeout; пользователь подтвердил, что стенды пока неактивны. KDE runtime
приёмка не зачтена и не заменена Weston.

Программные пункты дополнены Typography/Motion/Component/Material API,
версионированными tokens, readonly и составными контурами, сохранением transient
state при смене темы, расширенной галереей и exporter эталонов. Полная аппаратная
приёмка ТЗ остаётся открытой: сборка и automated checks не заменяют недоступные
аппаратные режимы и оптические проверки. AT-13 в этой работе исключён, а не зачтён.

Итоговый Windows root suite прошёл для 47 пакетов с race/cgocheck2, включая
регрессию переключения цвета иконок экспортера. Короткие native tests подтвердили
100 lifecycle/resize/backdrop циклов без пересоздания surface, реальные
maximize/minimize flags 100/100 и сохранение текста/фокуса/выделения/binding при
TextScale 100→150→200→225→100%. Галерея с установленным Segoe UI и режимами Mica,
Mica Alt и Desktop Acrylic запускается на Windows 26300 / NVIDIA RTX 4080 SUPER
(driver 32.0.16.1692); все три отчёта содержат SurfaceReady, NativeAccepted,
FirstFrameBeforeShow и 6 prepared frames. Это проверка native pipeline, её
границы оптической приёмки остаются прежними. 39 opaque PNG дважды экспортированы
и совпали по SHA-256; dark/high-contrast при 225% просмотрены.

Linux root/race, X11/Wayland GLES2 functional tests, material test, portal
startup, vet, gallery и compile checks прошли на f-test-bench до последней
правки экспортера эталонов. При переносе этой правки SSH через devproxy начал
сбрасываться до handshake. Пользователь подтвердил, что SSH-маршрут временно
не работает. Финальный Linux delta не применён; raw logs и
проверку идентичности remote sources пока нельзя выгрузить. Это не засчитано как
проверка окончательного exporter на Linux. Windows-артефакты и локальный source
snapshot сохранены в visual-completion-validation; Linux поля acceptance.json
оставлены недоступными. Для восстановления полного протокола используются
pack-core-validation.py, validate-visual-completion-linux.sh и
record-visual-completion.py без флага --local-only.
Окончательные WASM и macOS ARM64 software сборки дополнительно проверены локально
после правки экспортера; их метаданные сохранены в Windows-папке артефактов.

## Требование → реализация → доказательство

| Область ТЗ | Реализация | Доказательство и границы |
|---|---|---|
| RES-01, VIS-01–03 | ModernTokenVersion=1, NewModern, light/dark/HC, optional style providers | theme visual_style/modern/system tests; v1 PNG+SHA-256 manifest |
| RES-02, VIS-04–10 | Составные VisualState, высота 40/32 с ростом по тексту, две границы Entry, stock colors/radii, readonly | widget visual_profile/popup_visual tests; прежние snapshots остаются в suite |
| TXT-01–02 | Один resolved font/style/size для TextSegment draw/measure/selection/hit test; baseline и semantic line spacing | Русский/English/numbers/path/code, оба font policies; RenderedTextSize equality |
| TXT-03–06 | Opt-in installed Segoe UI, bundled fallbacks, trailing/code, vector icon cache, AccessibleName | SystemUIFont test Windows; portable fallback tests; gallery toolbar/icon-only |
| WDG-01–03 | ThemeChangeAware вокруг child reflow, Entry rune positions/viewport/binding; прежняя виртуализация | Live scale/density/HC test; 10k List/Table/Tree создают только viewport children |
| WDG-04 | Перенос labels/error helpers в modern Form, ширина измеряется до row height | TestModernFormWrapsLabelsAndErrorsWithoutReplacingInput |
| WDG-05–06 | Escape top surface, modal cancel callback, previous focus/selection; штатные Fyne dialogs | Popup software/native tests и существующий dialog suite; IME Escape сперва отменяет preedit |
| MOT-01–04 | Общий Runner, durations 100/80/180/220 ms, latest target from current color, color-only refresh | Policy/retarget tests; Button/PopUp; Navigation token доступен, Tabs switch мгновенный |
| MOT-05–06 | Most-restrictive policy, once completion, hidden/minimized scope pause, static busy indicator | Unit tests; native visibility на Windows, X11, Wayland; actual minimize только Windows |
| RES-03, DPI | ScaleExact, отдельные logical/window/framebuffer/text scales; WM_DPICHANGED у GLFW | Existing exact scale/display tests; 12 software DPI/TextScale combinations; physical mixed monitors остаются открыты |
| RES-04, ALP/POC | WGL/DWM, first hidden frames, premultiplied alpha, status/capability/fallback | WINDOW_VISUALS_VALIDATION.md; новый 100-cycle native test; Win26300/NVIDIA, 22621/22631 не проверены |
| RES-05, IAC | Cropped source, sharp foreground, bounded two-layer GPU pool, fallback | Native canvas/popup material X11; предыдущие Windows/GL/GLES checks; optical compositions на остальных GPU открыты |
| RES-06, SYS-01–04/08–09 | Windows WinRT/Win32 + Linux portal; синхронный startup snapshot; UI dispatch/cancellation, immutable overrides | Windows provider ранее; native Linux initial snapshot 0.01 s; не измерено 500 ms при изменении реальной ОС |
| SYS-05–07 | System HC roles, no shadow/blur/transparency/motion; status contrast, opaque tint fallback | Palette tests; 4.5:1 text/status on light/dark surfaces and tinted black/white/R/G/B source extrema |
| SYS-10–11 | Сохранены native accessibility bridges/roles; readonly state/edit guards; explicit icon names | Existing accessibility/IME suites; screen reader hardware acceptance прежняя, не полный аудит |
| TST-01–05, RES-07 | Root/race/cgocheck2, CI, vet, native lifecycle, gallery, deterministic opaque capture | См. артефакты visual-completion-validation; сборки WASM и macOS software являются только compile checks |
| DOC-01–02 | Compiling gallery/profile/capture examples, migration guide, source manifest | VISUAL_MODERNIZATION_MIGRATION.md; универсальный core отделён от OS files; commit/push/PR не выполнялись |

## Матрица приёмки

| AT | Статус этой итерации |
|---|---|
| 01 | Автоматическая совместимость legacy/custom и прежних эталонов проверяется полным suite |
| 02 | Все семейства представлены в галерее; автоматические state/type/layout tests и opaque references; визуальная аппаратная оценка всех состояний открыта |
| 03 | Открыт: Windows 22621/22631 отсутствуют |
| 04 | Code/native acceptance проверены на 26300/NVIDIA; оптические материалы остальных конфигураций открыты |
| 05 | Pure/native/software scale checks; физические fractional-scale/input проверки всей сетки открыты |
| 06 | Открыт: два физических монитора и 20 round trips не проверены |
| 07 | DPI 100/150/200 × TextScale 100/150/200/225 software; native live text/focus/selection/binding; реальный OS text scale open |
| 08 | Политики и synthetic snapshots проверены; реальное изменение пользовательских настроек ОС не выполнялось |
| 09 | 100 native show/hide/backdrop cycles; surface generation и first-frame status; оптическая непрерывная запись отсутствует |
| 10 | 100 resize/maximize/minimize requests; Windows actual native flags 100/100; headless Linux не даёт эквивалентной проверки minimize |
| 11 | Runtime policy/fallback tests; доступного RDP/VM GPU стенда нет |
| 12 | Native pipeline source/clip/cache/depth/budget tests; обязательная optical/GPU hardware quality матрица открыта |
| 13 | Исключён пользователем: длительные performance/stress runs не запускались и не добавлены в script |
| 14 | Linux X11/Wayland/GLES2 + software/tests; WASM build; macOS ARM64 software compile; macOS native runtime open |

## Воспроизведение и артефакты

Артефакты: [visual-completion-validation](visual-completion-validation/).
Эталоны: [references/manifest.json](visual-completion-validation/references/manifest.json).
Миграция/API: [VISUAL_MODERNIZATION_MIGRATION.md](VISUAL_MODERNIZATION_MIGRATION.md).
Итоговые package counts, source hashes и проверенные режимы записываются в
`acceptance.json` и `source-manifest.json` рядом с raw logs.

Windows root выполняется последовательно `go test -p 1 -race -tags migrated_fynedo
-timeout 4m ./...` с `GOEXPERIMENT=cgocheck2`, изолированными USERPROFILE,
GOCACHE/GOTMPDIR/TEMP/TMP. `FYNE_TEST_TZ_FUNCTIONAL=1` включает три коротких native
tests: visibility/Escape, 100 lifecycle/resize cycles и live TextScale.

Linux exporter `scripts/pack-core-validation.py` переносит рабочие sources и
offline module proxy. `scripts/validate-visual-completion-linux.sh` выполняет
portal, root/race, X11 и Wayland/GLES2 native tests, GPU material, vet, gallery,
WASM и macOS software compile. Weston/runtime принадлежат script и удаляются
его trap. Общий локальный/remote временный validation directory удаляется после
сохранения необходимых результатов.

Remote каталог `/tmp/fvc14-c83f` пока недоступен для очистки из-за сброса SSH;
удалять его следует после выгрузки logs и source hashes. Локальный временный
validation directory удаляется после сохранения Windows-артефактов. Состояние
очистки фиксируется в acceptance.json. KDE runtime остаётся отложенным до
активации предоставленных стендов.

`go vet` для изменённого core и Linux driver выполняется со всеми анализаторами.
Windows native package даёт семь предупреждений unsafeptr в существующих
Win32 clipboard/IMM uintptr→pointer bridges; raw log сохранён. Для этого package
остальные vet-анализаторы выполнены с `-unsafeptr=false`, а native functional
suite — с race, cgocheck2 и checkptr=2. Это явное исключение диагностики, а не
заявление о чистом полном native vet.

Тестовый portal service подтверждает wire/provider contract, а не настоящий
GNOME/KDE/Fly portal. Headless Xvfb/Weston не подтверждают физические мониторы,
системное размытие конкретного desktop environment или взаимодействие с
пользовательским GPU compositor. NativeAccepted и raw framebuffer alpha не
являются самостоятельным доказательством видимого Mica/Acrylic.

## Материалы пользователя: WinUI Gallery 3.0 и Plasma 6.8

14 октября 2026 пользователь передал статьи Comss о WinUI Gallery 3.0 и OpenNet
о Plasma 6.8. Проверены их первоисточники:

- [Microsoft, 13 октября: WinUI Gallery 3.0](https://devblogs.microsoft.com/ifdef-windows/winui-gallery-3-0-is-here-tableview-charts-inkcanvas-and-more/).
  Gallery использует Windows App SDK 2.5.4-experimental; новые TableView, Charts,
  WrapPanel/LayoutPanel, PagerControl, InkCanvas/InkToolbar и Window size APIs
  обозначены как экспериментальные. Это ориентиры для оценки дополнительных
  возможностей Fyne, а не новые закрытые пункты приёмки текущего ТЗ.
- [KDE, 14 октября: Plasma 6.8](https://kde.org/announcements/plasma/6/6.8.0/).
  Матрица дополнена native Wayland и XWayland отдельными строками; Plasma 6.8
  X11 session больше не предполагается. Runtime проверки остаются отложенными.

Для следующей итерации полезны следующие направления:

| Направление | Текущая база Fyne и предлагаемый следующий шаг |
|---|---|
| Каталог и доступность | Дополнить галерею поиском по controls/сценариям, примерами кода и демонстрацией состояний, focus/keyboard/accessible names |
| Табличные данные | Сохранить viewport virtualization, binding, multiple selection и существующие StickyRowCount/StickyColumnCount; отдельно оценить сортировку/фильтрацию, группировку и редактор ячеек с validation |
| Layout/paging | Сопоставить Flow/GridWrap с WrapPanel; определить переносимый Pager и сценарии неизвестного общего числа страниц |
| Charts/ink | Оценить самостоятельные canvas widgets, объём входных данных, redraw budget и pen/touch contract до нового API |
| Оконный API | Оценить optional min/max size constraints над существующим GLFW SetSizeLimits с учётом всех desktop платформ |
| KDE/Union | Native Wayland и XWayland проверять по capabilities; возможный адаптер темы строить вокруг Typography/Component/Motion/Material providers и immutable snapshots |

Расширенная DataGrid и собственный titlebar остаются дополнительным объёмом
раздела 16 исходного ТЗ. Экспериментальные APIs WinUI не копируются автоматически
в публичный Fyne API. Union требует отдельного output/adapter для Fyne:
[документация архитектуры](https://api.kde.org/overview.html). Эти статьи не
меняют статусы аппаратной приёмки или отложенной Linux-выгрузки.
