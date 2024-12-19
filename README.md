# audivo-media-downloader
A simple software to download media.

**Note:** This project is still under development. Download functionality is not available yet.


# Prerequisites

## For Apple Music Download
 - Python 3.8 or higher
 - The cookies file of your Apple Music browser session in Netscape format (requires an active subscription)
    - To export your cookies, use one of the following browser extensions while signed in to Apple Music:
        - Firefox: https://addons.mozilla.org/addon/export-cookies-txt
        - Chromium based browsers: https://chrome.google.com/webstore/detail/gdocmgbfkjnnpapoeobnolbbkoibbcif

 - FFmpeg on your system PATH
    - Older versions of FFmpeg may not work.
    - Up to date binaries can be obtained from the links below:
        - Windows: https://github.com/AnimMouse/ffmpeg-stable-autobuild/releases
        - Linux: https://johnvansickle.com/ffmpeg/
        - MacOS: `brew install ffmpeg`


## Used Libraries

- [wails](https://github.com/wailsapp/wails)
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [gamdl](https://github.com/glomatico/gamdl)
