# Прозрачные окна через DXGI и DirectComposition

2 ноября 2026 года. Совместимые окна Windows с заранее подготовленной alpha
используют аппаратный OpenGL → D3D11 → DXGI → DirectComposition. Поддерживаются
Transparent, Mica, Mica Alt, Desktop Acrylic и System Blur. Отрисовка OpenGL
сохраняется; кадр не читается в память CPU и не выводится через WGL SwapBuffers.

## Выбор поверхности

`GRANIO_WGL_PRESENT=auto` или отсутствие переменной выбирает этот путь для
подготовленной alpha-поверхности. `dxgi` включает тот же выбор явно. Обычное
непрозрачное окно использует DXGI swap chain, связанную непосредственно с HWND.
`wgl` выбирает прежний вывод. `GRANIO_RENDERER=software` сохраняет приоритет
и использует непрозрачный программный fallback.

Требуются desktop OpenGL, `WGL_NV_DX_interop2`, framebuffer blit и совместимое
аппаратное устройство D3D11. DirectComposition предоставляется начиная
с Windows 8. Системные материалы через документированный
`DWMWA_SYSTEMBACKDROP_TYPE` требуют Windows 11 build 22621 или новее.
GLES сохраняет прежний путь вывода. Без interop аппаратный DXGI не заявляется.

Материал или `PrepareAlpha: true` задаётся через `desktop.ConfigureWindowVisuals`
до создания окна. Позднее можно переключать материалы и Opaque на том же HWND.
Окно без предварительного резервирования возвращает `desktop.ErrVisualNeedsRecreate`
при позднем запросе alpha. Новая ветка обозначается
`desktop.VisualBackendDirectComposition`.

## Alpha и первый кадр

GLFW создаёт HWND с `Win32NoRedirectionBitmap` до установки WGL pixel format:
Windows отклоняет попытку добавить этот флаг к готовому окну. Старый WGL
blur-behind для такой поверхности не включается. Рамка, заголовок и системные
кнопки сохраняются.

Composition swap chain использует два буфера RGBA8, FLIP_SEQUENTIAL и
PREMULTIPLIED alpha. OpenGL back buffer копируется внутри GPU с восстановлением
read/draw framebuffer, read buffer и scissor. RGB уже умножен на alpha;
полностью прозрачный пиксель равен нулю по всем каналам.

Начальная служебная текстура 1×1 не присоединяется к visual tree. Перед Show
готовится полный кадр фактического размера, затем выполняются Present и первый
Commit. Последующие кадры требуют только Present. Нулевые и свёрнутые размеры
не заменяют последний кадр; resize пересоздаёт необходимые ресурсы.

## Fallback и восстановление

Недоступность DirectComposition обнаруживается до показа. Неизменяемый стиль
HWND требует повторного создания ещё скрытого окна с обычной redirection
surface. Оно получает непрозрачный DXGI либо WGL при отсутствии interop.
Глобальные GLFW hints сбрасываются после каждой попытки, включая отказ.

Контрастная тема, запрет прозрачности, RDP, отсутствие композиции или системного
материала дают непрозрачные пиксели и причину в VisualStatus. При смене политики
alpha-окно сохраняет DirectComposition и HWND, поэтому прозрачность можно
восстановить без пересоздания окна. VisualFallbackFail отражается состоянием Failed.

При потере устройства разрешена одна попытка восстановления до следующего
успешного кадра. Пересоздаются target, visual и composition swap chain.
Восстановление в первом скрытом Present завершается подготовкой нового полного
кадра. При постоянном отказе библиотека не запускает бесконечный цикл
и не объявляет неподготовленную поверхность готовой.

## Короткая проверка

Windows 11 build 26300, RTX 4080 SUPER, драйвер 616.92, Go 1.27.2:

- Нативные проверки с `-race`: скрытый первый кадр, alpha 0/64/128/192/255,
  порядок RGBA, resize, minimize/restore, Opaque ↔ Transparent, нулевой размер,
  предел ресурса, сохранение GL state и освобождение ресурсов.
- Реальный конфликт DirectComposition target проверяет отказ и непрозрачный
  fallback. Подменённый ответ DEVICE_REMOVED проверяет реальные новые
  GPU-ресурсы, включая восстановление первого скрытого кадра. Настоящий сброс
  видеокарты не выполняется.
- Оптическая проверка двух собственных окон: ошибка alpha не более 1/255,
  Mica и Mica Alt различаются, Acrylic реагирует на фон. Контраст шахматного
  фона: стандартное отклонение 119 без размытия и 0,449 с Acrylic.
- Первый показ содержит материал и корректную alpha, без чёрного кадра.
  Системная анимация Show отключена только у тестового окна, чтобы её
  масштабирование не подменяло проверку первого подготовленного кадра.

Другие GPU и Windows 22621/22631 требуют самостоятельной аппаратной приёмки.
Linux-проверки с Mesa llvmpipe подтверждают регрессию библиотек, а не эффекты
реальных GNOME/KDE/Fly. Длительные проверки производительности не выполняются.
Исходники приложений не изменяются.

[Состав проверок и результаты](windows-desktop-stall-validation/dcomp-checks.json).

Основание интерфейсов: [composition swap chain](https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nf-dxgi1_2-idxgifactory2-createswapchainforcomposition),
[DirectComposition](https://learn.microsoft.com/en-us/windows/win32/api/dcomp/nf-dcomp-dcompositioncreatedevice),
[системные материалы DWM](https://learn.microsoft.com/en-us/windows/win32/api/dwmapi/ne-dwmapi-dwm_systembackdrop_type).
