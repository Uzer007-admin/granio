# Реализация следующей очереди GL / GLFW / Glow / D-Bus

12 октября 2026 года, Go 1.27.1, Windows amd64 и Linux `f-test-bench` amd64.
Это библиотечные primitives для задач из
[NATIVE_VISUAL_REQUIREMENTS.md](NATIVE_VISUAL_REQUIREMENTS.md).
Production-код ядра Fyne в этой итерации не менялся.

## Реализовано

| Библиотека | Изменение | Контракт |
| --- | --- | --- |
| `gl` | Material-профиль GL 2.1: 78 functions / 68 enums, separate blend, core/ARB и EXT texture FBO/renderbuffer | 53 required и 25 optional functions. FBO extensions не повышают минимальную обязательную версию; default Fyne profile остаётся 50 functions |
| `gl` | `capabilities` и `capabilities/native` | Version/API, vendor/renderer, extensions, max texture size; выбор FBO требует version/extension и всех необходимых procedure addresses. Snapshots не используют глобальную таблицу GL; native query делает три resolver lookups на текущем context thread |
| `glfw` 3.3/3.4/3.5 | `GetSurfaceGeometry`, `SurfaceGeometry`, `PixelScale` | Window и framebuffer dimensions, content scale отдельно; три native geometry queries в одном cgo-вызове без output allocations; focus/visibility/compositor flags не запрашиваются |
| `glow` | Schema-2 symbol manifests и сохранение неизменённых файлов | Return и ordered parameter C/Go types; одинаковые Go files, JSON, headers и vendor stubs сохраняют content и mtime. Header copy читает источник до изменения destination; changed writes staged |
| `dbus` | `Conn.SubscribeSignals` | Owned FIFO capacity 1..65536, drop-newest при overflow, monotonic counter и coalesced notification; нет goroutine на deferred delivery. Работает рядом с custom SignalHandler; legacy `Signal` не изменён |
| `dbus/portal` | `WatchAppearance` | Один control goroutine, максимум один активный read worker, один pending snapshot; уникальный owner, restart/overflow re-read, отмена старого чтения, suppress unchanged snapshots; caller-owned connection остаётся открытым |

### GL

```go
info, err := native.Query(glfw.GetProcAddress)
if err != nil {
    return err
}
path := info.Framebuffers(gl.IsFunctionAvailable)
separateBlend := info.SeparateBlend(gl.IsFunctionAvailable)
```

Здесь `native` — `github.com/go-gl/gl/capabilities/native`,
`gl` — `github.com/go-gl/gl/v2.1/material`. Выбор CoreFramebuffer учитывает
GL 3+, ARB FBO, GLES 2+ и WebGL; EXT path проверяется отдельно. Consumer выполняет
вызовы выбранного family и проверяет actual framebuffer completeness. Наличие
entry point само по себе не включает feature. Pure-Go Queries допускает browser
extension provider; C-adapter остаётся отдельным.

GL 3+ использует indexed extension queries; legacy GL/GLES 2 — extension string.
RGB и alpha blend factors проверены реальным render/readback. Основания API:
[Khronos glGetString](https://github.com/KhronosGroup/OpenGL-Refpages/blob/main/gl4/glGetString.xml),
[Khronos GLES glBlendFuncSeparate](https://github.com/KhronosGroup/OpenGL-Refpages/blob/main/es2.0/glBlendFuncSeparate.xml).

### D-Bus

```go
watcher, err := portal.WatchAppearance(ctx, conn, portal.WatchOptions{})
if err != nil {
    return err
}
defer watcher.Close()
for snapshot := range watcher.Updates() {
    // Передать snapshot в UI scheduler; user overrides применяются отдельно.
}
```

Watcher устанавливает свои match rules, читает существующий owner через daemon,
запрашивает ReadAll по уникальному имени и проверяет owner после ответа. Сигналы
других senders на shared connection не передают UI значения. При raw queue loss
выполняется resync; даже queued owner-loss требует authoritative read, поскольку
следующий owner-acquisition мог быть потерян. Одинаковые snapshots не порождают
новые UI updates. Отсутствующий portal не auto-activates, наблюдение продолжается.

Source preference keys и wire contract:
[XDG Settings](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.Settings.html).
Снятие rules при завершении имеет отдельные bounded deadlines; при недоступном
daemon cleanup является best-effort. Owned subscription закрывается локально
независимо от результата cleanup RPC. Snapshot хранит последнее наблюдение;
после Done его нельзя считать текущим состоянием работающего сервиса.

## Проверка

| Проверка | Результат |
| --- | --- |
| Glow / capability parsing / optional material extensions | Windows и Linux PASS, race/cgocheck2 |
| GLFW native suites 3.3/3.4/3.5 | Windows и Linux Xvfb PASS, race/cgocheck2; Geometry соответствует отдельным queries |
| Bounded subscriptions | Overflow/order, concurrent delivery/close и closed registration PASS, race |
| D-Bus full normal + godbus_no_tcp | Linux настоящий isolated dbus-daemon, оба full suites PASS, race |
| Portal watcher | Initial unavailable -> owner -> settings change -> restart; чужой sender на shared match; cancel blocked old-owner read; connection remains open after Close |
| Current watcher повторно | 20 runs normal и 5 godbus_no_tcp, PASS, race; initial full matrix тоже включала five repeats |
| Desktop material FBO | NVIDIA GL 4.6 Windows и Mesa GL 4.5 Linux/X11/Wayland, core/ARB и EXT paths PASS |
| Strict GLES 2.0 | Mesa override 2.0, 142 core lookups, native capability query и texture-FBO separate alpha blend PASS |
| Native C bridges / GUID | Linux ASan/UBSan PASS; Geometry fixture включён во все три версии |
| Fyne integration | app / internal/painter/gl / internal/driver/glfw PASS с race и migrated_fynedo на обоих хостах |
| Pure capability portability | GOOS=js / GOARCH=wasm / CGO_ENABLED=0 test binary compiled; browser runtime не проверен |
| Unchanged regeneration | Fyne и material, по 8 файлов: content и mtime полностью сохранены, включая headers/stubs |

Полный `go vet` выбранных packages возвращает две existing diagnostics в
generated `Ptr(uintptr)` / deprecated `PtrOffset`: integer-to-pointer conversion
incompatible с unsafeptr analysis. Эти helpers сохранены для API compatibility;
рекомендуются SlicePtr и typed WithOffset overloads. Это не green full-vet gate.
Новый handwritten capability, watcher/subscription и Glow код проверяется
отдельным vet; live GPU checks используют safe overloads и проходят checkptr=2.

Основной тест пикселей: premultiplied green backdrop `(0,.25,0,.25)` и straight
red source alpha 0/.25/.5/.75/1. Separate alpha blend даёт примерно
`[0,64,0,64]`, `[64,48,0,112]`, `[128,32,0,159]`, `[191,16,0,207]`,
`[255,0,0,255]` с допуском 2 для quantization. Оба desktop FBO families дают
те же результаты. GLES 2 midpoint — `[128,32,0,160]`. Проверяется storage/blending
offscreen target; compositor materials этим не приняты.

Material probe использует скрытое окно: до Wayland output enter его scale может
быть 1 при output scale 2. Отдельная проверка видимых окон после event processing
показывает framebuffer 384x256 для 192x128 и 448x320 для 224x160. Geometry API
не pump events и не подменяет значения предполагаемым DPI.

## Измерения

Три последовательных sample runs, median; Windows i5-13400, Linux Xeon Gold 5320.

| Запрос | Старые три getters | GetSurfaceGeometry | Go allocations |
| --- | ---: | ---: | --- |
| Windows | 2 156 ns | 2 069 ns | 3 / 24 B -> 0 / 0 B |
| Linux / X11 / Xvfb | 49 593 ns | 44 867 ns | 3 / 24 B -> 0 / 0 B |

Median меньше примерно на 4% и 9.5% соответственно; sample scatter не даёт
оснований обещать стабильный выигрыш времени кадра. Изменение additive: старый
renderer loop не переключён на новые queries. Состояние рекомендуется собирать
после size/scale events и использовать повторно.

Bounded delivery в один полный буфер с уже существующим Signal: median 85.49 ns
Windows и 107.5 ns Linux, 0 B / 0 allocations. Это стоимость dispatch/drop и
notification, без message decode, IPC или создания входного Signal. Размер
очереди и количество delivery goroutines не растут при медленном consumer.

Материал-профиль использует 78 resolver lookups вместо полного GL 2.1 profile
с 2 848. Это разные наборы функций; сравнение FPS на основании lookup counts
не производится. Snapshot/capability query выполняется при создании/смене
context, а не в каждом draw call.

Raw samples, GPU probe JSON и SHA-256 11 exact Linux sources и всех 25 manifests:
[library-development.json](graphics-assessment/library-development.json).
Logs: [library-development](graphics-assessment/library-development).
Доказательство unchanged generation:
[unchanged-generation.json](graphics-assessment/unchanged-generation.json).
Все пять exported fork patches применены к чистым pinned upstream archives и
сверены по содержимому: 211 / 699 / 548 / 117 / 20 файлов. Reverse applicability
также прошла; [hashes](graphics-assessment/fork-patch-replay.json).

## Повторение и границы

```powershell
python scripts/update-material-gl-profile.py
go generate -tags gen ./deps/gl
go test -race ./deps/glow ./deps/gl/capabilities/... ./deps/gl/v2.1/material ./deps/dbus/portal
go test -race -run 'Test(Bounded|Store|Watcher)' ./deps/dbus
go test -race -tags glfw_integration ./deps/glfw/v3.5/glfw
go run -tags graphics_probe ./scripts/material-gl-probe/main.go
python scripts/verify-unchanged-gl-generation.py
```

Linux: `pack-graphics-validation.py` -> unpack sources/proxy в собственный
validation directory -> из `source` выполнить
`sh scripts/validate-library-next-linux.sh /absolute/validation-directory`.
Используются отдельные GOCACHE/GOTMPDIR/GOMODCACHE, file proxy и собственный Weston
runtime. Logs сохраняются перед удалением общего directory.

Namespace-auth case остаётся SKIP (`ENOSPC` user/PID namespaces). Не выполнены
Windows daemon IPC, реальные desktop portal implementations, GNOME/KWin/wlroots,
physical fractional-DPI, macOS/BSD/ARM/mobile runtime и browser rendering.
Native DWM/AppKit materials, Fyne material graph, memory budget, scene present
latency и соответствующие AT-gates из ТЗ остаются следующей интеграционной работой.
