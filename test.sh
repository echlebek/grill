# set -e

go test .

export PATH=`pwd`:$PATH
export LANG=C

rm -f examples/*.err

cram examples/*.t tests/*.t
