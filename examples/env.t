Check environment variables:

  $ echo "$LANG"
  C
  $ echo "$LC_ALL"
  C
  $ echo "$LANGUAGE"
  C
  $ echo "$TZ"
  GMT
  $ echo "$CDPATH"
  
  $ echo "$GREP_OPTIONS"
  
  $ echo "$CRAMTMP"
  .+ (re)
  $ echo "$TESTDIR"
  **/examples (glob)
  $ ls "$TESTDIR"
  bare.t
  empty.t
  env.t
  missingeol.t
  test.t

# TODO re-enable when files are added back
# fail.t
# skip.t

  $ echo "$TESTFILE"
  env.t

  $ pwd
  **/grilltests*/examples/env.t (glob)
