# Системная визуальная политика Fyne

Дата: 2026-10-13. Продолжение ТЗ `Fyne_Visual_Modernization_TZ.md`:
RES-06, SYS-01–SYS-05, SYS-08–SYS-09 и DPI-16.
База workspace: `7f99a87b0e1903ebc17f3c8c8f478b840a6eabe5`, Go 1.27.1.
Это протокол реализованного блока; полная приёмка ТЗ остаётся отдельной задачей.

## Реализация

Windows получает один провайдер на приложение. `UISettings` читает тему,
accent, `TextScaleFactor`, `AnimationsEnabled` и `AdvancedEffectsEnabled`;
`AccessibilitySettings` читает high contrast. Семантическая палитра берётся через
`GetSysColor`: Window/WindowText, ButtonFace/ButtonText, Highlight/HighlightText,
GrayText и Hotlight. Каждое свойство имеет собственный флаг доступности.

Начальный снимок применяется в `App.Run` до GL loop. При запуске loop
инвалидируются уже созданные layout/font/theme caches. Изменения ОС читаются на
отдельном MTA-потоке; native callbacks только устанавливают событие. Ни GLFW,
ни GL, ни Go-объекты из них не вызываются. UI применяет последний неизменяемый
снимок через один ожидающий callback.

На Windows 11 Pro build 26300 четыре подписки WinRT зарегистрировались.
`HighContrastChanged` вернул `0x80070490`. Для этого случая работает принадлежащее
провайдеру скрытое top-level Win32 окно, принимающее `WM_SETTINGCHANGE`,
`WM_SYSCOLORCHANGE`, `WM_THEMECHANGED`. Message-only HWND не используется, так
как он не получает broadcasts. Окно не показывается, не имеет GL context и
не перехватывает DPI. WinRT mask = `0x1b`, Win32 mask = `0x04`, итоговый = `0x1f`.
Причину отказа WinRT нельзя выводить только из этого HRESULT.

При полном наборе событий нет периодического опроса: используется отменяемое
ожидание события/сообщений. При отсутствии подписки для доступного свойства
включается резервное чтение раз в 500 мс. Неизменившееся значение не увеличивает
generation и не обновляет UI. Закрытие отменяет ожидание, отписывает события,
уничтожает HWND/class, освобождает COM/delegates/event и завершает WinRT apartment
на том же OS thread. Поздние callbacks владеют отдельной ссылкой на wake state;
отмена синхронизируется до освобождения native памяти.

```go
modern, err := theme.NewModern(theme.ModernOptions{FollowSystem: true})
if err != nil { panic(err) }
a.Settings().SetTheme(modern)
```

Нулевой `TextScale` следует доступному системному масштабу текста, `nil Accent`
следует системному accent. Явные значения сохраняются. `FollowSystem: false`
сохраняет прежние параметры accent/text scale. Системный high contrast имеет
приоритет в modern-теме независимо от этого флага. Обычный `Theme` не получает
обязательных методов: `SystemAppearanceTheme` и `SystemColorTheme` опциональны.

High contrast сохраняет пользовательскую палитру. Button использует
ButtonFace/ButtonText; Entry/Select/menu — Window/WindowText. Focus и выбор
текста/элементов List, Table, Tree, GridWrap получают контуры цвета Highlight,
сохраняя исходный текст и shaping пользовательских дочерних виджетов. Это
проекция палитры в существующие renderers, а не полное совпадение с нативной
заполненной подсветкой. Shortcut-текст имеет непрозрачный GrayText; декоративные
hover/pressed fills, overlay dimming и тени отключены. При выходе из high contrast
исходные параметры восстанавливаются.

Материалы окна/canvas используют снимок для запрета прозрачности и high contrast;
настройки анимаций соблюдают reduced motion. Новая тема не читает ОС в `Color`/
`Size`: параметры разрешаются при обновлении снимка. TextScale меняет типографику,
не умножая размер иконки или canvas.

## Проверки и артефакты

Логи и JSON: [system-appearance-validation](system-appearance-validation/).
Linux-матрица воспроизводится `scripts/validate-system-appearance-linux.sh`
после экспорта workspace/module proxy через `scripts/pack-core-validation.py`.

| Проверка | Результат |
|---|---|
| Windows root, `go test -p 1 -race -tags migrated_fynedo ./...`, cgocheck2 | 46 пакетов PASS |
| Linux/X11 root, race + cgocheck2, `f-test-bench` | 48 пакетов PASS |
| Windows native provider + shared theme/palette tests | race, cgocheck2, checkptr=2 PASS |
| Linux isolated portal provider lifecycle | PASS; чужая D-Bus connection остаётся открытой |
| Windows/Linux scoped vet | PASS |
| Windows software/CI: app, desktop, theme, canvas, widget | PASS |
| Windows Gallery: Mica, Mica Alt, Desktop Acrylic | PASS; снимок до loop, Ready, FirstFrameBeforeShow |
| Linux Gallery: X11 и Wayland/GLES2 под Weston | PASS; Transparent |
| WASM gallery build | PASS; browser runtime не проверялся |
| macOS ARM64 software/CI, без cgo | PASS; нативный AppKit не проверялся |

Native-проба читает реальные настройки, регистрирует настоящие события и
выполняет 30 циклов open/read/notify/cancel/close. Счётчики provider/delegate/wake
возвращаются к исходным значениям. Проверены объединение 1000 уведомлений,
сообщение в собственный HWND, отмена и 1.1 с без изменения generation в idle.
Это не 30-минутный ресурсный тест из ТЗ.

Фактические настройки Windows: dark, accent `#0063b1`, TextScale 1, transparency
разрешена, high contrast/reduced motion выключены. Gallery записала generation 1
до driver loop и при capture, размер текста 14, иконки 20. Материалы сохраняют
`SurfaceReady`, `NativeAccepted`, `FirstFrameBeforeShow`. Smoke подтверждает
интеграцию; оптические свойства DWM проверялись раньше в
[WINDOW_VISUALS_VALIDATION.md](WINDOW_VISUALS_VALIDATION.md).

Начальные логи сохранены отдельно. Исправлены повторный цвет новой роли в
тестовой теме и отсутствие роли в markup renderer. Один параллельный Windows
прогон не прошёл чувствительный к нагрузке тест 280 Гц; полный последовательный
прогон прошёл без изменения порогов. В compile gate macOS добавлен отсутствовавший
`DefaultVariant` для software без cgo: он возвращает явный fallback.

## Границы приёмки и оставшаяся работа

Настройки рабочего стола не менялись. Реальные OS changes с измерением UI latency
≤500 мс, переключение contrast палитр, text scale 150/200/225%, transparency и
animations через Settings требуют выделенного интерактивного стенда. Синтетические
снимки и сообщения в собственный HWND проверяют путь приложения, но не заменяют
эту приёмку.

Синхронный начальный Windows snapshot реализован. Linux использует существующий
асинхронный XDG provider; стандартный portal не задаёт text scale/permission для
transparency. Собственные macOS/mobile providers, нативные AppKit материалы и
browser runtime ещё не реализованы/не проверены этим блоком.

До завершения ТЗ остаются семантическая Typography/Motion и полный набор widget
states/эталонов, таблица DPI × TextScale для обязательных компонентов, физические
mixed-DPI переносы, Windows 22621/22631 и Intel/AMD, точные desktop environments
GNOME/KDE/Fly, пять performance runs по 60 секунд и 30-минутный stress/resource
gate. См. [LINUX_WINDOW_VISUALS_VALIDATION.md](LINUX_WINDOW_VISUALS_VALIDATION.md).

## Первичные основания API

- [UISettings](https://learn.microsoft.com/en-us/uwp/api/windows.ui.viewmanagement.uisettings),
  [AdvancedEffectsEnabled](https://learn.microsoft.com/en-us/uwp/api/windows.ui.viewmanagement.uisettings.advancedeffectsenabled).
- [AccessibilitySettings](https://learn.microsoft.com/en-us/uwp/api/windows.ui.viewmanagement.accessibilitysettings).
- [Win32 high contrast parameter](https://learn.microsoft.com/en-us/windows/win32/winauto/high-contrast-parameter).
- Windows SDK 10.0.26100.0 `winrt/windows.ui.viewmanagement.h`: ABI/IID optional
  IUISettings4/6 и typed event handlers сверены с установленным SDK.
