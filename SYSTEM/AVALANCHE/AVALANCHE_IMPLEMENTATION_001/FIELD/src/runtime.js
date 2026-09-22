import { loadConfig } from "./config.js";
import { FieldEngine } from "./engine.js";
import { AppendOnlyJournal } from "./journal.js";
import { loadAdapter } from "./adapter.js";
import { VMClient } from "./vm-client.js";

// createRuntime wires the FIELD from a config file: a validated config, a
// validated scoped adapter, an append-only journal, a read/write VM client, and
// the fail-closed engine. Restart safety comes from the journal being reloaded
// and the engine re-reading accepted VM state; nothing is fabricated on reload.
export async function createRuntime(configPath) {
  const config = await loadConfig(configPath);
  const adapter = await loadAdapter(config.adapter_path);
  const journal = await new AppendOnlyJournal(config.journal_path).initialize();
  const client = new VMClient(config);
  const engine = new FieldEngine({ config, adapter, client, journal });
  return { config, adapter, journal, client, engine };
}
