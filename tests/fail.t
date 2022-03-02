
  $ cp $TESTDIR/fail/a.t .

TODO can remove || true after fixing the bug for error codes > 9
  $ grill a.t > a.diff || true

  $ diff $TESTDIR/fail/expected.diff a.diff
  $ diff $TESTDIR/fail/expected.err a.t.err

The output .err file is a valid passing test with applied changes:

  $ cp a.t.err b.t
  $ grill b.t
  .
  # Ran 1 test, 0 skipped, 0 failed.

Applying diff back to the source .t file makes it pass:

  $ patch -p0 < a.diff
  patching file a.t

  $ grill a.t
  .
  # Ran 1 test, 0 skipped, 0 failed.
