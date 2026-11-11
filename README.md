# host-wails

The Spartan Native host adapter for desktop apps built with **Go + [Wails v3](https://v3.wails.io)**.

Spartan Native UI components take their look from the operating system: accent color, light/dark, contrast, reduced transparency and motion, text size, and the OS "skin" (Windows 11/10, macOS, GTK, Qt). In a browser, those values come from CSS media queries, which often can't tell what the OS really uses. In a Wails app, this adapter reads them **from the OS itself** and pushes changes to the frontend as they happen. It also gives the frontend the native window controls, application and context menus, and file dialogs and alerts.

It has two halves that are released together from one git tag:

| Half | Where | What it does | Installed with |
| --- | --- | --- | --- |
| **Go module** `github.com/spartan-native/host-wails` | repository root | A Wails service that reads the OS settings, watches for changes and shows native menus and dialogs | `go get github.com/spartan-native/host-wails@v1.2.3` |
| **npm package** `@spartan-native/host-wails` | `npm/` | `wailsHost()`, the `NativeHost` implementation the Angular app talks to, which calls the Go service over the Wails runtime | `npm install @spartan-native/host-wails@1.2.3` (GitHub Packages) |

How to use it in an app, and what each OS reports: [`npm/projects/host-wails/README.md`](npm/projects/host-wails/README.md). Instructions for AI agents building apps with it: [`npm/projects/host-wails/AGENTS.md`](npm/projects/host-wails/AGENTS.md) (it ships in the npm package). Instructions for agents working on this repository: [`AGENTS.md`](AGENTS.md).

## Quick look

```go
// main.go
import spartannative "github.com/spartan-native/host-wails"

app := application.New(application.Options{
	Services: []application.Service{application.NewService(spartannative.New())},
})
```

```ts
// app.config.ts
import { isWails, wailsHost } from '@spartan-native/host-wails';
import { webHost } from '@spartan-native/host-web';

provideSpartanNative({ host: isWails() ? wailsHost() : webHost() });
```

## Repository layout

```text
.
├── go.mod, go.sum             Go module github.com/spartan-native/host-wails
├── doc.go                     Package documentation (package spartannative)
├── service.go                 Service: the Wails-bound methods and lifecycle, wiring the packages below
├── types.go                   Public aliases of the contract types and event names
├── service_test.go            Service tests (change detection, JSON shape, live read)
├── internal/
│   ├── contract/              The JSON types shared with @spartan-native/theme, and event names
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
└── .github/workflows/         CI (ci.yml) and release (release.yml)
```

The public Go package stays at the repository root: Wails names bound methods after the Go package path (`github.com/spartan-native/host-wails.Service.Appearance`), and the npm adapter calls them by that name. `Service` is a thin layer; the work happens in the `internal/` packages, which apps can't import.

## Dependencies

### Go module

| Dependency | Used for |
| --- | --- |
| [`github.com/wailsapp/wails/v3`](https://github.com/wailsapp/wails) | The service lifecycle, events, window, menu and dialog APIs |
| [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) | Windows: registry, `SystemParametersInfo`, `RtlGetVersion` |
| [`github.com/godbus/dbus/v5`](https://github.com/godbus/dbus) | Linux: the XDG desktop portal (`org.freedesktop.portal.Settings`) |

macOS reads AppKit through **cgo**, so building on macOS needs the Xcode Command Line Tools. Windows and Linux need no C toolchain for this module (Wails itself has its own platform requirements).

### npm package

| Dependency | Kind | Used for |
| --- | --- | --- |
| `@spartan-native/theme` | peer | The host contract (`NativeHost`, `AppearanceState`, `MenuItem`, …) |
| `@wailsio/runtime` | peer | Calling the Go service and receiving its events |
| `tslib` | runtime | TypeScript helpers |
| `@spartan-native/host-conformance` | dev | The shared test suite every host adapter must pass |
| Angular 22, ng-packagr, Vitest, jsdom | dev | Building and testing the package |

The `@spartan-native/*` packages come from the spartan-native repository and are published to **GitHub Packages**, not npmjs.org.

## Develop

You need:

- **Go 1.26+**
- **Node 24.15+ or 26** (`npm/.nvmrc`) and **npm 11** (`engine-strict` is on)
- On macOS, the **Xcode Command Line Tools** (for cgo)
- A GitHub token with `read:packages`, to install `@spartan-native/*`

Log in to GitHub Packages once, then install:

```bash
cd npm
npm login --scope=@spartan-native --auth-type=legacy --registry=https://npm.pkg.github.com
npm install
```

| Command | Where | What it does |
| --- | --- | --- |
| `make test` | root | Go tests (change detection, skin table, JSON shape, menus, dialogs) |
| `make vet` | root | `go vet` for darwin, windows and linux, from any OS |
| `make live` | root | Prints what this machine's OS reports |
| `npm test` | `npm/` | Adapter tests and the host-conformance suite, against a fake Wails runtime |
| `npm run build` | `npm/` | Builds the npm package into `npm/dist/host-wails` |

Set `SPARTAN_NATIVE_DEBUG=1` in an app to log every call and change the Go service sees.

To try a change in a Wails app before releasing it, point the app at your clone:

```bash
# in the app
go mod edit -replace github.com/spartan-native/host-wails=../host-wails
npm install ../host-wails/npm/dist/host-wails     # after npm run build in host-wails/npm
```

## Release

Push a tag. `.github/workflows/release.yml` runs CI, then publishes the npm package to GitHub Packages. The tag itself is the Go module's version.

```bash
git tag v1.2.3 && git push origin v1.2.3                 # npm "latest"
git tag v1.3.0-beta.1 && git push origin v1.3.0-beta.1   # npm "beta"
```

Only `vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-beta[.N]` tags publish. The `v` is required, because Go ignores tags without it. From v2 on, Go needs the module path to end in `/v2`.
