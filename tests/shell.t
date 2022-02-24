-shell option sets the interpreter used to run the test.

  $ cat > check-sh.t <<EOF
  >   $ echo \$0
  >   /bin/sh (glob)
  > EOF

  $ cat > check-bash.t <<EOF
  >   $ echo \$0
  >   /bin/bash (glob)
  > EOF

  $ grill check-sh.t
  .
  # Ran 1 test, 0 skipped, 0 failed.

  $ grill -shell=/bin/bash check-bash.t
  .
  # Ran 1 test, 0 skipped, 0 failed.
