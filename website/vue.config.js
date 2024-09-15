const path = require('path');

module.exports = {
  publicPath: process.env.WEBSITE_PUBLIC_PATH || '/',
  configureWebpack: {
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    }
  }
};
