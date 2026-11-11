# @spartan-native/host-wails

**Layer 04 · host adapter for Go + Wails v3.** A Go service reads the **real** OS settings and pushes changes live. The TypeScript adapter implements `NativeHost` over the Wails runtime and owns the window, so desktop-ui's window buttons minimize, maximize and close it.

> Status: **implemented**. Passes `@spartan-native/host-conformance`, the Go tests pass, and it was verified end to end on macOS (Harbor in a Wails window read the system accent and dark mode through the service). Windows and Linux readers compile and vet, but haven't been run on those systems yet. Spec: [`07-host-adapters.md`](../../../docs/07-host-adapters.md).

## Use it

```go
// main.go
import spartannative "github.com/spartan-native/host-wails/native"

app := application.New(application.Options{
	Services: []application.Service{application.NewService(spartannative.New())},
	// …
})
```

```ts
// app.config.ts
import { isWails, wailsHost } from '@spartan-native/host-wails';
import { webHost } from '@spartan-native/host-web';

provideSpartanNative({ host: isWails() ? wailsHost() : webHost() })
```

For your own title bar, make the window frameless on Windows/Linux and use the hidden-inset title bar on macOS (see `apps/harbor-desktop/main.go`). `<sn-title-bar>` and `<sn-window-controls>` then act on the real window, and their drag regions use `--wails-draggable`.

## What the Go service reads

| | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Accent | `Explorer\Accent\AccentPalette` (the Settings app's accent), else DWM `AccentColor` | `NSColor.controlAccentColor` | portal `org.freedesktop.appearance` `accent-color` |
| Light / dark | `AppsUseLightTheme` | `AppleInterfaceStyle` | portal `color-scheme` |
| Contrast | contrast themes (`SPI_GETHIGHCONTRAST`) → `forced` | Increase contrast → `more` | portal `contrast` → `more` |
| Reduce transparency | `EnableTransparency` | `accessibilityDisplayShouldReduceTransparency` | — (none) |
| Reduce motion | `SPI_GETCLIENTAREAANIMATION` | `accessibilityDisplayShouldReduceMotion` | portal `reduced-motion` |
| Text size | `Accessibility\TextScaleFactor` | — (apps set their own) | GNOME `text-scaling-factor` via the portal |
| OS / skin | `RtlGetVersion` (build ≥ 22000 → Windows 11) | `NSProcessInfo.operatingSystemVersion` | `XDG_CURRENT_DESKTOP` (KDE → Qt skin), `/etc/os-release` |

Changes are detected by Wails' `common:ThemeChanged` event and a 1-second poll for settings that send no notification (`WithPollInterval`). The service emits `spartan-native:appearance` only when something actually changed. `Capabilities()` reports what each OS really supports.

**Window:** focus from `common:WindowFocus` / `WindowLostFocus`; `minimize`, `toggleMaximize`, `close`, `isMaximized` and the maximized state from the Wails window API.

**Menus:** `SetApplicationMenu(items)` and `ShowContextMenu(items, x, y)` build Wails menus from the contract's `MenuItem` data (labels, checkbox and radio items, submenus, `enabled`, accelerators in the portable `mod+shift+s` notation, and roles such as `appMenu`, `editMenu`, `quit`). Roles the OS doesn't have are dropped (the app menu outside macOS). A chosen item comes back as the `spartan-native:menu-action` event with its id, which `menus.onAction` delivers. desktop-ui's `<sn-menubar>` uses this for the macOS global menu bar, and `[snContextMenu]` for OS context menus. `nativeMenus` is `native` on macOS and `partial` on Windows and Linux: context menus are native there, but Wails 3.1 gives the application menu only to windows created after it is set, and frameless windows don't show it.

**Dialogs:** `OpenFile`, `SaveFile`, `PickFolder` and `Message` show the OS's own dialogs, attached to the current window (a sheet on macOS), and back `host.dialogs`. Files come back with their `path`. A cancelled dialog gives `[]` / `null`. `Message` returns the pressed button's id; GTK lets the user close an alert without pressing a button, and that call then only ends when the frontend cancels it, so always pass a `cancelId`.

**Options:** `spartannative.New(spartannative.WithSkin("gtk"), spartannative.WithPollInterval(500*time.Millisecond))`. Set `SPARTAN_NATIVE_DEBUG=1` to log every call and change.

## Layout

| Path | What it is |
| --- | --- |
| `src/lib/wails-host.ts` | `wailsHost()`, `isWails()`; the Wails runtime is imported lazily (or injected for tests) |
| `native/` | Go module `github.com/spartan-native/host-wails/native`: `service.go`, `menus.go`, `dialogs.go`, one `appearance_<os>.go` per OS, tests |

## Tests

```bash
npx ng test host-wails                                   # conformance suite + adapter tests (fake Wails runtime)
cd packages/hosts/host-wails/native && go test ./...     # change detection, skin table, JSON shape, menus, dialogs
SN_LIVE=1 go test -run TestLiveRead -v ./...             # print what this machine reports
GOOS=windows go vet ./... ; GOOS=linux go vet -tags purego ./...
```
