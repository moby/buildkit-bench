const path = require('path');

module.exports = {
  publicPath: process.env.WEBSITE_PUBLIC_PATH || '/',
  configureWebpack: {
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    },
    module: {
      rules: [
        {
          test: /\.html$/,
          use: [
            {
              loader: 'html-loader',
              options: {
                minimize: true,
              },
            },
          ],
        },
      ],
    },
  },
  chainWebpack: config => {
    config.module.rule('text').test(/\.txt$/).type('asset/source');
  }
};
