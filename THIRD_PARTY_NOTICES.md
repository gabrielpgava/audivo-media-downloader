# Third-party notices

Audivo Media Downloader is distributed under GPL-3.0. This file identifies the principal third-party projects used by the source tree or expected in a packaged distribution. Each release must update this list together with its engine manifest and checksums.

## Application and build dependencies

| Project | Use | License/source |
| --- | --- | --- |
| Wails v2 | Desktop shell, bindings, native runtime | [MIT](https://github.com/wailsapp/wails/blob/v2.14.0/LICENSE) · [repository](https://github.com/wailsapp/wails/tree/v2.14.0) |
| React / React DOM | Frontend UI runtime | [MIT](https://github.com/facebook/react/blob/main/LICENSE) · [repository](https://github.com/facebook/react) |
| MUI, Emotion | UI components and styling | [MIT](https://github.com/mui/material-ui/blob/master/LICENSE) · [repository](https://github.com/mui/material-ui) |
| Go modules and npm packages | Transitive application/build dependencies | Versions are pinned in `go.sum` and `frontend/package-lock.json`; consult each package's license before redistribution. |

## Download engines and runtimes

| Project | Use | License/source |
| --- | --- | --- |
| yt-dlp | YouTube metadata and download engine | [Unlicense and release licensing notes](https://github.com/yt-dlp/yt-dlp/blob/master/README.md#licensing) · [source](https://github.com/yt-dlp/yt-dlp) |
| gamdl | Apple Music download CLI | [MIT](https://github.com/glomatico/gamdl/blob/master/LICENSE) · [source](https://github.com/glomatico/gamdl) |
| FFmpeg / ffprobe | Merge, remux and audio conversion | [LGPL-2.1+ by default, optional GPL components](https://github.com/FFmpeg/FFmpeg/blob/master/LICENSE.md) · [source](https://github.com/FFmpeg/FFmpeg) |
| Deno | JavaScript runtime used by yt-dlp EJS | [MIT](https://github.com/denoland/deno/blob/main/LICENSE.md) · [source](https://github.com/denoland/deno) |
| Python | Isolated runtime used to prepare gamdl | [PSF License](https://docs.python.org/3/license.html) · [source](https://www.python.org/) |
| uv (optional) | Local virtualenv/bootstrap accelerator | [MIT or Apache-2.0](https://github.com/astral-sh/uv/blob/main/LICENSE-APACHE) · [source](https://github.com/astral-sh/uv) |

The exact binary, archive, version and SHA-256 used by a release are defined by that release's engine manifest. Do not assume that a system-installed binary carries the same licensing composition as a packaged binary; FFmpeg builds may enable optional components, and yt-dlp release formats can include code under additional licenses.

## User-supplied cookies and media

Apple Music cookie files and downloaded media are user data, not Audivo third-party assets. They are never committed to this repository and are excluded by `.gitignore`.
