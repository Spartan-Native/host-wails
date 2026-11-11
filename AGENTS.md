# AGENTS.md

Instructions for AI coding agents working in this repository. (`npm/AGENTS.md` is a different file: it ships in the npm package, for agents building apps **with** the adapter.)

## Layout

- **Repository root:** the Go module `github.com/spartan-native/host-wails` (package `spartannative`).
- **`npm/`:** the npm package `@spartan-native/host-wails`, built with ng-packagr (`angular.json` project `host-wails`).

## Keep the two halves in sync

The halves talk through names the compilers can't check. Change both sides in the same commit:

- **Bound methods.** Wails names a bound method `<Go package path>.<Type>.<Method>`. `DEFAULT_SERVICE` in `npm/src/lib/wails-host.ts` must be `github.com/spartan-native/host-wails.Service`, matching the `module` line in `go.mod`. Renaming a Go method means changing the `call('…')` name in the adapter.
- **Events.** `spartan-native:appearance` and `spartan-native:menu-action` (`EventAppearance`, `EventMenuAction` in Go; `APPEARANCE_EVENT`, `MENU_ACTION_EVENT` in TypeScript).
- **Data shapes.** The structs in `types.go`, `menus.go` and `dialogs.go` must keep producing the JSON the contract types in `@spartan-native/theme` describe.

The fake runtime in `npm/src/lib/wails-host.spec.ts` stands in for the Go service. Update it when the Go side changes.

## Checks

| Command | Use it for |
| --- | --- |
| `npm test` | Adapter tests and the host-conformance suite |
| `npm run build` | Checking that the npm package compiles |
| `go test ./...` | Go tests |
| `npm run vet:go` | Vetting the Windows and Linux code from any OS |

## Docs

When you add, rename or remove a public export, option, Go option or host service, update `npm/AGENTS.md` and `npm/README.md` in the same change. `npm/AGENTS.md` must only refer to other packages by name (`@spartan-native/theme`), never by path, because the repository isn't there once the package is installed.

The spartan-native repository's agent skill (`tools/claude-plugin`) describes how apps set up this host. If a change affects app setup, say so in your summary so it can be updated there.
