# Экосистема Granio

После объединения библиотек, инструментов и сайтов рабочий набор содержит **11 репозиториев вместо 28** на предыдущем этапе. Все 11 опубликованы, основная ветка — `dev/granio`. Источники 38 исходных компонентов сохраняются в этих репозиториях.

На этом этапе перенесены 17 компонентов: 14 отдельных репозиториев удалены после проверок, создание ещё трёх отменено. Ранее в общий набор уже были перенесены десять наборов примеров. Код, истории Git и имеющиеся лицензионные файлы сохранены.

[Отчёт об объединении и проверках](REPOSITORY_CONSOLIDATION.md).

| Репозиторий | Основной исходник | Назначение |
| --- | --- | --- |
| [granio](https://github.com/Uzer007-admin/granio) | [fyne](https://github.com/fyne-io/fyne) | Ядро Granio и встроенные модули графического и системного стека: OpenGL/WebGL, окна, трей, SVG и изображения. |
| [granio-apps](https://github.com/Uzer007-admin/granio-apps) | [apps](https://github.com/fyne-io/apps) | Каталог приложений Granio и проверка совместимости ядра. |
| [granio-xgb](https://github.com/Uzer007-admin/granio-xgb) | [xgb](https://github.com/fyne-io/xgb) | Общий Go-модуль протокола X11, EWMH, ICCCM, событий, клавиатуры и вспомогательной графики. |
| [granio-x](https://github.com/Uzer007-admin/granio-x) | [fyne-x](https://github.com/fyne-io/fyne-x) | Общие расширения Granio: виджеты, терминал, описания интерфейсов, облачные провайдеры и локаль. |
| [granio-tools](https://github.com/Uzer007-admin/granio-tools) | [tools](https://github.com/fyne-io/tools) | Инструменты Granio: упаковка, подготовка окружения, Studio, кросс-сборка, образы и мобильный мост. |
| [granio-examples](https://github.com/Uzer007-admin/granio-examples) | [examples](https://github.com/fyne-io/examples) | Общий модуль приложений и примеров Granio: терминал, виджеты, игры, локаль, облако и упаковка. |
| [granio-docs](https://github.com/Uzer007-admin/granio-docs) | [docs.fyne.io](https://github.com/fyne-io/docs.fyne.io) | Документация Granio и исходники каталогов приложений и расширений. |
| [gl](https://github.com/Uzer007-admin/gl) | [gl](https://github.com/go-gl/gl) | Обёртки OpenGL/GLES, диагностика функций и управление ресурсами Granio. |
| [glfw](https://github.com/Uzer007-admin/glfw) | [glfw](https://github.com/go-gl/glfw) | GLFW 3.5.1, окна, ввод, Wayland/X11 и оконные материалы Granio. |
| [glow](https://github.com/Uzer007-admin/glow) | [glow](https://github.com/go-gl/glow) | Воспроизводимая генерация согласованных обёрток OpenGL/GLES Granio. |
| [dbus](https://github.com/Uzer007-admin/dbus) | [dbus](https://github.com/godbus/dbus) | D-Bus, системные порталы, наблюдение за темой и палитрой KDE для Granio. |

## Размещение объединённых компонентов

| Прежнее имя | Новый каталог | Исходник |
| --- | --- | --- |
| `granio-7guis` | [granio-examples/7guis](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/7guis) | [7guis](https://github.com/fyne-io/7guis) |
| `granio-flatpak-demo` | [granio-examples/packaging/flatpak](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/packaging/flatpak) | [flatpak_demo](https://github.com/fyne-io/flatpak_demo) |
| `granio-tutorials` | [granio-examples/tutorials](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/tutorials) | [tutorials](https://github.com/fyne-io/tutorials) |
| `granio-workshops` | [granio-examples/workshops](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/workshops) | [workshops](https://github.com/fyne-io/workshops) |
| `granio-demo` | [granio-examples/demo](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/demo) | [demo](https://github.com/fyne-io/demo) |
| `granio-pixeledit` | [granio-examples/pixeledit](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/pixeledit) | [pixeledit](https://github.com/fyne-io/pixeledit) |
| `granio-calculator` | [granio-examples/calculator](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/calculator) | [calculator](https://github.com/fyne-io/calculator) |
| `granio-life` | [granio-examples/life](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/life) | [life](https://github.com/fyne-io/life) |
| `granio-solitaire` | [granio-examples/solitaire](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/solitaire) | [solitaire](https://github.com/fyne-io/solitaire) |
| `granio-web-demo` | [granio-examples/packaging/web](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/packaging/web) | [demo.fyne.io](https://github.com/fyne-io/demo.fyne.io) |
| `granio-gl-js` | [granio/deps/gl-js](https://github.com/Uzer007-admin/granio/tree/dev/granio/deps/gl-js) | [gl-js](https://github.com/fyne-io/gl-js) |
| `granio-glfw-js` | [granio/deps/glfw-js](https://github.com/Uzer007-admin/granio/tree/dev/granio/deps/glfw-js) | [glfw-js](https://github.com/fyne-io/glfw-js) |
| `granio-systray` | [granio/deps/systray](https://github.com/Uzer007-admin/granio/tree/dev/granio/deps/systray) | [systray](https://github.com/fyne-io/systray) |
| `granio-image` | [granio/deps/image](https://github.com/Uzer007-admin/granio/tree/dev/granio/deps/image) | [image](https://github.com/fyne-io/image) |
| `granio-oksvg` | [granio/deps/oksvg](https://github.com/Uzer007-admin/granio/tree/dev/granio/deps/oksvg) | [oksvg](https://github.com/fyne-io/oksvg) |
| `granio-terminal` | [granio-x/terminal](https://github.com/Uzer007-admin/granio-x/tree/dev/granio/terminal) | [terminal](https://github.com/fyne-io/terminal) |
| `granio-ui` | [granio-x/ui](https://github.com/Uzer007-admin/granio-x/tree/dev/granio/ui) | [refyne](https://github.com/fyne-io/refyne) |
| `granio-cloud` | [granio-x/cloud](https://github.com/Uzer007-admin/granio-x/tree/dev/granio/cloud) | [cloud](https://github.com/fyne-io/cloud) |
| `granio-locale` | [granio-x/locale](https://github.com/Uzer007-admin/granio-x/tree/dev/granio/locale) | [go-locale](https://github.com/fyne-io/go-locale) |
| `granio-setup` | [granio-tools/setup](https://github.com/Uzer007-admin/granio-tools/tree/dev/granio/setup) | [setup](https://github.com/fyne-io/setup) |
| `granio-cross` | [granio-tools/cross](https://github.com/Uzer007-admin/granio-tools/tree/dev/granio/cross) | [fyne-cross](https://github.com/fyne-io/fyne-cross) |
| `granio-cross-images` | [granio-tools/cross-images](https://github.com/Uzer007-admin/granio-tools/tree/dev/granio/cross-images) | [fyne-cross-images](https://github.com/fyne-io/fyne-cross-images) |
| `granio-mobile-bridge` | [granio-tools/mobile-bridge](https://github.com/Uzer007-admin/granio-tools/tree/dev/granio/mobile-bridge) | [gomobile-bridge](https://github.com/fyne-io/gomobile-bridge) |
| `granio-studio` | [granio-tools/studio](https://github.com/Uzer007-admin/granio-tools/tree/dev/granio/studio) | [defyne](https://github.com/fyne-io/defyne) |
| `granio-xgbutil` | [granio-xgb/xgbutil](https://github.com/Uzer007-admin/granio-xgb/tree/dev/granio/xgbutil) | [xgbutil](https://github.com/fyne-io/xgbutil) |
| `granio-catalog` | [granio-docs/sites/catalog](https://github.com/Uzer007-admin/granio-docs/tree/dev/granio/sites/catalog) | [apps.fyne.io](https://github.com/fyne-io/apps.fyne.io) |
| `granio-addons` | [granio-docs/sites/addons](https://github.com/Uzer007-admin/granio-docs/tree/dev/granio/sites/addons) | [addons.fyne.io](https://github.com/fyne-io/addons.fyne.io) |

Библиотека терминала находится в `granio-x/terminal`, приложение — в `granio-examples/terminal`. Studio использует этот же терминальный виджет и `granio-x/ui`. Кросс-сборка и мобильный мост сохраняют отдельные Go-модули внутри `granio-tools`. Пять модулей `granio/deps` хранятся обычными исходниками вместе с ядром.

## Сопровождение

[repositories.json](../ecosystem/repositories.json) разделяет активные репозитории и объединённые источники. Удалённые репозитории не архивируются и не создаются повторно. `python ecosystem/publish.py --metadata` обновляет описания и ссылки на исходники; `--only ИМЯ` ограничивает набор. `--create`, `--push`, `--set-default` применяются только к активным репозиториям.

[modules.json](../ecosystem/modules.json) содержит проверенные опубликованные версии. Полный локальный workspace подключает 22 модуля. Независимые установки CLI и Studio проверены при `GOWORK=off`. Старые локальные копии источников сохраняются для истории, в активный workspace они не включены.

Аппаратная приёмка, публикация сайтов и образов, мобильные SDK и длительные измерения производительности не входят в этот этап.

## Что исключено

- [fyne-io.github.io](https://github.com/fyne-io/fyne-io.github.io): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [tour.fyne.io](https://github.com/fyne-io/tour.fyne.io): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [conf.fyne.io](https://github.com/fyne-io/conf.fyne.io): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [mobile](https://github.com/fyne-io/mobile): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [developer.fyne.io](https://github.com/fyne-io/developer.fyne.io): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [chat.fyne.io](https://github.com/fyne-io/chat.fyne.io): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [proposals](https://github.com/fyne-io/proposals): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [fin](https://github.com/fyne-io/fin): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [go-image](https://github.com/fyne-io/go-image): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
- [glfw](https://github.com/fyne-io/glfw): Сайты исходного проекта, архивный компонент или дублирующая зависимость; в рабочий набор не включён.
