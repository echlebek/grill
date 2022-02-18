go test .

rm -f examples/*.err
./cram --quiet examples/env.t
