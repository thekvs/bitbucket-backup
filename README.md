## Compilation

[![Build Status](https://travis-ci.org/thekvs/bitbucket-backup.svg?branch=master)](https://travis-ci.org/thekvs/bitbucket-backup)

Run `make` command in the repository's root, compiled binary will be located in the `bin` subfolder. You need Go programming language version >= 1.11 to compile this project.

## Usage

```
$ bitbucket-backup -h
Usage:
  bitbucket-backup [OPTIONS]

Application Options:
  -u, --username=      bitbucket username
  -p, --password=      bitbucket user's password
  -l, --location=      local backup location
  -a, --attempts=      number of attempts to make before giving up (default: 1)
  -i, --ignore=        repository to ignore, may be specified several times
  -b, --bare           clone bare repository (git only)
  -w, --with-wiki      also backup wiki
  -P, --prune          prune repo on remote update (git only)
  -h, --http           clone/update via https instead of ssh
  -d, --dry-run        do nothing, just print commands
  -v, --verbose        be more verbose
  -s, --show-progress  show progress bar

Help Options:
  -h, --help           Show this help message
```

It is highly recommended to use separate application password (how to create one explayned [here](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html)) with limited access rights instead of your normal Bitbucket password.
