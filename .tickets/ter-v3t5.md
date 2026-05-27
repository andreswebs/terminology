---
id: ter-v3t5
status: closed
deps: []
links: [ter-qxrg]
created: 2026-05-25T14:19:42Z
type: bug
priority: 2
assignee: Andre Silva
tags: [e1, bug, ux]
---
# E1.BUG — Unknown subcommand not identified by name in error message

When an unknown subcommand is given (e.g. "terminology bogus"), the error
envelope returns error.code "no_subcommand" with message "no subcommand
specified" — it does not mention "bogus" by name.

An agent cannot distinguish between "no command given" and "unknown command
given" from the error response alone.

## Reproduction

```sh
TT="./bin/terminology-$(go env GOOS)-$(go env GOARCH)"
$TT bogus 2>&1
```

Returns exit 2 (correct), but message says "no subcommand specified" instead
of something like "unknown subcommand: bogus".

## Affected test case

TC-ROOT-006

## Fix direction

Detect unknown subcommands in the error handler and include the unrecognized
name in the error message. Consider a distinct error code (e.g.
"unknown_subcommand") to differentiate from the bare-invocation case.


## Notes

**2026-05-25T14:39:21Z**

Fixed rootAction to check cmd.Args() before returning the error sentinel. If args are present, the first arg is treated as an unknown subcommand and a new terr.New with code 'unknown_subcommand' is returned, including the name in the message. The bare-invocation case (no args) still returns the original 'no_subcommand' sentinel. Added TestUnknownSubcommand and TestNoSubcommand_StillWorks tests.
