Don't sterilize environment:

  $ cat > env.t <<EOF
  >   \$ echo "\$LANG"
  >   C
  >   \$ echo "\$TZ"
  >   foo
  >   \$ echo "\$CDPATH"
  >   bar
  >   \$ echo "\$GREP_OPTIONS"
  >   baz
  > EOF

  $ export TZ=foo
  $ export CDPATH=bar
  $ export GREP_OPTIONS=baz
  $ grill -preserve-env env.t
  .
  # Ran 1 test, 0 skipped, 0 failed.
