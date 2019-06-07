# SSMEB

A simple tool to get ssm parameters to an .ebextensions file.

## Usage

Create a template like the one in `example/template.yaml`

### Example

```bash
ssmeb -i example/template.yaml -o .ebextensions/env_variables.config
```

### Help

```text
Usage of ./ssmeb:
-e environment
    environment flag shorthand
-environment string
    environment name used as prefix for the ssm parameters (e.g. codacy)
-i input
    input flag shorthand
-input string
    input template environment variables config
-m mode
    mode flag shorthand (default "get")
-mode string
    enable set or get mode (default "get")
-o output
    output flag shorthand
-output string
    destination of the resulting elastic beanstalk data
```

## What is Codacy

[Codacy](https://www.codacy.com) is an Automated Code Review Tool
that monitors your technical debt, helps you improve your code quality,
teaches best practices to your developers, and helps you save time in
Code Reviews.

### Among Codacy’s features

- Identify new Static Analysis issues
- Commit and Pull Request Analysis with GitHub, BitBucket/Stash, GitLab
  (and also direct git repositories)
- Auto-comments on Commits and Pull Requests
- Integrations with Slack, HipChat, Jira, YouTrack
- Track issues in Code Style, Security, Error Proneness, Performance,
  Unused Code and other categories

Codacy also helps keep track of Code Coverage, Code Duplication, and
Code Complexity.

Codacy supports PHP, Python, Ruby, Java, JavaScript, and Scala, among
others.

## Free for Open Source

Codacy is free for Open Source projects.
