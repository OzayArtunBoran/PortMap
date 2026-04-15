# Skill: GitHub Repo Setup

## Amaç
GitHub repository'sini son haline getir: issue template'leri, CONTRIBUTING.md, LICENSE, etiketler.

## Girdiler
- `project-spec.yml` → `project`, `repo` bölümleri

---

## Dosyalar

### CONTRIBUTING.md

```markdown
# Contributing to {display_name}

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/{author.github}/{name}.git
cd {name}
{kurulum komutları — dile göre}
```

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Ensure all tests pass: `{quality.test_command}`
5. Ensure linting passes: `{quality.lint_command}`
6. Commit with conventional commits (see below)
7. Open a Pull Request

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `test:` — Adding or updating tests
- `chore:` — Maintenance (CI, deps, config)
- `refactor:` — Code change that neither fixes a bug nor adds a feature

Examples:
```
feat: add Slack notification support
fix: handle empty config file gracefully
docs: update CLI reference in README
test: add edge case tests for config parser
chore: update CI workflow to Go 1.22
```

## Code Style

{Go projeler:}
- Follow standard Go conventions
- Run `golangci-lint run` before committing
- Keep functions small and focused

{Python projeler:}
- Follow PEP 8
- Run `ruff check .` and `ruff format .` before committing
- Type hints for all function signatures

## Reporting Issues

- Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) template for bugs
- Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) template for ideas
- Check existing issues before creating a new one

## License

By contributing, you agree that your contributions will be licensed under the project's [{license}](LICENSE) license.
```

---

### .github/ISSUE_TEMPLATE/bug_report.md

```markdown
---
name: Bug Report
about: Report a bug to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Describe the Bug

A clear and concise description of what the bug is.

## To Reproduce

Steps to reproduce the behavior:

1. Run `{name} ...`
2. With config `...`
3. See error: `...`

## Expected Behavior

What you expected to happen.

## Actual Behavior

What actually happened. Include error messages and stack traces if applicable.

## Environment

- **OS:** [e.g., Ubuntu 22.04, macOS 14.2]
- **{display_name} Version:** [e.g., v1.0.0]
- **Go/Python Version:** [e.g., go1.22.0]
- **Install Method:** [go install / binary / docker]

## Config File

```yaml
# Relevant parts of your .{name}.yml (remove sensitive values)
```

## Additional Context

Any other context, screenshots, or logs that might help.
```

---

### .github/ISSUE_TEMPLATE/feature_request.md

```markdown
---
name: Feature Request
about: Suggest a new feature or improvement
title: '[FEATURE] '
labels: enhancement
assignees: ''
---

## Problem

What problem does this feature solve? What's the current limitation?

## Proposed Solution

How should this feature work? Describe the desired behavior.

## Alternatives Considered

Have you considered any alternative approaches? Why is this one preferred?

## Additional Context

Any mockups, examples from other tools, or additional information.
```

---

### .github/ISSUE_TEMPLATE/config.yml (opsiyonel)

```yaml
blank_issues_enabled: false
contact_links:
  - name: Questions & Discussions
    url: https://github.com/{author.github}/{name}/discussions
    about: Ask questions and discuss ideas
```

---

### LICENSE

```bash
# MIT lisansı oluştur
# NOT: Claude Code content filtering hatası verebilir. curl ile indir:
curl -sL https://raw.githubusercontent.com/spdx/license-list-data/main/text/MIT.txt -o LICENSE

# Yıl ve isim güncelle:
sed -i "s/<year>/$(date +%Y)/" LICENSE
sed -i "s/<copyright holders>/{author.name}/" LICENSE
```

**AGPL-3.0 için:**
```bash
curl -sL https://raw.githubusercontent.com/spdx/license-list-data/main/text/AGPL-3.0-only.txt -o LICENSE
```

---

## Git Setup & İlk Push

```bash
# Tag
git tag v1.0.0

# Remote ekle
git remote add origin git@github.com:{author.github}/{name}.git

# Push
git push -u origin main --tags
```

---

## GitHub Repo Ayarları (manuel — GitHub UI'dan)

Aşağıdaki ayarlar GitHub web arayüzünden yapılmalı:

1. **Description:** {project.tagline}
2. **Website:** {author.website}/projects/{name} (varsa)
3. **Topics:** İlgili etiketler (dile göre: `go`, `cli`, `devops`, vs.)
4. **Features:**
   - Issues ✅
   - Discussions ✅ (opsiyonel)
   - Wiki ❌
5. **Branch protection (main):**
   - Require pull request reviews ✅
   - Require status checks ✅ (CI workflow)
   - Require branches be up to date ✅

---

## Doğrulama
- CONTRIBUTING.md mevcut ve eksiksiz
- Issue template'leri `.github/ISSUE_TEMPLATE/` altında
- LICENSE dosyası doğru lisans metniyle
- `git tag v1.0.0` oluşturulmuş
- `git remote -v` doğru URL'ye işaret ediyor
