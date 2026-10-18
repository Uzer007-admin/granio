# Реализация локальных GL / GLFW / Glow / D-Bus форков

Дата: 12 октября 2026 года. Это результат исправлений после [исходной оценки](GRAPHICS_STACK_ASSESSMENT.md). Все 11 `go.mod` и `go.work` используют Go 1.27.1.

## Версии и подключение

- `deps/dbus`: полный upstream history, 664 commits, база `4ad75260bdbdd09d51af62b316a17dd60c6aa0e3`. Локальная версия 5.3.0, `Version` и [changelog](../deps/dbus/CHANGELOG.md). Import path v5 сохранён.
- `deps/glfw/v3.5/glfw`: native GLFW 3.5.1, commit `d9d6f0f1f967807ffade6598ea9a631ebaf37a56`. Новая константа `UnlimitedMouseButtons`. Модули 3.3 и 3.4 сохранены.
- Fyne/apps используют GLFW 3.5 и D-Bus 5.3 через workspace и локальные `replace`. Для автономного apps указан локальный Fyne; таким образом в одном executable используется одна GLFW версия.
- Исходные upstream remotes сохранены как `upstream`. Теги и внешние релизы не опубликованы. Нового опубликованного Go GLFW v3.5 модуля в upstream нет: локальное требование `v0.0.0` разрешается через `replace`.

## Исправления исходного отчёта

| Область | Итог |
| --- | --- |
| GL strings | `Strs` добавляет недостающий NUL; таблица и строки находятся в одном C allocation. Пустые строки, existing/embedded NUL проверены. `StrsWithLengths` сохраняет точные длины; `Str` использует `unsafe.StringData`. |
| Upload helpers | `SlicePtr[T Scalar]`, `SliceBytes[T Scalar]`: фиксированная ширина чисел, пустые slices возвращают nil, нет boxing/reflection. Raw API сохранён; lifetime остаётся ответственностью caller. |
| Buffer offsets | Добавлены `DrawElementsInstancedWithOffset`, `ReadPixelsWithOffset` в XML overrides; все bindings перегенерированы. |
| GLES | Новый strict GLES 2.0 профиль без расширений; Fyne импортирует его вместо GLES 3.1 при запросе контекста 2.0. |
| Vendor | Генератор создаёт `dummy.go` в header directories и module-aware `build_cgo_hack.go`, включая новые GLES/EGL packages. |
| GLFW lifecycle | `Terminate` очищает windows map и Go monitor/joystick callback holders после native cleanup; повторная инициализация проверена. |
| BSD tags | Go native X11/Wayland wrappers и `WAYLAND` согласованы с C selection: Linux поддерживает оба, FreeBSD/NetBSD выбирают по `wayland`, OpenBSD — X11. BSD runtime не проверен. |
| Keyboard queries | Непечатаемые внешние scancodes с `KeyUnknown` и `invalidValue` возвращают пустое имя. Другие programmer errors сохраняют старую обработку. |
| ContextNoError | Константа добавлена в 3.3; в 3.4/3.5 уже существует. |
| Generator errors | Ошибки templates возвращаются без panic; stderr gofmt сохраняется; неполная последняя строка flush-ится; package loader проверяет результат и diagnostics. |
| EGL 1.5 | Исправлены struct names/typedef dependencies, scalar/native handle types, EGL_CAST constants и C pointer returns. Добавлен generated `egl1.5/egl` с EGL resolver. |
| Native update | XML Wayland protocols совпадает с прежним snapshot; генерируемые headers переиспользованы после byte equality check. Сохранён локальный Win32 high-resolution event wait patch; native hashes обновлены. |

Проверенная обработка GUID охватывает Linux/Win32/Cocoa во всех трёх версиях GLFW: форматирование беззнаковых байтов, ограниченные буферы, нулевая инициализация имени, NUL после Linux ioctl и сохранение прежних сопоставлений. Подробные свидетельства и результаты отдельного анализа сохранены отдельно; аппаратные проверки джойстиков остаются необходимыми.

Архитектурные предложения исходного отчёта — no-cgo backend, context-specific Dispatch/capabilities, selective noescape/nocallback и полный IME/platform parity — остаются отдельными задачами. Они не являются подтверждёнными исправленными дефектами этого изменения.

## D-Bus 5.3.0

Исправлены Windows return-value regression, чтение при `(0,nil)` Reader result, short writes, malformed string/signature encoding, object paths, nonzero padding, array boundaries, uint32 length overflow, protocol version и body/signature mismatch. Чужие `panic(error)` из Reader/Writer больше не поглощаются recovery кодека.

Числа пишутся через reusable scratch buffer и ByteOrder PutUint; `float64` читается по битам. Byte arrays проходят целиком, без reflect.Append на каждом байте, с собственным storage у результата. Пустой byte array сохраняет non-nil slice. String converter выделяется только при необходимости. Обработка FD и API `BusObject` сохранены; high uint32 FD indices проверены на Windows 386 без знакового переполнения.

Добавлены `GetPropertyWithContext`, `StorePropertyWithContext`, `SetPropertyWithContext` как функции для любого `BusObject` и методы конкретного `Object`. Не пришлось расширять существующий интерфейс и ломать его реализации. Добавлен поиск socket под `XDG_RUNTIME_DIR`. Снимок 36 открытых upstream issues: [реестр](dbus-issue-audit/INVENTORY.md), [полные тексты и комментарии](dbus-issue-audit/open-issues.json).

## Производительность

Windows amd64, Go 1.27.1, i5-13400. Один и тот же benchmark запускался по три раза на оригинальном кодеке и fork; общий overhead Encoder+Decoder включён. Тайминги не являются измерением IPC/FPS и зависят от фоновой нагрузки.

| Benchmark | Upstream baseline | Fork |
| --- | --- | --- |
| Encode/decode byte array 35 КиБ, median | 3.88 ms/op | 8.02 µs/op |
| Byte array allocations | 71 702/op, около 1.07 MB/op | 8/op, 41 224 B/op |
| Существующий mixed interface benchmark, median | 6.02 µs/op | 4.84 µs/op |
| Mixed interface allocations | 34/op, 4 824 B/op | 25/op, 4 948 B/op |

Mixed workload уменьшил число allocations, но bytes/op немного выросли из-за scratch state. Результат byte arrays основан на устранении работы по каждому байту; его нельзя переносить на все D-Bus сообщения.

## Следующая очередь развития: 12 октября 2026

Fyne теперь использует `github.com/go-gl/gl/v2.1/fyne`: 50 функций и 40 enum вместо загрузки 2 848 функций полного GL 2.1 пакета. Полный API сохранён. `profiles/fyne-gl21.json` строится из вызовов native renderer скриптом `update-fyne-gl-profile.py`; генератор проверяет наличие каждого символа. Диагностика renderer/version и компиляции шейдеров включена в профиль. Локальный GL `replace` добавлен в Fyne и apps для автономной работы с `GOWORK=off`.

Все 24 GL/GLES/EGL пакета перегенерированы. Native `Init` передаёт статические NUL-terminated literals напрямую, устраняя C allocation/free на каждом имени. Пользовательский `InitWithProcAddrFunc` получает прежние имена без NUL. WGL loader отбрасывает `1`, `2`, `3`, `-1` и корректно обрабатывает неудачу загрузки DLL; это соответствует [описанию загрузки функций Khronos](https://wikis.khronos.org/opengl/Load_OpenGL_Functions).

Glow отклоняет неизвестные JSON поля, неизвестные enum/functions, дубликаты символов и второй JSON document до записи bindings. Отсутствующий список оставляет весь соответствующий набор, явный пустой список `[]` удаляет его полностью. Историческая семантика `Package.Filter` сохранена для прямых callers.

Во всех GLFW 3.3/3.4/3.5 исправлен `SetGammaRamp`: равные ненулевые RGB slices передаются через C-stack структуру с правильным `size`. По [контракту GLFW](https://www.glfw.org/docs/latest/group__monitor.html) данные копируются до возврата. Nil/разные/пустые channels вызывают `InvalidValue` до native call. Добавлены `GetAxesInto`, `GetButtonsInto`, `GetHatsInto`, `GetGamepadStateInto`, `GetGammaRampInto`; результаты не заимствуют native память. Убраны вызовы cgo на каждый элемент, а monitor/video-mode wrappers используют общий Go backing array.

Шесть getter C-мостов возвращают count/state по значению из C stack. Это устраняет скрытые Go output-buffer allocations и сохраняет callbacks. Эксперимент с отдельным `noescape` не дал выигрыша: Go 1.27 cgo использует keepalive вместо escape только вместе с `nocallback`, а joystick polling/error paths могут вызвать Go callback. В итоговом коде эти аннотации отсутствуют.

D-Bus 5.3.0 сохраняет `FlagAllowInteractiveAuthorization`. No-reply calls возвращают фактическую ошибку транспорта, игнорируют caller channel и публикуют результат в собственном закрытом `Done`; заранее отменённый context ничего не отправляет. Serial освобождается после no-reply, ошибок/closed connection и pre-cancel; `0` остаётся зарезервированным после wraparound. Отмена уже отправленного reply-expected call сохраняет serial до позднего ответа, чтобы не смешивать вызовы.

Числовые arrays `int16/uint16/int32/uint32/int64/uint64/float64` кодируются/декодируются пакетно с обоими byte orders. Проверены независимые wire vectors, alignment пустых массивов, truncated/oversized input, собственное хранение результата, NaN/signed-zero bits и сохранение FD semantics. Найден и исправлен прежний reflection panic при декодировании UnixFD в arrays/nested arrays/maps: тип контейнера теперь соответствует декодированным descriptors, invalid index возвращает ошибку. Массивы и словари FD прошли передачу через настоящий Linux daemon.

| Замер Windows amd64, i5-13400 | До этой очереди | После |
| --- | --- | --- |
| D-Bus encode/decode 16 `float64`, median | 2.96 µs, 43 allocs, 1 224 B | 0.823 µs, 9 allocs, 644 B |
| D-Bus encode/decode 4 096 `float64`, median | 503 µs, 8 212 allocs, около 230 KB | 43.9 µs, 9 allocs, 98 566 B |
| Копирование 64 axes, без device polling, median | 2 520 ns | 5.52 ns, 0 allocs |
| Полный GL native Init, repeated average | 7.78 ms | 6.71 ms |
| Компактный GL native Init, repeated average | — | 53.3 µs, 50 lookups, 0 Go allocs |
| Четыре reusable joystick queries на Null platform | 4 Go allocations | 0 |

Raw samples и границы измерений: [evolution-benchmarks.json](graphics-assessment/evolution-benchmarks.json), [логи проверок](graphics-assessment/evolution-validation). GL timing не является серией из трёх benchmark runs; копирование arrays не измеряет IPC, polling или FPS. Ранее оптимизированный 35 KiB byte-array путь сохранил 8 allocations / 41 224 B.

Окончательные Windows/Linux GLFW suites прошли с race/cgocheck2, gamma round-trip выполнен на Null platform без изменения физических мониторов. C fixtures проверяют три gamma и 18 array/state bridges плюс WGL failure cases; Linux fixtures прошли ASan/UBSan. Полные Linux D-Bus suites с race, обычная и `godbus_no_tcp`, прошли; namespace case явно skipped (`ENOSPC`). На обоих хостах прошли `app`, `internal/painter/gl`, `internal/driver/glfw` с race. Linux GL shader/readback и строгий GLES 2.0 readback прошли. Физические joystick/gamepad, macOS/BSD/ARM runtime и Windows D-Bus daemon не проверены.

Windows D-Bus codec/send/container tests прошли с race на amd64 и отдельно на 386. Повторный Fyne race-run с `GOWORK=off` и offline proxy прошёл; apps транзитивно разрешает компактный профиль без workspace. Свежий standalone consumer с `-mod=vendor` собрался и запустился: DBus 5.3.0, GLFW 3.5.1, compact GL/GLES 2.0/3.0. Новый `native_array.h` и KHR headers сохранены в vendor.

Полные пять fork patches обновлены; `verify-fork-patches.py` применяет их к чистым pinned archives в собственных temporary Git repositories, сравнивает все файлы и проверяет reverse applicability. [Актуальные hashes](graphics-assessment/fork-patch-replay.json) подтверждают совпадение; temporary directories удаляются.

## Проверка и повторение

- Windows: native GLFW 3.5.1 hidden window/reinit и invalid-scancode tests с race; GL conversions с race; shaders из отдельных ненулевых строк, PBO readback, synchronous debug callback с `cgocheck2`/`checkptr=2`. GPU RTX 4080 SUPER, OpenGL 4.6 / driver 616.92.
- Linux `f-test-bench`: нормальная и `godbus_no_tcp` D-Bus suites с настоящим dbus-daemon и race; GLFW 3.3/3.4 compile, GLFW 3.5 native lifecycle под Xvfb; X11-only/Wayland-only builds.
- EGL smoke: EGL 1.5, Mesa 25.0.5; первоначальный ES 2-compatible context вернул OpenGL ES 3.2. Затем проверен `MESA_GLES_VERSION_OVERRIDE=2.0`: версия OpenGL ES 2.0, 142 core lookups, pixel `[0,255,0,255]`. Fyne окно 64x64 и Capture дали зелёный центр. GLES адаптер не вызывает отсутствующий ReadBuffer, как mobile/WebGL. Runtime другого 2.0-only драйвера не заявлен.
- Свежий Windows consumer с `GOWORK=off` и `-mod=vendor` собирается и запускается: KHR headers сохранены для GLES 2.0/3.0. Все пять patches применены к чистым upstream архивам в временных Git repositories; каждый файл сверён после нормализации LF: [доказательство](graphics-assessment/fork-patch-replay.json).
- GUID runner выполняет реальные production formatting blocks и legacy converters с signed/unsigned char и ASan/UBSan. Имеется независимый review.
- Fyne интеграция: пакеты `internal/driver/glfw`, `internal/painter/gl`, `app` проверяются с race и `migrated_fynedo`.

На Linux один namespace-auth test явно пропущен: kernel отказывает в создании user/PID namespace (`ENOSPC`). Windows daemon IPC, macOS/BSD/ARM runtime и реальные джойстики не проверены. Проверка shader/GPU работы не означает приёмку всех IME/window-management сценариев.

Fyne GLES probe проверяет render/capture и завершает процесс сразу после результата; graceful application shutdown им не проверяется. Попытка обычного Close/Quit зависала в текущем изменяемом Fyne tree. Самостоятельные native GLFW Init/Terminate/reinit tests проходят; это отдельная граница.

```powershell
python scripts/export-fork-patches.py
python scripts/verify-fork-patches.py
python scripts/apply-patches.py
./scripts/build.ps1 -Task All -Race -NativeGLFW
go -C deps/dbus test -race -run 'Test(Codec|Windows|PropertyContext|MessageLength|Encode|Decode|ReadFull|SigByteSize|Proto)' .
go -C deps/glow test ./...
go -C deps/glfw/v3.5/glfw test -race -tags glfw_integration .
```

```sh
dbus-run-session -- sh -c 'export DBUS_SYSTEM_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS; go -C deps/dbus test -race ./...; go -C deps/dbus test -race -tags godbus_no_tcp ./...'
xvfb-run -a go -C deps/glfw/v3.5/glfw test -race -tags glfw_integration .
xvfb-run -a go run -tags egl ./scripts/egl-probe/main.go
python3 scripts/test-joystick-guid.py --sanitize
python3 scripts/test-native-bridges.py --sanitize
```

Полные fork patches применяются к pinned upstream HEAD и пропускаются при повторном запуске. Conflicts не сбрасываются. `prepare-glfw-35.py` — средство первичного импорта upstream; сохранённый `glfw-fork.patch` является итоговым reproducible state, включая fixes после импорта.


Дополнительная проверка ядра: зависание штатного Close/Quit воспроизведено и исправлено в очереди lifecycle. Windows, Linux/X11 и Linux/Wayland теперь проходят native render/Capture и graceful shutdown; подробности и границы приёмки — [CORE_VALIDATION.md](CORE_VALIDATION.md).

Следующая итерация добавила полную перезагрузку GL procedure table и availability
queries, безопасную генерацию с symbol manifests, SurfaceMetrics во всех трёх
GLFW modules и XDG Settings client. Native portal проверка также выявила и
исправила потерю wire signatures Variant в Store containers. Подробности и
новые измерения, включая отсутствие подтверждённого timing ускорения X11:
[VISUAL_STACK_VALIDATION.md](VISUAL_STACK_VALIDATION.md). Отдельный platform-aware
backlog по визуальному ТЗ: [NATIVE_VISUAL_REQUIREMENTS.md](NATIVE_VISUAL_REQUIREMENTS.md).
