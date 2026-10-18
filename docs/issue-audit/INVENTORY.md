# Реестр открытых issues Fyne и apps

Снимок от 2026-10-12 01:09 МСК содержит все 655 открытых issues Fyne и 3 issues apps.

Тема назначена по заголовку и меткам. Метки передают состояние upstream; подтверждение конкретного дефекта в нашем форке требует воспроизведения. Полные описания сохранены в [open-issues.json](open-issues.json). Анализ приоритетов находится в [ISSUES_AUDIT.md](../ISSUES_AUDIT.md).

## fyne-io/fyne

| Issue | Заголовок | Тема | Метки upstream |
| --- | --- | --- | --- |
| [#6596](https://github.com/fyne-io/fyne/issues/6596) | Window content jumps vertically during quick resize | Платформы окна и интеграция с ОС | unverified |
| [#6589](https://github.com/fyne-io/fyne/issues/6589) | TextGrid under ThemeOverride draws letters in mixed fonts and sizes | Графика шрифты и масштабирование | нет |
| [#6587](https://github.com/fyne-io/fyne/issues/6587) | Android: Press and hold backspace in entry deletes only newly entered text | Платформы окна и интеграция с ОС | OS:Android, unverified |
| [#6583](https://github.com/fyne-io/fyne/issues/6583) | mobile: Context menu for selectable seems misplaced | Ввод фокус и навигация | Driver:mobile, unverified |
| [#6582](https://github.com/fyne-io/fyne/issues/6582) | Wayland touchpad gestures | Ввод фокус и навигация | нет |
| [#6579](https://github.com/fyne-io/fyne/issues/6579) | macOS: every new theme-override scope runs `defaults read -g AppleLocale` on the main thread | Локализация и языки | нет |
| [#6570](https://github.com/fyne-io/fyne/issues/6570) | binding.ExternalTree.Reload does not respect binding&#x27;s nested structure | Виджеты компоновка и привязки данных | unverified |
| [#6560](https://github.com/fyne-io/fyne/issues/6560) | Update golang.org/x/image to address TIFF resource exhaustion vulnerability | Сборка зависимости и установка | нет |
| [#6552](https://github.com/fyne-io/fyne/issues/6552) | Preferences directory created on app startup even when preferences are not used | Файлы диалоги и хранилище | нет |
| [#6541](https://github.com/fyne-io/fyne/issues/6541) | Web backend calls preventDefault on every key, so the browser cannot handle Cmd+1–9 or reload | Ввод фокус и навигация | нет |
| [#6540](https://github.com/fyne-io/fyne/issues/6540) | Wayland: SIGSEGV in PollEvents when the window&#x27;s monitor is switched away | Стабильность и потоки | unverified |
| [#6531](https://github.com/fyne-io/fyne/issues/6531) | Auto-resize dialogs when resizing windows | Платформы окна и интеграция с ОС | нет |
| [#6527](https://github.com/fyne-io/fyne/issues/6527) | Modal PopUp/dialog content re-centers on every canvas resize (regression since 2.8.0) — Android soft keyboard causes buttons to shift under finger mid-tap | Ввод фокус и навигация | unverified |
| [#6509](https://github.com/fyne-io/fyne/issues/6509) | Cache entries touched after a long idle are immediately expired | Производительность и память | нет |
| [#6500](https://github.com/fyne-io/fyne/issues/6500) | perf(test): support reusable image buffer in test.Canvas.Capture() to reduce heap churn in test suites | Тесты и диагностика | нет |
| [#6492](https://github.com/fyne-io/fyne/issues/6492) | Accessibility: macOS system tray item has no AX label or description | Доступность | нет |
| [#6491](https://github.com/fyne-io/fyne/issues/6491) | Accessibility: widget.Card content is omitted from macOS AX tree in v2.8.0 | Доступность | bug, blocker |
| [#6486](https://github.com/fyne-io/fyne/issues/6486) | Problem with Modifier keys on Windows | Ввод фокус и навигация | bug |
| [#6483](https://github.com/fyne-io/fyne/issues/6483) | Windows on ARM: desktop builds request OpenGL ES over WGL, with no way to select desktop GL or EGL | Сборка зависимости и установка | нет |
| [#6481](https://github.com/fyne-io/fyne/issues/6481) | MultipleWindows also overwrites OnDragged/OnResized set before a window is added | Ввод фокус и навигация | нет |
| [#6474](https://github.com/fyne-io/fyne/issues/6474) | Badge for app icon and more | API документация и прочие запросы | нет |
| [#6473](https://github.com/fyne-io/fyne/issues/6473) | Replace Wayland borders from lib decor with our own | Платформы окна и интеграция с ОС | blocker |
| [#6464](https://github.com/fyne-io/fyne/issues/6464) | Desktop: migrate apps config and cache root out of `fyne` subdirectory | Производительность и память | нет |
| [#6455](https://github.com/fyne-io/fyne/issues/6455) | Add Disableable support to widget.Hyperlink | Виджеты компоновка и привязки данных | нет |
| [#6450](https://github.com/fyne-io/fyne/issues/6450) | For some CJK strings, fallback font selection oscillates between two fonts every character or few | Ввод фокус и навигация | bug |
| [#6447](https://github.com/fyne-io/fyne/issues/6447) | wayland: text renders with wrong font size on startup until hover (regression from v2.7.4) | Графика шрифты и масштабирование | unverified, Wayland |
| [#6433](https://github.com/fyne-io/fyne/issues/6433) | NewMultipleWindows constructor overwrites OnTappedBar callback which may have been user-set | Ввод фокус и навигация | bug |
| [#6423](https://github.com/fyne-io/fyne/issues/6423) | Adding a widget to a container doesn&#x27;t resize it | Виджеты компоновка и привязки данных | unverified |
| [#6420](https://github.com/fyne-io/fyne/issues/6420) | Linux wayland window resize not working when called after window.Show but before app.Run | Платформы окна и интеграция с ОС | bug, Wayland |
| [#6410](https://github.com/fyne-io/fyne/issues/6410) | Need a way to query for the current locale&#x27;s text direction | Локализация и языки | нет |
| [#6403](https://github.com/fyne-io/fyne/issues/6403) | Systray not registering silently causes crash on application quit on Linux | Стабильность и потоки | bug |
| [#6383](https://github.com/fyne-io/fyne/issues/6383) | Disabled button does not fade emoji (bitmap glyph) text content | Локализация и языки | нет |
| [#6374](https://github.com/fyne-io/fyne/issues/6374) | Update back button implementation in GoNativeActivity | Виджеты компоновка и привязки данных | unverified |
| [#6373](https://github.com/fyne-io/fyne/issues/6373) | Intel/Arc graphics: Drag resizing on Windows causes application freezes | Стабильность и потоки | OS:Windows, unverified |
| [#6365](https://github.com/fyne-io/fyne/issues/6365) | Allow Go code using Fyne to make dynamic changes to icon state in taskbar | Стабильность и потоки | future |
| [#6359](https://github.com/fyne-io/fyne/issues/6359) | Wayland: canvas keeps stale layout after window auto-expands for content min-size change | Стабильность и потоки | unverified |
| [#6354](https://github.com/fyne-io/fyne/issues/6354) | Failure to run on iOS 27 if built using Xcode/sdk 27 due to API change | Стабильность и потоки | unverified, after-release |
| [#6353](https://github.com/fyne-io/fyne/issues/6353) | Support for fyne.TextTruncateMiddle | API документация и прочие запросы | нет |
| [#6316](https://github.com/fyne-io/fyne/issues/6316) | widget.Entry: Button as ActionItem is misaligned on 1920x1080 display | Виджеты компоновка и привязки данных | bug |
| [#6310](https://github.com/fyne-io/fyne/issues/6310) | A colour picker widet (could derive from dialog&#x27;s `colorAdvancedPcker`) | Графика шрифты и масштабирование | нет |
| [#6306](https://github.com/fyne-io/fyne/issues/6306) | widget.List: Keep list scrolled to bottom when appending list item, if list is already scrolled to bottom | Виджеты компоновка и привязки данных | нет |
| [#6303](https://github.com/fyne-io/fyne/issues/6303) | ubuntu + wayland windows resize lag with dialog | Производительность и память | OS:Linux, optimization, Wayland |
| [#6301](https://github.com/fyne-io/fyne/issues/6301) | *** Error in Fyne call thread, this should have been called in fyne.Do[AndWait] *** - raised | API документация и прочие запросы | unverified, information-needed |
| [#6293](https://github.com/fyne-io/fyne/issues/6293) | Allow widget.RichText to be scrolled via keyboard | Ввод фокус и навигация | нет |
| [#6269](https://github.com/fyne-io/fyne/issues/6269) | Cannot build under macOS for macOS due to old deployment target | Сборка зависимости и установка | unverified, information-needed |
| [#6251](https://github.com/fyne-io/fyne/issues/6251) | [Wayland] Small window content leaves transparent gutter | Графика шрифты и масштабирование | bug, OS:Linux, Wayland |
| [#6235](https://github.com/fyne-io/fyne/issues/6235) | Calling obj.MinSize() in software.Render() breaks TestRender | Графика шрифты и масштабирование | unverified |
| [#6219](https://github.com/fyne-io/fyne/issues/6219) | Scrollbars are very hard to hit on Android | Платформы окна и интеграция с ОС | unverified |
| [#6217](https://github.com/fyne-io/fyne/issues/6217) | Button icon is not grayed out when disabled if PNG content | Виджеты компоновка и привязки данных | enhancement |
| [#6214](https://github.com/fyne-io/fyne/issues/6214) | Mouse wheel scrolling disabled when the pointer is over a sticky column in widget.Table | Ввод фокус и навигация | unverified |
| [#6198](https://github.com/fyne-io/fyne/issues/6198) | Window does not show when launching on a secondary monitor on Windows | Платформы окна и интеграция с ОС | unverified, information-needed |
| [#6192](https://github.com/fyne-io/fyne/issues/6192) | Inconsistency in RunNative context | API документация и прочие запросы | unverified |
| [#6189](https://github.com/fyne-io/fyne/issues/6189) | wasm: add local font support | Графика шрифты и масштабирование | нет |
| [#6185](https://github.com/fyne-io/fyne/issues/6185) | FileOpenDialog on mobile - extension filter | Файлы диалоги и хранилище | unverified |
| [#6174](https://github.com/fyne-io/fyne/issues/6174) | Rich Text - custom colors | Графика шрифты и масштабирование | нет |
| [#6172](https://github.com/fyne-io/fyne/issues/6172) | Rich Text - Markdown makes nonsense the first time | API документация и прочие запросы | unverified |
| [#6168](https://github.com/fyne-io/fyne/issues/6168) | Card styles Elevated, Filled, Outlined | API документация и прочие запросы | нет |
| [#6167](https://github.com/fyne-io/fyne/issues/6167) | Button styles Elevated and Outlined | Виджеты компоновка и привязки данных | нет |
| [#6164](https://github.com/fyne-io/fyne/issues/6164) | Themed ressourceb | API документация и прочие запросы | unverified |
| [#6160](https://github.com/fyne-io/fyne/issues/6160) | Allow vertically expanded tables | Виджеты компоновка и привязки данных | нет |
| [#6147](https://github.com/fyne-io/fyne/issues/6147) | Window.Show() silently fails after Windows deep sleep/resume (stale GLFW viewport) | Платформы окна и интеграция с ОС | unverified, information-needed |
| [#6142](https://github.com/fyne-io/fyne/issues/6142) | Android - add own amnifest | Платформы окна и интеграция с ОС | information-needed |
| [#6141](https://github.com/fyne-io/fyne/issues/6141) | wasm: missing cache support | Производительность и память | bug, Driver:web |
| [#6137](https://github.com/fyne-io/fyne/issues/6137) | Proposal: UX/A11y - Make widget/Entry select all text when tabbed into | Доступность | нет |
| [#6129](https://github.com/fyne-io/fyne/issues/6129) | Anroid - Softkeyboard | Ввод фокус и навигация | unverified |
| [#6124](https://github.com/fyne-io/fyne/issues/6124) | Creation of scroll container with 30 paragraphs is super slow on Android | Производительность и память | unverified |
| [#6118](https://github.com/fyne-io/fyne/issues/6118) | Removing items from a bound list spawns out-of-bound error | Виджеты компоновка и привязки данных | unverified |
| [#6110](https://github.com/fyne-io/fyne/issues/6110) | file dialog crashs | Стабильность и потоки | unverified, information-needed |
| [#6099](https://github.com/fyne-io/fyne/issues/6099) | TextGrid: Chinese characters will overlap | Локализация и языки | unverified |
| [#6092](https://github.com/fyne-io/fyne/issues/6092) | macOS systray (menubar) items don&#x27;t work with menubar organizers | Платформы окна и интеграция с ОС | unverified |
| [#6084](https://github.com/fyne-io/fyne/issues/6084) | macOS: re-opening a systrayed/menubared should open hidden windows | Платформы окна и интеграция с ОС | нет |
| [#6073](https://github.com/fyne-io/fyne/issues/6073) | document app.Storage() from `storage` API docs | Файлы диалоги и хранилище | information-needed |
| [#6072](https://github.com/fyne-io/fyne/issues/6072) | Select widget usability improvement: Match via character | Виджеты компоновка и привязки данных | нет |
| [#6071](https://github.com/fyne-io/fyne/issues/6071) | Keyboard usability issues in Select | Ввод фокус и навигация | unverified |
| [#6070](https://github.com/fyne-io/fyne/issues/6070) | Unable to input Greek characters with Greek keyboard in WASM builds on Ubuntu 24.04 | Сборка зависимости и установка | unverified |
| [#6069](https://github.com/fyne-io/fyne/issues/6069) | widget.Table: Headers disappear on size updates and fails to resize properly | Виджеты компоновка и привязки данных | unverified |
| [#6063](https://github.com/fyne-io/fyne/issues/6063) | Accordion in a Scroll Container | Виджеты компоновка и привязки данных | unverified |
| [#6061](https://github.com/fyne-io/fyne/issues/6061) | Range slider widget | Виджеты компоновка и привязки данных | нет |
| [#6055](https://github.com/fyne-io/fyne/issues/6055) | linux-arm: Build fails with &#x27;EGL/egl.h&#x27; file not found when using -tags wayland | Сборка зависимости и установка | unverified |
| [#6050](https://github.com/fyne-io/fyne/issues/6050) | FileOpenDialog - Directories on Desktop should be on top of the list | Файлы диалоги и хранилище | нет |
| [#6049](https://github.com/fyne-io/fyne/issues/6049) | FileOpen | Файлы диалоги и хранилище | unverified, information-needed |
| [#6048](https://github.com/fyne-io/fyne/issues/6048) | My project is to use a dropdown component on a touch screen Linux system, but the scrollbar of widget.Select is too thin | Ввод фокус и навигация | нет |
| [#6047](https://github.com/fyne-io/fyne/issues/6047) | lang: unable to translate &quot;description&quot; | Локализация и языки | bug |
| [#6042](https://github.com/fyne-io/fyne/issues/6042) | Focus events cause multiline entry to change scroll position to show cursor | Стабильность и потоки | unverified |
| [#6039](https://github.com/fyne-io/fyne/issues/6039) | Explore making MacOS CFBundleVersion be `{{.Version}}.{{.Build}}` | Сборка зависимости и установка | нет |
| [#6030](https://github.com/fyne-io/fyne/issues/6030) | Resizing a window that contains a popup menu causes the window to jiggle | Платформы окна и интеграция с ОС | нет |
| [#6021](https://github.com/fyne-io/fyne/issues/6021) | ShowAndRun sometimes fails to return on Windows | Платформы окна и интеграция с ОС | unverified |
| [#6016](https://github.com/fyne-io/fyne/issues/6016) | OS font lookup not working for some text | Графика шрифты и масштабирование | bug |
| [#6011](https://github.com/fyne-io/fyne/issues/6011) | DocTabs side menu not showing icons. | Виджеты компоновка и привязки данных | unverified |
| [#6005](https://github.com/fyne-io/fyne/issues/6005) | Open systray stops/pauses running app on macOS | Платформы окна и интеграция с ОС | bug |
| [#6004](https://github.com/fyne-io/fyne/issues/6004) | [Linux] Window icon is ignored on Gnome taskbar | Платформы окна и интеграция с ОС | unverified |
| [#6003](https://github.com/fyne-io/fyne/issues/6003) | Support removing, or overriding builtin application shortcuts | Ввод фокус и навигация | нет |
| [#5992](https://github.com/fyne-io/fyne/issues/5992) | gridWrap.ScrollToBottom() keeps highlight and jumps back to it | Виджеты компоновка и привязки данных | bug |
| [#5983](https://github.com/fyne-io/fyne/issues/5983) | window Icon does not work on wayland | Платформы окна и интеграция с ОС | bug, OS:Linux, Wayland |
| [#5982](https://github.com/fyne-io/fyne/issues/5982) | go build/fyne package succeed but .exe can&#x27;t run on Windows 11 | Сборка зависимости и установка | unverified |
| [#5978](https://github.com/fyne-io/fyne/issues/5978) | Reflect sum of all tags and metadata in compiled | Сборка зависимости и установка | нет |
| [#5976](https://github.com/fyne-io/fyne/issues/5976) | Add support for XDG Desktop Portal FileTransfer drag-and-drop with application/vnd.portal.filetransfer MIME type | Ввод фокус и навигация | bug, OS:Linux, flatpak |
| [#5963](https://github.com/fyne-io/fyne/issues/5963) | Table Widget: Inconsistent visual highlight behavior | Виджеты компоновка и привязки данных | unverified |
| [#5962](https://github.com/fyne-io/fyne/issues/5962) | Allow widget.Select to always show PlaceHolder at the top | Виджеты компоновка и привязки данных | нет |
| [#5949](https://github.com/fyne-io/fyne/issues/5949) | Fyne should support SVG&#x27;s `mask`, `symbol`, and `use` tags. | Графика шрифты и масштабирование | unverified |
| [#5940](https://github.com/fyne-io/fyne/issues/5940) | Support centering the `*container.Scroll` content on a position (when content can scroll) | Виджеты компоновка и привязки данных | enhancement |
| [#5916](https://github.com/fyne-io/fyne/issues/5916) | Mobile web build: Unable to paste in entry fields | Сборка зависимости и установка | unverified |
| [#5915](https://github.com/fyne-io/fyne/issues/5915) | Broken NewMultiLineEntry scrolling on Android | Платформы окна и интеграция с ОС | unverified |
| [#5914](https://github.com/fyne-io/fyne/issues/5914) | Many widgets in a `container.Grid` make it not usable on mobile (buggy and very laggy) | Производительность и память | unverified |
| [#5905](https://github.com/fyne-io/fyne/issues/5905) | Some widgets do not update when setting theme on android | Платформы окна и интеграция с ОС | unverified |
| [#5898](https://github.com/fyne-io/fyne/issues/5898) | When the window is minimized and restored again, there is a deviation in the UI position | Платформы окна и интеграция с ОС | OS:Windows, unverified |
| [#5894](https://github.com/fyne-io/fyne/issues/5894) | Update internal widgets to use `theme.XYZForWidget` helpers | Производительность и память | optimization, refactor |
| [#5885](https://github.com/fyne-io/fyne/issues/5885) | Ability to Enable/Disable fyne.Menu top level menus | Виджеты компоновка и привязки данных | нет |
| [#5865](https://github.com/fyne-io/fyne/issues/5865) | Cannot override DateEntry Validator | Виджеты компоновка и привязки данных | unverified |
| [#5864](https://github.com/fyne-io/fyne/issues/5864) | Support all keys in mobile.NumberKeyboard and normal Keyboard on Android | Ввод фокус и навигация | unverified |
| [#5861](https://github.com/fyne-io/fyne/issues/5861) | Buttons in notifications | Платформы окна и интеграция с ОС | нет |
| [#5859](https://github.com/fyne-io/fyne/issues/5859) | Support Data Binding in Widgets Defined Outside of fyne | Виджеты компоновка и привязки данных | нет |
| [#5856](https://github.com/fyne-io/fyne/issues/5856) | Rework tab containers to handle refresh in a better way | Производительность и память | optimization, refactor |
| [#5852](https://github.com/fyne-io/fyne/issues/5852) | List: Selected item is not cleared when list goes from populated -&gt; empty -&gt; populated | Виджеты компоновка и привязки данных | unverified |
| [#5848](https://github.com/fyne-io/fyne/issues/5848) | OnEnteredForeground lifecycle callback invoked before run loop begins on Windows | Платформы окна и интеграция с ОС | bug, OS:Windows |
| [#5847](https://github.com/fyne-io/fyne/issues/5847) | AssertImageMatches does not always produce the same image | Тесты и диагностика | unverified |
| [#5844](https://github.com/fyne-io/fyne/issues/5844) | Names in `layout` package stutter | Сборка зависимости и установка | future |
| [#5837](https://github.com/fyne-io/fyne/issues/5837) | can not start a demo,panic,font problem | Стабильность и потоки | unverified |
| [#5827](https://github.com/fyne-io/fyne/issues/5827) | Widgets should not use cache.Renderer | Производительность и память | optimization, refactor |
| [#5823](https://github.com/fyne-io/fyne/issues/5823) | ActionItem cannot be added to bound entry | Виджеты компоновка и привязки данных | bug |
| [#5814](https://github.com/fyne-io/fyne/issues/5814) | boundTreeItem data race with boundTree | Стабильность и потоки | unverified, information-needed |
| [#5810](https://github.com/fyne-io/fyne/issues/5810) | Spinner Widget | Виджеты компоновка и привязки данных | нет |
| [#5809](https://github.com/fyne-io/fyne/issues/5809) | Keyboard Access Keys for menus | Ввод фокус и навигация | нет |
| [#5800](https://github.com/fyne-io/fyne/issues/5800) | Support for biometrics | API документация и прочие запросы | нет |
| [#5799](https://github.com/fyne-io/fyne/issues/5799) | Soft password keyboard not showing password manager | Ввод фокус и навигация | нет |
| [#5798](https://github.com/fyne-io/fyne/issues/5798) | Mobile keyboard capitalisation | Ввод фокус и навигация | нет |
| [#5801](https://github.com/fyne-io/fyne/issues/5801) | Application Not Responsive Until Resize(?) Event | Стабильность и потоки | нет |
| [#5770](https://github.com/fyne-io/fyne/issues/5770) | Consider using weak pointers in internal caches | Производительность и память | optimization |
| [#5768](https://github.com/fyne-io/fyne/issues/5768) | Issue rendering SVG file - (path outside radius?) | Графика шрифты и масштабирование | unverified |
| [#5763](https://github.com/fyne-io/fyne/issues/5763) | Set EGL Context when GLES is used | Графика шрифты и масштабирование | нет |
| [#5755](https://github.com/fyne-io/fyne/issues/5755) | Touchscreen scrolling on desktop | Ввод фокус и навигация | нет |
| [#5751](https://github.com/fyne-io/fyne/issues/5751) | Support more than just an array of strings as options for a select | API документация и прочие запросы | enhancement |
| [#5750](https://github.com/fyne-io/fyne/issues/5750) | Add the option to disable item tabs | Виджеты компоновка и привязки данных | нет |
| [#5749](https://github.com/fyne-io/fyne/issues/5749) | Add a footer to Table widget | Виджеты компоновка и привязки данных | enhancement |
| [#5748](https://github.com/fyne-io/fyne/issues/5748) | Data race in Fyne API during app shutdown with multiple goroutines | Стабильность и потоки | bug |
| [#5744](https://github.com/fyne-io/fyne/issues/5744) | MultipleWindows: Tap content to bring to top | Ввод фокус и навигация | bug |
| [#5743](https://github.com/fyne-io/fyne/issues/5743) | Improved handling of heterogeneous data in tables | Виджеты компоновка и привязки данных | нет |
| [#5708](https://github.com/fyne-io/fyne/issues/5708) | Android theme does not update immediately when changed while app is focused | Стабильность и потоки | unverified |
| [#5700](https://github.com/fyne-io/fyne/issues/5700) | Request: Measure the last row of RichText during word wrapping | Виджеты компоновка и привязки данных | нет |
| [#5695](https://github.com/fyne-io/fyne/issues/5695) | Mobile tap not working on bottom aligned buttons when soft keyboard is visible | Ввод фокус и навигация | bug, blocker, Driver:mobile |
| [#5693](https://github.com/fyne-io/fyne/issues/5693) | Adjusted popup position does not consider window padding | Платформы окна и интеграция с ОС | bug |
| [#5691](https://github.com/fyne-io/fyne/issues/5691) | storage.Delete() fails to delete non-empty folders | Файлы диалоги и хранилище | unverified |
| [#5683](https://github.com/fyne-io/fyne/issues/5683) | Animate Dialogs and PopUps when showing/dismissing | Виджеты компоновка и привязки данных | enhancement |
| [#5663](https://github.com/fyne-io/fyne/issues/5663) | Selectable labels do not work properly on Android | Ввод фокус и навигация | unverified |
| [#5660](https://github.com/fyne-io/fyne/issues/5660) | Add Text Alignment to Entry Widget | Виджеты компоновка и привязки данных | нет |
| [#5656](https://github.com/fyne-io/fyne/issues/5656) | Add support for Wrapping for RadioGroup option strings | Виджеты компоновка и привязки данных | enhancement |
| [#5634](https://github.com/fyne-io/fyne/issues/5634) | Support notifications and theme watching on mobile simulator | Платформы окна и интеграция с ОС | enhancement, mobile-simulator |
| [#5607](https://github.com/fyne-io/fyne/issues/5607) | A scroll widget without a scroll bar | Виджеты компоновка и привязки данных | нет |
| [#5604](https://github.com/fyne-io/fyne/issues/5604) | Mobile tags fullscreen | Платформы окна и интеграция с ОС | unverified |
| [#5585](https://github.com/fyne-io/fyne/issues/5585) | Incorrect item height in list with label Wrapping | Виджеты компоновка и привязки данных | unverified |
| [#5568](https://github.com/fyne-io/fyne/issues/5568) | Implement a single threaded pool to replace sync.Pool | Производительность и память | optimization |
| [#5543](https://github.com/fyne-io/fyne/issues/5543) | Radio Items take width of longest Item when set horizontally | Виджеты компоновка и привязки данных | unverified |
| [#5522](https://github.com/fyne-io/fyne/issues/5522) | Use navigator.userAgentData?.mobile for mobile lookup | Платформы окна и интеграция с ОС | enhancement, Driver:web |
| [#5517](https://github.com/fyne-io/fyne/issues/5517) | Add fast path for move using URI file repository | Производительность и память | optimization, Storage |
| [#5502](https://github.com/fyne-io/fyne/issues/5502) | Replace untyped bindings with generic implementations | Производительность и память | enhancement, optimization, data binding |
| [#5501](https://github.com/fyne-io/fyne/issues/5501) | Form dialog with pre-filled text Entry widgets doesn&#x27;t resize to allow for that text | Виджеты компоновка и привязки данных | unverified |
| [#5492](https://github.com/fyne-io/fyne/issues/5492) | [3.0] Consider making the `storage/repository` package internal (or merging into `storage`) | Сборка зависимости и установка | future, refactor |
| [#5486](https://github.com/fyne-io/fyne/issues/5486) | Unbounded internal cache sizes can lead to unbounded RAM use | Производительность и память | bug, optimization |
| [#5473](https://github.com/fyne-io/fyne/issues/5473) | FyneApp.toml section for Windows metadata | Сборка зависимости и установка | нет |
| [#5458](https://github.com/fyne-io/fyne/issues/5458) | Expand test suite with fuzz testing | Тесты и диагностика | enhancement, testing |
| [#5457](https://github.com/fyne-io/fyne/issues/5457) | Some tests are flaky on develop | Тесты и диагностика | bug, testing |
| [#5449](https://github.com/fyne-io/fyne/issues/5449) | AppTabs OnSelected fires a lot of times on Disable and Enable | Виджеты компоновка и привязки данных | not an issue |
| [#5447](https://github.com/fyne-io/fyne/issues/5447) | Consider switching test assertions to alecthomas/assert | Тесты и диагностика | enhancement, testing |
| [#5446](https://github.com/fyne-io/fyne/issues/5446) | Merge coverage from Linux, macOS and Windows for coverage report | Платформы окна и интеграция с ОС | enhancement, Workflows |
| [#5410](https://github.com/fyne-io/fyne/issues/5410) | mobile web paints the canvas too high up in chrome android | Графика шрифты и масштабирование | unverified |
| [#5404](https://github.com/fyne-io/fyne/issues/5404) | Rework Linux/BSD mobile simulator to use GLFW | Платформы окна и интеграция с ОС | enhancement, OS:Linux, OS:FreeBSD, mobile-simulator |
| [#5402](https://github.com/fyne-io/fyne/issues/5402) | There is no way to get the beginning of the selected text in entry | Виджеты компоновка и привязки данных | нет |
| [#5393](https://github.com/fyne-io/fyne/issues/5393) | Investigate suspicious preference write delay | API документация и прочие запросы | information-needed, races |
| [#5391](https://github.com/fyne-io/fyne/issues/5391) | Change widget.Entry.onFocusChanged to widget.Entry.OnFocusChanged | Стабильность и потоки | нет |
| [#5389](https://github.com/fyne-io/fyne/issues/5389) | Minor incorrect triggering of &quot;Failed to convert systray icon&quot; error | Платформы окна и интеграция с ОС | unverified |
| [#5369](https://github.com/fyne-io/fyne/issues/5369) | Once the filedialog opened and file selected, the window will lost all the binding events handling | Файлы диалоги и хранилище | unverified, information-needed |
| [#5356](https://github.com/fyne-io/fyne/issues/5356) | TextGrid - Random Errors when running the application | Виджеты компоновка и привязки данных | unverified |
| [#5440](https://github.com/fyne-io/fyne/issues/5440) | Selection highlighting in the calendar application | Ввод фокус и навигация | нет |
| [#5333](https://github.com/fyne-io/fyne/issues/5333) | Lang package - Set the desired language | Сборка зависимости и установка | нет |
| [#5323](https://github.com/fyne-io/fyne/issues/5323) | Text strings become blurry when dragging window from low DPI to high DPI monitor | Ввод фокус и навигация | bug |
| [#5319](https://github.com/fyne-io/fyne/issues/5319) | [Windows] Fyne chooses app language based on system *region* and not language setting | Локализация и языки | unverified |
| [#5299](https://github.com/fyne-io/fyne/issues/5299) | Widgets added to Overlay have incorrect Size until window Resize event occurs. | Платформы окна и интеграция с ОС | unverified |
| [#5297](https://github.com/fyne-io/fyne/issues/5297) | Long text in multiline word wrapped entries causes massive performance issues | Производительность и память | unverified |
| [#5289](https://github.com/fyne-io/fyne/issues/5289) | Add Support For Printing | API документация и прочие запросы | нет |
| [#5284](https://github.com/fyne-io/fyne/issues/5284) | NewFileOpen needs file path information in the callback | Файлы диалоги и хранилище | information-needed |
| [#5269](https://github.com/fyne-io/fyne/issues/5269) | Allow disabling of all translations | Локализация и языки | Internationalisation |
| [#5257](https://github.com/fyne-io/fyne/issues/5257) | Replacement character for C0 control characters is measured wrong | API документация и прочие запросы | bug |
| [#5256](https://github.com/fyne-io/fyne/issues/5256) | Android/iOS: Touch/mouse coordinates are wrong | Ввод фокус и навигация | OS:Android, OS:iOS, unverified |
| [#5255](https://github.com/fyne-io/fyne/issues/5255) | Android: Tapping Entry with stylus creates a sticky selection area | Ввод фокус и навигация | unverified, information-needed |
| [#5253](https://github.com/fyne-io/fyne/issues/5253) | Replace strcpy and strcat with safer alternatives in C-code on Android | Платформы окна и интеграция с ОС | enhancement, OS:Android |
| [#5237](https://github.com/fyne-io/fyne/issues/5237) | color picker callback isn&#x27;t called in case of dismissal | Графика шрифты и масштабирование | unverified |
| [#5235](https://github.com/fyne-io/fyne/issues/5235) | Menu shortcuts should be disabled while a dialog is open | Ввод фокус и навигация | нет |
| [#5233](https://github.com/fyne-io/fyne/issues/5233) | the app was built for an older version of android.... install anyway | Сборка зависимости и установка | unverified |
| [#5232](https://github.com/fyne-io/fyne/issues/5232) | set binding&#x27;s value only after successful validation | Виджеты компоновка и привязки данных | data binding |
| [#5216](https://github.com/fyne-io/fyne/issues/5216) | webasm app run very slowly  | Производительность и память | unverified |
| [#5211](https://github.com/fyne-io/fyne/issues/5211) | dialog.New(Folder, File, etc)Open(callback, parent) + Show() does not work in javascript / wasm | Файлы диалоги и хранилище | unverified |
| [#5207](https://github.com/fyne-io/fyne/issues/5207) | func (f *fileDialog) setLocation(dir fyne.URI) error { does not work with non &quot;file&quot; repository | Файлы диалоги и хранилище | bug |
| [#5205](https://github.com/fyne-io/fyne/issues/5205) | Fyne app crashes on unfocus | Стабильность и потоки | unverified |
| [#5204](https://github.com/fyne-io/fyne/issues/5204) | dialog.New(File/Folder/...)Open() only works with &quot;file&quot; repository | Файлы диалоги и хранилище | bug, FileDialog |
| [#5203](https://github.com/fyne-io/fyne/issues/5203) | Add fyne repository implementation for WebDav | Файлы диалоги и хранилище | нет |
| [#5199](https://github.com/fyne-io/fyne/issues/5199) | widget.Card does not like big text | Виджеты компоновка и привязки данных | unverified |
| [#5194](https://github.com/fyne-io/fyne/issues/5194) | In widget.Form, Validateable items don&#x27;t show their error text unless they are widget.Entry or have HintText set | Виджеты компоновка и привязки данных | unverified |
| [#5189](https://github.com/fyne-io/fyne/issues/5189) | [android] dlopen failed: library &quot;*&quot; not found | Платформы окна и интеграция с ОС | unverified |
| [#5181](https://github.com/fyne-io/fyne/issues/5181) | Vertically Aligned Toolbar | Виджеты компоновка и привязки данных | нет |
| [#5179](https://github.com/fyne-io/fyne/issues/5179) | font load error | Графика шрифты и масштабирование | unverified, information-needed |
| [#5178](https://github.com/fyne-io/fyne/issues/5178) | fyne_demo not runing  on FreeBSD | API документация и прочие запросы | unverified |
| [#5172](https://github.com/fyne-io/fyne/issues/5172) | Password entry shifts window on wlroots | Платформы окна и интеграция с ОС | unverified |
| [#5147](https://github.com/fyne-io/fyne/issues/5147) | Storing data during development in a different location | API документация и прочие запросы | нет |
| [#5130](https://github.com/fyne-io/fyne/issues/5130) | [3.0] Consider removing render cache and make widgets maintain a reference to their renderer | Производительность и память | нет |
| [#5111](https://github.com/fyne-io/fyne/issues/5111) | Ability to add CanvasObjects to the right part of Tab containers | Виджеты компоновка и привязки данных | нет |
| [#5110](https://github.com/fyne-io/fyne/issues/5110) | RichText rendered from markdown renders an extra space after underscore (MacOS) | Графика шрифты и масштабирование | bug |
| [#5106](https://github.com/fyne-io/fyne/issues/5106) | select menu choices | Виджеты компоновка и привязки данных | unverified, information-needed |
| [#5100](https://github.com/fyne-io/fyne/issues/5100) | MacOS Notification: Wrong icon + opens script editor | Платформы окна и интеграция с ОС | not an issue |
| [#5090](https://github.com/fyne-io/fyne/issues/5090) | Nested Draggable Objects Priority Issue | Ввод фокус и навигация | Driver:mobile |
| [#5089](https://github.com/fyne-io/fyne/issues/5089) | Add Access to Table&#x27;s Current Focus | Ввод фокус и навигация | нет |
| [#5088](https://github.com/fyne-io/fyne/issues/5088) | Table Keyboard Navigation with Header Row Hides First Data Row | Ввод фокус и навигация | bug |
| [#5081](https://github.com/fyne-io/fyne/issues/5081) | Add tooltip to system tray | Платформы окна и интеграция с ОС | нет |
| [#5077](https://github.com/fyne-io/fyne/issues/5077) | fyne_demo logs Fyne error:  Failed to load image when click Widgets/Text | Графика шрифты и масштабирование | bug |
| [#5070](https://github.com/fyne-io/fyne/issues/5070) | fyne package is trying to bundle 32bit resource file | Сборка зависимости и установка | unverified, information-needed |
| [#5062](https://github.com/fyne-io/fyne/issues/5062) | Capability to intercept Quit from Menu/Dock for Mac | Платформы окна и интеграция с ОС | нет |
| [#5059](https://github.com/fyne-io/fyne/issues/5059) | Handle huge amount of Text in Entry Widget | Виджеты компоновка и привязки данных | нет |
| [#5051](https://github.com/fyne-io/fyne/issues/5051) | Showing an overlay widget using the canvas.Overlays().Add() API blocks interaction with rest of window | Платформы окна и интеграция с ОС | documentation |
| [#5046](https://github.com/fyne-io/fyne/issues/5046) | UTF-8 Characters are not displayed correctly on the ubuntu&#x27;s &quot;all windows&quot; section | Платформы окна и интеграция с ОС | unverified |
| [#5044](https://github.com/fyne-io/fyne/issues/5044) | It would be awesome if RichTextFromMarkdown supported footnotes | Виджеты компоновка и привязки данных | нет |
| [#5035](https://github.com/fyne-io/fyne/issues/5035) | Chinese garbled characters in web | Локализация и языки | unverified |
| [#5025](https://github.com/fyne-io/fyne/issues/5025) | Only icons of MediumImportance widgets adhere to ThemedResource declarations with Fyne 2.5 | Виджеты компоновка и привязки данных | unverified |
| [#5024](https://github.com/fyne-io/fyne/issues/5024) | Steam deck apps tap area limited in game mode | Ввод фокус и навигация | unverified |
| [#5017](https://github.com/fyne-io/fyne/issues/5017) | FyneApp.toml not loaded with GOTMPDIR set and running `go run` | API документация и прочие запросы | unverified |
| [#5008](https://github.com/fyne-io/fyne/issues/5008) | Built binary is too large | API документация и прочие запросы | unverified, information-needed |
| [#5006](https://github.com/fyne-io/fyne/issues/5006) | Form dialog validation becomes valid if one input field is valid | Ввод фокус и навигация | bug, Validation |
| [#4997](https://github.com/fyne-io/fyne/issues/4997) | Scrolling is laggy with high speed gaming mice | Производительность и память | bug |
| [#4993](https://github.com/fyne-io/fyne/issues/4993) | Unrelated Button Disable Causes Entry ActionItem to Not Be Displayed | Виджеты компоновка и привязки данных | unverified, information-needed |
| [#4989](https://github.com/fyne-io/fyne/issues/4989) | MacOS notifications not working as expected: showing fallback notifications if explicitly disallowed, showing no notifications if allowed | Платформы окна и интеграция с ОС | unverified |
| [#4985](https://github.com/fyne-io/fyne/issues/4985) | DocTabs.OnUnselected not called | Виджеты компоновка и привязки данных | bug |
| [#4982](https://github.com/fyne-io/fyne/issues/4982) | Objects added to a container that is part of a theme override don&#x27;t get the overridden theme | Виджеты компоновка и привязки данных | bug |
| [#4978](https://github.com/fyne-io/fyne/issues/4978) | Annoying extra output when tests are failing | Тесты и диагностика | bug, testing |
| [#4976](https://github.com/fyne-io/fyne/issues/4976) | Menu tests are failing on some developer computers | Тесты и диагностика | bug, testing |
| [#4971](https://github.com/fyne-io/fyne/issues/4971) | iOS notifications only work in the foreground | Платформы окна и интеграция с ОС | нет |
| [#4970](https://github.com/fyne-io/fyne/issues/4970) | Mobile simulator should follow light/dark mode | Платформы окна и интеграция с ОС | enhancement, OS:Linux, mobile-simulator |
| [#4953](https://github.com/fyne-io/fyne/issues/4953) | Notification actions/callbacks | Платформы окна и интеграция с ОС | нет |
| [#4930](https://github.com/fyne-io/fyne/issues/4930) | Build fails with transparent image for iOS icon | Сборка зависимости и установка | нет |
| [#4915](https://github.com/fyne-io/fyne/issues/4915) | UpdateHeader func of a table with header is being called multiple times whenever a cell in the table&#x27;s body is refreshed | Виджеты компоновка и привязки данных | unverified |
| [#4908](https://github.com/fyne-io/fyne/issues/4908) | Custom fonts use memory that is never freed | Производительность и память | bug, optimization |
| [#4900](https://github.com/fyne-io/fyne/issues/4900) | FormLayout calls MinSize twice | Производительность и память | optimization |
| [#4890](https://github.com/fyne-io/fyne/issues/4890) | Rework MinSize implementation for richtext with wrapping | Виджеты компоновка и привязки данных | bug, races |
| [#4886](https://github.com/fyne-io/fyne/issues/4886) | Incorrect window scaling in Gnome Boxes VMs | Доступность | bug, OS:Linux, Accessibility, Virtual machine |
| [#4868](https://github.com/fyne-io/fyne/issues/4868) | Fyne crashes when resizing by touching Window 10&#x27;s border. | Стабильность и потоки | unverified |
| [#4867](https://github.com/fyne-io/fyne/issues/4867) | RichTextSegment | Виджеты компоновка и привязки данных | нет |
| [#4859](https://github.com/fyne-io/fyne/issues/4859) | Strange behavior on NewAppTabs in a container border | Виджеты компоновка и привязки данных | unverified |
| [#4856](https://github.com/fyne-io/fyne/issues/4856) | Mobile: open pop-up (or any other overlay) does not play well with window-head | Платформы окна и интеграция с ОС | bug |
| [#4854](https://github.com/fyne-io/fyne/issues/4854) | Add a lower-level EditableText widget | Виджеты компоновка и привязки данных | future, Design |
| [#4852](https://github.com/fyne-io/fyne/issues/4852) | GridWrapLayout doesn&#x27;t play nicely with parent containers/layouts | Виджеты компоновка и привязки данных | unverified |
| [#4844](https://github.com/fyne-io/fyne/issues/4844) | TextGrid background color shows visual artifacts when FYNE_SCALE is not 1.0, 1.5, or 2.0 | Графика шрифты и масштабирование | bug |
| [#4836](https://github.com/fyne-io/fyne/issues/4836) | Add a type of overlay/popup container that doesn&#x27;t block interaction with the underlying layer | Виджеты компоновка и привязки данных | нет |
| [#4826](https://github.com/fyne-io/fyne/issues/4826) | Add an API to get the &quot;min&quot; width of text widgets with wrapping that would be needed to show the text without wrapping | Виджеты компоновка и привязки данных | нет |
| [#4819](https://github.com/fyne-io/fyne/issues/4819) | blured scaled zoom gui | Графика шрифты и масштабирование | unverified |
| [#4812](https://github.com/fyne-io/fyne/issues/4812) | Automatically remove focus from a widget that is no longer in the object tree | Ввод фокус и навигация | нет |
| [#4801](https://github.com/fyne-io/fyne/issues/4801) | fyne package does not work with main package not in the current directory | Сборка зависимости и установка | unverified |
| [#4784](https://github.com/fyne-io/fyne/issues/4784) | Licence has self conflicting wording, possible typo | API документация и прочие запросы | unverified |
| [#4782](https://github.com/fyne-io/fyne/issues/4782) | Support Software Rendering for X11 by allowing to disable GLX | Графика шрифты и масштабирование | unverified |
| [#4775](https://github.com/fyne-io/fyne/issues/4775) | container.NewScroll does not support horizontal scroll for widget tree object | Виджеты компоновка и привязки данных | unverified |
| [#4754](https://github.com/fyne-io/fyne/issues/4754) | Add mobile web support | Платформы окна и интеграция с ОС | blocker, Driver:web |
| [#4738](https://github.com/fyne-io/fyne/issues/4738) | Toolbar does not display in both markup and image test rendering | Тесты и диагностика | unverified |
| [#4734](https://github.com/fyne-io/fyne/issues/4734) | Widget SliderWithData doesn&#x27;t work with OnChangeEnded() functionality | Стабильность и потоки | unverified, information-needed |
| [#4731](https://github.com/fyne-io/fyne/issues/4731) | Render of widgets implementing layout.SpacerObject look bad | Графика шрифты и масштабирование | unverified |
| [#4727](https://github.com/fyne-io/fyne/issues/4727) | Ability to Refresh only TabItem (without Refreshing its content) | Виджеты компоновка и привязки данных | enhancement |
| [#4718](https://github.com/fyne-io/fyne/issues/4718) | NewListWithData too many iterations | API документация и прочие запросы | unverified |
| [#4716](https://github.com/fyne-io/fyne/issues/4716) | Access icons by names that describe the icon image | Графика шрифты и масштабирование | нет |
| [#4706](https://github.com/fyne-io/fyne/issues/4706) | Paste multiline text into a multilineEntry sometimes ignores newlines | Ввод фокус и навигация | unverified |
| [#4695](https://github.com/fyne-io/fyne/issues/4695) | Collections widgets should scroll with Home/End/Page Up/Page Down when focused | Ввод фокус и навигация | enhancement |
| [#4692](https://github.com/fyne-io/fyne/issues/4692) | Add default file name extension support to file dialogs | Файлы диалоги и хранилище | нет |
| [#4683](https://github.com/fyne-io/fyne/issues/4683) | The &quot;tree&quot; disappears in the &quot;split&quot; after a certain time | Виджеты компоновка и привязки данных | unverified |
| [#4679](https://github.com/fyne-io/fyne/issues/4679) | SetOnDropped functionallity missing when running app in a browser | Платформы окна и интеграция с ОС | нет |
| [#4675](https://github.com/fyne-io/fyne/issues/4675) | Make bound(External)Untyped.Set more useful by making value comparison more capable | API документация и прочие запросы | нет |
| [#4673](https://github.com/fyne-io/fyne/issues/4673) | Hello author, I use fyne-io/fyne-cross for packaging... | Сборка зависимости и установка | нет |
| [#4660](https://github.com/fyne-io/fyne/issues/4660) | CTRL+A shortcut not working (CTRL+S etc works though) | Ввод фокус и навигация | unverified |
| [#4653](https://github.com/fyne-io/fyne/issues/4653) | Complete internationalisation support | API документация и прочие запросы | future |
| [#4648](https://github.com/fyne-io/fyne/issues/4648) | [Internal optimization] Add a static labelMinSize function that can get the min size of a label without actually creating one | Производительность и память | optimization |
| [#4646](https://github.com/fyne-io/fyne/issues/4646) | Mouse becomes misaligned | Ввод фокус и навигация | unverified |
| [#4645](https://github.com/fyne-io/fyne/issues/4645) | Fyne app freezes with file dialog in /dev/ | Стабильность и потоки | unverified |
| [#4638](https://github.com/fyne-io/fyne/issues/4638) | Conflict between button and progress bar located on two adjacent columns. Regression in fyne 2.4.4 (and 2.5.2) vs 2.4.3 | Виджеты компоновка и привязки данных | unverified |
| [#4621](https://github.com/fyne-io/fyne/issues/4621) | Android Fullscreen | Платформы окна и интеграция с ОС | enhancement |
| [#4605](https://github.com/fyne-io/fyne/issues/4605) | URIs for Windows (c:) are handled incorrectly | Платформы окна и интеграция с ОС | unverified |
| [#4580](https://github.com/fyne-io/fyne/issues/4580) | Option to &quot;Share to...&quot; to your Fyne app | API документация и прочие запросы | future, Design |
| [#4564](https://github.com/fyne-io/fyne/issues/4564) | Win7 32bit use v2.4.3 can&#x27;t create window with lower opengl version ( opengl 1.1) | Сборка зависимости и установка | unverified |
| [#4562](https://github.com/fyne-io/fyne/issues/4562) | Theme variant broken for `theme.ColorNameHeaderBackground` on Android | Графика шрифты и масштабирование | unverified |
| [#4555](https://github.com/fyne-io/fyne/issues/4555) | When packaging for the web. I get an error when it comes to js packaging | Сборка зависимости и установка | unverified, information-needed |
| [#4554](https://github.com/fyne-io/fyne/issues/4554) | widget.Scroll doesn&#x27;t seem to calculate properly the size of its content | Виджеты компоновка и привязки данных | unverified |
| [#4541](https://github.com/fyne-io/fyne/issues/4541) | Rewrite Windows notification logic in C | Платформы окна и интеграция с ОС | enhancement, OS:Windows |
| [#4540](https://github.com/fyne-io/fyne/issues/4540) | Provide public API for manipulating focus of items in a collection widget | Ввод фокус и навигация | нет |
| [#4513](https://github.com/fyne-io/fyne/issues/4513) | A Hoverable and Cursorable widget should have mouse moves processed before cursor updates | Ввод фокус и навигация | bug |
| [#4502](https://github.com/fyne-io/fyne/issues/4502) | ld: warning: ignoring duplicate libraries: &#x27;-lobjc&#x27; | API документация и прочие запросы | unverified |
| [#4499](https://github.com/fyne-io/fyne/issues/4499) | Android folder open dialog returning unusable URI | Файлы диалоги и хранилище | OS:Android, unverified |
| [#4478](https://github.com/fyne-io/fyne/issues/4478) | Attach desktop portals to Wayland windows | Платформы окна и интеграция с ОС | OS:Linux, Wayland |
| [#4472](https://github.com/fyne-io/fyne/issues/4472) | Programmatically selecting an item in List should set it up to be the next-focused item | Ввод фокус и навигация | unverified, information-needed |
| [#4469](https://github.com/fyne-io/fyne/issues/4469) | Fyne can support the development of macOS widgets and can be installed into the macOS notification center. | Сборка зависимости и установка | information-needed |
| [#4454](https://github.com/fyne-io/fyne/issues/4454) | Add API for gamepads and joysticks | API документация и прочие запросы | enhancement, Driver:desktop |
| [#4453](https://github.com/fyne-io/fyne/issues/4453) | Button reports wrong MinSize() if Hide() called before first Refresh() | Виджеты компоновка и привязки данных | unverified |
| [#4431](https://github.com/fyne-io/fyne/issues/4431) | Multiple new windows with mainmenu can mess up the layout | Платформы окна и интеграция с ОС | unverified |
| [#4410](https://github.com/fyne-io/fyne/issues/4410) | AssertRendersToMarkup does not invoke Canvas.EnsureMinSize | Тесты и диагностика | unverified |
| [#4382](https://github.com/fyne-io/fyne/issues/4382) | Add Entry.CursorPosition() fyne.Position API | Ввод фокус и навигация | нет |
| [#4379](https://github.com/fyne-io/fyne/issues/4379) | storage.Move() must keep file modification time unchanged | Стабильность и потоки | unverified |
| [#4374](https://github.com/fyne-io/fyne/issues/4374) | [Regression] calling Hide() on a widget.PopUp does not call FocusLost() on focused widgets contained in it | Ввод фокус и навигация | bug |
| [#4368](https://github.com/fyne-io/fyne/issues/4368) | Improve documentation about `fyne package` flag passthrough | Сборка зависимости и установка | нет |
| [#4367](https://github.com/fyne-io/fyne/issues/4367) | Support XDG accent colour setting  | Графика шрифты и масштабирование | enhancement, OS:Linux |
| [#4366](https://github.com/fyne-io/fyne/issues/4366) | Improve documentation about &quot;fyne package --executable&quot; | Сборка зависимости и установка | нет |
| [#4365](https://github.com/fyne-io/fyne/issues/4365) | FyneApp.toml Details.Build increments even when other value is injected | Сборка зависимости и установка | нет |
| [#4364](https://github.com/fyne-io/fyne/issues/4364) | Build numbers &gt; 65535 procude broken windows executables | Сборка зависимости и установка | bug |
| [#4363](https://github.com/fyne-io/fyne/issues/4363) | Can&#x27;t See AppTab Selection When Hidden Behind Ellipses | Ввод фокус и навигация | нет |
| [#4359](https://github.com/fyne-io/fyne/issues/4359) | Add widget.Switch for toggle switches | Виджеты компоновка и привязки данных | нет |
| [#4353](https://github.com/fyne-io/fyne/issues/4353) | Cannot test if SetCloseIntercept function is set | Тесты и диагностика | нет |
| [#4343](https://github.com/fyne-io/fyne/issues/4343) | In macO,fill in the widget.Entry in Chinese, deleting candidate words will also delete the content | Локализация и языки | unverified |
| [#4334](https://github.com/fyne-io/fyne/issues/4334) | When I debug a Fyne 2.4.1 program on Windows 11, my computer becomes very slow and the mouse barely moves. Keystrokes are also very delayed. | Производительность и память | unverified |
| [#4317](https://github.com/fyne-io/fyne/issues/4317) | Symbol lookup isn’t working for not bundled glyphs | Локализация и языки | unverified |
| [#4315](https://github.com/fyne-io/fyne/issues/4315) | Support display version in metadata | Сборка зависимости и установка | нет |
| [#4300](https://github.com/fyne-io/fyne/issues/4300) | File Dialog - file name area scrolling wrong behaviour  | Файлы диалоги и хранилище | bug |
| [#4299](https://github.com/fyne-io/fyne/issues/4299) | Using custom fonts resulted in deviation in the position of (widget. Entry). ActionItem when using ButtonWithIcon | Графика шрифты и масштабирование | unverified |
| [#4291](https://github.com/fyne-io/fyne/issues/4291) | Fyne Setup.app: WARNING: Secure coding is not enabled for restorable state | API документация и прочие запросы | bug, OS:macOS |
| [#4289](https://github.com/fyne-io/fyne/issues/4289) | Mobile simulation mode does not allow pasting of text into widgets | Платформы окна и интеграция с ОС | unverified |
| [#4287](https://github.com/fyne-io/fyne/issues/4287) | RichText: Support for line height | Виджеты компоновка и привязки данных | нет |
| [#4280](https://github.com/fyne-io/fyne/issues/4280) | Table actions ScrollsToTop when clicking between cell borders | Виджеты компоновка и привязки данных | unverified |
| [#4273](https://github.com/fyne-io/fyne/issues/4273) | widget.NewEntry has unwanted scrollbar with a custom-scaling Theme  | Графика шрифты и масштабирование | unverified |
| [#4272](https://github.com/fyne-io/fyne/issues/4272) | Introduce fuzz test in this project | Тесты и диагностика | нет |
| [#4238](https://github.com/fyne-io/fyne/issues/4238) | Splash window behaves like regular one in i3 | Платформы окна и интеграция с ОС | unverified, information-needed |
| [#4237](https://github.com/fyne-io/fyne/issues/4237) | The &quot;NewRichTextFromMarkdown&quot; widget cannot display an image in &#x27;.gif&#x27; format. | Графика шрифты и масштабирование | unverified |
| [#4218](https://github.com/fyne-io/fyne/issues/4218) | Inconsistent Validation Status Display on Entry widgets when binding/unbinding | Виджеты компоновка и привязки данных | unverified |
| [#4209](https://github.com/fyne-io/fyne/issues/4209) | calling Driver.AbsolutePositionForObject(entry) inside entry.OnChanged can hang the app | Стабильность и потоки | unverified |
| [#4195](https://github.com/fyne-io/fyne/issues/4195) | Add file filter input to FileOpen dialog | Ввод фокус и навигация | enhancement, FileDialog |
| [#4169](https://github.com/fyne-io/fyne/issues/4169) | Filedialog : add column view | Файлы диалоги и хранилище | FileDialog |
| [#4156](https://github.com/fyne-io/fyne/issues/4156) | High cpu usage when drawing a table with 288 items | Производительность и память | unverified |
| [#4155](https://github.com/fyne-io/fyne/issues/4155) | Calling window.SetTitle inside OnEnteredForeground callback will freeze the app | Стабильность и потоки | unverified |
| [#4147](https://github.com/fyne-io/fyne/issues/4147) | Form dialog validation does not work with extended entry | Виджеты компоновка и привязки данных | bug |
| [#4145](https://github.com/fyne-io/fyne/issues/4145) | Better error handling for OpenGL errors | Графика шрифты и масштабирование | нет |
| [#4144](https://github.com/fyne-io/fyne/issues/4144) | Crash in paintImage | Стабильность и потоки | unverified, information-needed |
| [#4143](https://github.com/fyne-io/fyne/issues/4143) | infinite progress bar animation speed size dependent | API документация и прочие запросы | unverified |
| [#4141](https://github.com/fyne-io/fyne/issues/4141) | Simplify theme API | API документация и прочие запросы | enhancement |
| [#4138](https://github.com/fyne-io/fyne/issues/4138) | form validation message not working on unfocus | Ввод фокус и навигация | bug |
| [#4134](https://github.com/fyne-io/fyne/issues/4134) | close sub window crashed on mac os Ventura (intel chip) | Стабильность и потоки | unverified |
| [#4131](https://github.com/fyne-io/fyne/issues/4131) | Focus for collection widgets always lands on first item, even if scrolled | Ввод фокус и навигация | bug |
| [#4127](https://github.com/fyne-io/fyne/issues/4127) | Build date in app metadata | Сборка зависимости и установка | нет |
| [#4106](https://github.com/fyne-io/fyne/issues/4106) | Control Focus Order aka Tab Order | Ввод фокус и навигация | enhancement |
| [#4099](https://github.com/fyne-io/fyne/issues/4099) | container.NewScroll(myCheckGroup) issue when the number of check boxes increases | Виджеты компоновка и привязки данных | unverified |
| [#4098](https://github.com/fyne-io/fyne/issues/4098) | WASI  Support | API документация и прочие запросы | нет |
| [#4096](https://github.com/fyne-io/fyne/issues/4096) | Binding Conversions should support custom methods | Сборка зависимости и установка | нет |
| [#4095](https://github.com/fyne-io/fyne/issues/4095) | Preference Bindings should support Defaults/Fallbacks | Виджеты компоновка и привязки данных | нет |
| [#4078](https://github.com/fyne-io/fyne/issues/4078) | Calling Refresh after Bind on widget.Entry hangs | Стабильность и потоки | unverified |
| [#4063](https://github.com/fyne-io/fyne/issues/4063) | Entry widget caret is hard to see | Виджеты компоновка и привязки данных | нет |
| [#4060](https://github.com/fyne-io/fyne/issues/4060) | make List scroll in both directions | Виджеты компоновка и привязки данных | нет |
| [#4058](https://github.com/fyne-io/fyne/issues/4058) | widget.List is very slow on android | Производительность и память | unverified |
| [#4053](https://github.com/fyne-io/fyne/issues/4053) | Support for font collections in theme / user font | Графика шрифты и масштабирование | нет |
| [#4048](https://github.com/fyne-io/fyne/issues/4048) | Bad console error on not existing directories | Файлы диалоги и хранилище | bug |
| [#4031](https://github.com/fyne-io/fyne/issues/4031) | Resizing a window glitches if its content increases MinSize during the process | Платформы окна и интеграция с ОС | OS:Linux, unverified |
| [#4025](https://github.com/fyne-io/fyne/issues/4025) | ChromeOS IMEs don&#x27;t register keydown event in Entry widgets | Виджеты компоновка и привязки данных | unverified |
| [#4023](https://github.com/fyne-io/fyne/issues/4023) |  Repeatedly moving the mouse through AppTabs causes memory rise | Производительность и память | unverified |
| [#3982](https://github.com/fyne-io/fyne/issues/3982) | Disabled entry doesn&#x27;t &quot;dim&quot; like other widgets. | Виджеты компоновка и привязки данных | unverified |
| [#3968](https://github.com/fyne-io/fyne/issues/3968) | Can&#x27;t set SVG icon in tray | Графика шрифты и масштабирование | OS:Linux, unverified |
| [#3946](https://github.com/fyne-io/fyne/issues/3946) | FileDialog freezing on CentOS 6 | Стабильность и потоки | unverified, information-needed |
| [#3945](https://github.com/fyne-io/fyne/issues/3945) | Bitmap font not support in mac | Графика шрифты и масштабирование | unverified |
| [#3928](https://github.com/fyne-io/fyne/issues/3928) | customWidget.Show() method not called when displayed in a PopUp | Виджеты компоновка и привязки данных | unverified |
| [#3923](https://github.com/fyne-io/fyne/issues/3923) | text render blend func may use BlendFuncSeparate better | Графика шрифты и масштабирование | нет |
| [#3918](https://github.com/fyne-io/fyne/issues/3918) | top Table elements become invisible after ScrollToBottom() | Виджеты компоновка и привязки данных | unverified |
| [#3917](https://github.com/fyne-io/fyne/issues/3917) | Resizable PopUp/Dialogs | Виджеты компоновка и привязки данных | нет |
| [#3907](https://github.com/fyne-io/fyne/issues/3907) | Add NewFolderSave dialog for creating new folder | Файлы диалоги и хранилище | enhancement, FileDialog |
| [#3901](https://github.com/fyne-io/fyne/issues/3901) | DoubleTapped should not &quot;eat&quot; Tapped message | Ввод фокус и навигация | нет |
| [#3899](https://github.com/fyne-io/fyne/issues/3899) | Add displayName option for fyne package command | Сборка зависимости и установка | нет |
| [#3896](https://github.com/fyne-io/fyne/issues/3896) | theme.*Resource wrapper types call svg.Colorize() on raster resources | Графика шрифты и масштабирование | unverified |
| [#3883](https://github.com/fyne-io/fyne/issues/3883) | Accordion item title Wrapping | Виджеты компоновка и привязки данных | нет |
| [#3874](https://github.com/fyne-io/fyne/issues/3874) | Intermittent unexpected signal when closing window from code | Платформы окна и интеграция с ОС | unverified |
| [#3868](https://github.com/fyne-io/fyne/issues/3868) | When in small window mode, moving the Table scroll bar to the last position and then clicking on maximize will cause some data to display bugs. | Платформы окна и интеграция с ОС | unverified |
| [#3858](https://github.com/fyne-io/fyne/issues/3858) | Make tabs of AppTabs container focusable to allow keyboard navigation | Ввод фокус и навигация | enhancement, Keyboard |
| [#3856](https://github.com/fyne-io/fyne/issues/3856) | Submenu of pop-up menu shown near bottom of app window is clipped | Платформы окна и интеграция с ОС | unverified |
| [#3853](https://github.com/fyne-io/fyne/issues/3853) | GUI refresh rate is low on Windows based laptop with 144Hz screen | Платформы окна и интеграция с ОС | enhancement |
| [#3848](https://github.com/fyne-io/fyne/issues/3848) | Provide adaptive window size functionality for widget.Table | Платформы окна и интеграция с ОС | нет |
| [#3845](https://github.com/fyne-io/fyne/issues/3845) | Mac OS: clicking dock icon of system tray-enabled app with hidden window should re-show window | Платформы окна и интеграция с ОС | bug, OS:macOS |
| [#3842](https://github.com/fyne-io/fyne/issues/3842) | The memory occupied by the fyne program is so huge,that made me don&#x27;t wanne use it | Производительность и память | unverified, information-needed |
| [#3836](https://github.com/fyne-io/fyne/issues/3836) | SVG App Window Icon not displaying | Графика шрифты и масштабирование | unverified |
| [#3834](https://github.com/fyne-io/fyne/issues/3834) | add richtext binding to list item | Виджеты компоновка и привязки данных | нет |
| [#3828](https://github.com/fyne-io/fyne/issues/3828) | Tree - get selected IDs | Виджеты компоновка и привязки данных | нет |
| [#3823](https://github.com/fyne-io/fyne/issues/3823) | Cleanup call to canvas.Refresh() | API документация и прочие запросы | нет |
| [#3818](https://github.com/fyne-io/fyne/issues/3818) | Suspected memory leak/OOM error on fyne.Container refresh | Производительность и память | unverified |
| [#3813](https://github.com/fyne-io/fyne/issues/3813) | Markdown: 2 spaces or more at line end should force line break inside paragraph | API документация и прочие запросы | bug |
| [#3812](https://github.com/fyne-io/fyne/issues/3812) | API to receive mouse back/forward button events (mouse buttons 4 and 5) | Ввод фокус и навигация | blocker |
| [#3809](https://github.com/fyne-io/fyne/issues/3809) | fine install command fails on Windows 10 | Сборка зависимости и установка | bug |
| [#3808](https://github.com/fyne-io/fyne/issues/3808) | Markdown: vertical space after list items | Виджеты компоновка и привязки данных | enhancement |
| [#3800](https://github.com/fyne-io/fyne/issues/3800) | Add ability to limit the size of a Select widget dropdown before it scrolls | Виджеты компоновка и привязки данных | нет |
| [#3788](https://github.com/fyne-io/fyne/issues/3788) | Support hiding/removing system tray menu | Платформы окна и интеграция с ОС | нет |
| [#3786](https://github.com/fyne-io/fyne/issues/3786) | CheckGroup has multiple slices of items | Производительность и память | bug, optimization |
| [#3776](https://github.com/fyne-io/fyne/issues/3776) | Systray on Windows is not refreshed while expanded | Платформы окна и интеграция с ОС | bug |
| [#3771](https://github.com/fyne-io/fyne/issues/3771) | Simple video playback and local recording | API документация и прочие запросы | нет |
| [#3758](https://github.com/fyne-io/fyne/issues/3758) | macOS: OrderNSWindow: unsupported window ordering op | Платформы окна и интеграция с ОС | unverified |
| [#3754](https://github.com/fyne-io/fyne/issues/3754) | Build support for wear OS  | Сборка зависимости и установка | нет |
| [#3753](https://github.com/fyne-io/fyne/issues/3753) | Moving split slider repeatedly may cause app to freeze | Стабильность и потоки | unverified |
| [#3749](https://github.com/fyne-io/fyne/issues/3749) | Add Alignment Options For AppTabs. | Виджеты компоновка и привязки данных | enhancement |
| [#3746](https://github.com/fyne-io/fyne/issues/3746) | Add specific theme Size that can modify tab text size | Виджеты компоновка и привязки данных | нет |
| [#3742](https://github.com/fyne-io/fyne/issues/3742) | Failed to uplaod .appx packaged file to Microsoft Store | Сборка зависимости и установка | unverified |
| [#3728](https://github.com/fyne-io/fyne/issues/3728) | If the screen is locked for a long time, panic occurs in fyne program | Стабильность и потоки | unverified |
| [#3720](https://github.com/fyne-io/fyne/issues/3720) | Tests not running on CI are failing locally | Тесты и диагностика | bug |
| [#3713](https://github.com/fyne-io/fyne/issues/3713) | Images should use less memory | Производительность и память | нет |
| [#3708](https://github.com/fyne-io/fyne/issues/3708) | BoundStruct DataListener callback is not invoked when a value in the struct changes | Стабильность и потоки | unverified |
| [#3690](https://github.com/fyne-io/fyne/issues/3690) | Toolbar inside a Border get resized to a zero height in the test | Тесты и диагностика | bug |
| [#3681](https://github.com/fyne-io/fyne/issues/3681) | Text draw in software driver not consistent with arm64 processor | API документация и прочие запросы | bug |
| [#3660](https://github.com/fyne-io/fyne/issues/3660) | Content is shaking when resizing | API документация и прочие запросы | bug, OS:macOS |
| [#3659](https://github.com/fyne-io/fyne/issues/3659) | Image size is changed to smaller by the &quot;Small&quot; Scale size of fyne_settings command but the window size remains the &quot;Normal&quot; Size | Стабильность и потоки | unverified |
| [#3656](https://github.com/fyne-io/fyne/issues/3656) | Add support for push notifications | Платформы окна и интеграция с ОС | нет |
| [#3653](https://github.com/fyne-io/fyne/issues/3653) | Additional Window not loading properly | Платформы окна и интеграция с ОС | OS:Windows, unverified |
| [#3644](https://github.com/fyne-io/fyne/issues/3644) | Fyne demo: Expanded tree doesn&#x27;t refresh properly when entering/exiting full screen | Виджеты компоновка и привязки данных | unverified |
| [#3635](https://github.com/fyne-io/fyne/issues/3635) | glfw libs not works on virtual MacOS/ | API документация и прочие запросы | unverified |
| [#3630](https://github.com/fyne-io/fyne/issues/3630) | Fyne doesn&#x27;t recognize my screens DPI | Графика шрифты и масштабирование | OS:Windows, unverified |
| [#3623](https://github.com/fyne-io/fyne/issues/3623) | android sharing/receive  | Платформы окна и интеграция с ОС | нет |
| [#3603](https://github.com/fyne-io/fyne/issues/3603) | Fyne app freezes if window is hidden while it&#x27;s in fullscreen | Стабильность и потоки | unverified |
| [#4436](https://github.com/fyne-io/fyne/issues/4436) | Support Ubuntu Touch platform | Ввод фокус и навигация | Driver:mobile |
| [#3582](https://github.com/fyne-io/fyne/issues/3582) | Splash window on Sway still shows title | Платформы окна и интеграция с ОС | unverified, information-needed, Wayland |
| [#3543](https://github.com/fyne-io/fyne/issues/3543) | Remove slow deduplication on each refresh of CheckGroup | Производительность и память | bug, optimization |
| [#3537](https://github.com/fyne-io/fyne/issues/3537) | DPI/scaling on external monitor | Графика шрифты и масштабирование | OS:Linux, unverified |
| [#3536](https://github.com/fyne-io/fyne/issues/3536) | Ability to Quit Systray Application with Keybind  | Ввод фокус и навигация | нет |
| [#3506](https://github.com/fyne-io/fyne/issues/3506) | Call MultiLineEntry.SetText() with long text will result in no response | Виджеты компоновка и привязки данных | unverified |
| [#3499](https://github.com/fyne-io/fyne/issues/3499) | v2.3.0 Performance and Memory issues | Производительность и память | unverified |
| [#3490](https://github.com/fyne-io/fyne/issues/3490) | Additional label properties. | Виджеты компоновка и привязки данных | нет |
| [#3466](https://github.com/fyne-io/fyne/issues/3466) | Focusing a widget inside AppTab\|DocTabs.OnSelected doesn&#x27;t work | Ввод фокус и навигация | unverified |
| [#3465](https://github.com/fyne-io/fyne/issues/3465) | Random crash and stack trace &quot;signal arrived during external code execution&quot; | Стабильность и потоки | unverified |
| [#3445](https://github.com/fyne-io/fyne/issues/3445) | widget.Entry.InsertText | Виджеты компоновка и привязки данных | enhancement |
| [#3442](https://github.com/fyne-io/fyne/issues/3442) | Surface Scrolling events | Виджеты компоновка и привязки данных | нет |
| [#3434](https://github.com/fyne-io/fyne/issues/3434) | Hiding a widget will let empty space ; Layout is not recalculated (until Refresh() is called on parent container) | Виджеты компоновка и привязки данных | bug |
| [#3427](https://github.com/fyne-io/fyne/issues/3427) | Table size iteration could be faster | Производительность и память | bug, optimization |
| [#3412](https://github.com/fyne-io/fyne/issues/3412) | Add ability to use vector graphics for real time animation | Графика шрифты и масштабирование | enhancement |
| [#3401](https://github.com/fyne-io/fyne/issues/3401) | Drastic increase in memory usage on develop | Производительность и память | bug, optimization |
| [#3378](https://github.com/fyne-io/fyne/issues/3378) | Use Flatpak portals for notifications | Платформы окна и интеграция с ОС | enhancement, OS:Linux |
| [#3377](https://github.com/fyne-io/fyne/issues/3377) | Use Flatpak portals for file handling | Платформы окна и интеграция с ОС | enhancement, OS:Linux |
| [#3363](https://github.com/fyne-io/fyne/issues/3363) | CanvasObject in AccordionItem&#x27;s Title | Виджеты компоновка и привязки данных | нет |
| [#3360](https://github.com/fyne-io/fyne/issues/3360) | NewMultiLineEntry component crash problem | Стабильность и потоки | unverified, information-needed |
| [#3342](https://github.com/fyne-io/fyne/issues/3342) | Running app in C:\Windows\System32 gives the error `rundll32 resolves to executable in current directory (.\rundll32.exe)` | Файлы диалоги и хранилище | bug, Hacktoberfest |
| [#3339](https://github.com/fyne-io/fyne/issues/3339) | Android - Entry - it is impossible to fire OnSubmitted on MultiLineEntry | Платформы окна и интеграция с ОС | unverified |
| [#3329](https://github.com/fyne-io/fyne/issues/3329) | Add validator Function and/or Required Flag to Select Widget | Виджеты компоновка и привязки данных | enhancement |
| [#3324](https://github.com/fyne-io/fyne/issues/3324) | Add VHResizeCursor and GrapCursor | Ввод фокус и навигация | enhancement, Driver:desktop |
| [#3316](https://github.com/fyne-io/fyne/issues/3316) | Unfocus using escape key  | Доступность | enhancement, Keyboard, Accessibility |
| [#3293](https://github.com/fyne-io/fyne/issues/3293) | crash in widget.(*RichText).insertAt | Стабильность и потоки | unverified |
| [#3278](https://github.com/fyne-io/fyne/issues/3278) | Calling `SetValidationError` does not work on new Entry. | Виджеты компоновка и привязки данных | unverified |
| [#3256](https://github.com/fyne-io/fyne/issues/3256) | Access to (FTDI) USB serial device (on Android) | Платформы окна и интеграция с ОС | нет |
| [#3255](https://github.com/fyne-io/fyne/issues/3255) | Android always returns true from CanWrite | Платформы окна и интеграция с ОС | bug, OS:Android |
| [#3242](https://github.com/fyne-io/fyne/issues/3242) | Go version upgrade checklist | Сборка зависимости и установка | enhancement, future, meta |
| [#3212](https://github.com/fyne-io/fyne/issues/3212) | NewCheckWithData().OnChanged not synchronized with its binding data | Стабильность и потоки | unverified |
| [#3197](https://github.com/fyne-io/fyne/issues/3197) | Can open hided window from macOS dock with no response | Платформы окна и интеграция с ОС | bug, help wanted, OS:macOS |
| [#3196](https://github.com/fyne-io/fyne/issues/3196) | table.ScrollToBottom shows list glitch when called from OnSelected | Виджеты компоновка и привязки данных | bug |
| [#3183](https://github.com/fyne-io/fyne/issues/3183) | add code widget that highlights syntax and shows line numbers etc | Виджеты компоновка и привязки данных | enhancement, future |
| [#3182](https://github.com/fyne-io/fyne/issues/3182) | Annoying select problem for list or tree widget | Виджеты компоновка и привязки данных | unverified |
| [#3181](https://github.com/fyne-io/fyne/issues/3181) | Entry OnValidationChanged behavior | Стабильность и потоки | unverified |
| [#3160](https://github.com/fyne-io/fyne/issues/3160) | Tree item bad indentation when changing level of an item | Стабильность и потоки | bug |
| [#3156](https://github.com/fyne-io/fyne/issues/3156) | Hide the icon in taskbar on Windows or hide the icon in Dockbar on Mac | Платформы окна и интеграция с ОС | enhancement, OS:macOS, OS:Linux |
| [#3149](https://github.com/fyne-io/fyne/issues/3149) | Auto Complete on SelectEntry | Виджеты компоновка и привязки данных | enhancement |
| [#3148](https://github.com/fyne-io/fyne/issues/3148) | Documentation addition about testing bindings | Виджеты компоновка и привязки данных | нет |
| [#3146](https://github.com/fyne-io/fyne/issues/3146) | Pressing Alt while a modal PopUp is displayed should not show the window&#x27;s MainMenu | Платформы окна и интеграция с ОС | bug |
| [#3114](https://github.com/fyne-io/fyne/issues/3114) | SVG Image Rendering/ Refresh too slow e.g. vs. Web browser | Производительность и память | unverified |
| [#3095](https://github.com/fyne-io/fyne/issues/3095) | Support to control visibility of system tray MenuItems | Платформы окна и интеграция с ОС | нет |
| [#3090](https://github.com/fyne-io/fyne/issues/3090) | Font corruption for Google&#x27;s WorkSans-Black | Графика шрифты и масштабирование | bug |
| [#3079](https://github.com/fyne-io/fyne/issues/3079) | Notification blocked by user security software on Windows | Платформы окна и интеграция с ОС | OS:Windows |
| [#3063](https://github.com/fyne-io/fyne/issues/3063) | Unable to run application twice, even after quitting first one | API документация и прочие запросы | bug |
| [#3054](https://github.com/fyne-io/fyne/issues/3054) | Select widget: changing selection with keyboard (Up/Down) not scrolling | Стабильность и потоки | bug |
| [#3044](https://github.com/fyne-io/fyne/issues/3044) | Add app skeleton container | Виджеты компоновка и привязки данных | enhancement, Design |
| [#3043](https://github.com/fyne-io/fyne/issues/3043) | Merge fyne-cross into Fyne command as &quot;cross&quot; sub-command | API документация и прочие запросы | enhancement |
| [#3038](https://github.com/fyne-io/fyne/issues/3038) | Propagate Shortcut and KeyPress events to parent Canvas when a Widget is focused and it&#x27;s not handling the Key/Shortcut | Ввод фокус и навигация | нет |
| [#3032](https://github.com/fyne-io/fyne/issues/3032) | Add option to hide or show file extension in filedialog | Файлы диалоги и хранилище | enhancement, FileDialog |
| [#3029](https://github.com/fyne-io/fyne/issues/3029) | Method Scroll.MakeChildVisible(CanvasObject) | Виджеты компоновка и привязки данных | нет |
| [#3016](https://github.com/fyne-io/fyne/issues/3016) | Vet and cleanup module dependencies | Сборка зависимости и установка | enhancement, meta |
| [#2999](https://github.com/fyne-io/fyne/issues/2999) | Switching window with opened MainMenu will steal focus visual indication from other Widget | Ввод фокус и навигация | unverified |
| [#2997](https://github.com/fyne-io/fyne/issues/2997) | Keyboard navigation: possibility to activate buttons (or other activable widgets) with Enter (not only with Spacebar) | Ввод фокус и навигация | enhancement |
| [#2993](https://github.com/fyne-io/fyne/issues/2993) | Improve layout function | Виджеты компоновка и привязки данных | enhancement, future |
| [#2982](https://github.com/fyne-io/fyne/issues/2982) | Consistent user-faced fyne error reporting | API документация и прочие запросы | нет |
| [#2981](https://github.com/fyne-io/fyne/issues/2981) | FileDialog NewFileSave can&#x27;t create folder | Файлы диалоги и хранилище | нет |
| [#2964](https://github.com/fyne-io/fyne/issues/2964) | Issues with Validation | Виджеты компоновка и привязки данных | нет |
| [#2960](https://github.com/fyne-io/fyne/issues/2960) | Improvement on handling image decoding inside Fyne | Производительность и память | enhancement, Design, optimization |
| [#2935](https://github.com/fyne-io/fyne/issues/2935) | Customize the icon and name of the notification message | Платформы окна и интеграция с ОС | duplicate |
| [#2934](https://github.com/fyne-io/fyne/issues/2934) | FileDialog hangs on directories if contains hundreds of Subdirectories | Стабильность и потоки | unverified |
| [#2926](https://github.com/fyne-io/fyne/issues/2926) | How to change branchIcon in tree widget ? | Стабильность и потоки | нет |
| [#2923](https://github.com/fyne-io/fyne/issues/2923) | Listener is called too many times | API документация и прочие запросы | unverified, data binding |
| [#2917](https://github.com/fyne-io/fyne/issues/2917) | func (s *Scroll) ScrollToBottom() not scrolling | Виджеты компоновка и привязки данных | unverified |
| [#2912](https://github.com/fyne-io/fyne/issues/2912) | Increase paragraph padding in RichText widget | Виджеты компоновка и привязки данных | нет |
| [#2902](https://github.com/fyne-io/fyne/issues/2902) | [Cross-domain integration] Fyne CLI should support exporting the native sourcebase template into Android studio &amp;&amp; Xcode projects | Платформы окна и интеграция с ОС | нет |
| [#2897](https://github.com/fyne-io/fyne/issues/2897) | Font scale doesn&#x27;t respect Gnome settigs | Графика шрифты и масштабирование | нет |
| [#2891](https://github.com/fyne-io/fyne/issues/2891) | Add a way to hide the phone&#x27;s status bar on mobile versions of fyne | Сборка зависимости и установка | enhancement, future |
| [#2883](https://github.com/fyne-io/fyne/issues/2883) | Calling Bind on a widget clobbers the Validator and OnChanged fields, and this behavior is not documented clearly or at all. | Стабильность и потоки | bug |
| [#2876](https://github.com/fyne-io/fyne/issues/2876) | [Wayland] Touch on touch-panel is not working within the application | Ввод фокус и навигация | bug, OS:Linux, Wayland |
| [#2867](https://github.com/fyne-io/fyne/issues/2867) | label text stop updating after a few seconds | Виджеты компоновка и привязки данных | unverified |
| [#2865](https://github.com/fyne-io/fyne/issues/2865) | Support for Metal on MacOS/iOS | Платформы окна и интеграция с ОС | future |
| [#2842](https://github.com/fyne-io/fyne/issues/2842) | App, Window and System tray icons do not update with theme | Платформы окна и интеграция с ОС | bug |
| [#2840](https://github.com/fyne-io/fyne/issues/2840) | Sequential animations | API документация и прочие запросы | нет |
| [#2837](https://github.com/fyne-io/fyne/issues/2837) | Wait for a form to be validated and Submit to be clicked to propagate all the data binding change. | Стабильность и потоки | data binding |
| [#2836](https://github.com/fyne-io/fyne/issues/2836) | Add data binding support to Select | Виджеты компоновка и привязки данных | data binding |
| [#2834](https://github.com/fyne-io/fyne/issues/2834) | Add data binding support to RadioGroup | Виджеты компоновка и привязки данных | data binding |
| [#2833](https://github.com/fyne-io/fyne/issues/2833) | Control the text of Check via data binding | Виджеты компоновка и привязки данных | data binding |
| [#2832](https://github.com/fyne-io/fyne/issues/2832) | Control Enable/Disable from a data binding | Виджеты компоновка и привязки данных | data binding |
| [#2820](https://github.com/fyne-io/fyne/issues/2820) | Form with multiple columns | Виджеты компоновка и привязки данных | enhancement |
| [#2815](https://github.com/fyne-io/fyne/issues/2815) | For Android, does this support &quot;Custom Tabs&quot; for oidc? | Платформы окна и интеграция с ОС | question, information-needed |
| [#2809](https://github.com/fyne-io/fyne/issues/2809) | Window on more than one monitor | Платформы окна и интеграция с ОС | question |
| [#2803](https://github.com/fyne-io/fyne/issues/2803) | Improve web backend performance by using inkwasm | Производительность и память | enhancement, Driver:web |
| [#2790](https://github.com/fyne-io/fyne/issues/2790) | android intent call and read response | Платформы окна и интеграция с ОС | enhancement |
| [#2788](https://github.com/fyne-io/fyne/issues/2788) | After lots of rows table shows separators in wrong position | Виджеты компоновка и привязки данных | bug |
| [#2777](https://github.com/fyne-io/fyne/issues/2777) | Chinese input method string truncated when typing into `widget.NewEntryWithData&#x27;  | Ввод фокус и навигация | unverified |
| [#2769](https://github.com/fyne-io/fyne/issues/2769) | Svg is not displayed | Графика шрифты и масштабирование | unverified |
| [#2766](https://github.com/fyne-io/fyne/issues/2766) | some entry problem | Виджеты компоновка и привязки данных | нет |
| [#2759](https://github.com/fyne-io/fyne/issues/2759) | Single Line Entry Shows Multiple Lines | Виджеты компоновка и привязки данных | нет |
| [#2743](https://github.com/fyne-io/fyne/issues/2743) | Handle context loss with the web driver | Платформы окна и интеграция с ОС | bug, Driver:web |
| [#2742](https://github.com/fyne-io/fyne/issues/2742) | windows virtual keyboard can not work with widget.Entry | Ввод фокус и навигация | unverified |
| [#2739](https://github.com/fyne-io/fyne/issues/2739) | Full mobile support for the web driver and desktop driver. | Платформы окна и интеграция с ОС | blocker, Driver:desktop, Driver:web |
| [#2738](https://github.com/fyne-io/fyne/issues/2738) | Handle file saving dialog as a file download in the web driver | Платформы окна и интеграция с ОС | Driver:web |
| [#2737](https://github.com/fyne-io/fyne/issues/2737) | Add file open dialog for the web driver | Файлы диалоги и хранилище | Driver:web |
| [#2730](https://github.com/fyne-io/fyne/issues/2730) | Rune is not handled on data binding functions of all input components | Ввод фокус и навигация | unverified |
| [#2721](https://github.com/fyne-io/fyne/issues/2721) | Complete support for PinePhone | API документация и прочие запросы | OS:Linux, Driver:mobile |
| [#2712](https://github.com/fyne-io/fyne/issues/2712) | ThemedResource documentation answer the 101 | API документация и прочие запросы | нет |
| [#2699](https://github.com/fyne-io/fyne/issues/2699) | List with databinding has issues when the list gets too long and the user is scrolling fast | Виджеты компоновка и привязки данных | bug |
| [#2680](https://github.com/fyne-io/fyne/issues/2680) | Adapter from Repository to fs.FS | Файлы диалоги и хранилище | enhancement, Storage |
| [#2656](https://github.com/fyne-io/fyne/issues/2656) | Support Image hyperlinks | Графика шрифты и масштабирование | information-needed |
| [#2645](https://github.com/fyne-io/fyne/issues/2645) | The macOS application packaged in the M1 MacBook cannot be installed on other computers | Сборка зависимости и установка | unverified, information-needed |
| [#2637](https://github.com/fyne-io/fyne/issues/2637) | Add support for building without bundled fonts | Сборка зависимости и установка | enhancement |
| [#2636](https://github.com/fyne-io/fyne/issues/2636) | Don&#x27;t bundle fonts where possible | Графика шрифты и масштабирование | enhancement |
| [#2635](https://github.com/fyne-io/fyne/issues/2635) | Support for Android Auto and Apple CarPlay | Платформы окна и интеграция с ОС | future, OS:Android, OS:iOS |
| [#2632](https://github.com/fyne-io/fyne/issues/2632) | Odd behaviour with canvas.Circle  | API документация и прочие запросы | information-needed |
| [#2629](https://github.com/fyne-io/fyne/issues/2629) | Add support for markdown tables | Виджеты компоновка и привязки данных | нет |
| [#2618](https://github.com/fyne-io/fyne/issues/2618) | Support an editable variant of TextGrid | Виджеты компоновка и привязки данных | enhancement |
| [#2607](https://github.com/fyne-io/fyne/issues/2607) | Struct binding fails to reload with slice field | Виджеты компоновка и привязки данных | bug |
| [#2599](https://github.com/fyne-io/fyne/issues/2599) | Binding &quot;Must*&quot; variants | Виджеты компоновка и привязки данных | нет |
| [#2596](https://github.com/fyne-io/fyne/issues/2596) | Fyne CLI - Inconsistent behavior with --src flag | API документация и прочие запросы | bug |
| [#2576](https://github.com/fyne-io/fyne/issues/2576) | Add custom content to system tray support (maybe) | Платформы окна и интеграция с ОС | enhancement |
| [#2575](https://github.com/fyne-io/fyne/issues/2575) | Add multiple selection | Ввод фокус и навигация | нет |
| [#2574](https://github.com/fyne-io/fyne/issues/2574) | Add global menu support on Linux/BSD | Виджеты компоновка и привязки данных | enhancement, OS:Linux, OS:FreeBSD |
| [#2565](https://github.com/fyne-io/fyne/issues/2565) | Should Unbind() guarantee that widgets stop responding to data changes? | Стабильность и потоки | data binding |
| [#2554](https://github.com/fyne-io/fyne/issues/2554) | Horizontal NewRadioGroup text width not adjusting | Виджеты компоновка и привязки данных | enhancement |
| [#2551](https://github.com/fyne-io/fyne/issues/2551) | Extending a Tappable with DoubleTappable causes a loss of Tappability for containing widget | Ввод фокус и навигация | bug |
| [#2515](https://github.com/fyne-io/fyne/issues/2515) | [cmd/fyne] update macOS package to specify a target architecture | Сборка зависимости и установка | нет |
| [#2506](https://github.com/fyne-io/fyne/issues/2506) | An idle application with no visible windows still uses noticeable CPU time | Производительность и память | bug, optimization |
| [#2491](https://github.com/fyne-io/fyne/issues/2491) | WatchableRepository | Файлы диалоги и хранилище | enhancement, Storage |
| [#2479](https://github.com/fyne-io/fyne/issues/2479) | SearchableRepository | Файлы диалоги и хранилище | enhancement, Storage |
| [#2472](https://github.com/fyne-io/fyne/issues/2472) | Window double close cause panic | Стабильность и потоки | bug, Driver:mobile |
| [#2470](https://github.com/fyne-io/fyne/issues/2470) | Custom widgets implementing TappedSecondary() are not selectable in widget.List | Ввод фокус и навигация | bug |
| [#2468](https://github.com/fyne-io/fyne/issues/2468) | Can u support MacOS Two fingers swipe with event | API документация и прочие запросы | question, information-needed |
| [#2464](https://github.com/fyne-io/fyne/issues/2464) | Rich text cannot switch from scrolling after initially shown | Виджеты компоновка и привязки данных | bug, enhancement |
| [#2455](https://github.com/fyne-io/fyne/issues/2455) | Control+Escape key combination on Mac OS is filtered out by GLFW | Ввод фокус и навигация | bug, OS:macOS |
| [#2454](https://github.com/fyne-io/fyne/issues/2454) | Support mobile simulator on Windows | Платформы окна и интеграция с ОС | enhancement, OS:Windows, mobile-simulator |
| [#2453](https://github.com/fyne-io/fyne/issues/2453) | Table Widget fill space mode | Виджеты компоновка и привязки данных | enhancement |
| [#2438](https://github.com/fyne-io/fyne/issues/2438) | Add text to toolbar widget (positioned under icon) | Виджеты компоновка и привязки данных | enhancement |
| [#2418](https://github.com/fyne-io/fyne/issues/2418) | Android app does not show Chinese (default) keyboard | Ввод фокус и навигация | unverified |
| [#2399](https://github.com/fyne-io/fyne/issues/2399) | SIGSEGV: Fyne v2.0.4 on Pop OS panics using README example | Стабильность и потоки | unverified |
| [#2373](https://github.com/fyne-io/fyne/issues/2373) | panic: FormatUnavailable: NSGL: Failed to find a suitable pixel format | Стабильность и потоки | OS:macOS, unverified |
| [#2372](https://github.com/fyne-io/fyne/issues/2372) | iOS: Old devices don&#x27;t reserve space for status bar | Платформы окна и интеграция с ОС | bug, OS:iOS |
| [#2365](https://github.com/fyne-io/fyne/issues/2365) | Add support for temporary notifications (like Android Toast) | Платформы окна и интеграция с ОС | enhancement, information-needed |
| [#2352](https://github.com/fyne-io/fyne/issues/2352) | Fyne API storage implement io.Seeker (mainly for mobile) | Файлы диалоги и хранилище | enhancement |
| [#2327](https://github.com/fyne-io/fyne/issues/2327) | app can&#x27;t lunch on android 4.1.2 jelly bean (api 16) | Платформы окна и интеграция с ОС | unverified |
| [#2325](https://github.com/fyne-io/fyne/issues/2325) | Feature Request: Spell check in MultiLineEntry | Виджеты компоновка и привязки данных | enhancement |
| [#2314](https://github.com/fyne-io/fyne/issues/2314) | Fyne application doesn&#x27;t terminate after all windows are closed. | Платформы окна и интеграция с ОС | information-needed |
| [#2313](https://github.com/fyne-io/fyne/issues/2313) | [develop] Select with &lt;=3 options may show a blank option at the tail when hover the popup options list | Виджеты компоновка и привязки данных | bug |
| [#2309](https://github.com/fyne-io/fyne/issues/2309) | widgets with scroll may stop responding to mouse | Ввод фокус и навигация | bug |
| [#2304](https://github.com/fyne-io/fyne/issues/2304) | add feature to hook global key bindings into (desktop?) OS | Ввод фокус и навигация | enhancement |
| [#2287](https://github.com/fyne-io/fyne/issues/2287) | Scroll Image output to file can be truncated when image is not aligned on top left | Графика шрифты и масштабирование | bug |
| [#2285](https://github.com/fyne-io/fyne/issues/2285) | crash-when-original-fill-used-fresh | Стабильность и потоки | bug |
| [#2274](https://github.com/fyne-io/fyne/issues/2274) | Resizing with tab dropdown open causes issues | Виджеты компоновка и привязки данных | bug |
| [#2271](https://github.com/fyne-io/fyne/issues/2271) | Allow arrow and TAB navigation in the file/folder dialogs  | Файлы диалоги и хранилище | enhancement, Hacktoberfest, FileDialog |
| [#2259](https://github.com/fyne-io/fyne/issues/2259) | Add StringList.Insert method | API документация и прочие запросы | enhancement |
| [#2233](https://github.com/fyne-io/fyne/issues/2233) | Add time infos in the storage package | Сборка зависимости и установка | enhancement, Storage |
| [#2228](https://github.com/fyne-io/fyne/issues/2228) | [Feature] Chart/Graph widget | Виджеты компоновка и привязки данных | question |
| [#2215](https://github.com/fyne-io/fyne/issues/2215) | Add form button to clear content of form widgtes/fields | Виджеты компоновка и привязки данных | нет |
| [#2206](https://github.com/fyne-io/fyne/issues/2206) | Text is squiggly when writing in any input field | Ввод фокус и навигация | bug |
| [#2202](https://github.com/fyne-io/fyne/issues/2202) | 3.0: Removal of deprecated APIs | API документация и прочие запросы | meta |
| [#2173](https://github.com/fyne-io/fyne/issues/2173) | Unrelated dialogs should not be dismissed | Виджеты компоновка и привязки данных | bug |
| [#2164](https://github.com/fyne-io/fyne/issues/2164) | Drop-down menu with many items is slow | Производительность и память | bug, optimization |
| [#2151](https://github.com/fyne-io/fyne/issues/2151) | Dynamic Form&#x27;s | Виджеты компоновка и привязки данных | enhancement |
| [#2141](https://github.com/fyne-io/fyne/issues/2141) | Android: menu icon overwrites first tab item text | Платформы окна и интеграция с ОС | нет |
| [#2134](https://github.com/fyne-io/fyne/issues/2134) | storage: Support for getting an os.FileInfo (or similar) from URI | Файлы диалоги и хранилище | enhancement, Storage |
| [#2099](https://github.com/fyne-io/fyne/issues/2099) | focus-responsive Edit Menu (for use in MainMenu) | Ввод фокус и навигация | нет |
| [#2098](https://github.com/fyne-io/fyne/issues/2098) | iOS Save dialog exporting into DropBox app does not seem to work | Платформы окна и интеграция с ОС | bug |
| [#2083](https://github.com/fyne-io/fyne/issues/2083) | iPhone 12 mini cannot tap on widgets in right-most 12dp elements | Ввод фокус и навигация | bug, OS:iOS, Driver:mobile |
| [#2066](https://github.com/fyne-io/fyne/issues/2066) | Android combining character input (using MessagEase) doesn&#x27;t work correctly | Ввод фокус и навигация | bug, Internationalisation |
| [#1967](https://github.com/fyne-io/fyne/issues/1967) | Form not rendering correct Items | Графика шрифты и масштабирование | bug |
| [#1950](https://github.com/fyne-io/fyne/issues/1950) | Support fill set by groups in SVG | Графика шрифты и масштабирование | нет |
| [#1949](https://github.com/fyne-io/fyne/issues/1949) | Support SVG Polyline | Графика шрифты и масштабирование | нет |
| [#1927](https://github.com/fyne-io/fyne/issues/1927) | mobile app simulator freezes if move to external screen (macOS) | Стабильность и потоки | bug |
| [#1914](https://github.com/fyne-io/fyne/issues/1914) | [Question/Suggestion for docs] widget.Tree Nodes with the same name/uuid | Виджеты компоновка и привязки данных | question |
| [#1883](https://github.com/fyne-io/fyne/issues/1883) | Repository Bindings | Файлы диалоги и хранилище | Storage |
| [#1853](https://github.com/fyne-io/fyne/issues/1853) | binding.StringListList | Виджеты компоновка и привязки данных | enhancement, data binding |
| [#1849](https://github.com/fyne-io/fyne/issues/1849) | BindingData in Entry should not modify the content while user is typing | Виджеты компоновка и привязки данных | bug |
| [#1847](https://github.com/fyne-io/fyne/issues/1847) | Register URL Protocol Handler | API документация и прочие запросы | enhancement |
| [#1845](https://github.com/fyne-io/fyne/issues/1845) | Device Orientation Hints / Lock | API документация и прочие запросы | question |
| [#1844](https://github.com/fyne-io/fyne/issues/1844) | Native Widgets [Ads, Monetization Libraries, system widgets, etc.] | Виджеты компоновка и привязки данных | нет |
| [#1841](https://github.com/fyne-io/fyne/issues/1841) | `test.AbsolutePositionForObject` not finding some widgets | Тесты и диагностика | bug, information-needed |
| [#1834](https://github.com/fyne-io/fyne/issues/1834) | Add storage.Walk as a uri-based filepath.Walk alternative | Файлы диалоги и хранилище | Storage |
| [#1800](https://github.com/fyne-io/fyne/issues/1800) | Color Pipette in Picker  | Графика шрифты и масштабирование | enhancement, future |
| [#1776](https://github.com/fyne-io/fyne/issues/1776) | META: Improvements to button tap animation | Ввод фокус и навигация | enhancement, Design |
| [#1775](https://github.com/fyne-io/fyne/issues/1775) | Add option to turn off HintText for validator | API документация и прочие запросы | нет |
| [#1743](https://github.com/fyne-io/fyne/issues/1743) | High contrast theme variant | Доступность | enhancement, Design, Accessibility |
| [#1742](https://github.com/fyne-io/fyne/issues/1742) | Highlight interactible widgets | Доступность | enhancement, Design, Accessibility |
| [#1741](https://github.com/fyne-io/fyne/issues/1741) | Visualization Colors | Доступность | enhancement, Design, Accessibility |
| [#1705](https://github.com/fyne-io/fyne/issues/1705) | Specify Keystore Password via File | Сборка зависимости и установка | нет |
| [#1704](https://github.com/fyne-io/fyne/issues/1704) | Data Binding Validation | Виджеты компоновка и привязки данных | Hacktoberfest, data binding |
| [#1674](https://github.com/fyne-io/fyne/issues/1674) | Improve alignment of colour sliders and submenu arrows | Графика шрифты и масштабирование | нет |
| [#1672](https://github.com/fyne-io/fyne/issues/1672) | Improve precision of markup test XML | Тесты и диагностика | нет |
| [#1660](https://github.com/fyne-io/fyne/issues/1660) | Text Wrapping Hyphenation | API документация и прочие запросы | enhancement |
| [#1650](https://github.com/fyne-io/fyne/issues/1650) | Add tooltip | API документация и прочие запросы | enhancement, Design |
| [#1648](https://github.com/fyne-io/fyne/issues/1648) | Searchable menus | Виджеты компоновка и привязки данных | enhancement, Design, Hacktoberfest |
| [#1646](https://github.com/fyne-io/fyne/issues/1646) | Support to deliver over the web :) | API документация и прочие запросы | blocker |
| [#1625](https://github.com/fyne-io/fyne/issues/1625) | mobile driver behaviour lagging desktop driver | Производительность и память | нет |
| [#1623](https://github.com/fyne-io/fyne/issues/1623) | Text RTL and locale hinting for Unicode strings | Локализация и языки | Internationalisation |
| [#1603](https://github.com/fyne-io/fyne/issues/1603) | Gnome/XWayland : Crash on standby/resume | Стабильность и потоки | bug, help wanted, Hacktoberfest |
| [#1590](https://github.com/fyne-io/fyne/issues/1590) | Fyne&#x27;s handling of keys doesn&#x27;t allow implementation of a remote control client | Ввод фокус и навигация | нет |
| [#1570](https://github.com/fyne-io/fyne/issues/1570) | Resize ignored in popup containing wrapped text. | Виджеты компоновка и привязки данных | question |
| [#1566](https://github.com/fyne-io/fyne/issues/1566) | Change Title/Message in a Dialog | Стабильность и потоки | enhancement, Hacktoberfest |
| [#1515](https://github.com/fyne-io/fyne/issues/1515) | Make keyboard controls a first class citizen | Ввод фокус и навигация | Keyboard |
| [#1493](https://github.com/fyne-io/fyne/issues/1493) | Accordion Event on Open/Close | Виджеты компоновка и привязки данных | enhancement, Hacktoberfest |
| [#1395](https://github.com/fyne-io/fyne/issues/1395) | Consistent Selection API | Ввод фокус и навигация | нет |
| [#1391](https://github.com/fyne-io/fyne/issues/1391) | Get current state of CapsLock key? | Ввод фокус и навигация | Keyboard |
| [#1384](https://github.com/fyne-io/fyne/issues/1384) | Also support GNUstep .app bundles | API документация и прочие запросы | enhancement, OS:Linux, OS:FreeBSD |
| [#1333](https://github.com/fyne-io/fyne/issues/1333) | Bind keybindings to exit application by default | Ввод фокус и навигация | Keyboard, information-needed |
| [#1289](https://github.com/fyne-io/fyne/issues/1289) | Add .SetText() method to tab containers | Виджеты компоновка и привязки данных | enhancement |
| [#1268](https://github.com/fyne-io/fyne/issues/1268) | META: List of validators to consider adding | Виджеты компоновка и привязки данных | meta |
| [#1265](https://github.com/fyne-io/fyne/issues/1265) | How to let mac open a windows ask for administrator privileges | Платформы окна и интеграция с ОС | question, future |
| [#1246](https://github.com/fyne-io/fyne/issues/1246) | Smoother anti-aliasing on canvas.Line | API документация и прочие запросы | enhancement |
| [#1238](https://github.com/fyne-io/fyne/issues/1238) | Inter-operate with platform-specific &quot;favorite directories&quot; | Файлы диалоги и хранилище | enhancement, Hacktoberfest, FileDialog |
| [#1229](https://github.com/fyne-io/fyne/issues/1229) | Move to outline icons | API документация и прочие запросы | Design |
| [#1217](https://github.com/fyne-io/fyne/issues/1217) | Documentation about building iOS without Apple developer account | Сборка зависимости и установка | question, documentation |
| [#1184](https://github.com/fyne-io/fyne/issues/1184) | WSL: Window does not fully render on startup after updating to fyne 1.3.1 or 1.3.2 | Графика шрифты и масштабирование | bug, OS:Windows |
| [#1171](https://github.com/fyne-io/fyne/issues/1171) | Cannot call the right-click menu continuously | Виджеты компоновка и привязки данных | information-needed |
| [#1160](https://github.com/fyne-io/fyne/issues/1160) | XLib warning printed to terminal on application close | API документация и прочие запросы | OS:Linux, information-needed |
| [#1157](https://github.com/fyne-io/fyne/issues/1157) | Proposal: Move Window.Resize() and Window.CenterOnScreen() to desktop package | Сборка зависимости и установка | нет |
| [#1135](https://github.com/fyne-io/fyne/issues/1135) | fyne_demo slow on PowerShell with administrative privileges | Производительность и память | bug, help wanted, OS:Windows |
| [#1127](https://github.com/fyne-io/fyne/issues/1127) | Fyne package missing some Android options | Сборка зависимости и установка | bug |
| [#1121](https://github.com/fyne-io/fyne/issues/1121) | Add option for tabs to fill the tab bar | Виджеты компоновка и привязки данных | information-needed |
| [#1114](https://github.com/fyne-io/fyne/issues/1114) | Lag and dark flashing when resizing | Производительность и память | bug |
| [#1113](https://github.com/fyne-io/fyne/issues/1113) | Add 2D transforms | Виджеты компоновка и привязки данных | information-needed |
| [#1103](https://github.com/fyne-io/fyne/issues/1103) | Black flicker when launching application on external display on Thinkpad T430 | Графика шрифты и масштабирование | bug, help wanted, OS:Linux |
| [#1098](https://github.com/fyne-io/fyne/issues/1098) | Text-based path input in file browser | Ввод фокус и навигация | enhancement, FileDialog |
| [#1097](https://github.com/fyne-io/fyne/issues/1097) | Large APK Size | API документация и прочие запросы | нет |
| [#1082](https://github.com/fyne-io/fyne/issues/1082) | Support selecting multiple files at once | Файлы диалоги и хранилище | enhancement, FileDialog |
| [#1065](https://github.com/fyne-io/fyne/issues/1065) | ScrollContainer on RPI touchscreen | Ввод фокус и навигация | нет |
| [#1062](https://github.com/fyne-io/fyne/issues/1062) | Android apps are lagging when enter text and scrolling | Производительность и память | bug, OS:Android, Bounty |
| [#1044](https://github.com/fyne-io/fyne/issues/1044) | Add Validation For Saving Files From Save Dialog With Filter Enabled | Файлы диалоги и хранилище | enhancement, FileDialog |
| [#1024](https://github.com/fyne-io/fyne/issues/1024) | Mac Function Key InvalidValue: Invalid scancode | Ввод фокус и навигация | upstream |
| [#964](https://github.com/fyne-io/fyne/issues/964) | Add an OnIdle handler to the app object | API документация и прочие запросы | нет |
| [#936](https://github.com/fyne-io/fyne/issues/936) | Android notification uses fallback icon instead of app icon | Платформы окна и интеграция с ОС | bug, Hacktoberfest |
| [#920](https://github.com/fyne-io/fyne/issues/920) | When the target is a shared library, the windows UI object cannot be created for the second time | Платформы окна и интеграция с ОС | нет |
| [#915](https://github.com/fyne-io/fyne/issues/915) | fyne demo: fullscreen: second monitor flickers when entering/exiting full screen on main monitor  | Графика шрифты и масштабирование | bug, OS:Windows |
| [#911](https://github.com/fyne-io/fyne/issues/911) | Remove CGo on Windows | Сборка зависимости и установка | enhancement, OS:Windows |
| [#888](https://github.com/fyne-io/fyne/issues/888) | Add a date/time input widget | Ввод фокус и навигация | enhancement, Design, blocker, Hacktoberfest |
| [#843](https://github.com/fyne-io/fyne/issues/843) | Tidy Android / gomobile code | Производительность и память | optimization |
| [#837](https://github.com/fyne-io/fyne/issues/837) | Add NewTableWithStrings | Виджеты компоновка и привязки данных | enhancement |
| [#685](https://github.com/fyne-io/fyne/issues/685) | SVG Text Not Showing | Графика шрифты и масштабирование | bug, enhancement, Hacktoberfest |
| [#653](https://github.com/fyne-io/fyne/issues/653) | fyne package will not build Android applications when sources are not in same directory | Сборка зависимости и установка | bug |
| [#650](https://github.com/fyne-io/fyne/issues/650) | clipboard should support data that is not utf8 string | Ввод фокус и навигация | enhancement |
| [#618](https://github.com/fyne-io/fyne/issues/618) | The input method is located outside the app | Ввод фокус и навигация | Internationalisation, information-needed |
| [#585](https://github.com/fyne-io/fyne/issues/585) | SVG images using &quot;class&quot; to lookup style do not work | Графика шрифты и масштабирование | help wanted |
| [#564](https://github.com/fyne-io/fyne/issues/564) | Support for Tizen (Linux OS used for Samsung wearable) | API документация и прочие запросы | future |
| [#563](https://github.com/fyne-io/fyne/issues/563) | Draw canvas primitives using OpenGL? | Производительность и память | help wanted, optimization |
| [#554](https://github.com/fyne-io/fyne/issues/554) | Remember last window position (plus size and maximized state) | Платформы окна и интеграция с ОС | enhancement, Hacktoberfest |
| [#497](https://github.com/fyne-io/fyne/issues/497) | Drag and drop to application icon on macOS | Ввод фокус и навигация | enhancement, OS:macOS |
| [#449](https://github.com/fyne-io/fyne/issues/449) | Multimedia support for fyne | API документация и прочие запросы | нет |
| [#428](https://github.com/fyne-io/fyne/issues/428) | allow partial texture refresh of CanvasObject | Графика шрифты и масштабирование | enhancement, future |
| [#359](https://github.com/fyne-io/fyne/issues/359) | Release 1.1 - Double click + drag doesn&#x27;t select word and expand | Ввод фокус и навигация | bug |
| [#191](https://github.com/fyne-io/fyne/issues/191) | Windows dual GPU: invalid memory address or nil pointer dereference | Производительность и память | bug, OS:Windows, information-needed |
| [#181](https://github.com/fyne-io/fyne/issues/181) | Transparent window background | Графика шрифты и масштабирование | enhancement |
| [#142](https://github.com/fyne-io/fyne/issues/142) | Add drag and drop support | Ввод фокус и навигация | Design, Bounty |
| [#123](https://github.com/fyne-io/fyne/issues/123) | Propose new raster scroll/zoom widget | Виджеты компоновка и привязки данных | нет |
| [#90](https://github.com/fyne-io/fyne/issues/90) | Add support for Right to Left text | API документация и прочие запросы | future |
| [#47](https://github.com/fyne-io/fyne/issues/47) | Flexbox - like layout API | Виджеты компоновка и привязки данных | нет |
| [#4](https://github.com/fyne-io/fyne/issues/4) | Add AutoLayout (cassowary) option | Виджеты компоновка и привязки данных | enhancement, help wanted |

## fyne-io/apps

| Issue | Заголовок | Тема | Метки upstream |
| --- | --- | --- | --- |
| [#5](https://github.com/fyne-io/apps/issues/5) | permission failure on Linux / BSD installs | Сборка зависимости и установка | нет |
| [#4](https://github.com/fyne-io/apps/issues/4) | Add uninstall option | Сборка зависимости и установка | нет |
| [#2](https://github.com/fyne-io/apps/issues/2) | Install based on tags when available | Сборка зависимости и установка | нет |
