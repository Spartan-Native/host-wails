# host-wails

The Spartan Native host adapter for **Go + Wails v3**. Both halves of the adapter live here and are released together:

| Half | Where | Installed with |
| --- | --- | --- |
| Go service | repository root (`*.go`, `go.mod`) | `go get github.com/spartan-native/host-wails@v1.2.3` |
| TypeScript adapter | `npm/` | `npm install @spartan-native/host-wails@1.2.3` (GitHub Packages) |

What the adapter does and how to use it in an app: [`npm/README.md`](npm/README.md). Instructions for agents building apps with it: [`npm/AGENTS.md`](npm/AGENTS.md), which ships in the npm package.

The adapter implements the host contract from `@spartan-native/theme` and is tested against `@spartan-native/host-conformance`. Both come from the spartan-native repository and are installed from GitHub Packages.

## Develop

Needs Node 24.15+ or 26 (`.nvmrc`), npm 11 and Go 1.26+.

`@spartan-native/*` packages come from GitHub Packages. Log in once with a token that has `read:packages`:

```bash
npm login --scope=@spartan-native --auth-type=legacy --registry=https://npm.pkg.github.com
npm install
```

| Command | What it does |
| --- | --- |
| `npm test` | Adapter tests and the host-conformance suite, against a fake Wails runtime |
| `npm run build` | Builds the npm package into `dist/host-wails` |
| `go test ./...` | Go tests (change detection, skin table, JSON shape, menus, dialogs) |
| `npm run vet:go` | `go vet` for darwin, windows and linux |
| `SN_LIVE=1 go test -run TestLiveRead -v ./...` | Prints what this machine's OS reports |

To try a change in a Wails app before releasing it, point the app at your clone:

```bash
# in the app
go mod edit -replace github.com/spartan-native/host-wails=../host-wails
npm install ../host-wails/dist/host-wails     # after npm run build here
```

## Release

Push a tag. `.github/workflows/release.yml` runs CI, then publishes the npm package. The tag itself is the Go module's version.

```bash
git tag v1.2.3 && git push origin v1.2.3            # npm "latest"
git tag v1.3.0-beta.1 && git push origin v1.3.0-beta.1   # npm "beta"
```

Only `vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-beta[.N]` tags publish. The `v` is required, because Go ignores tags without it. From v2 on, Go needs the module path to end in `/v2`.
