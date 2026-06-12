const fs = require('fs');
const path = require('path');

const resultsDir = process.env.RESULTS_DIR || path.join(__dirname, 'public', 'result');
const keepResultsPerDay = Number.parseInt(process.env.KEEP_RESULTS_PER_DAY || '1', 10);
const monthsToKeep = Number.parseInt(process.env.MONTHS_TO_KEEP || '3', 10);

function parseDate(dateStr) {
  const year = Number.parseInt(dateStr.substring(0, 4), 10);
  const month = Number.parseInt(dateStr.substring(4, 6), 10) - 1;
  const day = Number.parseInt(dateStr.substring(6, 8), 10);
  return new Date(year, month, day);
}

function isScheduledResult(dir) {
  const envFilePath = path.join(resultsDir, dir, 'env.txt');
  if (!fs.existsSync(envFilePath)) {
    return false;
  }
  return fs.readFileSync(envFilePath, 'utf8').includes('GITHUB_EVENT_NAME=schedule');
}

function removeResult(dir, reason) {
  const dirPath = path.join(resultsDir, dir);
  fs.rmSync(dirPath, { recursive: true, force: true });
  console.log(`Removed ${dirPath}${reason ? ` (${reason})` : ''}`);
}

async function main() {
  const entries = await fs.promises.readdir(resultsDir, { withFileTypes: true });
  const resultsByDate = entries
    .filter(entry => entry.isDirectory())
    .map(entry => entry.name)
    .reduce((acc, dir) => {
      const date = dir.split('-')[0];
      acc[date] = acc[date] || [];
      acc[date].push(dir);
      return acc;
    }, {});

  const cutoff = new Date();
  cutoff.setMonth(cutoff.getMonth() - monthsToKeep);
  cutoff.setHours(0, 0, 0, 0);

  Object.keys(resultsByDate).forEach(date => {
    const dirs = resultsByDate[date];
    const pdate = parseDate(date);
    if (Number.isNaN(pdate.getTime())) {
      console.warn(`Skipping unrecognized date format in ${date}`);
      return;
    }

    if (pdate < cutoff) {
      dirs.forEach(dir => removeResult(dir, `older than ${monthsToKeep} months`));
      return;
    }

    const removeDirs = dirs.filter(dir => !isScheduledResult(dir));
    removeDirs.sort().reverse();
    removeDirs.slice(keepResultsPerDay).forEach(dir => removeResult(dir));
  });
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
