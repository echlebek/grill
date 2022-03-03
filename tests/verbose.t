Verbose mode:

  $ mkdir sub/

  $ echo '  $ true' > sub/a.t
  $ echo '  $ false' > sub/b.t
  $ echo '' > sub/c.t

  $ grill -verbose -quiet sub/*
  sub/a.t: passed
  sub/b.t: failed
  sub/c.t: skipped
  # Ran 3 tests, 1 skipped, 1 failed.
  [1]
