[![CI](https://github.com/nugget/netskel/workflows/Go/badge.svg)](https://github.com/nugget/netskel/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nugget/netskel)](https://goreportcard.com/report/github.com/nugget/netskel)
[![codecov](https://codecov.io/gh/nugget/netskel/branch/master/graph/badge.svg)](https://codecov.io/gh/nugget/netskel)
[![Release](https://img.shields.io/github/release/nugget/netskel.svg)](https://github.com/nugget/netskel/releases)

```text
                                    _       _        _ 
 David McNett            _ __   ___| |_ ___| | _____| |
                        | '_ \ / _ \ __/ __| |/ / _ \ |
 http://macnugget.org/  | | | |  __/ |_\__ \   <  __/ |
                        |_| |_|\___|\__|___/_|\_\___|_|

                        netskel environment synchronizer
```
# INTRODUCTION

netskel is an ssh based file synchronization tool which can be used to 
mirror a central repository of files to multiple UNIX shell accounts.  It was
built to create a simple and automated mechanism for users to push out their
various shell environment config files to all the machines where they have
shell accounts.

In particular it's useful for pushing out a common .cshrc, .bashrc, or ssh
authorized_keys file to all your hosts.  Using netskel will allow you to make
centralized changes and updates to your unix enviornment without suffering
through the tedium of applying those changes to the multitude of hosts where
you have an account.

# VERSION 3.0 BREAKING CHANGES 

If you are a current Netskel user with a previous version, please be aware
the current v3.0 release is a complete and total overhaul of the Netskel
system.  The old releases required Apache web server and used http/https as
the deployment mechanism.  This update re-write is entirely ssh-based and 
no longer uses or requires a web server to host the back end service

## SERVER REQUIREMENTS

* [Go](https://golang.org) compiler to build the server binaries
* A server that can be reached via ssh from your client installations
* A dedicated user account on that server to operate the backend service

## CLIENT REQUIREMENTS

* A reasonably POSIX-flavoured system that has Bourne shell (`/bin/sh`) and 
  a modest assortment of the usual Unix tools.
* The `xxd` binary (part of vim, should be on any Linux and macOS box but will
  require the vim port/package on FreeBSD.  This replaces the older Netskel's
  reliance on `uuencode` and `uudecode` which is no longer reliably present on
  modern machines.

# INSTALLATION

* Create a user on your server to hold the server code and your userland
  database.  This user's homedirectory will be your install path.  Suggested
  `/usr/local/netskel` with the user's shell as `/usr/local/netskel/bin/server`
  and no password.
* Copy your personal `authorized_keys` file to the netskel user's `.ssh` 
  directory with the proper permissions.
* Run `make && make install` from the `server` directory of this repo
* Place any files and directories you want to deploy in the `./db/` folder of
  your Netskel installation.  By default this is a git repo so you can use
  version control to track changes and additions to it.
* Run `make userzero` from the `server` directory of this repo, which will
  bootstrap your current login on your host as the first deployment of the
  client.  Verify that the server info in `~/.netskel/config` makes sense
  to you.

You should now be able to use `netskel push <hostname>` to deploy the netskel
client from your current account to other hosts.
