# Contributing

## Philosophy

GDA is built for one thing: versioning research data without the
complexity of git-annex. Every feature decision should pass this test:

> Does this help a neuroscientist on an HPC cluster manage 700 GB of
> BIDS data?

If it doesn't, it probably doesn't belong.

## Pull requests

- Keep scope small. One feature per PR.
- No AI-generated code in commits. Every line should be intentional.
- Write tests that exercise the feature with real directory structures,
  not mocks.
- Run the full test suite before opening:

      go test ./...
      go build -o gda ./cmd/gda/

## Code style

- No comments that restate what the code says.
- No doc comments on unexported functions unless the logic is
  non-obvious.
- Error messages are lowercase, no punctuation.
- Go 1.26 idioms.

## What needs help

- rsync/rclone remote backend for push/pull
- Windows compatibility (not a priority, but patches welcome)
- Integration testing on real OpenNeuro datasets
- Performance profiling on 100 GB+ datasets
