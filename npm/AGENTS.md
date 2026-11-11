# AGENTS.md: @spartan-native/host-wails

Instructions for AI coding agents building a Go + Wails v3 desktop app with `@spartan-native/host-wails`.

## What this package is

The `NativeHost` adapter for Wails v3 (Layer 04). It has two halves that work together:

- **Go service** (Go module `github.com/spartan-native/host-wails`). It reads the real OS settings, pushes changes live, and shows native menus and dialogs.
- **TypeScript adapter** (this npm package). `wailsHost()` implements the contract from `@spartan-native/theme` over the Wails runtime (`@wailsio/runtime`, imported lazily).

Both halves are released together from one tag, so **use the same version of each**: `@spartan-native/host-wails@1.2.3` with `github.com/spartan-native/host-wails@v1.2.3`. They talk through bound-method names and events, which can change between versions.

## Set up

The Go module lives in a private GitHub repository, so tell Go not to use the public module proxy for it and give git a token that can read it (the same token as for the npm package works):

```bash
go env -w GOPRIVATE=github.com/spartan-native/*
git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
go get github.com/spartan-native/host-wails@v1.2.3    # same version as the npm package
```

```go
// main.go: register the service; nothing else to wire
import spartannative "github.com/spartan-native/host-wails"

app := application.New(application.Options{
	Services: []application.Service{application.NewService(spartannative.New())},
})
// options: spartannative.WithSkin("gtk"), spartannative.WithPollInterval(500*time.Millisecond)
// debug:   SPARTAN_NATIVE_DEBUG=1
```

```ts
// app.config.ts: same frontend in the browser and in Wails
import { isWails, wailsHost } from '@spartan-native/host-wails';
import { webHost } from '@spartan-native/host-web';

providers: [provideSpartanNative({ host: isWails() ? wailsHost() : webHost() })]
```

For a custom title bar, use a frameless window on Windows and Linux and the hidden-inset title bar on macOS. `<sn-title-bar>` is already a `--wails-draggable` region.

## Public API (`import … from '@spartan-native/host-wails'`)

| Symbol | Use |
| --- | --- |
| `wailsHost({ service?, runtime? })` | The host. `runtime` injects a fake Wails runtime for tests |
| `isWails(location?)` | True inside a Wails webview |
| `DEFAULT_SERVICE`, `APPEARANCE_EVENT`, `MENU_ACTION_EVENT` | The bound-method prefix and event names, for custom setups and tests |
| `WailsRuntimeLike` | The runtime surface the adapter uses |

## What the app gets

| Service | Notes |
| --- | --- |
| Appearance | Real OS values: accent, scheme, contrast, transparency, motion, text size; live |
| `window` | Focus, `minimize()`, `toggleMaximize()`, `close()`, `isMaximized()`, `onMaximizedChange()`. desktop-ui's window buttons use them automatically |
| `menus` | `setApplicationMenu`, `showContextMenu`, `onAction`. Native on macOS. Partial on Windows and Linux: context menus are native, but the app menu only reaches windows created after it is set. Use `<sn-menubar>` and `[snContextMenu]`, not the raw service |
| `dialogs` | The OS's own `openFile`, `saveFile`, `pickFolder` and `message`. Files come back with a `path`. Cancelled dialogs give `[]` / `null` |
| `notifications` | Not specified yet |

```ts
const host = inject(NATIVE_HOST);
const [file] = await host.dialogs!.openFile({ filters: [{ name: 'Backups', extensions: ['hbk'] }] });
const answer = await host.dialogs!.message!({
  kind: 'question', message: 'Save changes?',
  buttons: [{ id: 'save', label: 'Save' }, { id: 'discard', label: "Don't save" }, { id: 'cancel', label: 'Cancel' }],
  defaultId: 'save', cancelId: 'cancel',
});
```

## Rules

1. **Always pass `cancelId` to `message()`.** GTK lets the user close an alert without pressing a button, and then the call doesn't resolve.
2. Don't call `@wailsio/runtime` or the Go service's bound methods directly from components. Go through `SpartanNativeTheme` or `NATIVE_HOST`, so the app keeps working in the browser.
3. Keep the Go module and the npm package on the same version. A mismatch fails at runtime, not at build time.
4. Don't register a second appearance or menu service in Go. `spartannative.New()` owns the `spartan-native:*` events.
5. To test without Wails, pass `wailsHost({ runtime: fake })` and run `describeHostConformance` from `@spartan-native/host-conformance`.
