# WindowVisuals: прозрачность и системное размытие Linux

13 октября 2026. Реализованы переносимые `BackdropTransparent` и
`BackdropSystemBlur` в настоящем Fyne, с X11/Wayland-механизмами GLFW 3.5.
Численные значения прежних Opaque/Mica/Mica Alt/Desktop Acrylic сохранены.
На Windows SystemBlur использует существующий Desktop Acrylic путь, а
Transparent отключает backdrop и сохраняет подготовленную alpha-поверхность.

## Использование

Настроить формат необходимо до первого создания нативного окна:

```go
w := a.NewWindow("Application")
_, err := desktop.ConfigureWindowVisuals(w, desktop.VisualOptions{
    Backdrop: desktop.BackdropSystemBlur, // либо BackdropTransparent
    PrepareAlpha: true,
    Fallback: desktop.VisualFallbackOpaque,
    Frame: desktop.VisualFramePreserve,
})
// Обработать err, установить содержимое, затем вызвать Show.
```

`PrepareAlpha` позволяет создать непрозрачное окно с подходящим форматом,
затем менять режимы без пересоздания. Запрос прозрачности у уже созданного
окна без резервирования возвращает `ErrVisualNeedsRecreate`. Configure безопасен
с любой goroutine; нативные запросы и observers выполняются на UI goroutine.
Прозрачность относится к пикселям фона; ввод, рамка и caption не заменяются.
Непрозрачные controls сохраняют собственный фон и читаемость.

`AlphaSurface` описывает пригодность созданного формата. `TransparentSurface`
дополнительно учитывает доступность композиции. `CompositorAvailable` и маски
`OSBackdrops`/`RendererBackdrops` разделяют состояние композитора и рендерера.
`VisualStatus` сообщает Requested/Selected, причины `CompositorUnavailable`,
`BlurUnavailable`, high contrast и запрет прозрачности. NativeAccepted означает
принятый запрос, а не оптическое подтверждение результата композитора.

При недоступности SystemBlur выбирается полностью непрозрачный фон, alpha=255.
Полупрозрачность без blur можно запросить явно через Transparent. Политика
`VisualFallbackFail` сохраняет прежний контракт: Failed с безопасным Opaque,
когда рендерер работоспособен; отказ самого renderer оставляет окно скрытым.
Linux учитывает уже имеющийся SystemAppearanceProvider и явный HighContrast
темы. Настройки окружения и пользовательский композитор не изменяются.

## Нативные механизмы и производительность

X11 проверяет реальный alpha visual и владельца `_NET_WM_CM_Sn`. Для SystemBlur
требуется объявленное корневое свойство `_KDE_NET_WM_BLUR_BEHIND_REGION`.
Наличие одного interned atom поддержкой не считается. Пустое CARDINAL/32
свойство окна запрашивает полный blur region, поэтому resize не требует
повторной записи геометрии. Смена владельца compositor selection вызывает
повторное применение, включая восстановление после отключения композитора.
Без композитора используется Opaque; alpha visual сохраняется для восстановления.
Контракт KWin проверен по [исходникам 5.27](https://raw.githubusercontent.com/KDE/kwin/Plasma/5.27/src/effects/blur/blur.cpp).

Wayland использует [ext-background-effect-v1](https://gitlab.freedesktop.org/wayland/wayland-protocols/-/blob/main/staging/ext-background-effect/ext-background-effect-v1.xml)
на принадлежащих GLFW display/surface. Стандарт поддерживается
[KWin 6.7](https://raw.githubusercontent.com/KDE/kwin/Plasma/6.7/src/wayland/backgroundeffect_v1.cpp)
и [Mutter 51](https://gnome.pages.gitlab.gnome.org/mutter/meta/method.Compositor.set_background_blur_params.html).
Состояние следует событиям capabilities и удалению/повторному объявлению global,
без определения рабочего стола по имени или версии. Регион задаётся в
surface-local coordinates и обрезается композитором; framebuffer density в
координаты blur не умножается. Это сохраняет покрытие при дробном масштабе.
Поверхность владеет единственным effect object; его освобождение предшествует
уничтожению wl_surface. EGL alpha проверяется по фактическому EGLConfig и
GL_ALPHA_BITS, в том числе когда выбран NATIVE_CONTEXT_API.

Исходный XML сохранён в форке, SHA256
`e463b7863c97d7be05489b52b15cd0b3a5d8290b51340f62023b47d03b4217e3`.
Client header/private code сгенерированы wayland-scanner 1.23.0; лицензии
исходного протокола сохранены. Go wrappers проверяют уничтоженное окно и
GLFW initialization thread. Второго Wayland соединения, нового event loop,
roundtrip при запросе blur, захвата рабочего стола и CPU readback в production
пути нет. Проверка EGLConfig не требует текущего GL-контекста.

Новые режимы используют уже реализованное premultiplied GL/GLES смешивание,
прозрачный clear RGB=0/A=0 и инвалидирование cache материалов при смене clear.
Обычные окна проходят одно atomic чтение. Opt-in policy проверяется не чаще
двух раз в секунду; неизменное состояние не готовит дополнительные кадры.
Wayland query читает состояние уже обработанных событий и локальные EGL
метаданные; X11 query объединён в один cgo переход.

## Первый кадр и переходы

Скрытая подготовка X11/Windows делает один swap и оставляет последнюю полную
отрисовку в back buffer. Финальный лишний swap удалён: драйвер с несколькими
буферами мог после него выдать неинициализированный back buffer. Это также
сокращает количество подготовительных swaps.

GLFW подавляет swap скрытого Wayland окна. До Show выполняется полная подготовка
back buffer; после создания shell role/получения configure первый видимый swap
выполняется после полной отрисовки. PreparedFrames считает подготовленные
отрисовки, а не scanout или реально отправленные скрытые Wayland buffers.
При fallback удаление Wayland blur ставится в pending state перед opaque swap:
удаление эффекта и непрозрачные пиксели применяются одним surface commit.
X11/Windows сначала публикуют непрозрачные кадры, затем отключают эффект.

## Проверки

Стенд f-test-bench: OSnova Linux Onyx3 3.1, kernel 6.12.60, Go 1.27.1,
Xvfb, xcompmgr 1.1.8, headless Weston 10.0.1, Wayland 1.23.0, Mesa software
rendering. xcompmgr распакован только внутри временного каталога; системные
пакеты не установлены и чужие сеансы не затронуты.

| Проверка | Результат |
| --- | --- |
| X11 без композитора | Opaque fallback, actual framebuffer alpha=255 |
| X11 с xcompmgr, GL/GLES | Transparent, premultiplied RGBA, подготовленный скрытый back buffer |
| X11 KWin property fixture | Применение/удаление свойства, отключение/возврат capability и композитора без нового окна |
| Weston без blur protocol, scale=2 | Transparent работает, SystemBlur выбирает Opaque/BlurUnavailable |
| Wayland standard protocol fixture, GL/GLES2 | Запрос на owning surface, смена capabilities, global remove/rebind, возврат прежнего requestID |
| Дробный масштаб | Реальные fractional-scale events 120→150→240→120; framebuffer density 1→1.25→2→1, alpha сохраняется |
| Первый видимый Wayland buffer | Независимое compositor-side чтение; есть A=0, premultiplied RGB, invalid=0 |
| Fallback/high contrast | Непрозрачные пиксели, прежний surface lifetime и нативное окно |
| Неизменные capabilities | PreparedFrames не увеличивается за несколько policy intervals |
| GLFW lifetime/thread preflight | Wrong-thread, opaque-format и destroyed-window ошибки проверены |
| Windows regression | Transparent/SystemBlur/Opaque, прежние Mica/Mica Alt/Acrylic, alpha и optical DWM gates прошли |

Нативные тесты выполняются с race, cgocheck2 и checkptr=2. Alpha swatches
0/64/128/192/255 и RGB проверяются с допуском 1/255. CaptureBack читает уже
подготовленный буфер без перерисовки. Галерея запускалась через обычный Fyne
API в Transparent и SystemBlur. Подробные логи и Windows снимки находятся в
[linux-backgrounds-validation](linux-backgrounds-validation/).

Общие root race suites прошли: 48 Linux и 46 Windows test packages. Windows
native suite включает оптическую регрессию прежних DWM материалов. `go vet`
desktop API, painter, gallery и GLFW 3.5 прошёл. Exported patches всех пяти
локальных форков воспроизведены на чистых закреплённых upstream-базах;
GLFW patch содержит 739 совпадающих файлов. Результаты и SHA256 находятся в
`fork-patch-replay.json`, `source-manifest.json` и `acceptance.json`.

Software/CI compatibility checks (`no_glfw,ci`) прошли для desktop API,
GL painter, canvas и theme; WASM галерея с новыми enum/API скомпилирована.

Первый Windows запуск общего набора в ограниченной песочнице получил отказы
записи тестовых preferences/storage в пользовательском каталоге. Его лог
сохранён отдельно; повторный запуск с доступом для тестовых процессов прошёл.
Это ограничение запуска, а не исправление production preferences/storage.

Воспроизведение на экспортированном workspace с локальным file proxy:

```sh
sh scripts/validate-linux-backgrounds.sh /path/to/owned-work-directory
```

Галерея:

```sh
go run -tags migrated_fynedo ./cmd/fyne_gallery -backdrop transparent
go run -tags migrated_fynedo ./cmd/fyne_gallery -backdrop system-blur
```

## Границы приёмки

Wayland module и X11 root-property fixture проверяют протокол и жизненный цикл,
но не реализуют blur renderer. xcompmgr обеспечивает настоящую alpha-композицию;
нативное размытие именно в GNOME 51, KDE 5.27/6.7 и Fly 1.8.6 на этих стендах
не измерялось. Для них остаётся оптическая приёмка на реальных окружениях/GPU,
проверка настроек эффектов и длительные FPS/present/энергопотребление измерения.

Старый GNOME без этого протокола получает Transparent либо явный SystemBlur
fallback. Fly поддерживается через X11 alpha/compositor capability; blur не
объявляется без соответствующего свойства композитора. Не используются GNOME
Shell extensions и не устанавливается заменяющий compositor. macOS/mobile
backends новых режимов здесь не реализованы; browser runtime не проверен.
Отключение blur/global в живом Wayland соединении обработано. Завершение самого
Wayland compositor разрывает соединение и уничтожает все окна; сохранить тот
же native handle и продолжить показ после такого разрыва невозможно.
