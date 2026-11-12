# host-wails

[![CI](https://github.com/Spartan-Native/host-wails/actions/workflows/ci.yml/badge.svg)](https://github.com/Spartan-Native/host-wails/actions/workflows/ci.yml) [![npm](https://img.shields.io/npm/v/@spartan-native/host-wails?label=%40spartan-native%2Fhost-wails)](https://www.npmjs.com/package/@spartan-native/host-wails) [![Go Reference](https://pkg.go.dev/badge/github.com/spartan-native/host-wails.svg)](https://pkg.go.dev/github.com/spartan-native/host-wails)

**Go + Wails, reading the real OS.**

The [Spartan Native](https://spartan-native.github.io/spartan-native/) host adapter for desktop apps built with **Go + [Wails v3](https://v3.wails.io)**. A Go service reads the accent color, light/dark, contrast, transparency, motion, text size and window focus **from the operating system itself**, and pushes every change to your Angular app as it happens. It also gives the frontend native window controls, application and context menus, and file dialogs and alerts.

[Website](https://spartan-native.github.io/host-wails/) · [Spartan Native](https://github.com/Spartan-Native/spartan-native) · [npm package](https://www.npmjs.com/package/@spartan-native/host-wails) · [Go reference](https://pkg.go.dev/github.com/spartan-native/host-wails)

## Why

In a browser, OS settings come from CSS media queries, which often can't tell what the OS really uses: Chromium hides the accent color, and text size, transparency and the desktop (GNOME or KDE) aren't visible at all. In a Wails app you have a whole Go process next to the webview, so this adapter asks the OS directly, and Spartan Native components look and behave like they belong there.

Your frontend doesn't change: the same app still runs in a browser with `@spartan-native/host-web`.

## Quick start

Set up [Spartan Native](https://github.com/Spartan-Native/spartan-native#quick-start) in your Angular frontend first. Then install both halves of this adapter **at the same version**: they're released together from one git tag.

```bash
npm install @spartan-native/host-wails @spartan-native/host-web @wailsio/runtime
go get github.com/spartan-native/host-wails@v0.2.0
```

Register the Go service:

```go
// main.go
import spartannative "github.com/spartan-native/host-wails"

app := application.New(application.Options{
	Services: []application.Service{application.NewService(spartannative.New())},
	// …
})
```

Give the theme the Wails host, and fall back to the web host in a browser:

```ts
// app.config.ts
import { provideSpartanNative } from '@spartan-native/theme';
import { isWails, wailsHost } from '@spartan-native/host-wails';
import { webHost } from '@spartan-native/host-web';

export const appConfig: ApplicationConfig = {
  providers: [provideSpartanNative({ host: isWails() ? wailsHost() : webHost() })],
};
```

That's it. For your own title bar, make the window frameless on Windows and Linux and use the hidden-inset title bar on macOS: `<sn-title-bar>` and `<sn-window-controls>` from `@spartan-native/desktop-ui` then drag, minimize, maximize and close the real window.

## What it reads

| Setting | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Accent | `AccentPalette` (the Settings app's accent), else DWM `AccentColor` | `NSColor.controlAccentColor` | portal `accent-color` |
| Light / dark | `AppsUseLightTheme` | `AppleInterfaceStyle` | portal `color-scheme` |
| Contrast | contrast themes → `forced` | Increase contrast → `more` | portal `contrast` → `more` |
| Reduce transparency | `EnableTransparency` | `accessibilityDisplayShouldReduceTransparency` | — |
| Reduce motion | `SPI_GETCLIENTAREAANIMATION` | `accessibilityDisplayShouldReduceMotion` | portal `reduced-motion` |
| Text size | `TextScaleFactor` | — (apps set their own) | GNOME `text-scaling-factor` |
| OS & skin | `RtlGetVersion` (build ≥ 22000 → Windows 11) | `operatingSystemVersion` | `XDG_CURRENT_DESKTOP` (KDE → Qt skin) |
| Window focus | Wails window events | Wails window events | Wails window events |

Changes arrive through Wails' theme-changed event and a 1-second poll for settings that send no notification. The frontend hears about a change only when a value actually changed. `Capabilities()` reports what each OS really supports, as `native`, `partial` or `none`, and Spartan Native components fall back where something is missing.

## What the frontend gets

| | |
| --- | --- |
| **Appearance** | The values above, live, as one `AppearanceState` |
| **Window** | Focus, `minimize()`, `toggleMaximize()`, `close()`, `isMaximized()`. desktop-ui's window buttons use them automatically |
| **Menus** | Application and context menus from plain `MenuItem` data, with checkbox and radio items, submenus, roles and portable accelerators (`mod+shift+s`). Native on macOS (the global menu bar); on Windows and Linux, context menus are native and the app menu reaches only windows created after it is set |
| **Dialogs** | The OS's own `openFile`, `saveFile`, `pickFolder` and `message`, attached to the window (a sheet on macOS). A cancelled dialog gives `[]` / `null` |

```ts
const host = inject(NATIVE_HOST);
const [file] = await host.dialogs!.openFile({ filters: [{ name: 'Backups', extensions: ['hbk'] }] });
```

Always pass a `cancelId` to `message()`: GTK lets the user close an alert without pressing a button.

## Options

```go
spartannative.New(
	spartannative.WithSkin("gtk"),                         // force a skin
	spartannative.WithPollInterval(500*time.Millisecond), // default 1 s; 0 turns polling off
)
```

Set `SPARTAN_NATIVE_DEBUG=1` to log every call and every change the Go service sees.

More detail, including the full API, is in the [package README](npm/projects/host-wails/README.md). Instructions for AI agents building apps with this adapter are in [`npm/projects/host-wails/AGENTS.md`](npm/projects/host-wails/AGENTS.md), which ships in the npm package.

## Status

Verified end to end on macOS with the [Harbor](https://github.com/Spartan-Native/harbor) sample app. The Windows and Linux readers compile, vet and pass their tests in CI, but haven't been tried in a real app window yet: reports from those systems are very welcome.

The project is before `1.0`, so a minor version may contain breaking changes. Keep the npm package and the Go module on the same version: they talk through method and event names that can change between versions, and a mismatch fails at runtime, not at build time.

## How it works

| Half | Where | What it does |
| --- | --- | --- |
| **Go module** `github.com/spartan-native/host-wails` | repository root | A Wails service that reads the OS settings, watches for changes and shows native menus and dialogs |
| **npm package** `@spartan-native/host-wails` | `npm/` | `wailsHost()`, the `NativeHost` implementation from `@spartan-native/theme`, which calls the Go service over the Wails runtime |

The frontend calls the service's bound methods (`github.com/spartan-native/host-wails.Service.Appearance`, …), and the service emits `spartan-native:appearance` and `spartan-native:menu-action` events back. Both halves are tested against the same contract fixtures in [`testdata/contract/`](testdata/contract), and the adapter passes `@spartan-native/host-conformance`, the suite every Spartan Native host must pass.

## Contributing

You need:

- **Go 1.26+**
- **Node 24.15+ or 26** (`npm/.nvmrc`) and **npm 11** (`engine-strict` is on)
- On macOS, the **Xcode Command Line Tools** (the macOS reader uses cgo)

```bash
cd npm && npm install
```

| Command | Where | What it does |
| --- | --- | --- |
| `make test` | root | Go tests with the race detector (change detection, skin table, contract fixtures, menus, dialogs) |
| `make vet` | root | `go vet` for darwin, windows and linux, from any OS |
| `make live` | root | Prints what this machine's OS reports |
| `npm test` | `npm/` | Adapter tests and the host-conformance suite, against a fake Wails runtime |
| `npm run build` | `npm/` | Builds the npm package into `npm/dist/host-wails` |

To try a change in a Wails app before releasing it, point the app at your clone:

```bash
# in the app
go mod edit -replace github.com/spartan-native/host-wails=../host-wails
npm install ../host-wails/npm/dist/host-wails     # after npm run build in host-wails/npm
```

Notes for contributors, human or AI agent, are in [`AGENTS.md`](AGENTS.md). The most important rule: the Go and TypeScript halves talk through names the compilers can't check, so change both sides, and the fixtures in `testdata/contract/`, in the same commit.

### Repository layout

```text
.
├── go.mod, go.sum             Go module github.com/spartan-native/host-wails
├── doc.go                     Package documentation (package spartannative)
├── service.go                 Service: the Wails-bound methods and lifecycle, wiring the packages below
├── types.go                   Public aliases of the contract types and event names
├── service_test.go            Service tests (change detection, JSON shape, live read)
├── contract_test.go           Checks names and JSON against testdata/contract
├── testdata/contract/         Names and JSON values both halves are tested against
├── internal/
│   ├── contract/              The JSON types shared with @spartan-native/theme, event names, capability keys
│   ├── appearance/            OS readers (one reader_<os>.go per OS), skin table, change watcher
│   ├── menus/                 Builds native application and context menus from MenuItem data
│   ├── dialogs/               Native file dialogs and alerts
│   └── debug/                 SPARTAN_NATIVE_DEBUG logging
├── Makefile                   Go test / vet shortcuts
├── npm/                       Angular CLI workspace for the npm package
│   ├── package.json           Workspace scripts and dev dependencies
│   ├── angular.json, tsconfig.json, .npmrc, .nvmrc
│   ├── scripts/
│   │   └── stamp-release.mjs  Sets the release version before publishing
│   └── projects/host-wails/   The npm package @spartan-native/host-wails
│       ├── package.json       Published manifest (peer dependencies)
│       ├── src/               wailsHost(), isWails() and tests
│       ├── README.md          Package documentation
│       └── AGENTS.md          Agent instructions shipped with the package
├── site/                      The project page (plain HTML, CSS and JS), published to GitHub Pages
└── .github/workflows/         CI (ci.yml), release (release.yml) and the website (pages.yml)
```

The public Go package stays at the repository root: Wails names bound methods after the Go package path, and the npm adapter calls them by that name. `Service` is a thin layer; the work happens in the `internal/` packages, which apps can't import.

### Dependencies

| Dependency | Half | Used for |
| --- | --- | --- |
| [`github.com/wailsapp/wails/v3`](https://github.com/wailsapp/wails) | Go | The service lifecycle, events, window, menu and dialog APIs |
| [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) | Go | Windows: registry, `SystemParametersInfo`, `RtlGetVersion` |
| [`github.com/godbus/dbus/v5`](https://github.com/godbus/dbus) | Go | Linux: the XDG desktop portal |
| `@spartan-native/theme` | npm, peer | The host contract (`NativeHost`, `AppearanceState`, `MenuItem`, …) |
| `@wailsio/runtime` | npm, peer | Calling the Go service and receiving its events |
| `@spartan-native/host-conformance` | npm, dev | The shared test suite every host adapter must pass |

macOS reads AppKit through cgo, so building on macOS needs the Xcode Command Line Tools. Windows and Linux need no C toolchain for this module (Wails itself has its own platform requirements).

### Release

Push a tag. [`release.yml`](.github/workflows/release.yml) runs CI, then publishes the npm package to npmjs.org. The tag itself is the Go module's version.

```bash
git tag v1.2.3 && git push origin v1.2.3                 # npm "latest"
git tag v1.3.0-beta.1 && git push origin v1.3.0-beta.1   # npm "beta"
```

Only `vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-beta[.N]` tags publish. The `v` is required, because Go ignores tags without it. From v2 on, Go needs the module path to end in `/v2`.

The website in `site/` is deployed by [`pages.yml`](.github/workflows/pages.yml) on every push to `main` that changes it.
