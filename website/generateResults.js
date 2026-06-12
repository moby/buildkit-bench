const fs = require('fs');
const path = require('path');

const resultsDir = path.join(__dirname, 'public', 'result');
const outputFilePath = path.join(__dirname, 'src', 'assets', 'results.json');
const manifestFilePath = path.join(resultsDir, 'manifest.json');

console.log('Running generateResults.js...');

function toPosixPath(filePath) {
  return filePath.split(path.sep).join('/');
}

async function listFiles(dir, prefix = '') {
  const entries = await fs.promises.readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const entryPath = path.join(dir, entry.name);
    const relativePath = prefix ? path.join(prefix, entry.name) : entry.name;

    if (entry.isDirectory()) {
      files.push(...await listFiles(entryPath, relativePath));
    } else if (entry.isFile()) {
      files.push(toPosixPath(relativePath));
    }
  }

  return files.sort();
}

async function main() {
  const files = await fs.promises.readdir(resultsDir, { withFileTypes: true });
  const resultEntries = await Promise.all(files
    .filter(file => file.isDirectory())
    .map(async file => {
      return {
        name: file.name,
        files: await listFiles(path.join(resultsDir, file.name)),
      };
    }));
  resultEntries.sort((a, b) => a.name.localeCompare(b.name));

  const results = { results: resultEntries.map(result => result.name) };
  const manifest = { version: 1, results: resultEntries };

  await fs.promises.writeFile(outputFilePath, JSON.stringify(results, null, 2));
  console.log(`${outputFilePath} has been generated successfully.`);

  await fs.promises.writeFile(manifestFilePath, JSON.stringify(manifest, null, 2));
  console.log(`${manifestFilePath} has been generated successfully.`);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
