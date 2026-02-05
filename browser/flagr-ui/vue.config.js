const { execSync } = require('child_process');

try {
  process.env.VUE_APP_VERSION = execSync(
    'git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD',
    { encoding: 'utf-8', shell: true }
  ).trim();
} catch {
  // git not available
}

module.exports = {
  assetsDir: 'static',
  publicPath: process.env.BASE_URL
}
