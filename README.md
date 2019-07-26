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
  -i, --ignore=        list of repositories to ignore
  -m, --mirror         git only: clone bare repository
  -w, --with-wiki      also backup wiki
      --prune          prune repo on remote update
      --http           fetch via https instead of ssh
      --dry-run        do nothing, just print commands
      --verbose        be more verbose
      --show-progress  show progressbar

Help Options:
  -h, --help           Show this help message
```

It is highly recommended to use separate application password (how to create one explayned [here](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html)) with limited access rights instead of your normal Bitbucket password.
