import assert from "node:assert/strict";

// Re-implement nothing: import the real module via a TS->JS strip path.
// Use a tiny inline copy of the formatter is NOT acceptable; instead we verify
// the logic via node's TS strip-types against the actual source file.
// We mirror the source path so node can load it.
import { formatRupiah } from "../lib/format.ts";

const cases = [
  [0, "Rp0"],
  [1500, "Rp1.500"],
  [150000, "Rp150.000"],
  [999999, "Rp999.999"],
  [1000000, "Rp1.000.000"],
];

for (const [input, expected] of cases) {
  const got = formatRupiah(input);
  assert.equal(got, expected, `formatRupiah(${input}) => ${got}, want ${expected}`);
}
console.log("smoke: formatRupiah OK");