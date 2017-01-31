ghr -t $GITHUB_TOKEN -u $GITHUB_USERNAME -r $GITHUB_REPONAME -replace `git describe --tags --abbrev=0` ./bin
