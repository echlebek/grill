set -e

go test ./cmd/grill ./internal/...

export PATH=`pwd`:$PATH
export LANG=C

rm -f examples/*.err

grill examples/*.t tests/*.t
