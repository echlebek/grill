go test .

export LANG=C
rm -f examples/*.err
./cram --quiet examples/env.t
