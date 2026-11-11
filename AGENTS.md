# AGENTS.md

Instructions for AI coding agents working in this repository. (`npm/projects/host-wails/AGENTS.md` is a different file: it ships in the npm package, for agents building apps **with** the adapter.)

## Layout

- **Repository root:** the Go module `github.com/spartan-native/host-wails`. The root package `spartannative` is the public API: the Wails-bound `Service` (`service.go`), and aliases of the contract types (`types.go`). `Service` must stay at the root, because Wails derives bound method names from its package path. Keep it thin: it delegates to the internal packages.
- **`internal/contract/`:** the JSON types shared with the frontend and the event names. Add new contract types here and alias them in `types.go`.
- **`internal/appearance/`:** the per-OS readers (`reader_<os>.go`), the skin table and the change `Watcher`.
- **`internal/menus/`, `internal/dialogs/`:** native menus and dialogs, as functions that take the Wails `*application.App`.
- **`internal/debug/`:** `SPARTAN_NATIVE_DEBUG` logging.
- **`npm/`:** the Angular CLI workspace (`package.json`, `angular.json`, `.nvmrc`, release script in `scripts/`). Run npm commands from here.
- **`npm/projects/host-wails/`:** the npm package `@spartan-native/host-wails`, built with ng-packagr into `npm/dist/host-wails`.

## Keep the two halves in sync

The halves talk through names the compilers can't check. Change both sides in the same commit:

- **Bound methods.** Wails names a bound method `<Go package path>.<Type>.<Method>`. `DEFAULT_SERVICE` in `npm/projects/host-wails/src/lib/wails-host.ts` must be `github.com/spartan-native/host-wails.Service`, matching the `module` line in `go.mod`. Renaming a Go method means changing the `call('…')` name in the adapter.
- **Events.** `spartan-native:appearance` and `spartan-native:menu-action` (`EventAppearance`, `EventMenuAction` in Go; `APPEARANCE_EVENT`, `MENU_ACTION_EVENT` in TypeScript).
- **Data shapes.** The structs in `internal/contract/contract.go` must keep producing the JSON the contract types in `@spartan-native/theme` describe.

`testdata/contract/` holds the service name, event names, capability keys, menu roles and one JSON value of each contract type. `contract_test.go` (and a test in `internal/menus`) checks the Go side against it, and `wails-host.spec.ts` checks the adapter and `@spartan-native/theme` against it. When the contract changes, change the fixture and both sides together. Capability keys live once, in `contract.Capabilities`; the readers list only the keys that aren't `none`.

The fake runtime in `npm/projects/host-wails/src/lib/wails-host.spec.ts` stands in for the Go service. Update it when the Go side changes.

## Checks

| Command | Use it for |
| --- | --- |
| `npm test` (in `npm/`) | Adapter tests and the host-conformance suite |
| `npm run build` (in `npm/`) | Checking that the npm package compiles |
| `make test` | Go tests, with the race detector (CI also runs them on Linux with `-tags purego`, and on Windows) |
| `make vet` | Vetting the Windows and Linux code from any OS |

## Docs

When you add, rename or remove a public export, option, Go option or host service, update `npm/projects/host-wails/AGENTS.md` and `npm/projects/host-wails/README.md` in the same change. That `AGENTS.md` must only refer to other packages by name (`@spartan-native/theme`), never by path, because the repository isn't there once the package is installed.

The spartan-native repository's agent skill (`tools/claude-plugin`) describes how apps set up this host. If a change affects app setup, say so in your summary so it can be updated there.
