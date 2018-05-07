# Graviton Release Process

1. Make the release branch from `master`
1. Update `CHANGELOG.md`
1. Tag the release in the branch
1. Run `make`
1. Run `make test`
1. Run: `python release/release.py <path to license> <path to stardog release zip> <path to ssh key> <matching aws ssh key name>`
1. The release files can be found at `release/darwin_amd64/stardog-graviton_<version>_darwin_amd64.zip` and `release/linux_amd64/stardog-graviton_<version>_linux_amd64.zip`.
