APP_NAME=audivo-media-downloader
BUILD_DIR=build/bin

all: windows windows-installer linux mac 

windows:
	wails build -platform windows/amd64 -obfuscated -o $(BUILD_DIR)/$(APP_NAME).exe

windows-installer:
	wails build -platform windows/amd64 -nsis -obfuscated -o $(BUILD_DIR)/$(APP_NAME).exe

linux:
	wails build -platform linux/amd64 -obfuscated -o $(BUILD_DIR)/$(APP_NAME)-linux

mac:
	wails build -platform darwin/amd64 -obfuscated -o $(APP_NAME)-macos

mac-arm:
	wails build -platform darwin/arm64 -obfuscated -o $(APP_NAME)-macos-arm

clean:
	rm -rf $(BUILD_DIR)/*

.PHONY: all windows linux darwin darwin-arm clean