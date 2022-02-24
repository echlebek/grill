Output one status summary entry per single test file/suite.

  $ mkdir -p sub

  $ echo '  $ true' > sub/pass.t
  $ echo '  $ true' >> sub/pass.t

  $ echo '  $ true' > sub/fail.t
  $ echo '  $ false' >> sub/fail.t

  $ echo 'No command' > sub/skip.t

  $ grill -quiet sub/*.t
  !.s
  # Ran 3 tests, 1 skipped, 1 failed.
  [1]
