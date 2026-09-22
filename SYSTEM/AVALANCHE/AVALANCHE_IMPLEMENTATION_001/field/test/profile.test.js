import test from "node:test";
import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { validateProfile } from "../src/profile.js";

test("all four reference media satisfy the common profile contract", async () => {
  const directory = resolve("profiles");
  const files = (await readdir(directory)).filter((file) => file.endsWith(".json")).sort();
  assert.deepEqual(files, ["AI_FIELD_001.json", "DEVICE_FIELD_001.json", "HUMAN_FIELD_001.json", "LOCALITY_FIELD_001.json"]);
  const kinds = [];
  for (const file of files) {
    kinds.push(validateProfile(JSON.parse(await readFile(resolve(directory, file), "utf8"))).embodiment_kind);
  }
  assert.deepEqual(kinds.sort(), ["AI", "DEVICE", "HUMAN", "LOCALITY"]);
});

test("a profile cannot silently drop an effect ceiling", async () => {
  const profile = JSON.parse(await readFile(resolve("profiles/AI_FIELD_001.json"), "utf8"));
  profile.effect_ceiling.pop();
  assert.throws(() => validateProfile(profile), /effect_ceiling is missing/);
});

test("FIELD contract references resolve to the shared implementation artifacts", async () => {
  const url = new URL("../contract/PRESENCE_AVALANCHE_FIELD_CONTRACT_001.json", import.meta.url);
  const contract = JSON.parse(await readFile(url, "utf8"));
  for (const path of [contract.public_origin.binding, contract.vm.operation_contract]) {
    const document = JSON.parse(await readFile(new URL(path, url), "utf8"));
    assert.equal(document.implementation, contract.implementation);
  }
  for (const reference of contract.reference_profiles) {
    const profile = validateProfile(JSON.parse(await readFile(new URL(reference.path, url), "utf8")));
    assert.equal(profile.profile_id, reference.profile_id);
    assert.equal(profile.embodiment_kind, reference.embodiment_kind);
  }
});
