const path = require('path');
const { VueLoaderPlugin } = require('vue-loader');
const CopyPlugin = require('copy-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
  entry: {
    ticket: './src/ticket.js',
    organization: './src/organization.js'
  },
  output: {
    filename: '[name].js',
    path: path.resolve(__dirname, 'dist/assets'),
    publicPath: ''
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader'
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-env']
          }
        }
      },
      {
        test: /\.css$/,
        use: [
          'style-loader',
          'css-loader',
          'postcss-loader'
        ]
      }
    ]
  },
  plugins: [
    new VueLoaderPlugin(),
    new HtmlWebpackPlugin({
      template: './src/templates/ticket.html',
      filename: 'ticket.html',
      chunks: ['ticket']
    }),
    new HtmlWebpackPlugin({
      template: './src/templates/organization.html',
      filename: 'organization.html',
      chunks: ['organization']
    }),
    new CopyPlugin({
      patterns: [
        { from: 'manifest.json', to: '../' },
        { from: 'translations', to: '../translations' },
        { from: 'assets', to: './', globOptions: { ignore: ['**/*.js', '**/*.html'] } }
      ],
    }),
  ],
  resolve: {
    extensions: ['.js', '.vue']
  }
};
