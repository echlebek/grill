--keep-tmpdir leaves behind temporary directories. Ensure
temporary directories are distinct for .t files with the
same name but in different branches.

  $ mkdir -p sub1/sub2
  $ echo '  $ true' > sub1/abc.t
  $ echo '  $ true' > sub1/sub2/abc.t
  $ grill --keep-tmpdir sub1/abc.t sub1/sub2/abc.t >log 2>&1

  $ cat log
  .
  # Ran 1 test, 0 skipped, 0 failed.
  .
  # Ran 1 test, 0 skipped, 0 failed.
  # Kept temporary directory: **/grilltests*/** (glob)

  $ find `cat log | grep -oP "(?<=directory: ).*"` -name '*.t'
  **/grilltests*/sub1/sub2/abc.t (glob)
  **/grilltests*/sub1/abc.t (glob)
