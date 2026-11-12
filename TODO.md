# TODO: first release of host-wails

Left over from moving host-wails out of spartan-native. The repository is created and `main` is pushed. Delete this file when everything is done.

## Before releasing

- [ ] Wait for spartan-native's `0.2.0` packages to be on npmjs.org (`npm view @spartan-native/theme@0.2.0`).
- [ ] Make this repository public on GitHub (the Go module and npm provenance both rely on it).
- [ ] Add the repository secret `NPM_TOKEN`: a granular npm access token with publish rights on the `@spartan-native` scope. After the first release, set up trusted publishing for `@spartan-native/host-wails` on npmjs.com and delete the secret.
- [ ] In `npm/`: `npm install`, commit `npm/package-lock.json`, push. CI (`ci.yml`) should go green: Go tests + vet, npm tests + build.

## Release

- [ ] Tag `v0.2.0` on `main` and push the tag (same minor as spartan-native; the peer range is `^0.2.0`). `release.yml` publishes `@spartan-native/host-wails@0.2.0` to npmjs.org; the tag itself releases the Go module `github.com/spartan-native/host-wails@v0.2.0`.
- [ ] Check `npm view @spartan-native/host-wails` shows the version with provenance, and `go get github.com/spartan-native/host-wails@v0.2.0` works from another module (no `GOPRIVATE` needed).

## After

- [ ] Harbor: drop its `@spartan-native` GitHub Packages `.npmrc` line and any `HOST_WAILS_TOKEN` / `GOPRIVATE` setup, and move to `^0.2.0`.
