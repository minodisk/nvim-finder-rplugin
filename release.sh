mkdir -p ./bin && \
gox -arch="amd64" -output="./bin/nvim-finder_{{.OS}}_{{.Arch}}" . && \
ghr -t $GITHUB_TOKEN -u $GITHUB_USERNAME -r $GITHUB_REPONAME -replace `git describe --tags --abbrev=0` ./bin
