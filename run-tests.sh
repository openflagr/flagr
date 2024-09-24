echo 'Running unit tests...'
# Run unit tests and generate code coverage report
go test -v ./... -covermode=count -coverprofile=coverage.out
echo 'Unit tests completed.'

echo 'Running tests for SonarQube...'
# Run additional tests for SonarQube and generate code coverage report
go test -v ./... -covermode=count -coverprofile=sonar_coverage.out
echo 'Tests for SonarQube completed.'
