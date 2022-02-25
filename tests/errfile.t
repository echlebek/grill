
  $ mkdir sub
  $ echo '  $ true' > sub/pass.t
  $ echo '  $ false' > sub/fail.t

  $ grill -quiet sub/*.t
  !.
  # Ran 2 tests, 0 skipped, 1 failed.
  [1]

Error files are written out only for failed suites.

  $ ls sub/
  fail.t
  fail.t.err
  pass.t

If test starts passing, err file is removed

  $ echo '  $ true' > sub/fail.t
  $ grill -quiet sub/*.t
  ..
  # Ran 2 tests, 0 skipped, 0 failed.

  $ ls sub/
  fail.t
  pass.t
