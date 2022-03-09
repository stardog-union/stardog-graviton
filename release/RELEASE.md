# Graviton Release Process

The graviton build process must use go `go version go1.15.15`.

1. Make the release branch from `develop`, e.g.: `release/1.0.9`
1. Update `CHANGELOG.md` manually with a helpful description of any commits that have landed since the last release.
1. Tag the release: `git tag 1.0.9`
1. Run `make clean && make`
1. Verify the version string matches the new release and latest git hash (just with `stardog-graviton --version` on the new binary)
1. Run `make test`
1. Run: `python release/release.py <path to license> <path to stardog release zip> <path to ssh key> <matching aws ssh key name>`. This runs in us-west-1 so make sure the key exists in that region.
1. Once the tests pass, push up the release branch and tag: `git push origin release/1.0.9` and `git push origin <tag name>`
1. The release files can be found at `release/darwin_amd64/stardog-graviton_<version>_darwin_amd64.zip` and `release/linux_amd64/stardog-graviton_<version>_linux_amd64.zip`. Upload them to the newly found release here: https://github.com/stardog-union/stardog-graviton/releases
1. PR the release branch to develop and merge it
