# Экосистема Granio

Источник отбора — [44 репозитория fyne-io](https://github.com/fyne-io). В рабочий набор включены 34 из них и четыре библиотеки графического и системного стека.

На 18 октября 2026 года созданы 20 из 38 удалённых репозиториев. Создание оставшихся временно ограничено GitHub. Подготовленные рабочие копии сохраняются локально.

Пути Go-модулей и импорты переведены на Granio. Это исходный перенос: сборка и проверка функций для каждого компонента отмечаются отдельно. Контейнеры, мобильные SDK, сайты и среды разработки требуют собственной приёмки.

| Компонент | Исходник | Назначение | Публикация |
| --- | --- | --- | --- |
| [granio](https://github.com/Uzer007-admin/granio) | [fyne](https://github.com/fyne-io/fyne) | Ядро независимого GUI-фреймворка Granio: системная тема, Wayland/X11 и виртуализированные компоненты. | Репозиторий создан |
| [granio-apps](https://github.com/Uzer007-admin/granio-apps) | [apps](https://github.com/fyne-io/apps) | Каталог приложений Granio и проверка совместимости ядра. | Репозиторий создан |
| [granio-gl-js](https://github.com/Uzer007-admin/granio-gl-js) | [gl-js](https://github.com/fyne-io/gl-js) | OpenGL, GLES и WebGL для графического драйвера Granio. | Репозиторий создан |
| [granio-glfw-js](https://github.com/Uzer007-admin/granio-glfw-js) | [glfw-js](https://github.com/fyne-io/glfw-js) | Браузерные окна и ввод для Granio. | Репозиторий создан |
| [granio-systray](https://github.com/Uzer007-admin/granio-systray) | [systray](https://github.com/fyne-io/systray) | Системный трей и меню приложений Granio. | Репозиторий создан |
| [granio-image](https://github.com/Uzer007-admin/granio-image) | [image](https://github.com/fyne-io/image) | Кодеки изображений для интерфейса и упаковки приложений Granio. | Репозиторий создан |
| [granio-oksvg](https://github.com/Uzer007-admin/granio-oksvg) | [oksvg](https://github.com/fyne-io/oksvg) | Отрисовка SVG для значков и ресурсов Granio. | Репозиторий создан |
| granio-locale | [go-locale](https://github.com/fyne-io/go-locale) | Определение системного языка и локали для Granio. | Подготовлен локально; ожидает создания |
| granio-xgb | [xgb](https://github.com/fyne-io/xgb) | Низкоуровневый протокол X11 для Linux-интеграции Granio. | Подготовлен локально; ожидает создания |
| granio-xgbutil | [xgbutil](https://github.com/fyne-io/xgbutil) | EWMH, ICCCM и вспомогательные средства X11 для Granio. | Подготовлен локально; ожидает создания |
| granio-mobile-bridge | [gomobile-bridge](https://github.com/fyne-io/gomobile-bridge) | Мост мобильных библиотек и приложений Granio. | Подготовлен локально; ожидает создания |
| [granio-x](https://github.com/Uzer007-admin/granio-x) | [fyne-x](https://github.com/fyne-io/fyne-x) | Дополнительные компоненты и интеграции Granio. | Репозиторий создан |
| [granio-tools](https://github.com/Uzer007-admin/granio-tools) | [tools](https://github.com/fyne-io/tools) | Сборка, упаковка и обслуживание приложений Granio. | Репозиторий создан |
| [granio-cross](https://github.com/Uzer007-admin/granio-cross) | [fyne-cross](https://github.com/fyne-io/fyne-cross) | Кросс-компиляция приложений Granio. | Репозиторий создан |
| [granio-cross-images](https://github.com/Uzer007-admin/granio-cross-images) | [fyne-cross-images](https://github.com/fyne-io/fyne-cross-images) | Рецепты контейнеров кросс-компиляции Granio. | Репозиторий создан |
| [granio-setup](https://github.com/Uzer007-admin/granio-setup) | [setup](https://github.com/fyne-io/setup) | Подготовка окружения разработчика Granio. | Репозиторий создан |
| [granio-ui](https://github.com/Uzer007-admin/granio-ui) | [refyne](https://github.com/fyne-io/refyne) | Сериализация и чтение описаний интерфейса Granio. | Репозиторий создан |
| granio-cloud | [cloud](https://github.com/fyne-io/cloud) | Провайдеры синхронизации и облачных данных Granio. | Подготовлен локально; ожидает создания |
| [granio-examples](https://github.com/Uzer007-admin/granio-examples) | [examples](https://github.com/fyne-io/examples) | Рабочие примеры приложений Granio. | Репозиторий создан |
| granio-7guis | [7guis](https://github.com/fyne-io/7guis) | Семь базовых сценариев интерфейса на Granio. | Подготовлен локально; ожидает создания |
| granio-flatpak-demo | [flatpak_demo](https://github.com/fyne-io/flatpak_demo) | Пример упаковки приложения Granio в Flatpak. | Подготовлен локально; ожидает создания |
| granio-tutorials | [tutorials](https://github.com/fyne-io/tutorials) | Адаптация учебных примеров к Granio. | Подготовлен локально; ожидает создания |
| granio-workshops | [workshops](https://github.com/fyne-io/workshops) | Практические занятия и учебные проекты Granio. | Подготовлен локально; ожидает создания |
| [granio-demo](https://github.com/Uzer007-admin/granio-demo) | [demo](https://github.com/fyne-io/demo) | Демонстрационное приложение компонентов Granio. | Репозиторий создан |
| granio-pixeledit | [pixeledit](https://github.com/fyne-io/pixeledit) | Редактор пиксельных изображений на Granio. | Подготовлен локально; ожидает создания |
| granio-terminal | [terminal](https://github.com/fyne-io/terminal) | Графический терминал на Granio. | Подготовлен локально; ожидает создания |
| granio-calculator | [calculator](https://github.com/fyne-io/calculator) | Калькулятор на Granio. | Подготовлен локально; ожидает создания |
| granio-life | [life](https://github.com/fyne-io/life) | Клеточный автомат «Жизнь» на Granio. | Подготовлен локально; ожидает создания |
| granio-solitaire | [solitaire](https://github.com/fyne-io/solitaire) | Пасьянс на Granio. | Подготовлен локально; ожидает создания |
| granio-studio | [defyne](https://github.com/fyne-io/defyne) | Среда разработки приложений Granio. | Подготовлен локально; ожидает создания |
| [granio-docs](https://github.com/Uzer007-admin/granio-docs) | [docs.fyne.io](https://github.com/fyne-io/docs.fyne.io) | Документация независимого проекта Granio. | Репозиторий создан |
| granio-catalog | [apps.fyne.io](https://github.com/fyne-io/apps.fyne.io) | Исходники каталога приложений для экосистемы Granio. | Подготовлен локально; ожидает создания |
| granio-addons | [addons.fyne.io](https://github.com/fyne-io/addons.fyne.io) | Исходники каталога расширений Granio. | Подготовлен локально; ожидает создания |
| granio-web-demo | [demo.fyne.io](https://github.com/fyne-io/demo.fyne.io) | Браузерная демонстрация Granio. | Подготовлен локально; ожидает создания |
| [gl](https://github.com/Uzer007-admin/gl) | [gl](https://github.com/go-gl/gl) | Обёртки OpenGL/GLES, диагностика функций и управление ресурсами Granio. | Репозиторий создан |
| [glfw](https://github.com/Uzer007-admin/glfw) | [glfw](https://github.com/go-gl/glfw) | GLFW 3.5.1, окна, ввод, Wayland/X11 и оконные материалы Granio. | Репозиторий создан |
| [glow](https://github.com/Uzer007-admin/glow) | [glow](https://github.com/go-gl/glow) | Воспроизводимая генерация согласованных обёрток OpenGL/GLES Granio. | Репозиторий создан |
| [dbus](https://github.com/Uzer007-admin/dbus) | [dbus](https://github.com/godbus/dbus) | D-Bus, системные порталы, наблюдение за темой и палитрой KDE для Granio. | Репозиторий создан |

## Локальные копии

Основные библиотеки находятся в `deps/`, каталог приложений — в `extras/apps`. Остальная экосистема сохранена в `ecosystem/work/` отдельными Git-репозиториями. Этот каталог не включается в коммит ядра; изменения каждого компонента сохраняются в его собственной ветке `dev/granio`.

Реестр [repositories.json](../ecosystem/repositories.json) задаёт исходник, цель и путь каждого компонента. Команда `python ecosystem/publish.py` проверяет доступность репозиториев. Флаги `--create`, `--push`, `--set-default` выполняют соответствующий этап; `--only ИМЯ` ограничивает выбор. Скрипт не перезаписывает рабочие файлы, не создаёт коммитов и останавливается при ограничении GitHub.

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
