mkdir -p ./bin && \
gox -arch="amd64" -output="./bin/nvim-finder_{{.OS}}_{{.Arch}}" . && \
ls -al ./bin
