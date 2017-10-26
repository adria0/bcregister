var CopyWebpackPlugin = require('copy-webpack-plugin');
var HtmlWebpackPlugin = require('html-webpack-plugin');
var HTMLWebpackPluginConfig = new HtmlWebpackPlugin({
  template: __dirname + '/app/index.html',
  filename: 'index.html'
});
var CopyWebpackPluginConfig = new CopyWebpackPlugin([
  { from: './app/static' }
]);

module.exports = {
  entry: [
    './app/index.js'
  ],
  module: {    
    loaders: [
      {
        test: /\.js$/, 
        include:  __dirname + '/app', 
        loader: 'babel-loader?presets[]=es2017'
      },
      {
          test: /\.css$/, loader: 'style-loader!css-loader'
      },      
    ]
  },
  output: {
    filename: 'index_bundle.js',
    path: __dirname + '/dist'
  },
  plugins: [
    HTMLWebpackPluginConfig,
    CopyWebpackPluginConfig
  ]
};
