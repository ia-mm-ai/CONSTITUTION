import { loadConfig } from "./config.js";
import { MediumEngine } from "./engine.js";
import { AppendOnlyJournal } from "./journal.js";
import { loadProfile } from "./profile.js";
import { VMClient } from "./vm-client.js";

export async function createRuntime(configPath) {
  const config = await loadConfig(configPath);
  const profile = await loadProfile(config.profile_path);
  const journal = await new AppendOnlyJournal(config.journal_path).initialize();
  const client = new VMClient(config);
  const engine = new MediumEngine({ config, profile, client, journal });
  return { config, profile, journal, client, engine };
}
