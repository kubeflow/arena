## Purpose of this PR

<!-- Provide a clear and concise description of the changes. Link to relevant issues. -->

**Proposed changes:**

-

## Commit Message Format

This project uses [conventional commits](https://www.conventionalcommits.org/) for commit hygiene. The release changelog is driven by PR labels, which are auto-assigned based on your branch name prefix and changed files.

**Example commit messages:**

```
feat: add GPU sharing support for MPIJob
fix: correct pytorch replica count in dry-run output
feat(serving): add vLLM as serving framework
docs: update installation guide for v2
```

> **Note:** PR labels are auto-assigned by the Labeler workflow based on your branch name prefix and changed files. This determines how your change appears in the release changelog — `feat/` branch → Features, `fix/` branch → Bug Fixes, `breaking/` branch → Breaking Changes, `doc/` or `docs/` branch or all files in `docs/`/`examples/` → Documentation, `chore/` or `chores/` branch → Maintenance, `refactor/` branch → Refactoring, `test/`/`ci/` branch, all files are `.github/` CI configs, or only `VERSION` changed → excluded.

## Checklist

- [ ] Code follows the project's style guidelines (`make v2-fmt`)
- [ ] Code passes vet (`make v2-vet`)
- [ ] Code passes lint (`make v2-lint`)
- [ ] Tests pass (`make v2-test`)
- [ ] `go mod tidy` has been run
- [ ] Commit messages follow conventional commit format
