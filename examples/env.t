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
  env.t
  missingeol.t

# TODO re-enable when files are added back
# empty.t
# fail.t
# skip.t
# test.t

  $ echo "$TESTFILE"
  env.t

  $ pwd
  **/grilltests*/env.t (glob)
