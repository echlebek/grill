# Grill - Interrogate your programs

Grill is a cram-like test runner (see https://github.com/brodie/cram).
The project aims to provide better performance and independence from Python runtime.

Grill tests are almost exactly like cram tests, and grill should work the same
way as cram in most, but not all, cases.

Notable differences are:
  * (re) keyword: PCRE is not supported. Instead, Go's regexp language is.
  * (glob) keyword: Use `**` to glob across directory separators.
  * Short flags are not supported.

Still TODO / under consideration:
  * Concurrent test execution
  * Interactive mode
  * Skipping
  * Variable command indent

Bug reports and test cases are always appreciated!
