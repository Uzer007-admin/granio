# Экосистема Granio

Источник отбора — [44 репозитория fyne-io](https://github.com/fyne-io). В рабочий набор включены 34 из них и четыре библиотеки графического и системного стека.

На 18 октября 2026 года созданы 30 из 38 репозиториев; в каждом опубликована основная ветка `dev/granio`. Создание оставшихся восьми временно ограничено GitHub. Их адаптированные исходники сохранены в локальных коммитах.

Пути Go-модулей и импорты переведены на Granio. Сборка и проверка функций отмечены отдельно в [отчёте](GRANIO_DEVELOPMENT.md). Контейнеры, мобильные SDK, сайты и среды разработки требуют собственной приёмки.

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
| [granio-examples](https://github.com/Uzer007-admin/granio-examples) | [examples](https://github.com/fyne-io/examples) | Рабочие примеры приложений Granio. | Ветка `dev/granio` опубликована |
| [granio-7guis](https://github.com/Uzer007-admin/granio-7guis) | [7guis](https://github.com/fyne-io/7guis) | Семь базовых сценариев интерфейса на Granio. | Ветка `dev/granio` опубликована |
| [granio-flatpak-demo](https://github.com/Uzer007-admin/granio-flatpak-demo) | [flatpak_demo](https://github.com/fyne-io/flatpak_demo) | Пример упаковки приложения Granio в Flatpak. | Ветка `dev/granio` опубликована |
| [granio-tutorials](https://github.com/Uzer007-admin/granio-tutorials) | [tutorials](https://github.com/fyne-io/tutorials) | Адаптация учебных примеров к Granio. | Ветка `dev/granio` опубликована |
| granio-workshops | [workshops](https://github.com/fyne-io/workshops) | Практические занятия и учебные проекты Granio. | Коммит сохранён локально; создание ограничено GitHub |
| [granio-demo](https://github.com/Uzer007-admin/granio-demo) | [demo](https://github.com/fyne-io/demo) | Демонстрационное приложение компонентов Granio. | Ветка `dev/granio` опубликована |
| granio-pixeledit | [pixeledit](https://github.com/fyne-io/pixeledit) | Редактор пиксельных изображений на Granio. | Коммит сохранён локально; создание ограничено GitHub |
| [granio-terminal](https://github.com/Uzer007-admin/granio-terminal) | [terminal](https://github.com/fyne-io/terminal) | Графический терминал на Granio. | Ветка `dev/granio` опубликована |
| [granio-calculator](https://github.com/Uzer007-admin/granio-calculator) | [calculator](https://github.com/fyne-io/calculator) | Калькулятор на Granio. | Ветка `dev/granio` опубликована |
| granio-life | [life](https://github.com/fyne-io/life) | Клеточный автомат «Жизнь» на Granio. | Коммит сохранён локально; создание ограничено GitHub |
| granio-solitaire | [solitaire](https://github.com/fyne-io/solitaire) | Пасьянс на Granio. | Коммит сохранён локально; создание ограничено GitHub |
| granio-studio | [defyne](https://github.com/fyne-io/defyne) | Среда разработки приложений Granio. | Коммит сохранён локально; создание ограничено GitHub |
| [granio-docs](https://github.com/Uzer007-admin/granio-docs) | [docs.fyne.io](https://github.com/fyne-io/docs.fyne.io) | Документация независимого проекта Granio. | Ветка `dev/granio` опубликована |
| granio-catalog | [apps.fyne.io](https://github.com/fyne-io/apps.fyne.io) | Исходники каталога приложений для экосистемы Granio. | Коммит сохранён локально; создание ограничено GitHub |
| granio-addons | [addons.fyne.io](https://github.com/fyne-io/addons.fyne.io) | Исходники каталога расширений Granio. | Коммит сохранён локально; создание ограничено GitHub |
| granio-web-demo | [demo.fyne.io](https://github.com/fyne-io/demo.fyne.io) | Браузерная демонстрация Granio. | Коммит сохранён локально; создание ограничено GitHub |
| [gl](https://github.com/Uzer007-admin/gl) | [gl](https://github.com/go-gl/gl) | Обёртки OpenGL/GLES, диагностика функций и управление ресурсами Granio. | Ветка `dev/granio` опубликована |
| [glfw](https://github.com/Uzer007-admin/glfw) | [glfw](https://github.com/go-gl/glfw) | GLFW 3.5.1, окна, ввод, Wayland/X11 и оконные материалы Granio. | Ветка `dev/granio` опубликована |
| [glow](https://github.com/Uzer007-admin/glow) | [glow](https://github.com/go-gl/glow) | Воспроизводимая генерация согласованных обёрток OpenGL/GLES Granio. | Ветка `dev/granio` опубликована |
| [dbus](https://github.com/Uzer007-admin/dbus) | [dbus](https://github.com/godbus/dbus) | D-Bus, системные порталы, наблюдение за темой и палитрой KDE для Granio. | Ветка `dev/granio` опубликована |

## Локальные копии

Основные библиотеки находятся в `deps/`, каталог приложений — в `extras/apps`. Остальная экосистема сохранена в `ecosystem/work/` отдельными Git-репозиториями. Этот каталог не включается в коммит ядра; изменения каждого компонента сохраняются в его собственной ветке `dev/granio`.

Реестр [repositories.json](../ecosystem/repositories.json) задаёт исходник, цель и путь каждого компонента. Команда `python ecosystem/publish.py` проверяет доступность репозиториев. Флаги `--create`, `--push`, `--set-default` можно объединить для последовательного создания, публикации и выбора основной ветки; `--only ИМЯ` ограничивает выбор. Скрипт не перезаписывает рабочие файлы, не создаёт коммитов и останавливается при ограничении GitHub.

При разработке полного набора используется локальный `ecosystem/work/go.work`. Для независимой установки приложений нужны согласованные опубликованные версии; проверка без workspace фиксируется в отчёте.

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
