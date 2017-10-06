var merge = require('webpack-merge')
var prodEnv = require('./prod.env')

const ENV = {
  NODE_ENV: 'development',
  API_URL: process.env.API_URL || 'http://127.0.0.1:18000/api'
}

const jsonedEnv = Object.keys(ENV).reduce((acc, key) => {
  const value = ENV[key]
  acc[key] = JSON.stringify(value)
  return acc
}, {});

module.exports = merge(prodEnv, jsonedEnv);
