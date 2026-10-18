# Визуальная модернизация: GL, GLFW, Glow и D-Bus

Оценка от 12 октября 2026 года для текущего рабочего дерева `X:\fyne`.
Входной документ — `F:/Загрузки/Fyne_Visual_Modernization_TZ.md`, версия 1.0,
11 октября 2026 года, SHA-256
`b17bc258a1281bd5f16144148029b07832d35cf9b72661b60e03fb572d30cc2c`.
ТЗ используется как источник требований. Этот разбор не означает реализации
всего ТЗ или приёмки нативных материалов.

База ТЗ — Fyne 2.8.1 / `3dc06f47137aa3709807f1bb78dacc5daf68d551`;
текущий HEAD — `7f99a87b0e1903ebc17f3c8c8f478b840a6eabe5`, с незакоммиченными
изменениями Fyne и локальных форков. Перед общей приёмкой нужно зафиксировать
единую ревизию и baseline. Уже присутствующие изменения ядра, включая
`desktop.ScaleExact`, accessibility, IME и исправления shutdown, учитываются как
отдельная работа; новые возможности библиотек пока не включены автоматически
в визуальную политику Fyne.

## Вывод

Эти четыре библиотеки дают транспорт и графические примитивы. Цветовые токены,
состояния виджетов, типографика, text scale, политика accessibility, планирование
кадров и публичный `WindowVisuals` должны жить в Fyne. Системный материал окна
создаёт оконный compositor ОС; InAppAcrylic обрабатывает ранее нарисованные
слои собственного canvas. Их возможности, ресурсы и тесты нужно разделить.

Первый архитектурный gate — E1 из ТЗ: прозрачная WGL-поверхность вместе с DWM
на поддерживаемых Windows, затем тот же опыт с настоящим Fyne painter. Если
получить требуемый результат нельзя, нужен D3D11/DXGI/DirectComposition путь.
Расширение набора GL-функций само по себе этот gate не закрывает.

## Первая итерация

| Компонент | Реализовано | Польза для ТЗ | Граница |
| --- | --- | --- | --- |
| `deps/gl` | Полная перезагрузка таблицы адресов даже при ошибке; `IsFunctionAvailable` во всех генерируемых профилях | После смены контекста отсутствующая функция не сохраняет старый адрес; можно диагностировать доступные entry points | Указатель не доказывает поддержку extension/API и не является проверкой материалов ОС; таблицы остаются глобальными |
| `deps/glfw`, 3.3/3.4/3.5 | `SurfaceMetrics`, `GetSurfaceMetrics`, `PixelScale` | Один cgo-вызов вместо восьми для размеров, scale и пяти атрибутов; UI scale отделён от плотности framebuffer | Последовательные native queries, только main thread, без event pumping; snapshot не атомарен относительно compositor |
| `deps/glow` | Staging, проверка форматирования до публикации, восстановление заменённых файлов при ошибке; удаление только старого generated debug; sorted symbol manifest | Воспроизводимые компактные профили для материалов и диагностика состава API | Восстановление при возвращённой ошибке публикации, без гарантии атомарности при crash; копирование headers имеет свой этап |
| `deps/dbus` / `portal` | Сохранение wire-signature Variant в Store containers; cancellable Settings.ReadAll, типизированные appearance preferences и parser SettingChanged | Исправлена реальная потеря сигнатуры `(ddd)` цвета accent при чтении; Linux-источник темы, contrast и reduced motion без подстановки вымышленных настроек | Не создаёт watcher, goroutines или UI policy; реальный desktop portal и его restart требуют отдельной интеграции |

Новый D-Bus клиент использует `ReadAll`, доступный в версиях 1 и 2 Settings,
различает отсутствие ключа и значение по умолчанию. Стандартные четыре appearance
ключа не покрывают text scale, запрет transparency и compositor blur. Неизвестные
enum значения нормализуются к отсутствию предпочтения, недопустимый accent
считается отсутствующим. Основание: [XDG Settings](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.Settings.html).

Результаты текущих проверок и raw logs:
[VISUAL_STACK_VALIDATION.md](VISUAL_STACK_VALIDATION.md).

## Следующая реализованная очередь

Добавлены material-профиль GL 2.1 (78 functions, optional core/ARB/EXT FBO),
независимые context capabilities и проверка support вместе с procedure addresses.
GLFW 3.3/3.4/3.5 получили лёгкий Geometry без пяти attribute queries. Glow schema 2
содержит C/Go signatures, а неизменённая регенерация сохраняет timestamps Go/C
inputs. D-Bus получил bounded subscriptions и portal watcher с owner/restart,
overflow recovery, отменой старого чтения и coalesced snapshots.
Проверки и границы: [LIBRARY_DEVELOPMENT_VALIDATION.md](LIBRARY_DEVELOPMENT_VALIDATION.md).

Следующая итерация добавила per-context dispatch в Fyne painter, RGBA8 material
pool/cache, GPU snapshot и blur/mask/tint/noise, optional DWM API во всех трёх
версиях GLFW, portal retry/readiness и один XDG system appearance provider Fyne.
Результаты, контракты и границы: [MATERIAL_RUNTIME_VALIDATION.md](MATERIAL_RUNTIME_VALIDATION.md).
Material graph/popups, native visual acceptance, typography и providers других
ОС остаются частью общего backlog; таблицы ниже включают уже реализованные основы.

Следующий этап соединяет crop/downsampling, adaptive budget и blur в единый
`material.Renderer`; добавляет batched C-stack queries Glow, поколения поверхности
GLFW и ordered portal deltas без повторного ReadAll при каждом изменении.
Результаты: [MATERIAL_PIPELINE_VALIDATION.md](MATERIAL_PIPELINE_VALIDATION.md).

## Соответствие ТЗ и недостающая работа

| Требования | Что есть в исходниках | Что ещё требуется |
| --- | --- | --- |
| RES-01/02, VIS, TXT, CMP | Обычная тема и painter Fyne; низкоуровневые функции GL | Опциональный style layer и versioned semantic tokens, состояния виджетов без изменения геометрии; согласованные font metrics, caret, selection и accessibility. Это работа Fyne, а не обёрток GL |
| RES-03, DPI | В текущем ядре есть `ScaleExact` и `WindowScaleInfo`; в GLFW добавлена раздельная геометрия окна и framebuffer | Подключить native metrics к per-window state; проверить обе оси, несколько мониторов, fractional scale, text scale отдельно от UI scale; поколения GPU/cache при resize/DPI |
| RES-04, WIN, POC, GPU | GLFW поддерживает pre-create alpha hint и native window handles; GL может хранить RGBA | DWM capability/OS-build/provider, alpha-контракт всех passes, реальный desktop POC, lifecycle и первый кадр; GPU-only D3D bridge при необходимости |
| RES-05, IAC | `drawBlur` делает GPU copy и два прохода, повторно использует snapshot texture и kernel | Снимок нужных нижних слоёв, отдельные FBO ping-pong passes, clip/mask, tint/noise, dirty tracking, ограниченная глубина и общий бюджет памяти; не читать собственный результат как исходный backdrop |
| RES-06, SYS, MOT | Новый portal reader/parser; часть native accessibility в ядре | Один provider настроек на приложение, подписки с cancel/restart, snapshots с generation, объединение bursts, сохранение user overrides; единый motion scheduler с остановкой скрытых окон |
| API, LIF, CMP | Базовые public interfaces не требуют новых методов | Опциональный WindowVisuals/observer, request ID и last-wins, Pending/Ready/Fallback/Failed/Closed; причины fallback и возможность фактической созданной поверхности |
| RES-07, PER, TST, AT | Native probes, race, C sanitizers, microbenchmarks | Gallery и кадры общей сцены; пять длительных прогонов, CPU/GPU memory, present/input latency, реальная аппаратная и compositor матрица, desktop capture материалов |

`internal/painter/gl/context.go` пока не содержит FBO или `BlendFuncSeparate`.
Компактный desktop-профиль содержит 50 функций, включая только то, что использует
текущий painter. Полный `deps/gl` уже содержит многие нужные GL/extension entry
points: прежде чем дописывать wrappers, нужно проверить существующие сигнатуры и
добавить отдельный профиль/adapter для материалов.

В `drawBlur` ограничение до 101 samples растягивает шаг выборки, но не создаёт
отдельную уменьшенную render target. Два `CopyTexSubImage2D` и промежуточная запись
в главный framebuffer не дают самостоятельного material graph. Нужны явные
source/pass/result targets и измерение их качества и стоимости. Перевод всех
операций в compute shaders не нужен для минимального GLES 2/WebGL пути.

Legacy scale продолжает округляться до десятых; это допустимо для compatibility
policy. `ScaleExact` уже сохраняет дробное значение. Следующая работа должна
подключать корректные native значения к каждому окну, а не повторно вводить этот
API. `ContentScale` нельзя безусловно умножать на `PixelScale`: их смысл зависит
от backend, и такой код способен дважды применить Retina/Wayland scaling.

## Приоритетный backlog библиотек

Приоритеты относятся к визуальной модернизации: P0 — контракт или gate до
интеграции; P1 — обязательная реализация после gate; P2 — отдельное расширение.

| Приоритет / задача | Владелец и место | Результат и проверяемое условие |
| --- | --- | --- |
| P0: RGBA и композиция | Fyne painter, `gl`/`glow` profiles, GLFW native backend | Явные straight/premultiplied и linear/sRGB форматы каждого pass; RGB=0 при A=0; корректный отдельный alpha blend; alpha patches 0/.25/.5/.75/1 на чёрном, белом и RGB фоне без halo; opaque mode A=1 |
| P0: Windows WGL/DWM POC | GLFW Win32 + отдельный Fyne native provider | Проверить фактический pixel format и alpha, не только HRESULT; 22621, 22631, 26100+; focus/unfocus, resize, minimization, две поверхности, RDP/VM fallback; решение о D3D bridge до основной разработки |
| P0: Surface/capability state | Fyne window state и GLFW | Requested, OS-supported, renderer-supported и available-on-created-surface отдельно; resource generation и отмена после Close; alpha hint задаётся до создания; recreate не скрывается от lifecycle |
| P1: Минимальный material profile | `deps/glow`, `deps/gl/profiles` и Fyne context adapters | `BlendFuncSeparate`, FBO create/bind/attach/check/delete и нужные enums; GL 2.1 EXT/ARB вариант, GLES 2 core и WebGL adapter; загрузка optional entry points не повышает минимальную обязательную GL-версию |
| P1: Capability query | `deps/gl` + Fyne renderer | Version/extensions/limits, framebuffer completeness, alpha bits, sample count, colorspace; pointer availability применяется только вместе с API/extension проверкой; контекст без FBO даёт объяснимый fallback |
| P1: System settings provider | `deps/dbus/portal` и Fyne app settings | ReadAll + SettingChanged, уникальный owner и NameOwnerChanged, re-read после restart, bounded signal queue, connection lifetime/cancel; внешний sender не меняет UI; overrides и отсутствующие ключи сохраняются |
| P1: Per-window DPI | GLFW callbacks + Fyne driver/cache | `SurfaceMetrics` собирается после обработки событий на main thread; logical/UI/framebuffer/text scales отдельно; нулевой framebuffer не запускает rendering; cache key включает scale и resource generation |
| P1: Material resources | Fyne painter/cache | Два FBO для separable blur, downsample/upscale по физическим пикселям, clip/corners; source generation + параметры + theme version; статичная сцена не повторяет blur/noise; resize peak входит в бюджет 64 MiB |
| P1: Presentation/latency | GLFW platform adapters + Fyne scheduler | Present timestamps и callback pacing, stop на hidden/iconified/zero framebuffer; metrics GPU timing только при наличии extension, input-to-present измеряется отдельно; CPU Render duration не выдаётся за present latency |
| P1: Lifecycle integration | GLFW native handle hooks + Fyne | Main/UI/render threads имеют явные границы; handler/observer removal до уничтожения HWND/NSWindow/wl_surface; первая готовая поверхность до Show; context recreation и закрытие двух окон без утечек |
| P2: Context dispatch | `deps/glow` и `deps/gl` | Опциональные неизменяемые procedure tables на context generation вместо одной mutable global table; тест двух устройств/контекстов и прежний глобальный API для compatibility |
| P2: Wayland backdrop | GLFW Wayland extension adapter | Исследовать `ext-background-effect-v1` актуальной wayland-protocols; binding только к объявленному global и проверенным capabilities, обработка их изменения/удаления; отсутствие extension означает fallback, не crash |
| P2: macOS native composition | GLFW Cocoa bridge + Fyne provider | AppKit material view за GL/Metal content, system accessibility preferences и жизненный цикл; macOS-материалы имеют собственные названия/поведение; проверить Intel и Apple Silicon |
| P2: Additional renderers | Отдельные Fyne adapters | D3D11/DirectComposition при провале WGL gate; Metal как долгосрочный путь macOS; Vulkan/HDR не являются обязательным условием первого material profile |

У GL 2.1 FBO может предоставляться extension; у GLES 2 FBO есть в core. Не нужно
делать все desktop extension функции обязательными в минимальном профиле.
Профиль генерируется из явно перечисленных символов, сохраняет manifest и
проверяется реальным context; lazy import существующих полных bindings нельзя
считать экономией загрузки, если затем всё равно вызывается полный Init.

## Платформы и визуальные возможности

Матрица задаёт объём разработки и приёмки. Указанные будущие provider/backdrop
возможности ещё не являются поддержкой в текущем fork.

| Платформа | Окно и renderer | Настройки / масштаб | Материалы и fallback | Обязательная приёмка |
| --- | --- | --- | --- | --- |
| Windows 11 x64 | GLFW Win32 + WGL; D3D11/DXGI резерв по E1 | Per-monitor DPI, отдельно text scale и accessibility; HWND generation | DWM Mica/Mica Alt/Desktop Acrylic при поддержке и разрешении; opaque fallback с причиной | Сборки 22621, 22631, 26100+; Intel/AMD/NVIDIA, два DPI, focus, minimize, first frame, RDP/VM |
| Windows ARM64 | Отдельно проверить C compiler, Go/cgo, GLFW Win32 ABI и renderer | Те же контракты, собственный стенд | Материалы только после native verification | Compile недостаточен; ARM64 не считается выполненным по x64 результату |
| Linux X11 / XWayland | GLFW X11 + GLX или EGL | Portal при наличии; content scale и фактическая геометрия; XRandR отдельно | Alpha зависит от compositor; универсального GLFW desktop blur API нет; opaque + InAppAcrylic | GNOME/KDE/Xfce или явно выбранные WM; compositor on/off, XWayland отдельно, resize, mixed DPI и ввод |
| Linux Wayland | GLFW Wayland + EGL; viewporter/fractional-scale при объявлении compositor | Portal; logical surface size отдельно от framebuffer; DPI без X11/Windows физического baseline | Alpha по surface; background effect только по конкретному поддержанному protocol/capability; opaque + InAppAcrylic | GNOME, KWin и один wlroots compositor, Weston smoke; 1/1.25/1.5/2, output enter/leave, popup, IME, decorations, suspend |
| BSD | GLFW X11; FreeBSD/NetBSD Wayland в соответствующем build | D-Bus/portal опциональны и требуют своего runtime | Непрозрачная тема; alpha и compositor extensions по capability | FreeBSD native сначала, другие BSD по заявленной матрице; отсутствие portal не блокирует запуск |
| macOS Intel / Apple Silicon | GLFW Cocoa + текущий GL; Metal — отдельный adapter | AppKit appearance/accessibility, backing scale отдельно от points/text | AppKit materials через native bridge; generic opaque/InAppAcrylic при недоступности | Оба CPU, Retina/non-Retina, два дисплея, reduce transparency/motion, sleep/wake, fullscreen, GL recreation |
| Android | Fyne mobile driver + GLES, без desktop GLFW | Android configuration/font scale/reduce motion provider | InAppAcrylic в пределах GLES capability; платформенный window backdrop отдельной функцией | Реальное устройство, surface/context loss, rotation, keyboard/insets, memory pressure |
| iOS / iPadOS | Fyne mobile driver; текущий GLES, Metal перспектива | UIKit traits, Dynamic Type, reduce transparency/motion | UIKit materials требуют native adapter; generic opaque/InAppAcrylic | Устройство и simulator отдельно; background/foreground, memory warning, scale и keyboard |
| Web / WASM | Browser/WebGL driver, без cgo `deps/gl`/GLFW и desktop D-Bus | Browser media queries, device pixel ratio и zoom; user override | Canvas material graph; DOM/CSS backdrop-filter не эквивалентен blur содержимого GL canvas | WebGL1 minimum / WebGL2 optional, context lost/restored, browser zoom, high DPI, reduced motion |
| Software / test | Software renderer / headless | Детерминированные заданные значения | Opaque либо явно определённый software material с budget | Visual snapshot/layout/input tests; отсутствие native handles не является ошибкой |

GLFW предоставляет прозрачность содержимого через pre-create hint и отдельный
атрибут фактической поверхности. Нужна включённая desktop composition;
декорации окна не входят в эту прозрачность. Это позволяет проверить alpha,
но не подтверждает доступность Mica или compositor blur.
Основание: [GLFW transparency](https://www.glfw.org/docs/latest/window_guide.html#window_transparency).

Wayland protocol support определяется сервером; например, GLFW требует
xdg-shell, а decoration/focus-related extensions зависят от compositor.
Нельзя моделировать Wayland как X11 с доступом к любым координатам и focus.
Официальное описание: [GLFW compatibility](https://www.glfw.org/docs/latest/compat_guide.html).
Смысл fractional scale и правил округления нужно сверять с
[fractional-scale-v1 XML](https://gitlab.freedesktop.org/wayland/wayland-protocols/-/blob/main/staging/fractional-scale/fractional-scale-v1.xml);
копия протокола уже находится в vendored GLFW.
Для будущего background-effect требуется отдельно проверить актуальный
[upstream XML](https://gitlab.freedesktop.org/wayland/wayland-protocols/-/blob/main/staging/ext-background-effect/ext-background-effect-v1.xml)
и реализации целевых compositor; его наличие у всех платформ не предполагается.

На Windows `DWMWA_SYSTEMBACKDROP_TYPE` доступен с build 22621; режимы main,
transient и tabbed соответствуют Mica, Desktop Acrylic и Mica Alt.
`DWMWA_REDIRECTIONBITMAP_ALPHA` имеет отдельную границу 26100 и требует
premultiplied bitmap. Поэтому его нельзя сделать единственным alpha-путём для
ранних Windows 11. Основания:
[DWM backdrop types](https://learn.microsoft.com/en-us/windows/win32/api/dwmapi/ne-dwmapi-dwm_systembackdrop_type),
[DWM attributes](https://learn.microsoft.com/en-us/windows/win32/api/dwmapi/ne-dwmapi-dwmwindowattribute).

Apple deprecated OpenGL с macOS 10.14 и рекомендует Metal для нового GPU-кода.
`NSVisualEffectView` использует AppKit materials, включая blending содержимого
за окном; это отдельный native bridge, а не новая GL-функция.
Основания: [Apple OpenGL guide](https://developer.apple.com/library/archive/documentation/GraphicsImaging/Conceptual/OpenGL-MacProgGuide/opengl_intro/opengl_intro.html),
[NSVisualEffectView](https://developer.apple.com/documentation/appkit/nsvisualeffectview).

## Порядок внедрения и оптимизации

1. E0: закрепить текущий Fyne revision, patches форков, versions и аппаратный
   baseline. Запускать legacy и modern сцены на одном стенде. Измерения Init и
   cgo не переносить на FPS всей сцены.
2. E1: native Windows POC и Fyne POC, alpha patches и frame decorations; затем
   решение WGL или D3D composition. Параллельно зафиксировать общие surface и
   settings contracts для остальных платформ.
3. E2: подключить SurfaceMetrics к optional window state, exact DPI, resource
   generation и settings providers. Portal subscription должна заменять
   polling-настройки, а не создавать второй updater на каждое окно.
4. E3: непрозрачные semantic themes и typography на всех drivers; проверки
   high contrast/reduced motion и старого API до добавления эффекта.
5. E4/E5: native materials и InAppAcrylic разными provider/pass. Начать с depth
   <=2 и ограниченного размера blur region; FBO reuse, downsampling, dirty-region
   updates и noise cache оптимизировать по GPU timing и качеству.
6. E6: gallery/matrix и параметры PER из ТЗ: 1920x1080 physical, 150%, 60 Hz,
   10k rows и material region, warm-up 10 s, пять прогонов по 60 s. Проверить
   P95 present <=20 ms, P99 <=34 ms, input <=100 ms и CPU/memory budgets.

Разумные первые оптимизации: сбор native metrics при изменении поверхности с
использованием сохранённого состояния в кадрах, reusable
texture/FBO buffers, пропуск невидимых и статичных passes, bounded latest-wins
очередь настроек и материалов, минимальные symbol profiles. `ReadPixels` оставить
для Capture/тестов; не использовать его как часть обычной передачи кадра в DWM
или D3D. Не добавлять `noescape/nocallback` к вызовам GLFW, способным вызвать
Go callbacks. Новые backend и SIMD нужны только после измерения узкого места.

Полный SurfaceMetrics запрашивает и native flags; X11 queries могут требовать
server round trips. Его нельзя автоматически вставлять в каждый render pass.
Если hot path требует только geometry, следующий API должен запрашивать только
нужные поля либо использовать состояние обновлённых callbacks. Текущий API
аддитивен и сам не меняет старый renderer loop; в Xvfb timing ускорение не
подтвердилось, несмотря на устранение Go allocations.

## Критерии завершения

Новые функции библиотек приняты после unit/race/native checks, генерации и
replay patches. Полная визуальная модернизация принята только после применимых
AT-01–AT-14, платформенных дополнений выше и raw present/input/memory metrics.
Headless Weston проверяет клиентский Wayland/EGL путь, но не blur GNOME/KWin,
physical fractional DPI, interactive input, HDR или desktop appearance policy.
D-Bus тестовый Settings service подтверждает wire interoperability, но не
поведение реального desktop portal. Нативные материалы Windows и macOS
нуждаются в собственных runtime стендах и видимом результате desktop composition.
