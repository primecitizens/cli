# Contribute to `cli`

We would encourage you to read our general contribution guidelines at [primecitizens/pcz](https://github.com/primecitizens/pcz/blob/master/CONTRIBUTING.md) (if you haven't done so).

No additional rules to this repository, but here are some extra help when doing code contribution:

## Code Contribution Checklist

To keep this project maintainable by peasants like us, we created this checklist:

- [ ] Limit package imports. (not applicable to testings)
  - [ ] No external modules other than the standard library.
  - [ ] No `fmt` and any packages depending on `fmt`.
  - [ ] Only import `reflect` in `*_reflect.go`.
- [ ] Ensure VP implementations in `vp_generic.go` are always zero size.
- [ ] Ensure minimum data escape to heap.
  - [ ] Allow necessary and zero-size data to escape.
  - [ ] Run `go build -gcflags '-m -l' ./` in the module root to check unnecessary escapes.
- [ ] When updating completion scripts (`scripts/*`), follow the [checklist for completion scripts](./scripts/README.md#maintenance-checklist)
- [ ] Documentation
  - [ ] Run `pkgsite` to inspect examples and comments (to install, run `go install golang.org/x/pkgsite/cmd/pkgsite@latest`)
