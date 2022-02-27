# Grill - interrogate your programs.
A work-in-progress cram-like test runner. (See https://github.com/brodie/cram)

Grill tests are almost exactly like cram tests, and grill should work the same
way as cram in most, but not all, cases.

DONE:
  * A command line tool (grill) that can execute tests in grill format.
  * Support for regex and glob line matching.
  * Shell variables are persisted between commands.

TODO:
  * Support cram's environment variables.
  * Gradually add support for cram's command-line flags.
  * Flesh out the tests for the test.t reader.

WONTFIX:
  * PCRE is not supported. Instead, Go's regexp language is.
  * Short flags are not supported.

Additional differences:
  * glob keyword: Use `**` to glob across directory separators.

There are probably lots of bugs at this point, bug reports and test cases would
be appreciated.
