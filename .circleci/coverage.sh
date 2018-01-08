# collect coverage result on the circleci environment.
GOTEST_COVERAGE_PKGS=`go list ./... | grep -v "/main\|/internal\|/wstest\|/appengine"`
goverage -coverprofile=${TEST_RESULTS}/go-test.coverage ${GOTEST_COVERAGE_PKGS}
