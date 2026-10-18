# KDE X11 / Wayland acceptance

14 октября 2026 пользователь предложил два стенда. DNS jump host devproxy
разрешает `f-test-osnovacustom` в 172.21.9.27 (пользователь указал X11),
`f-autotest-altkde` в 172.21.9.21. SSH через devproxy дал No route to host/timeout;
через f-test-bench оба адреса дали No route to host. Пользователь подтвердил:
стенды пока неактивны. Тесты на KDE не выполнены; версия KDE/GPU не установлена.

После активации использовать существующую пользовательскую desktop-сессию.
Не запускать Xvfb/Weston вместо KWin для этого протокола. Сначала сохранить:

```sh
cat /etc/os-release
uname -a
go version
kwin_x11 --version   # X11
kwin_wayland --version  # Wayland
loginctl list-sessions
kscreen-doctor -o
```

SSH и тестовые процессы должны иметь DISPLAY/XAUTHORITY для X11, либо
WAYLAND_DISPLAY/XDG_RUNTIME_DIR для Wayland, а также session bus той же сессии.
Не печатать полный environment и не менять чужие cookies/settings. Build
dependencies и GPU libraries проверяются до переноса/сборки. Для sources/offline
proxy подходит `scripts/pack-core-validation.py`; предпочтительна сборка на
самом стенде, иначе проверять glibc/shared-library compatibility бинарников.

В текущем workspace доступны короткие native tests:

```sh
FYNE_TEST_TZ_FUNCTIONAL=1 GOEXPERIMENT=cgocheck2 go test -p 1 -race \
  -tags migrated_fynedo -count=1 -v \
  -run '^TestNative(ModernScopeVisibilityAndEscape|VisualHundredLifecycleAndResizeCycles|ModernLiveTextScalePreservesState)$' \
  ./internal/driver/glfw
FYNE_TEST_WINDOW_VISUALS=1 GOEXPERIMENT=cgocheck2 go test -p 1 -race \
  -tags migrated_fynedo -count=1 -v \
  -run '^TestNativeWindowVisualsLinuxPortable$' ./internal/driver/glfw
FYNE_TEST_MATERIAL=1 GOEXPERIMENT=cgocheck2 go test -p 1 -race \
  -tags migrated_fynedo -count=1 -v \
  -run '^TestNativeCanvasAndPopupMaterials$' ./internal/driver/glfw
go run -tags migrated_fynedo ./cmd/fyne_gallery -profile materials -backdrop system-blur
```

Для изолированного Wayland build использовать `-tags migrated_fynedo,wayland`; для отдельного GLES2
варианта — `-tags migrated_fynedo,wayland,gles`, с зафиксированным actual GL/GLES
version. Runtime backend проверять по diagnostics, а не имени compile tag.

## Дополнение: Plasma 6.8

[Официальный анонс от 14 октября 2026](https://kde.org/announcements/plasma/6/6.8.0/)
подтверждает прекращение запуска Plasma X11 sessions. X11-приложения продолжают
работать через XWayland. Поэтому матрица содержит отдельные строки:

| Конфигурация | Проверяемый путь |
|---|---|
| KDE 5.27, X11 | Native X11 и KWin composition/blur hints |
| KDE 6.7, Wayland | Native Wayland: alpha, fractional scale, blur capabilities |
| KDE 6.8, Wayland | Native Wayland: те же проверки, input/IME, first frame, portal |
| KDE 6.8, XWayland | X11 backend в Wayland session: capabilities, scale и fallback |

Native X11 session для Plasma 6.8 не является целевой строкой. Доступность
конкретных протоколов определяется runtime registry/capabilities, а не номером KDE.
Предоставленные стенды пока неактивны; ни одна новая строка не зачтена.

Текущий deps/glfw/v3.5 по умолчанию компилирует оба Linux backend; build tags
`x11` и `wayland` изолируют соответствующий backend. GLFW учитывает
XDG_SESSION_TYPE и доступный display при автоматическом выборе. На реальном KDE
6.8 проверить default build и каждый изолированный вариант, записав фактический
backend из diagnostics. Сборка без X11 libraries проверяется отдельно через
Wayland-only вариант.

В 6.8 KWin добавляет правила для shadows/outlines/rounded corners у CSD windows.
Проверить отсутствие двойной рамки/тени, корректный hit test и resize на
прозрачной поверхности. Такая декорация не доказывает доступность SystemBlur:
его advertised capabilities, alpha и opaque fallback проверяются отдельно.
Настройки compositor/window rules пользователя автоматически не изменять.

Union beta оформляет Qt Quick и Qt Widgets. [Текущие output plugins](https://api.kde.org/applications.html)
не предоставляют автоматического оформления Fyne. Возможный адаптер стилевых
токенов требует самостоятельного проектирования и проверок; наличие Union
не засчитывается как поддержка системной темы в Fyne.

Фиксировать отдельно: alpha surface, compositor/blur capabilities, Selected
backdrop/reason, FirstFrameBeforeShow, controls opacity, premultiplied swatches,
native show/hide/resize/minimize/maximize flags и живой TextScale/input state.
Положить под окно известные светлый/тёмный/checker/photo sources и сделать desktop
capture для оптической прозрачности/blur. Canvas.Capture не содержит эффект KWin.
Отсутствие protocol/effect означает диагностируемый opaque fallback. Изменение
реальных compositor/blur/OS settings требует отдельного согласованного действия;
этот runbook не изменяет их автоматически. Длительные performance-прогоны в
текущий протокол не включены.

Raw results, source hashes, OS/KDE/GPU/session/backend и исключения добавлять к
[VISUAL_MODERNIZATION_COMPLETION.md](VISUAL_MODERNIZATION_COMPLETION.md).
