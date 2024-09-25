const fs = require('fs');
const path = require('path');

class GenerateResultsJsonPlugin {
  apply(compiler) {
    compiler.hooks.thisCompilation.tap('GenerateResultsJsonPlugin', (compilation) => {
      compilation.hooks.processAssets.tapAsync(
        {
          name: 'GenerateResultsJsonPlugin',
          stage: compiler.webpack.Compilation.PROCESS_ASSETS_STAGE_ADDITIONAL,
        },
        (assets, callback) => {
          const resultsDir = path.resolve(__dirname, 'public/result');
          const results = fs.readdirSync(resultsDir).filter(file => fs.statSync(path.join(resultsDir, file)).isDirectory());
          const newJsonContent = JSON.stringify({ results }, null, 2);
          const jsonFilePath = path.resolve(__dirname, 'src/assets/results.json');
          if (fs.existsSync(jsonFilePath)) {
            const existingJsonContent = fs.readFileSync(jsonFilePath, 'utf-8');
            if (existingJsonContent !== newJsonContent) {
              fs.writeFileSync(jsonFilePath, newJsonContent);
            }
          } else {
            fs.writeFileSync(jsonFilePath, newJsonContent);
          }
          callback();
        }
      );
    });
  }
}

module.exports = GenerateResultsJsonPlugin;
