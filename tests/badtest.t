Report test parsing errors:

  $ echo '  notacommand' > bad.t
  $ grill bad.t
  ** syntax error parsing line 1: expected '$ ' after two spaces (glob)
  [1]
