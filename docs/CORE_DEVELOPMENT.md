# Развитие ядра Fyne

Проверено 12 октября 2026 года. Рабочая база: Fyne `7f99a87b`, apps
`c977ab9a`; Go 1.27.1, локальные GLFW 3.5.1 и D-Bus 5.3.0. Изменения библиотек
из параллельной задачи сохранены. Коммиты и внешняя публикация не выполнялись.

## Исправления

- TIFF: `golang.org/x/image v0.47.0` содержит исправление
  [GO-2026-5032](https://pkg.go.dev/vuln/GO-2026-5032), впервые выпущенное в
  v0.41.0. Регрессионные тесты проверяют реальную регистрацию декодера, Canvas,
  оба порядка байтов, вредоносные входы и допустимые изображения на границе
  размера распакованного блока. На v0.24.0 воспроизведение падает, на v0.47.0
  проходит с race detector. Отчёт и свидетельства проверки сохранены отдельно
  от исходников.
- Исправлены потеря Control при повторе клавиши
  [#6486](https://github.com/fyne-io/fyne/issues/6486), преждевременное удаление
  недавно использованного кэша после простоя
  [#6509](https://github.com/fyne-io/fyne/issues/6509), восстановление layout
  уничтоженного renderer и загрузка RGBA с неплотным stride.
- Binding Tree больше не добавляет root собственным ребёнком. Обновление
  иерархии использует линейный индекс родителей вместо повторного поиска
  по всей структуре; это относится к маршруту больших деревьев из
  [#5814](https://github.com/fyne-io/fyne/issues/5814).
- Исправлено штатное завершение приложения: lifecycle callbacks исполняются
  на UI-потоке до закрытия его очереди. Очередь отвергает поздние события,
  завершает принятые callbacks и освобождает ожидающих отправителей.
- apps показывает окно до сетевой загрузки. Каталог сохраняется в настоящий
  пользовательский кэш после проверки JSON и восстанавливается при сетевой
  ошибке. HTTP status, timeout, размер ответа и отмена проверяются. Старый
  результат картинки не перезаписывает новую карточку; установка использует
  приложение, выбранное в момент нажатия кнопки.

Полный исходный реестр открытых проблем остаётся в [ISSUES_AUDIT.md](ISSUES_AUDIT.md)
и `issue-audit/open-issues.json`: 655 Fyne issues и 3 apps issues. Этот реестр
не означает, что все жалобы воспроизведены или исправлены.

## Производительность и API

`desktop.FrameRateDriver.SetFrameRate(0)` выбирает частоту автоматически, минимум
60 Гц; значения 1–1000 задают предел явно. Планирование привязано к следующему
кадру, пропущенные периоды не вызывают серию кадров. Native wait пробуждается
из очереди задач, обновления Canvas и анимации; при простое ожидает до 250 мс.
На Windows используется high-resolution waitable timer с fallback для старых ОС.
Проверки 120/144 Гц измеряют cadence анимации и событий: это не измерение FPS
физического монитора или задержки GPU/vsync.

List/Table используют индекс геометрии с префиксными суммами изменений размеров;
Tree индексирует открытые строки и ищет видимый диапазон. Повторный scroll не
обходит все 20 тысяч строк дерева; separator pool ограничен viewport. TextGrid
не пересоздаёт строки неизменившихся рун, перенос текста использует уже
измеренные фрагменты.

Пакет `cache` предоставляет `GetLimits`, `SetLimits`, `GetUsage`, `Clear`.
По умолчанию текст: 8192 записи / 8 МиБ; SVG: 256 записей / 64 МиБ. При изменении
лимитов лишние записи удаляются сразу по LRU. Нулевой лимит отключает кэш;
отрицательный возвращает ошибку. Учёт bytes включает удерживаемые данные, но
не allocator/map overhead; число записей ограничивает и этот overhead.
Живые изображения, renderers и GPU textures имеют отдельное время жизни.

`canvas.LoadImageWithOptions` декодирует в фоне, ограничивает encoded bytes
(64 МиБ по умолчанию) и pixels (64 миллиона), поддерживает context cancellation.
`NewImageFromURIAsync` применяет результат через UI-поток.

Добавлены `SelectionMultiple`, `Selected`, `SetSelection`, `SelectRange`,
`SelectAll` для List/Table/Tree; поддерживаются Control/Shift и клавиатура.
Маркеры Table создаются только для видимых выделенных ячеек. `NewTreeTable`
виртуализирует строки с колонками; первый столбец содержит отступы и expander.
`binding.AsyncList[T]`, `AsyncTree[T]` загружают snapshots с отменой предыдущего
запроса; устаревший ответ не заменяет свежий. `Wait` следует вызывать вне
UI-потока. `NewTableWithData` связывает Table с двухмерными данными.

## Windows и Linux: нативные интерфейсы

Entry различает предварительный текст IME и подтверждение. Preedit не изменяет
Text, binding, validation и Undo; подтверждение или замена surrounding text
создают одну операцию Undo. RichTextEntry сохраняет изменения стилей в своём
Undo. Native caret offsets переводятся из UTF-16/UTF-8 в rune offsets; отмена
и смена фокуса восстанавливают исходное выделение.

Windows использует IMM и positioning кандидатов; Linux — XIM callbacks и
Wayland text-input-v3, когда compositor его предоставляет. Password text не
передаётся в surrounding text; Windows настраивает password input scope.

При сборке с `-tags=accessibility` доступны UI Automation на Windows и AT-SPI
на Linux. Общая модель содержит роли, значения, состояния и действия видимых
элементов. UIA реализует Invoke/Value/Toggle/SelectionItem/ExpandCollapse,
поддерживает несколько окон и корректное COM-время жизни удалённых объектов.
AT-SPI работает через отдельный accessibility bus, передаёт иерархию, текст,
выделение и действия; изменения контролов исполняются на UI-потоке.

`fyne.ClipboardData` расширяет Clipboard методами `Formats`, `ReadData`,
`WriteData`. Реализованы Unicode, HTML, PNG, URI/file lists и бинарные MIME
типы; совокупный payload ограничен 64 МиБ. Windows не ожидает 150 мс для
пустого clipboard. X11 поддерживает INCR; Wayland использует data offers.
`desktop.DataDropWindow.SetOnDataDropped` задаёт предпочитаемые MIME-типы
приёма внешнего drop через OLE, Xdnd или Wayland; старый файловый callback
сохраняется.

Продолжение реализует Windows TSF TextStoreACP и IMM reconversion, UIA
Text/RangeValue/Selection/Grid, AT-SPI Text/Value/Selection/Table и события,
а также внешние OLE/Xdnd тесты между процессами. Это всё ещё экспериментальные
адаптеры: нужны реальные CJK engines, screen readers, полная виртуальная иерархия,
Table headers/relations, TextEdit patterns и исходящий drag source. Подробности
и границы приёмки: [DISPLAY_INPUT_VALIDATION.md](DISPLAY_INPUT_VALIDATION.md).
Weston 10 проверяет render/capture и закрытие окна; наличие text-input-v3 или
полноценную композицию на другом compositor этот результат не подтверждает.

## Повторение проверки

На Windows из корня workspace:

```powershell
go test -p 4 -race '-tags=accessibility,migrated_fynedo' -count=1 ./... ./extras/apps/...
go test -p 4 -race '-tags=no_glfw,migrated_fynedo' -count=1 ./... ./extras/apps/...
$env:FYNE_TEST_CLIPBOARD='1'
go test -p 4 -race '-tags=accessibility,migrated_fynedo' ./internal/driver/glfw -run TestClipboard -count=1
```

В PowerShell значение `-tags=...` с запятой нужно передавать в кавычках.
Clipboard test сохраняет доступные предыдущие представления в памяти;
при невозможности безопасного восстановления оставляет clipboard нетронутым.
`FYNE_TEST_CLIPBOARD_CHILD` активируется только у независимого test subprocess.

На Linux нужны X11/Wayland/GL development libraries, Xvfb, D-Bus, at-spi2-core,
fontconfig и CJK fonts для `TestHangul`:

```sh
xvfb-run -a dbus-run-session -- env FYNE_PLATFORM=x11 FYNE_TEST_ATSPI=1 FYNE_TEST_CLIPBOARD=1 \
  go test -p 4 -race -tags=accessibility,migrated_fynedo -count=1 ./... ./extras/apps/...
```

`ci` переключает app.New на software driver, поэтому для проверки нативного
приложения этот тег не используется. Результаты и ограничения каждой проверки
находятся в [CORE_VALIDATION.md](CORE_VALIDATION.md).
