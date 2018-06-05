# Graviton Release Process

1. Make the release branch from `master`, e.g.: `release/1.0.9`
1. Update `CHANGELOG.md` manually with a helpful description of any commits that have landed since the last release.
1. Run `make clean && make`
1. Verify the version string matches the new release and latest git hash (just with `stardog-graviton --version` on the new binary)
1. Run `make test`
1. Run: `python release/release.py <path to license> <path to stardog release zip> <path to ssh key> <matching aws ssh key name>`
1. Once the tests all pass, tag the release in the branch, e.g.: `git tag 1.0.9`
1. Push up the release branch and tag
1. The release files can be found at `release/darwin_amd64/stardog-graviton_<version>_darwin_amd64.zip` and `release/linux_amd64/stardog-graviton_<version>_linux_amd64.zip`. Upload them to the newly found release here: https://github.com/stardog-union/stardog-graviton/releases
1. PR the release branch to master and merge it
