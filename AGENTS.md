# AGENTS.md

This repository is a private Android-only fork of FlClash. Treat Android `arm64-v8a` as the only supported target.

## Project Scope

- Platform target: Android only
- ABI target: `arm64-v8a` only
- Flutter app code lives in `lib/`
- Android native project lives in `android/`
- Go core wrapper lives in `core/`
- Android Go shared library output lives in `libclash/android/arm64-v8a/`
- Local Flutter plugin build hook lives in `plugins/setup/`
- Wi-Fi SSID plugin lives in `plugins/wifi_ssid/`

Desktop platforms, desktop plugins, Rust helper IPC, system tray, desktop hotkeys, desktop system proxy, and distributor packaging have been removed.

## Local SDKs

| Tool | Path |
|------|------|
| Flutter SDK | `D:\Code\Tools\flutter` |
| Go SDK | `D:\Code\Tools\Go\go` |
| Android SDK | `D:\Code\Tools\Android\Sdk` |
| Android NDK | `D:\Code\Tools\Android\Sdk\ndk\28.2.13676358` |
| ADB | `D:\Code\Tools\Android\Sdk\platform-tools\adb.exe` |

## Environment

Use the local environment scripts before building:

```powershell
dev-env.bat
```

or in WSL:

```bash
source dev-env.sh
```

Important variables:

| Variable | Value |
|----------|-------|
| `GRADLE_USER_HOME` | `D:\Code\Clash myself\FlClash-dev\.dev-tools\gradle` |
| `GOPATH` | `D:\Code\Clash myself\FlClash-dev\.dev-tools\go-pkg` |
| `GOMODCACHE` | `D:\Code\Clash myself\FlClash-dev\.dev-tools\go-pkg\mod` |
| `PUB_CACHE` | `D:\Code\Clash myself\FlClash-dev\.dev-tools\pub-cache` |
| `ANDROID_HOME` | `D:\Code\Tools\Android\Sdk` |
| `ANDROID_NDK` | `D:\Code\Tools\Android\Sdk\ndk\28.2.13676358` |

## Common Commands

```powershell
flutter pub get
flutter analyze
flutter test
flutter build apk --debug --target-platform android-arm64
flutter build apk --release --target-platform android-arm64
```

The Android Gradle/buildkit setup defaults to `arm64-v8a`; use `--target-platform android-arm64` anyway so Flutter's own target selection is explicit.

Recommended local verification sequence for feature work:

```powershell
dev-env.bat
flutter test
flutter analyze
flutter build apk --debug --target-platform android-arm64
```

For Go core changes, also run:

```powershell
cd core
go test ./...
```

For a focused node batch detection check, run:

```powershell
flutter test test\views\profiles\media_check_test.dart
```

Install the latest debug APK on a connected Android device with:

```powershell
D:\Code\Tools\Android\Sdk\platform-tools\adb.exe devices
D:\Code\Tools\Android\Sdk\platform-tools\adb.exe install -r build\app\outputs\flutter-apk\app-debug.apk
```

`flutter analyze` may report existing Flutter deprecation `info` diagnostics. Treat new errors or warnings as blockers; do not treat the known deprecation info set as a failed build unless the task is to clean them up.

## Go Core Build

Android builds invoke `plugins/setup/buildkit/gradle/plugin.gradle`, which runs the Dart build tool in `plugins/setup/buildkit/build_tool/`.

The build tool now supports only:

```powershell
dart run build_tool android
dart run build_tool android --arch arm64
dart run build_tool android --target-platform android-arm64
```

It compiles the Go core as a CGO shared library and copies outputs to:

- `libclash/android/arm64-v8a/libclash.so`
- `android/core/src/main/jniLibs/arm64-v8a/libclash.so`
- `android/core/src/main/cpp/includes/arm64-v8a/`

## Code Generation

Run this after modifying generated models, providers, or Drift schema:

```powershell
dart run build_runner build --delete-conflicting-outputs
```

Generated code covers Riverpod providers, Freezed/JSON models, and Drift database files.

## Testing Notes

- Use `flutter test`, not `dart test`, because models and providers use Flutter types.
- Root tests live under `test/`.
- `CoreController.test(mock)` can inject a mocked `CoreHandlerInterface`.
- Call `CoreController.resetInstance()` in `tearDown` when testing the singleton.
- For nested Freezed model round trips, test through `jsonEncode`/`jsonDecode`.

## Node Batch Detection Notes

The node batch detection feature is intentionally scoped to the profiles/configuration area:

- UI entry and page live under `lib/views/profiles/`.
- Dart bridge changes live in `lib/core/controller.dart` and `lib/core/interface.dart`.
- Go core detection logic lives in `core/media_check.go`.
- Go request parameters are defined in `core/constant.go`.
- Tests live in `test/views/profiles/media_check_test.dart` and `core/media_check_test.go`.

Behavioral constraints:

- Opening the node detection page must not automatically start detection.
- Detection starts only from an explicit manual action, except health observation.
- One run should target one subscription and one function mode.
- Keep the three modes independent: `GPT`, `YouTube`, and `health`.
- Do not reintroduce a combined "test all modes for all subscriptions" action unless explicitly requested.
- Health mode is delay/HTTPS health sampling only; it should not run YouTube or GPT unlock checks.
- Health observation should run only when due and idle. The current idle wait is 5 minutes without UI interaction.
- Cache entries are mode-aware. Clearing one mode should not wipe other mode results for the same node.
- Result lists should remain bounded in height and scroll internally when many nodes are shown.
- Sorting should prefer nodes that satisfy the selected mode and then rank multi-condition quality, such as GPT unlock, YouTube CN unlock, and stable low latency.

UI wording conventions:

- Use `GPT`, not `AI`, for the ChatGPT unlock mode.
- Use `YouTube`, not `video` or `视频`, for the YouTube mode.
- Keep result text compact, for example `解锁(US)` or `阻断`, rather than long explanatory phrases inside result chips.

## Architecture Notes

Android uses Go core in library mode:

- Go builds `libclash.so` with `go build -buildmode=c-shared`
- Flutter talks to Android native service code through method channels and FFI-facing interfaces
- Dart Android core implementation is `lib/core/lib.dart`
- `lib/core/controller.dart` directly uses `coreLib`

Desktop `CoreService`, Rust IPC transport, named pipes, system tray, desktop windows, and desktop proxy managers are intentionally absent.

## Keep

- Keep `.dev-tools/`; it stores local build caches and speeds rebuilds.
- Keep `plugins/setup/`; it is still required for Android Go core builds.
- Keep `plugins/wifi_ssid/`; it has the Android Wi-Fi SSID implementation.
- Keep `core/Clash.Meta` submodule.

## Avoid

- Do not reintroduce desktop platform directories unless the project scope changes.
- Do not add non-arm64 Android ABIs unless the target device requirements change.
- Do not use `setup.dart`; it has been removed with distributor packaging.
