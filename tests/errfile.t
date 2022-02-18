
  $ cp $TESTDIR/errfile/*.t .

  $ grill pass.t >/dev/null 2>&1

  $ grill fail.t >/dev/null 2>&1
  [1]

Error files are written out only for failed suites.

  $ ls
  fail.t
  fail.t.err
  pass.t
