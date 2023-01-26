# Bumping versions

To bump versions you just need to run `make version-updates` and commit the
change to vendor and go.mod.

When building the binary, the Makefile will detect the versions from the go.mod and
generate a file in pkg/version.json which is used to show the different version
components.
