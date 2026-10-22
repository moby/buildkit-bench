// Prune results before CI builds the Pages site. Keep reports (JSON and legacy
// HTML), raw benchmark inputs, and configuration for 12 months so retained runs
// can still be regenerated. Remove logs.tar.gz and unpacked logs/ after 3 months.
// Within the retention window, keep all scheduled runs and the latest
// non-scheduled run per UTC day. Identify scheduled runs from metadata.event in
// JSON reports, falling back to env.txt. The age limit also applies to them.
//
// Preview from the repository root:
//   node .github/scripts/pruneResults.js --dry-run
// Omit --dry-run to apply. Settings and the entire removal plan are checked
// before deleting anything. Unrecognized and symlinked result directories are
// skipped. UTC calendar cutoffs clamp to the target month's last day and retain
// the cutoff day itself.
//
// Environment overrides:
//   RESULTS_DIR: defaults to website/public/result relative to the repository;
//     explicit relative paths resolve against the current working directory.
//   MONTHS_TO_KEEP: 12; LOGS_MONTHS_TO_KEEP: 3 (positive integers).
//   KEEP_RESULTS_PER_DAY: 1 (non-scheduled runs only; zero is allowed).
//
// Tests: node --test .github/scripts/pruneResults.test.js

const fs = require('node:fs');
const path = require('node:path');

function integerSetting(env, name, fallback, minimum) {
  const value = env[name] ?? String(fallback);
  if (!/^\d+$/.test(value) || !Number.isSafeInteger(Number(value)) || Number(value) < minimum) {
    throw new Error(`${name} must be an integer >= ${minimum}`);
  }
  return Number(value);
}

function retentionSettings(env = process.env) {
  return {
    monthsToKeep: integerSetting(env, 'MONTHS_TO_KEEP', 12, 1),
    logsMonthsToKeep: integerSetting(env, 'LOGS_MONTHS_TO_KEEP', 3, 1),
    keepResultsPerDay: integerSetting(env, 'KEEP_RESULTS_PER_DAY', 1, 0),
  };
}

// Clamp to the last day of the target month instead of overflowing into the next.
function cutoffDate(now, months) {
  const cutoff = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - months, 1));
  const lastDay = new Date(Date.UTC(cutoff.getUTCFullYear(), cutoff.getUTCMonth() + 1, 0)).getUTCDate();
  cutoff.setUTCDate(Math.min(now.getUTCDate(), lastDay));
  if (Number.isNaN(cutoff.getTime())) throw new Error('Retention cutoff is outside the supported date range');
  return cutoff;
}

function resultDate(name) {
  if (!/^\d{8}-\d{6}$/.test(name)) return null;
  const iso = `${name.slice(0, 4)}-${name.slice(4, 6)}-${name.slice(6, 8)}`;
  const date = new Date(`${iso}T${name.slice(9, 11)}:${name.slice(11, 13)}:${name.slice(13, 15)}Z`);
  if (Number.isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== iso || Number(name.slice(9, 11)) > 23) return null;
  date.setUTCHours(0, 0, 0, 0);
  return date;
}

function readOptional(file) {
  try {
    return fs.readFileSync(file, 'utf8');
  } catch (err) {
    if (err.code === 'ENOENT') return null;
    throw err;
  }
}

function isScheduledResult(dir) {
  const events = ['buildkit', 'buildx'].flatMap(project => {
    const contents = readOptional(path.join(dir, `${project}.json`));
    if (contents === null) return [];
    const event = JSON.parse(contents).metadata?.event;
    return event ? [event] : [];
  });
  if (events.length) return events.includes('schedule');
  const env = readOptional(path.join(dir, 'env.txt')) || '';
  return /^GITHUB_EVENT_NAME\s*=\s*(?:schedule|"schedule"|'schedule')\s*$/m.test(env);
}

// Plan everything before deleting, so invalid metadata cannot cause a partial prune.
function planPruning(resultsDir, settings, now = new Date()) {
  const reportCutoff = cutoffDate(now, settings.monthsToKeep);
  const logsCutoff = cutoffDate(now, settings.logsMonthsToKeep);
  const entries = fs.readdirSync(resultsDir, { withFileTypes: true });
  const days = new Map();
  const actions = [];
  for (const entry of entries.sort((a, b) => b.name.localeCompare(a.name))) {
    if (!entry.isDirectory()) continue;
    const date = resultDate(entry.name);
    if (!date) {
      console.warn(`Skipping unrecognized result directory: ${entry.name}`);
      continue;
    }
    if (date < reportCutoff) {
      actions.push({ name: entry.name, reason: `older than ${settings.monthsToKeep} months` });
      continue;
    }
    const dir = path.join(resultsDir, entry.name);
    if (!isScheduledResult(dir)) {
      const day = entry.name.slice(0, 8);
      const count = days.get(day) || 0;
      days.set(day, count + 1);
      if (count >= settings.keepResultsPerDay) {
        actions.push({ name: entry.name, reason: 'daily non-scheduled run limit' });
        continue;
      }
    }
    if (date < logsCutoff) {
      for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
        if (entry.name === 'logs' || entry.name === 'logs.tar.gz') {
          actions.push({ name: path.join(path.basename(dir), entry.name), reason: `logs older than ${settings.logsMonthsToKeep} months` });
        }
      }
    }
  }
  return actions;
}

function main() {
  const args = process.argv.slice(2);
  if (args.some(arg => arg !== '--dry-run')) throw new Error('Usage: node pruneResults.js [--dry-run]');
  const dryRun = args.includes('--dry-run');
  const settings = retentionSettings();
  const resultsDir = fs.realpathSync(process.env.RESULTS_DIR || path.join(__dirname, '../../website/public/result'));
  const actions = planPruning(resultsDir, settings);
  for (const action of actions) {
    const target = path.resolve(resultsDir, action.name);
    const relative = path.relative(resultsDir, target);
    if (!relative || relative.startsWith('..') || path.isAbsolute(relative)) throw new Error(`Unsafe pruning target: ${target}`);
    console.log(`${dryRun ? 'Would remove' : 'Removing'} ${target} (${action.reason})`);
    if (!dryRun) fs.rmSync(target, { recursive: true, force: true });
  }
  console.log(`${dryRun ? 'Would remove' : 'Removed'} ${actions.length} result(s) or log artifact(s).`);
}

if (require.main === module) {
  try {
    main();
  } catch (err) {
    console.error(err);
    process.exitCode = 1;
  }
}

module.exports = { retentionSettings, cutoffDate, resultDate, planPruning };
