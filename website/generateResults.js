const fs = require('fs');
const path = require('path');

const resultsDir = path.join(__dirname, 'public', 'result');
const outputFilePath = path.join(__dirname, 'src', 'assets', 'results.json');

console.log('Running generateResults.js...');

fs.readdir(resultsDir, { withFileTypes: true }, (err, files) => {
  if (err) {
    console.error(`Failed to read directory ${resultsDir}:`, err);
    process.exit(1);
  }
  const results = { results: files.filter(file => file.isDirectory()).map(file => file.name) };
  fs.writeFile(outputFilePath, JSON.stringify(results, null, 2), (err) => {
    if (err) {
      console.error(`Failed to write ${outputFilePath}:`, err);
      process.exit(1);
    }
    console.log(`${outputFilePath} has been generated successfully.`);
  });
});
