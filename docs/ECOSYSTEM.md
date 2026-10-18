# Экосистема Granio

Источник отбора — [fyne-io](https://github.com/fyne-io). Исходный набор включает 34 компонента этой организации и четыре библиотеки графического и системного стека. После объединения примеров он сопровождается в **28 активных репозиториях вместо 38**.

Демонстрации, учебные проекты и небольшие приложения собраны в [granio-examples](https://github.com/Uzer007-admin/granio-examples) с одним Go-модулем. История девяти наборов, лицензии и авторские сведения сохранены; веб-пример представлен рецептом воспроизводимой сборки. Подробности — [в отчёте об объединении](EXAMPLES_CONSOLIDATION.md).

Пути Go-модулей и импорты переведены на Granio. Результаты проверки функций описаны в [отчёте развития](GRANIO_DEVELOPMENT.md). Контейнеры, мобильные SDK, сайты и среды разработки требуют собственной приёмки.

| Компонент | Исходник | Назначение | Публикация |
| --- | --- | --- | --- |
| [granio](https://github.com/Uzer007-admin/granio) | [fyne](https://github.com/fyne-io/fyne) | Ядро независимого GUI-фреймворка Granio: системная тема, Wayland/X11 и виртуализированные компоненты. | Ветка `dev/granio` опубликована |
| [granio-apps](https://github.com/Uzer007-admin/granio-apps) | [apps](https://github.com/fyne-io/apps) | Каталог приложений Granio и проверка совместимости ядра. | Ветка `dev/granio` опубликована |
| [granio-gl-js](https://github.com/Uzer007-admin/granio-gl-js) | [gl-js](https://github.com/fyne-io/gl-js) | OpenGL, GLES и WebGL для графического драйвера Granio. | Ветка `dev/granio` опубликована |
| [granio-glfw-js](https://github.com/Uzer007-admin/granio-glfw-js) | [glfw-js](https://github.com/fyne-io/glfw-js) | Браузерные окна и ввод для Granio. | Ветка `dev/granio` опубликована |
| [granio-systray](https://github.com/Uzer007-admin/granio-systray) | [systray](https://github.com/fyne-io/systray) | Системный трей и меню приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-image](https://github.com/Uzer007-admin/granio-image) | [image](https://github.com/fyne-io/image) | Кодеки изображений для интерфейса и упаковки приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-oksvg](https://github.com/Uzer007-admin/granio-oksvg) | [oksvg](https://github.com/fyne-io/oksvg) | Отрисовка SVG для значков и ресурсов Granio. | Ветка `dev/granio` опубликована |
| [granio-locale](https://github.com/Uzer007-admin/granio-locale) | [go-locale](https://github.com/fyne-io/go-locale) | Определение системного языка и локали для Granio. | Ветка `dev/granio` опубликована |
| [granio-xgb](https://github.com/Uzer007-admin/granio-xgb) | [xgb](https://github.com/fyne-io/xgb) | Низкоуровневый протокол X11 для Linux-интеграции Granio. | Ветка `dev/granio` опубликована |
| [granio-xgbutil](https://github.com/Uzer007-admin/granio-xgbutil) | [xgbutil](https://github.com/fyne-io/xgbutil) | EWMH, ICCCM и вспомогательные средства X11 для Granio. | Ветка `dev/granio` опубликована |
| [granio-mobile-bridge](https://github.com/Uzer007-admin/granio-mobile-bridge) | [gomobile-bridge](https://github.com/fyne-io/gomobile-bridge) | Мост мобильных библиотек и приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-x](https://github.com/Uzer007-admin/granio-x) | [fyne-x](https://github.com/fyne-io/fyne-x) | Дополнительные компоненты и интеграции Granio. | Ветка `dev/granio` опубликована |
| [granio-tools](https://github.com/Uzer007-admin/granio-tools) | [tools](https://github.com/fyne-io/tools) | Сборка, упаковка и обслуживание приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-cross](https://github.com/Uzer007-admin/granio-cross) | [fyne-cross](https://github.com/fyne-io/fyne-cross) | Кросс-компиляция приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-cross-images](https://github.com/Uzer007-admin/granio-cross-images) | [fyne-cross-images](https://github.com/fyne-io/fyne-cross-images) | Рецепты контейнеров кросс-компиляции Granio. | Ветка `dev/granio` опубликована |
| [granio-setup](https://github.com/Uzer007-admin/granio-setup) | [setup](https://github.com/fyne-io/setup) | Подготовка окружения разработчика Granio. | Ветка `dev/granio` опубликована |
| [granio-ui](https://github.com/Uzer007-admin/granio-ui) | [refyne](https://github.com/fyne-io/refyne) | Сериализация и чтение описаний интерфейса Granio. | Ветка `dev/granio` опубликована |
| [granio-cloud](https://github.com/Uzer007-admin/granio-cloud) | [cloud](https://github.com/fyne-io/cloud) | Провайдеры синхронизации и облачных данных Granio. | Ветка `dev/granio` опубликована |
| [granio-examples](https://github.com/Uzer007-admin/granio-examples) | [examples](https://github.com/fyne-io/examples) | Общий модуль демонстраций, учебных проектов, игр и примеров упаковки Granio. | Ветка `dev/granio` опубликована |
| [granio-terminal](https://github.com/Uzer007-admin/granio-terminal) | [terminal](https://github.com/fyne-io/terminal) | Графический терминал на Granio. | Ветка `dev/granio` опубликована |
| granio-studio | [defyne](https://github.com/fyne-io/defyne) | Среда разработки приложений Granio. | Подготовлен локально; создание ограничено GitHub |
| [granio-docs](https://github.com/Uzer007-admin/granio-docs) | [docs.fyne.io](https://github.com/fyne-io/docs.fyne.io) | Документация независимого проекта Granio. | Ветка `dev/granio` опубликована |
| granio-catalog | [apps.fyne.io](https://github.com/fyne-io/apps.fyne.io) | Исходники каталога приложений для экосистемы Granio. | Подготовлен локально; создание ограничено GitHub |
| granio-addons | [addons.fyne.io](https://github.com/fyne-io/addons.fyne.io) | Исходники каталога расширений Granio. | Подготовлен локально; создание ограничено GitHub |
| [gl](https://github.com/Uzer007-admin/gl) | [gl](https://github.com/go-gl/gl) | Обёртки OpenGL/GLES, диагностика функций и управление ресурсами Granio. | Ветка `dev/granio` опубликована |
| [glfw](https://github.com/Uzer007-admin/glfw) | [glfw](https://github.com/go-gl/glfw) | GLFW 3.5.1, окна, ввод, Wayland/X11 и оконные материалы Granio. | Ветка `dev/granio` опубликована |
| [glow](https://github.com/Uzer007-admin/glow) | [glow](https://github.com/go-gl/glow) | Воспроизводимая генерация согласованных обёрток OpenGL/GLES Granio. | Ветка `dev/granio` опубликована |
| [dbus](https://github.com/Uzer007-admin/dbus) | [dbus](https://github.com/godbus/dbus) | D-Bus, системные порталы, наблюдение за темой и палитрой KDE для Granio. | Ветка `dev/granio` опубликована |

## Примеры в общем репозитории

| Прежнее имя | Каталог в granio-examples | Исходник |
| --- | --- | --- |
| `granio-7guis` | [7guis](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/7guis) | [7guis](https://github.com/fyne-io/7guis) |
| `granio-flatpak-demo` | [packaging/flatpak](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/packaging/flatpak) | [flatpak_demo](https://github.com/fyne-io/flatpak_demo) |
| `granio-tutorials` | [tutorials](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/tutorials) | [tutorials](https://github.com/fyne-io/tutorials) |
| `granio-workshops` | [workshops](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/workshops) | [workshops](https://github.com/fyne-io/workshops) |
| `granio-demo` | [demo](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/demo) | [demo](https://github.com/fyne-io/demo) |
| `granio-pixeledit` | [pixeledit](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/pixeledit) | [pixeledit](https://github.com/fyne-io/pixeledit) |
| `granio-calculator` | [calculator](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/calculator) | [calculator](https://github.com/fyne-io/calculator) |
| `granio-life` | [life](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/life) | [life](https://github.com/fyne-io/life) |
| `granio-solitaire` | [solitaire](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/solitaire) | [solitaire](https://github.com/fyne-io/solitaire) |
| `granio-web-demo` | [packaging/web](https://github.com/Uzer007-admin/granio-examples/tree/dev/granio/packaging/web) | [demo.fyne.io](https://github.com/fyne-io/demo.fyne.io) |

Отдельные репозитории для этих десяти наборов исключены из плана создания. Пять ранее опубликованных форков удалены после проверки переноса кода и истории в общий репозиторий. Ещё пять отдельных форков создавать не требуется. Подтверждения удаления сохранены в [отчёте](examples-consolidation/deleted-repositories.json).

Библиотека `granio-terminal` остаётся самостоятельной: её компоненты использует среда разработки. Галерея ядра находится в `cmd/granio_gallery` и отдельного репозитория не требует.

## Локальные копии

Основные библиотеки находятся в `deps/`, каталог приложений — в `extras/apps`, остальные рабочие копии — в `ecosystem/work/`. Каталог `work` не включается в коммит ядра. Прежние локальные копии примеров сохранены для истории; активная разработка ведётся в `granio-examples`.

Реестр [repositories.json](../ecosystem/repositories.json) разделяет активные `repositories` и объединённые `consolidated_sources`. Команда `python ecosystem/publish.py` обрабатывает только активный набор и не создаёт отдельные репозитории примеров. Флаги `--create`, `--push`, `--set-default` управляют публикацией; `--only ИМЯ` ограничивает выбор.

Локальный `ecosystem/work/go.work` подключает общий модуль примеров. Для независимой сборки задаётся `GOWORK=off`; опубликованные версии указаны в [modules.json](../ecosystem/modules.json).

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
