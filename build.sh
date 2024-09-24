# Configure Git with the provided GitHub Personal Access Token (GITHUB_PAT)
echo 'Configuring Git...'
git config --global --add safe.directory '*'
git config --global url.https://$GITHUB_PAT@github.com/.insteadOf https://github.com/

# Running linters on code
#echo 'Running linter...'
#make lint

# Build the project
echo 'Building the project...'
make build
