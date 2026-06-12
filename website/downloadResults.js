const fs = require('fs');
const path = require('path');

const repository = process.env.GITHUB_REPOSITORY || 'moby/buildkit-bench';
const [repositoryOwner, repositoryName] = repository.split('/');
const defaultPagesUrl = `https://${repositoryOwner}.github.io/${repositoryName}/`;
const pagesUrl = ensureTrailingSlash(process.env.PAGES_URL || defaultPagesUrl);
const resultsDir = process.env.RESULTS_DIR || path.join(__dirname, 'public', 'result');
const args = process.argv.slice(2);
const listOnly = args.includes('--list');
const requestedResults = new Set(args.filter(arg => arg && arg !== '--list'));

function ensureTrailingSlash(value) {
  return value.endsWith('/') ? value : `${value}/`;
}

function validatePathSegment(value, field) {
  if (!value || value.includes('/') || value.includes('\\') || value === '.' || value === '..') {
    throw new Error(`Invalid ${field}: ${value}`);
  }
  return value;
}

function validateRelativePath(value) {
  const normalized = value.replace(/\\/g, '/');
  const segments = normalized.split('/');

  if (normalized.startsWith('/') || segments.some(segment => {
    return !segment || segment === '.' || segment === '..';
  })) {
    throw new Error(`Invalid manifest file path: ${value}`);
  }

  return segments;
}

function resultUrl(resultName, filePath) {
  const encodedPath = ['result', resultName, ...validateRelativePath(filePath)]
    .map(segment => encodeURIComponent(segment))
    .join('/');
  return new URL(encodedPath, pagesUrl).toString();
}

async function fetchRequired(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to download ${url}: ${response.status} ${response.statusText}`);
  }
  return response;
}

async function downloadFile(resultName, filePath) {
  const response = await fetchRequired(resultUrl(resultName, filePath));
  const destination = path.join(resultsDir, resultName, ...validateRelativePath(filePath));

  await fs.promises.mkdir(path.dirname(destination), { recursive: true });
  await fs.promises.writeFile(destination, Buffer.from(await response.arrayBuffer()));
}

function normalizeManifest(manifest) {
  if (!manifest || !Array.isArray(manifest.results)) {
    throw new Error('Pages result manifest does not contain a results array');
  }

  return manifest.results.map(result => {
    if (!result || typeof result !== 'object') {
      throw new Error('Pages result manifest contains an invalid result entry');
    }
    if (!Array.isArray(result.files)) {
      throw new Error(`Pages result manifest entry for ${result.name} does not contain files`);
    }

    return {
      name: validatePathSegment(result.name, 'result name'),
      files: result.files.map(file => validateRelativePath(file).join('/')),
    };
  });
}

async function readManifest() {
  const manifestUrl = new URL('result/manifest.json', pagesUrl).toString();
  const response = await fetch(manifestUrl);

  if (response.status === 404) {
    console.warn(`No Pages result manifest found at ${manifestUrl}`);
    process.exitCode = 2;
    return null;
  }
  if (!response.ok) {
    throw new Error(`Failed to download ${manifestUrl}: ${response.status} ${response.statusText}`);
  }

  return normalizeManifest(await response.json());
}

async function main() {
  const manifest = await readManifest();
  if (!manifest) {
    return;
  }
  if (listOnly) {
    console.log(JSON.stringify(manifest.map(result => result.name), null, 2));
    return;
  }

  const selectedResults = requestedResults.size === 0
    ? manifest
    : manifest.filter(result => requestedResults.has(result.name));

  const missingResults = [...requestedResults].filter(name => {
    return !selectedResults.some(result => result.name === name);
  });
  if (missingResults.length > 0) {
    throw new Error(`Pages result manifest does not contain: ${missingResults.join(', ')}`);
  }

  await fs.promises.mkdir(resultsDir, { recursive: true });

  let fileCount = 0;
  for (const result of selectedResults) {
    for (const file of result.files) {
      await downloadFile(result.name, file);
      fileCount++;
    }
  }

  console.log(`Downloaded ${fileCount} file(s) for ${selectedResults.length} result(s) from ${pagesUrl}`);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
