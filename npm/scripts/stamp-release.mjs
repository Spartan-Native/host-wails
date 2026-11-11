#!/usr/bin/env node
// Prepares npm/dist/host-wails for a release. Used by .github/workflows/release.yml after
// `npm run build`, so projects/host-wails/package.json keeps its 0.0.0 placeholder version.
//
// It sets "version" to the release version, points publishConfig at the GitHub Packages
// npm registry, and adds "repository", which links the package to this repo (and its
// visibility) on GitHub. Peer dependency ranges (@spartan-native/theme) are left as written.
//
// Usage (from npm/): node scripts/stamp-release.mjs <version> [<owner>/<repo>]
//   version:    v1.2.3, v1.2.3-beta or v1.2.3-beta.4 (the same tag the Go module uses)
//   owner/repo: defaults to $GITHUB_REPOSITORY

import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const file = join(root, 'dist/host-wails/package.json');

const VERSION = /^v?(\d+\.\d+\.\d+(?:-beta(?:\.\d+)?)?)$/;

const [rawVersion, repoArg] = process.argv.slice(2);
const match = VERSION.exec(rawVersion ?? '');
if (!match) {
  console.error(`"${rawVersion ?? ''}" is not a release version (v1.2.3, v1.2.3-beta or v1.2.3-beta.4).`);
  process.exit(1);
}
const version = match[1];

const repo = repoArg ?? process.env.GITHUB_REPOSITORY;
if (!repo) {
  console.error('Pass <owner>/<repo> or set GITHUB_REPOSITORY.');
  process.exit(1);
}

if (!existsSync(file)) {
  console.error(`Missing build output (${file}). Run "npm run build" first.`);
  process.exit(1);
}

const pkg = JSON.parse(readFileSync(file, 'utf8'));
pkg.version = version;
pkg.repository = { type: 'git', url: `git+https://github.com/${repo}.git`, directory: 'npm/projects/host-wails' };
pkg.publishConfig = { ...pkg.publishConfig, registry: 'https://npm.pkg.github.com' };

writeFileSync(file, `${JSON.stringify(pkg, null, 2)}\n`);
console.log(`${pkg.name}@${version}`);
