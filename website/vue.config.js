const path = require('path');
const GenerateResultsJsonPlugin = require('./GenerateResultsJsonPlugin');

module.exports = {
  publicPath: process.env.WEBSITE_PUBLIC_PATH || '/',
  configureWebpack: {
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    },
    plugins: [
      new GenerateResultsJsonPlugin()
    ]
  },
  chainWebpack: config => {
    config.module.rule('text').test(/\.txt$/).type('asset/source');
  }
};
