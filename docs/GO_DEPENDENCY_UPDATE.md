# Go 1.27.1 и обновление зависимостей

Текущее состояние: теперь подключены локальные D-Bus 5.3.0 и GLFW 3.5.1, все 11 go.mod используют Go 1.27.1. Исходное исключение dbus снято через fork; полные patches: gl-fork.patch, glfw-fork.patch, glow-fork.patch, dbus-fork.patch и apps-fork.patch. [Итоговая реализация](NATIVE_STACK_IMPLEMENTATION.md). Ниже сохранено описание первого этапа обновления.

Дата: 12 октября 2026 года. Обновлены рабочий Fyne, apps и локальные GL/GLFW/Glow. Нативная C GLFW остаётся на 3.4 по выбору пользователя; переход 3.5.1 оценивается в [GRAPHICS_STACK_ASSESSMENT.md](GRAPHICS_STACK_ASSESSMENT.md).

## Сделанные изменения

- `go 1.27.1`: основной модуль, apps, GL, Glow, корневой GLFW, GLFW 3.3, GLFW 3.4, два CLI module fixtures — всего девять `go.mod`. Та же версия в `go.work`; старый `toolchain go1.24.4` apps удалён.
- Workspace связывает Fyne/apps/GL/Glow и оба вложенных GLFW-модуля. Fyne по-прежнему импортирует 3.4; 3.3 используется также инструментальным графом x/mobile/shiny.
- Прямые, транзитивные, тестовые и инструментальные версии подняты до latest на существующих module paths. Итоговый workspace содержит 83 модуля. Проверка каждого самостоятельного Fyne/apps/Glow graph и workspace оставила одно исключение, описанное ниже.
- `go.sum` и `go.work.sum` обновлены. Некоторые indirect `require` намеренно фиксируют актуальные версии полного upstream test/tool graph. Автоматический `go mod tidy` может удалить неиспользуемые самим модулем ограничения и снова выбрать старые upstream требования; это необходимо учитывать при следующем обновлении.
- GitHub Actions проверены через upstream release API и обновлены: checkout v7.0.1, setup-go v7.0.0, upload-artifact v7.0.2, golangci-lint-action v9.3.0, actions-goveralls v1.11.2. В старом GL CI вместо Go 1.12/1.17 задан Go 1.27.1 плюс stable. Остальные CI берут минимальную версию из go.mod и/или дополнительно проверяют stable.
- CI получает recursive submodules и применяет локальные patches до Go-команд. Formatting ограничен tracked Go-файлами основного репозитория, чтобы не переписывать generated sources сторонних модулей.
- Исправлена одна совместимость с современным `go vet`: `%q` для UTF-16 `[]uint16` в binres диагностике теперь получает `string(utf16.Decode(data))`. Декодирование бинарного формата не менялось.

Основные изменения версий:

| Зависимость | Было | Стало |
| --- | --- | --- |
| Glow: golang.org/x/tools | pseudo-version 2019 года | v0.52.0 |
| Fyne: golang.org/x/tools | v0.30.0 | v0.52.0 |
| golang.org/x/image | v0.24.0 | v0.47.0 |
| golang.org/x/text | v0.22.0 | v0.43.0 |
| golang.org/x/sys | v0.30.0 | v0.49.0 |
| golang.org/x/mod | v0.23.0 | v0.42.0 |
| golang.org/x/net | v0.35.0 | v0.61.0 |
| golang.org/x/sync | v0.11.0 | v0.24.0 |
| fsnotify | v1.9.0 | v1.10.1 |
| bild | v0.14.0 | v0.17.1 |
| go-runewidth | v0.0.24 | v0.0.31 |
| go-i18n/v2 | v2.5.1 | v2.6.1 |
| testify | v1.11.1 | v1.12.1 |
| goldmark | v1.8.2 | v1.8.6 |
| apps: fyne.io/fyne/v2 | v2.8.0 | v2.8.1; workspace использует локальный Fyne |
| apps: go-gl/glfw 3.4 | pre.1 pseudo-version | v0.1.0-pre.2 |

GL/GLFW/Glow upstream heads совпали с закреплёнными commits; их исходники runtime не обновлялись автоматически с других веток. Новые major import paths, например urfave/cli/v3, не вводились без отдельной API миграции.

## Единственное исключение latest

`github.com/godbus/dbus/v5 v5.2.3` опубликована, но Windows файл `conn_windows.go:10` возвращает одно значение из функции с двумя результатами. Обычная сборка Fyne завершилась:

```text
not enough return values
have (string)
want (string, error)
```

Запрос `@master` разрешился в ту же v5.2.3 / commit `4ad75260bdbdd09d51af62b316a17dd60c6aa0e3`. Для рабочего cross-platform проекта оставлена **v5.2.2**. Системный Go cache не патчился, отдельный сторонний fork dbus не добавлялся. После исправленного upstream релиза повторить Windows build и source/race tests перед снятием ограничения. Итоговые latest queries: [module-updates.json](graphics-assessment/module-updates.json).

## Воспроизводимость подмодулей

Обновления вложенных модулей сохраняются вместе с основным репозиторием:

| Подмодуль | Patch |
| --- | --- |
| deps/gl | patches/go127-gl.patch |
| deps/glfw | patches/go127-glfw.patch |
| deps/glow | patches/go127-glow.patch |
| extras/apps | patches/go127-apps.patch; прежний apps-tests.patch |

`scripts/build.ps1` и `scripts/apply-patches.py` проверяют reverse apply: уже применённые patches пропускаются, конфликтующие локальные изменения не сбрасываются. Проверены чистые временные копии всех четырёх исходных HEAD: forward apply, совпадение module/CI файлов с рабочими и reverse check прошли. Результат: [patch-replay.json](graphics-assessment/patch-replay.json). Commits/remote publication не выполнялись.

После `git submodule update --init --recursive` для Linux/CI:

```sh
python scripts/apply-patches.py
go test -race ./deps/glow/... ./deps/glfw/v3.3/glfw/... ./deps/glfw/v3.4/glfw/...
```

На Windows обычный `scripts/build.ps1` применяет те же patches через Git и не требует Python для сборки.

## Фактическая проверка

| Проверка | Результат |
| --- | --- |
| Windows build fyne_settings.exe/apps.exe | Успешно на окончательном Go-графе |
| Windows Glow tests, race | Pass |
| Windows GLFW 3.3/3.4 packages, race flag | Компилируются; upstream unit tests отсутствуют |
| Windows hidden OpenGL smoke | Pass: framebuffer `[0,255,0,255]`, synchronous debug callback вызван |
| GL 2.1 / GLES 3.1 regeneration | Все восемь Go-файлов совпадают после одинаковых LF/gofmt |
| GLES 2.0 generation | Успешно; полноценный GLES 2.0 runtime здесь не испытан |
| EGL 1.5 generation / Linux compilation | Генерация успешна; compile probe выявил существующие ошибки struct types |
| Fresh GLES 3.0 vendor consumer | KHR header потерян при vendor; Windows build прошёл благодаря доступным внешним headers — это не доказательство самодостаточного vendor tree |
| f-test-bench Linux Glow tests, race | Pass |
| f-test-bench Linux GLFW 3.3/3.4 packages | X11/Wayland default configuration компилируется; runtime desktop/IME не испытан |
| Windows весь Fyne/apps native + race | 46 пакетов с тестами проходят; 4 TestTable теста текущей правки Fyne не проходят |
| Linux весь Fyne/apps CI + race | На переданном snapshot: 46 пакетов проходят, 8 тестов в dialog/internal-app/widget не проходят; build errors отсутствуют |

В рабочем дереве параллельно менялся Fyne `internal/cache/widget.go`. Первый полный Windows прогон выявил восемь runtime test failures и отдельную vet-ошибку binres. Vet исправлен в этой работе. Все восемь выбранных тестов прошли с новыми зависимостями и временным overlay прежнего единственного cache-файла; рабочий cache-файл не заменялся. Позднейший окончательный Windows прогон оставил только четыре TestTable failures — часть параллельных изменений Fyne к тому времени была исправлена. Linux snapshot сохраняет более раннее состояние cache-правки. Общий набор не объявляется полностью зелёным.

Оставшиеся Windows tests: `TestTable_Hovered`, `TestTable_JustHeaders`, `TestTable_Selection`, `TestTable_Select`. В markup эталонах внутренние containers имеют `0x0`, после восстановления Layout фактически получают размеры. Это отдельная согласованность cache/layout и golden XML текущего Fyne; эталонные PNG/XML в этой работе не подменялись.

Журналы: [Windows tests](graphics-assessment/windows-tests.json), [Windows build](graphics-assessment/windows-build.log), [Linux tests](graphics-assessment/linux-tests.json), [Linux stack](graphics-assessment/linux-stack-tests.log), [сравнение cache overlay](graphics-assessment/dependency-regression-comparison.log). Негативные candidate probes относятся к оценке дефектов и не входят в pass-status модулей.

Linux проверялся на Go 1.27.1 в собственном `/tmp/fyne-graphics.*` с распакованными OpenGL/X11/Wayland development/runtime packages и публичным module proxy. Установленные системные пакеты не менялись. Перед передачей состав архивов проверен: Git/SSH metadata, home files, `.env` и credentials не включались. Временные копии, generated outputs, SDK и архивы удаляются по завершении; доказательства сохраняются в `docs/graphics-assessment`.

Полные исходные module declarations: [modules-before.json](graphics-assessment/modules-before.json); declared version changes: [dependency-changes.json](graphics-assessment/dependency-changes.json); итоговый build list: [dependencies-after.txt](graphics-assessment/dependencies-after.txt).
