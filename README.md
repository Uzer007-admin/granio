# granio-image

Компонент независимого проекта **Granio — Гранио**.

**Исходный репозиторий:** [image](https://github.com/fyne-io/image).
**Цель форка:** Кодеки изображений для интерфейса и упаковки приложений Granio.

История изменений, лицензии и сведения об исходных авторах сохранены.
Go-модули переведены на собственные пути `github.com/Uzer007-admin/…`;
минимальная версия Go — **1.27.1**. Ядро и общие API подключаются из
[`granio/v2`](https://github.com/Uzer007-admin/granio).

Состояние переноса и результаты проверок приведены в
[реестре экосистемы](https://github.com/Uzer007-admin/granio/blob/dev/granio/docs/ECOSYSTEM.md).
Наличие форка само по себе не означает проверку всех платформ и сценариев.

---

## Документация исходного проекта

Команды установки и ссылки в сохранённом тексте ниже относятся к исходнику.
Для Granio используются новые пути модулей, перечисленные выше.

[![Tests](https://github.com/fyne-io/image/actions/workflows/tests.yml/badge.svg)](https://github.com/fyne-io/glfw-js/actions/workflows/tests.yml)
[![Static Analysis](https://github.com/fyne-io/image/actions/workflows/analysis.yml/badge.svg)](https://github.com/fyne-io/glfw-js/actions/workflows/analysis.yml)

# image

Collection of image codecs useful in GUI apps.

## .ico

Go support for the Windows [.ico](https://en.wikipedia.org/wiki/ICO_(file_format)) file format.

This format is a combination of work from [image2ascii](https://github.com/zyxar/image2ascii)
and [golang-image-ico](https://github.com/Kodeworks/golang-image-ico).

### Reader

Extracted from [image2ascii](https://github.com/zyxar/image2ascii), an excellent project but with many unrelated dependencies.

*License*: [Apache 2.0](http://opensource.org/licenses/Apache-2.0)

### Writer

From [golang-image-ico](https://github.com/Kodeworks/golang-image-ico).

License: BSD 3-clause (the same as the Go project).
