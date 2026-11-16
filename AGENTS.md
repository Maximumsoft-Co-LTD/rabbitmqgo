# AGENTS.md

<!-- secrets-policy -->
## Secrets — never commit (required)

- Secret values (keys, tokens, passwords, connection strings with `user:pass@`) live only in AWS Secrets Manager / GCP Secret Manager and reach pods through ExternalSecret. Git holds only secret and property names.
- Before the first commit in this repo, check that the pre-commit hook is installed: `test -f .git/hooks/pre-commit`. If it is not, remind the user to run `brew install pre-commit && pre-commit install` (or `pipx install pre-commit`).
- Never bypass gitleaks with `git commit --no-verify`. If it flags a value, move the value to the secret store. Add a fingerprint to `.gitleaksignore` only for a real false positive.
- Backstop: an org ruleset scans every pull request and blocks new secrets. A secret pushed to any branch counts as leaked and must be rotated.
- Never print secret values in logs, chat or pull requests; compare by hash.
