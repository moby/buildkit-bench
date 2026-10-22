const fs = require('fs');
const path = require('path');

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

// Build the dashboard index and public download manifest from the same inventory.
// Paths are relative to the website, independent of the caller's working directory.
async function indexResults(websiteDir = path.join(__dirname, '..')) {
  const resultsDir = path.join(websiteDir, 'public', 'result');
  const outputFilePath = path.join(websiteDir, 'src', 'generated', 'results.json');
  const manifestFilePath = path.join(resultsDir, 'manifest.json');
  // A fresh checkout has no reports yet, but still needs valid empty indexes.
  await fs.promises.mkdir(resultsDir, { recursive: true });
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

  const results = {
    results: resultEntries.map(result => result.name),
    logArchives: resultEntries.filter(result => result.files.includes('logs.tar.gz')).map(result => result.name),
    jsonReports: Object.fromEntries(resultEntries.map(result => [
      result.name,
      ['buildkit', 'buildx'].filter(project => result.files.includes(`${project}.json`)),
    ])),
    htmlReports: Object.fromEntries(resultEntries.map(result => [
      result.name,
      Object.fromEntries([
        ['buildkit', result.files.includes('buildkit.html') ? 'buildkit.html' : 'index.html'],
        ['buildx', 'buildx.html'],
      ].filter(([, file]) => result.files.includes(file))),
    ])),
  };

  const manifest = { version: 1, results: resultEntries };

  await fs.promises.mkdir(path.dirname(outputFilePath), { recursive: true });
  await fs.promises.writeFile(outputFilePath, JSON.stringify(results, null, 2));
  console.log(`${outputFilePath} has been generated successfully.`);

  await fs.promises.writeFile(manifestFilePath, JSON.stringify(manifest, null, 2));
  console.log(`${manifestFilePath} has been generated successfully.`);
}

if (require.main === module) {
  indexResults().catch(err => {
    console.error(err);
    process.exitCode = 1;
  });
}

module.exports = { indexResults };
