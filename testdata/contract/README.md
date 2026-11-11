# Contract fixtures

The names and JSON both halves of the adapter must agree on. The Go tests and the npm spec
(`npm/projects/host-wails/src/lib/wails-host.spec.ts`) both check against these files, so a
change on one side that isn't made on the other fails a test.

- `names.json`: the bound-service name, the event names, the capability keys (in the order of
  `CAPABILITIES` in `@spartan-native/theme`) and the menu roles.
- `values.json`: one value of each contract type, in the JSON the frontend sends or receives.
  Every field is set, so a field missing on either side shows up.
