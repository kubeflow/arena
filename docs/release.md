# Releasing Arena

## Prerequisites

- [Write](https://docs.github.com/organizations/managing-access-to-your-organizations-repositories/repository-permission-levels-for-an-organization#permission-levels-for-repositories-owned-by-an-organization) permission for the [kubeflow/arena](https://github.com/kubeflow/arena) repository.

- Create a [GitHub Token](https://docs.github.com/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token).

- Install `PyGithub`:

    ```bash
    pip install PyGithub==2.3.0
    ```

## Versioning Policy

Arena version format follows [Semantic Versioning](https://semver.org/). Arena versions are in the format of `vX.Y.Z`, where `X` is the major version, `Y` is the minor version, and `Z` is the patch version. The patch version contains only bug fixes.

## Release Process

### Update Versions

1. Modify `VERSION` file in the root directory of the project:

    - For the RC tag as follows:

    ```bash
    vX.Y.Z-rc.N
    ```

    - For the official release tag as follows:

    ```bash
    vX.Y.Z
    ```

2. Modify `version` and `appVersion` in `Chart.yaml`:

    ```bash
    # Get version and remove the leading 'v'
    VERSION=$(cat VERSION | sed "s/^v//")

    # Change the version and appVersion in Chart.yaml
    # On Linux
    sed -i "s/^version.*/version: ${VERSION}/" arena-artifacts/Chart.yaml
    sed -i "s/^appVersion.*/appVersion: ${VERSION}/" arena-artifacts/Chart.yaml

    # On MacOS
    sed -i '' "s/^version.*/version: ${VERSION}/" arena-artifacts/Chart.yaml
    sed -i '' "s/^appVersion.*/appVersion: ${VERSION}/" arena-artifacts/Chart.yaml
    ```

3. Commit and push the changes to your own branch:

    ```bash
    git add VERSION
    git add arena-artifacts/Chart.yaml
    git commit -s -m "Release v${VERSION}"
    git push --set-upstream origin $(git rev-parse --abbrev-ref HEAD)
    ```

4. Open a new PR to the master branch.

### Publish release

After `VERSION` file is modified and pushed to the master branch, a release workflow will be triggered to create a new draft release with the arena installer packaged as artifacts. After modifying the release notes, then publish the release.

## Update Changelog

1. Update the `CHANGELOG.md` file by running:

    ```bash
    # Use your GitHub token.
    GH_TOKEN=<github-token>
    # The previous release version, e.g. v1.7.1
    PREVIOUS_RELEASE=vX.Y.Z
    # The current release version, e.g. v1.8.0
    CURRENT_RELEASE=vX.Y.Z

    python hack/generate-changelog.py \
        --token=${GH_TOKEN} \
        --range=${PREVIOUS_RELEASE}..<CURRENT_RELEASE>
    ```

2. Group PRs in the `CHANGELOG.md` file into **Features**, **Bug Fixes** and **Misc**, etc.

3. Finally, open a new PR to the master branch with the updated Changelog.
