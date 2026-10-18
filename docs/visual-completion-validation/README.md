# Visual completion validation — 2026-10-14

`acceptance.json` содержит статусы и границы приёмки, `source-manifest.json` —
SHA-256 2215 текущих Go/C/native/module inputs локального workspace.
Идентичность финальному Linux checkout пока не установлена: после успешного
предыдущего Linux прогона SSH через devproxy стал сбрасываться до handshake.
Папка linux пока не содержит remote raw logs; она не подтверждает Linux-приёмку.

В windows сохранены окончательные результаты:

- `windows-root-latest.log`: 47 пакетов root/race/cgocheck2, последовательный запуск.
- `windows-functional-final.log`: visibility/Escape, 100 native lifecycle/resize/
  backdrop/maximize/minimize cycles и live TextScale без потери состояния;
  race/cgocheck2/checkptr=2.
- `windows-software-ci.log`: app/theme/canvas/widget/software suite.
- `references-tests.log`: SVG light→dark→light с test.Settings и app.Settings.
- `gallery-windows-{mica,mica-alt,acrylic}.{log,json}`: настоящая Fyne gallery,
  системный font policy, SurfaceReady/NativeAccepted и FirstFrameBeforeShow.
- `environment.json`: Windows 26300, RTX 4080 SUPER, driver 32.0.16.1692, Go 1.27.1.
- `windows-vet.log`: семь существующих unsafeptr warnings в Win32 clipboard/IMM;
  `windows-vet-native.log` — остальные analyzers; `windows-vet-core.log` и
  `gallery-vet.log` — полный vet соответствующего core/gallery.
- `wasm-latest.log`, `macos-software-latest.log`: окончательные compile checks,
  не runtime-приёмка браузера или macOS.

references содержит 39 opaque PNG, bundled fonts и SHA-256 manifest. Два
независимых экспорта совпали; `windows/references-identity.json` фиксирует результат.
Dark/high-contrast при TextScale 225% просмотрены. Canvas snapshots не включают
системную композицию DWM. Обновление эталонов выполняется exporter из галереи;
побитовое сравнение native Mica/Acrylic между GPU/OS не используется.

Длительные performance/stress runs исключены по указанию пользователя.
KDE-стенды пока неактивны. Windows 22621/22631, Intel/AMD, native macOS,
физические mixed-DPI мониторы и реальные переключения настроек ОС открыты.
Полная аппаратная приёмка ТЗ не зачтена.
