# Modern visual profile v1

Реализация в этом форке остаётся opt-in. Обычные `Theme`, `Window`, `Canvas`,
`Settings` и custom widgets сохраняют свои интерфейсы. Основной desktop backend
использует локальный GLFW 3.5.1 и компактные GL/GLES bindings, генерируемые Glow.
Версия Go — 1.27.1. Работа с D-Bus выполняется локальным модулем v5.3.0.

## Подключение

```go
a := app.NewWithID("example.modern")
w := a.NewWindow("Modern Fyne")
if err := desktop.ApplyModernProfile(a, w, theme.ModernOptions{
    TokenVersion: theme.ModernTokenVersion,
    FollowSystem: true,
    Density: theme.DensityComfortable,
    FontPolicy: theme.FontBundled,
}); err != nil {
    panic(err)
}
entry := widget.NewEntry()
entry.SetText("Русский · English")
title := widget.NewLabel("Семантический заголовок")
title.TextRole = fyne.TextRoleTitle
w.SetContent(container.NewVBox(title, entry))
w.Resize(fyne.NewSize(640, 400))
w.ShowAndRun()
```

Импорты: `fyne.io/fyne/v2`, `app`, `container`, `driver/desktop`, `theme`, `widget`
из этого модуля. Такой же путь используется в компилируемой `cmd/fyne_gallery`.
`ApplyModernProfile` включает `ScaleExact` на `ScalableWindow` и устанавливает
общую тему приложения. Для software/mobile/web драйвера достаточно переносимой
темы. Без helper можно отдельно вызвать `theme.NewModern` и `SetScalePolicy`.

TextScale относится к шрифтам и межстрочному интервалу. Он не умножает canvas,
иконки, оконный DPI и framebuffer density. Нулевой TextScale и nil Accent следуют
доступным системным значениям только при `FollowSystem`. Явные значения приложения
сохраняются. Ограничения high contrast, reduced motion и прозрачности имеют
более высокий приоритет.

## Токены и состояния

`ModernTokenVersion = 1`; ноль в options выбирает текущую версию. Неизвестная
версия, density, font policy, NaN/Inf или TextScale вне диапазона возвращают ошибку.
Optional providers описаны в `visual_style.go`; обязательный Theme не расширялся.

| Provider | Контракт |
|---|---|
| `TypographyTheme` | Size, LineSpacing, TextStyle, Font для semantic role |
| `ComponentTheme` | Составные VisualState и цвета/контуры/min-height компонента |
| `MotionTheme` | Duration, Curve, Disabled по роли перехода |
| `MaterialTheme` | BlurRadius, Noise, TintOpacity, CornerRadius, Tint, Fallback, Opaque |
| `ThemeChangeAware` | Сохранение временного состояния до обхода дочерних рендереров и восстановление после reflow |

| TextRole | Размер до TextScale | Стиль |
|---|---:|---|
| Body | 14 | regular |
| Caption | 12 | regular |
| Emphasis | 16 | bold |
| Subtitle | 20 | bold |
| Title | 28 | bold |
| Code | 14 | monospace |

`TextRoleDefault` сохраняет прежние SizeName/TextStyle. Низкоуровневый
`canvas.Text` продолжает использовать явный TextSize. `theme.Typography` разрешает
шрифт и стиль вместе, добавляя явные Bold/Italic/Monospace/Underline/Strikethrough.
Label и TextSegment используют результат для draw, measure, hit test и selection.
Semantic LineSpacing одинаков в layout и caret geometry; legacy paragraphs
сохраняют старое поведение. Для чисел можно задать `TextAlignTrailing`, для логов
и путей — Code/Monospace, Wrapping или Truncation согласно задаче приложения.

VisualState включает hovered, pressed, focused, disabled, readonly, selected,
error, inactive. Disabled подавляет hover/pressed/focus. Readonly, error и focus
совместимы: Entry резервирует отдельные внутренний error border и внешний focus
border; текст и MinSize не сдвигаются при смене этих состояний. Readonly использует
обычный цвет текста и отдельную подложку. Иконка/толщина границы и error/hint текста
формы дополняют цвет. Минимальная высота Button/Entry/Check/Radio/Select — 40 для
comfortable, 32 для compact, либо больше по измерению текста.

`Entry.SetReadOnly(true)` сохраняет focus, navigation, copy, binding и undo history.
Клавиатурные изменения, Cut/Paste, Undo/Redo, native surrounding-text edits,
IME/reconversion и accessibility SetText блокируются. Незавершённый preedit
отменяется. SetText и входящие изменения binding остаются разрешены; SetText, как
и раньше, сбрасывает undo. RichTextEntry наследует тот же запрет редактирования.

Для icon-only действий доступны `Button.AccessibleName` и
`ToolbarAction.AccessibleName`: имя описывает действие, независимо от имени asset.
Новые декоративные контуры не создают дополнительных accessibility roles.

## Шрифты

По умолчанию используются bundled ресурсы Fyne. `FontSystem` на Windows читает
установленные Segoe UI regular/bold/italic/bold-italic один раз, проверяя формат и
латинские/кириллические glyphs. Ошибка чтения/проверки отдельного face возвращает
bundled face соответствующего стиля. Monospace и Symbol используют прежние ресурсы;
дальнейший glyph fallback выполняет существующий painter Fyne. На остальных
платформах FontSystem использует bundled fallback. Segoe UI Variable и произвольная
системная замена UI font в этой политике не реализованы.

Файлы Windows fonts не включаются в репозиторий, package или reference archive.
Использование установленных fonts и их распространение различаются согласно
[Microsoft Font redistribution FAQ](https://learn.microsoft.com/en-us/typography/fonts/font-faq).
Состав лиц указан в [Segoe UI family](https://learn.microsoft.com/en-us/typography/font-list/segoe-ui).
Bundled fonts/icons сохраняют исходные licenses; vector icons используют обычный
Fyne raster cache при текущем DPI. Эталоны используют FontBundled.

## Анимации

| MotionRole | Duration v1 |
|---|---:|
| Hover | 100 ms |
| Pressed | 80 ms |
| Surface | 180 ms |
| Navigation | 220 ms |

Кнопки и появление обычного popup используют общий Runner. Переход кнопки меняет
только background; новый target начинается от текущего цвета. Popup меняет
только background при сохранении geometry/foreground. Material popup не получает
дополнительного декоративного fade. Navigation token доступен приложениям;
переход навигации штатных Tabs остаётся мгновенным, с сохранением состояния content.

`Animation.Object` ограничивает движение видимым объектом; desktop driver учитывает
скрытых родителей, неактивные вкладки и свёрнутое/скрытое окно одним общим обходом.
При возвращении timeline продолжается с сохранённого положения. Unscoped legacy
animation сохраняет прежнюю область действия. `Disabled`, settings/build policy,
системный reduced motion, high contrast и theme.DisableMotion объединяются по
запрету. Текущий переход доходит до конечного значения; OnFinished вызывается
однократно при естественном или policy completion, Stop отменяет без completion.
Repeat cycles не вызывают OnFinished. Запуск одного handle повторно не создаёт
неограниченной очереди pending animations. ProgressBarInfinite остаётся видимым
статическим индикатором работы при запрете движения.

## Оконные и canvas материалы

До первого Show можно вызвать:

```go
_, err := desktop.ConfigureWindowVisuals(w, desktop.VisualOptions{
    Backdrop: desktop.BackdropMica,
    PrepareAlpha: true,
    Fallback: desktop.VisualFallbackOpaque,
})
```

`WindowVisuals` — optional interface. Проверяйте capability/status/request ID;
SurfaceReady и NativeAccepted описывают запрос и framebuffer, а не оптическую
приёмку эффекта. Формат alpha резервируется до создания native surface. Поздняя
смена opaque-формата требует нового окна и возвращает `ErrVisualNeedsRecreate`.
Режим PrepareAlpha с начальным Opaque разрешает поздние переключения материалов
на той же поверхности. Обычная рамка/системные caption buttons сохраняются.

Windows использует WGL/DWM для Mica/Mica Alt/Desktop Acrylic; unsupported API,
RDP, high contrast или запрет прозрачности дают непрозрачный fallback. Linux
Transparent/SystemBlur используют runtime-возможности X11 compositor/KWin hints
или Wayland background-effect protocol. Если compositor/blur отсутствует,
рендерер выбирает opaque. Реальные GNOME/KDE/Fly в текущей аппаратной матрице
не проверены. На macOS сохранён переносимый opaque-профиль; AppKit backdrop не
является обязательной частью исходного ТЗ.

InAppAcrylic создаётся `canvas.NewThemedMaterial(theme, fyne.SurfaceCard, variant)`
и размещается перед резким content в Stack/overlay. Для popup можно присвоить
`popup.Material`. NewThemedMaterial разрешает снимок tokens; при смене темы
приложение обновляет его параметры, как в странице галереи. Painter независимо
соблюдает системные ограничения. Software/web painter рисует opaque fallback.
Материалы не захватывают desktop и другие окна, не создают собственные tickers.
Пул ограничен двумя layers, 48 MiB targets и 16 MiB cropped capture, суммарно
64 MiB. Превышение depth/region/resource budget даёт opaque fallback. ReadPixels
используется только явным Capture/проверками, а не штатным material/present pass.

## Галерея и эталоны

```sh
go run -tags migrated_fynedo ./cmd/fyne_gallery -profile modern
go run -tags migrated_fynedo ./cmd/fyne_gallery -profile materials -backdrop mica
go run -tags migrated_fynedo ./cmd/fyne_gallery -profile legacy
go run -tags migrated_fynedo ./cmd/fyne_gallery -font-policy system -text-scale 2.25 -compact
go run -tags migrated_fynedo,ci ./cmd/fyne_gallery -references-dir ./references-v1
```

Галерея включает Button/Toolbar, Entry/Form/password/readonly/errors,
Check/Radio/Select, Slider/Progress, Label/RichText/Hyperlink, Table/List/Tree по
10 000 rows, Tabs/Accordion/Split/Scroll, Menu/PopUp, Dialog/FileDialog, Card,
оконные и in-canvas материалы. Профили, density, TextScale, font и motion
переключаются в живом интерфейсе.

Exporter создаёт 39 PNG и SHA-256 manifest: modern light/dark/high contrast ×
DPI 100/150/200 × TextScale 100/150/200/225, плюс legacy на трёх DPI.
При переключении профиля exporter сбрасывает theme cache самостоятельно,
поскольку event loop не запускается. Проверяется окраска SVG при light→dark→light
как с test.Settings, так и с настоящими app.Settings на software driver.
`canvas.CaptureOpaque` композитит alpha over opaque background, не меняя окно и
политику эффектов. Это изображение Fyne canvas; DWM и внешний compositor в нём
отсутствуют. Для native effects нужен отдельный desktop capture и аппаратная
приёмка. Источник/качество GPU material фиксируются отдельными native tests.

Текущие результаты и оставшиеся аппаратные границы находятся в
[VISUAL_MODERNIZATION_COMPLETION.md](VISUAL_MODERNIZATION_COMPLETION.md).
