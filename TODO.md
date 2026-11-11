# TODO: first release of host-wails

Left over from moving host-wails out of spartan-native. The repository is created and `main` is pushed. Delete this file when everything is done.

## Before releasing

- [ ] Wait for spartan-native's `0.1.0` release. **Check that it did not publish `@spartan-native/host-wails@0.1.0`**: the tag was made before the split, so its workflow still includes the old host-wails. If it did, delete that package version (or the whole package) on GitHub Packages, or release this repository as `v0.1.1` instead (and change Harbor's `^0.1.0` to match).
- [ ] spartan-native's packages (theme, host-conformance) give this repository **read** access under **Manage Actions access**.
- [ ] In `npm/`: `npm install` (needs a token with `read:packages`), commit `npm/package-lock.json`, push. CI (`ci.yml`) should go green: Go tests + vet on macOS, npm tests + build.

## Release

- [ ] Tag `v0.1.0` on `main` and push the tag. `release.yml` publishes `@spartan-native/host-wails@0.1.0` to GitHub Packages; the tag itself releases the Go module `github.com/spartan-native/host-wails@v0.1.0`.
- [ ] Check that the published package is linked to this repository (the `repository` field from `npm/scripts/stamp-release.mjs`), and give the `harbor` repository **read** access to it under **Manage Actions access**.
- [ ] Check `go get github.com/spartan-native/host-wails@v0.1.0` works from another module (with `GOPRIVATE=github.com/spartan-native/*`).

## After

- [ ] Harbor needs a fine-grained token with **Contents: read** on this repository as its `HOST_WAILS_TOKEN` secret, so its CI can fetch the Go module.
