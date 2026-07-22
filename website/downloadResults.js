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
const fetchAttempts = 5;
const fetchRetryBaseDelayMs = 1000;

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

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function isRetryableStatus(status) {
  return status === 408 || status === 429 || status >= 500;
}

function retryDelayMs(attempt) {
  return fetchRetryBaseDelayMs * (2 ** (attempt - 1));
}

function formatFetchError(err) {
  return err instanceof Error ? err.message : String(err);
}

async function waitBeforeRetry(url, reason, attempt) {
  const delay = retryDelayMs(attempt);
  console.warn(`Failed to download ${url}: ${reason}; retrying in ${delay}ms (${attempt + 1}/${fetchAttempts})`);
  await sleep(delay);
}

async function fetchRequired(url) {
  for (let attempt = 1; attempt <= fetchAttempts; attempt++) {
    let response;
    try {
      response = await fetch(url);
    } catch (err) {
      const reason = formatFetchError(err);
      if (attempt === fetchAttempts) {
        throw new Error(`Failed to download ${url}: ${reason}`);
      }
      await waitBeforeRetry(url, reason, attempt);
      continue;
    }

    if (response.ok) {
      return response;
    }

    const reason = `${response.status} ${response.statusText}`;
    if (!isRetryableStatus(response.status) || attempt === fetchAttempts) {
      throw new Error(`Failed to download ${url}: ${reason}`);
    }

    if (response.body) {
      await response.body.cancel().catch(() => {});
    }
    await waitBeforeRetry(url, reason, attempt);
  }

  throw new Error(`Failed to download ${url}`);
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
  const response = await fetchRequired(manifestUrl);
  return normalizeManifest(await response.json());
}

async function main() {
  const manifest = await readManifest();
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
