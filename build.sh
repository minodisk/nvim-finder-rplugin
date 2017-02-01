if [ ! -d vendor ]; then
  dep ensure
  dep status
fi

mkdir -p ./bin
OS='darwin freebsd linux netbsd openbsd windows'
for os in $OS; do
  CGO_ENABLED=0 GOOS=$os go build -o ./bin/nvim-finder_${os}_amd64 .
done
