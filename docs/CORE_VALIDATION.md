# Проверка изменений ядра

12 октября 2026 года, Go 1.27.1 / amd64 / GOAMD64=v3. Проверяются Fyne и
`extras/apps` вместе, включая локальные версии GL/GLFW/Glow/D-Bus.
История исходной проверки сохранена в [TEST_RESULTS.md](TEST_RESULTS.md).
Это базовый снимок перед продолжением работ с высокими частотами, разрешениями,
TSF и расширением native accessibility. Актуальные результаты этого продолжения
находятся в [DISPLAY_INPUT_VALIDATION.md](DISPLAY_INPUT_VALIDATION.md).

## Полные наборы

| Среда | Теги | Результат | Passed test actions | Пропущены |
| --- | --- | --- | ---: | --- |
| Windows, native GLFW + UIA, race | accessibility,migrated_fynedo | 47 packages PASS, 0 FAIL | 4904 | 3 upstream Android fixture tests и 2 gates независимого clipboard test |
| Windows, headless, race | no_glfw,migrated_fynedo | 47 packages PASS, 0 FAIL | 4747 | 3 upstream Android fixture tests |
| Linux f-test-bench, Xvfb + D-Bus + AT-SPI, native GLFW, race | accessibility,migrated_fynedo | 49 packages PASS, 0 FAIL | 5005 | 3 upstream Android fixture tests и gate clipboard child subprocess |

Test actions включают вложенные subtests; это не число независимых test functions.
Packages с `[no test files]` не включены в PASS. На Linux переменные
`FYNE_TEST_ATSPI=1` и `FYNE_TEST_CLIPBOARD=1` включали реального D-Bus клиента
и обмен clipboard между процессами; дочерний тест действительно выполнялся
в subprocess. На Windows этот обмен был выполнен отдельной проверкой с race.

В первом Linux run без `ci` падал только `TestHangul`: на стенде не было
системного CJK шрифта. Установлен `fonts-noto-cjk` без обновления существующих
пакетов; повторный полный run прошёл. Для native проверки также установлены
X11/GL/Wayland development dependencies, Xvfb, at-spi2-core, ibus и Weston.
Глобальные proxy/DNS настройки не менялись; использован временный локальный
public-module proxy для offline Go builds.

## Нативные проверки

- Windows и Linux/X11: отдельный процесс проверил точное содержимое Unicode,
  HTML и бинарного MIME payload 512 КиБ, включая X11 INCR. Проверяется возврат
  ошибки отсутствующего формата без ожидания трёхсекундного deadline.
- Linux AT-SPI: второй D-Bus connection прочитал Unicode по rune offsets,
  проверил отсутствие password text, вызвал кнопку, изменил Entry и получил
  ошибку после удаления элемента. Проверка проходила под race detector.
- Windows UIA: проверены Value pattern, COM references после удаления control,
  несколько окон и reference после закрытия окна. Это проверка provider ABI и
  времени жизни, а не приёмка Narrator/NVDA.
- IME: Entry и RichTextEntry проверены на preedit без изменения binding,
  замещение selection, подтверждение одним Undo, отмену, смену фокуса,
  surrounding edit и password redaction. Нативные IMM/XIM/Wayland обвязки
  собираются. Полная композиция с реальными CJK engines ещё не подтверждена.
- В 400 мс native animation наблюдались 48–49 ticks при 120 Гц и 57–58 при
  144 Гц. Запросы из очереди пробуждают idle wait до его 250 мс deadline.
- Обычное приложение без `ci`: Windows, Linux/X11 и Linux/Wayland под Weston
  отдали Capture 167×123 и штатно завершили Run после Close/Quit. Исходное
  зависание WaitForEvents было воспроизведено и исправлено.

Native MIME drop adapters скомпилированы на Windows/Linux. Прогон взаимодействия
с внешними drag sources, полные screen-reader patterns, TSF/reconversion и
конкретные IME engines остаются отдельными задачами приёмки; результат compile
не выдаётся за их проверку.

## Производительность

Одинаковый Windows CPU i5-13400, три повтора, `-benchmem`, без race detector.
Время зависит от фоновых задач; финальные measurements выполнены при общей
нагрузке workspace. Baseline взят до изменений алгоритмов.

| Операция | До | Финальное состояние |
| --- | --- | --- |
| TextGrid Refresh | 22.9–24.1 µs; 4066 B; 1001 allocations | 16.6–20.7 µs; 65 B; 1 allocation |
| splitLines | 28 allocations | 9 allocations; 10.8–11.8 µs |
| WrapWord | 13948 allocations; 1.696 MB | 7385 allocations; 1.718 MB; 1.22–1.41 ms |
| WrapBreak | 14617 allocations; 1.704 MB | 7767 allocations; 1.740 MB; 1.33–1.72 ms |
| Geometry viewport 1000 / 100000 / 1000000 rows | линейные обходы до индексации | 148–155 / 208–210 / 238–296 ns; 0 allocations |
| Cache touch | — | 8.7–9.3 ns; 0 allocations |

Wrap уменьшает количество allocations, но bytes/op немного выросли. Финальный
WrapBreak timing под нагрузкой перекрывается с baseline; универсальное ускорение
всего текстового движка не заявлено. Бенчмарк geometry измеряет поиск диапазона,
а не создание всех widgets, рендеринг кадров или скорость источника данных.

## Воспроизводимость и файлы

Все пять submodule patches применены к чистым Git archives upstream bases.
Сверены дерево файлов и содержимое после нормализации LF; reverse check тоже
прошёл: [fork-patch-replay-final.json](core-validation/fork-patch-replay-final.json).
Повторное применение `scripts/apply-patches.py` распознаёт уже применённые patches.
Новые C/header/protocol inputs входят в GLFW patch.

Windows GUI binaries пересобраны в `bin/fyne_settings.exe` и `bin/apps.exe` с
`-tags=accessibility,migrated_fynedo -ldflags=-H=windowsgui`.

Свидетельства сохранены в [core-validation](core-validation/evidence.json):
полные Go JSON logs, отдельные native checks, shutdown, benchmarks и replay.
SHA-256 перечислены в evidence index. Временные рабочие каталоги, offline proxy
и тестовый Weston удаляются после сохранения этих результатов.
