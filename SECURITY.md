# Security policy

TrashDay deletes files with administrator rights. A bug that deletes the wrong thing counts as a security issue.

## Reporting

Please **don't open a public issue**. Use [GitHub private vulnerability reporting](https://github.com/Sannho/trashday/security/advisories/new) instead, and include:

- what got deleted, or could be deleted, that shouldn't be
- the Windows version and the TrashDay version
- the environment variables or folder layout that trigger it, if known

You'll get an answer as soon as possible. A fix ships as a new release, with credit if you want it.

## Supported versions

Only the [latest release](https://github.com/Sannho/trashday/releases/latest) gets fixes.

## Official builds

Download TrashDay only from this repository's releases. Each `trashday.exe` is built by GitHub Actions from the tagged source and ships with a SHA-256 checksum.
