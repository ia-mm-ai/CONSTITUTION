import test from "node:test";
import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { validateProfile } from "../src/profile.js";

test("all four reference media satisfy the common profile contract", async () => {
  const directory = resolve("profiles");
  const files = (await readdir(directory)).filter((file) => file.endsWith(".json")).sort();
  assert.deepEqual(files, ["AI_MEDIUM_001.json", "DEVICE_MEDIUM_001.json", "HUMAN_MEDIUM_001.json", "PRESENCE_FIELD_HOST_001.json"]);
  const kinds = [];
  for (const file of files) {
    kinds.push(validateProfile(JSON.parse(await readFile(resolve(directory, file), "utf8"))).embodiment_kind);
  }
  assert.deepEqual(kinds.sort(), ["AI", "DEVICE", "HUMAN", "LOCALITY"]);
});

test("a profile cannot silently drop an effect ceiling", async () => {
  const profile = JSON.parse(await readFile(resolve("profiles/AI_MEDIUM_001.json"), "utf8"));
  profile.effect_ceiling.pop();
  assert.throws(() => validateProfile(profile), /effect_ceiling is missing/);
});
