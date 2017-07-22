```
                                    _       _        _ 
 David McNett            _ __   ___| |_ ___| | _____| |
                        | '_ \ / _ \ __/ __| |/ / _ \ |
 http://macnugget.org/  | | | |  __/ |_\__ \   <  __/ |
                        |_| |_|\___|\__|___/_|\_\___|_|

                        netskel environment synchronizer
```
INTRODUCTION

netskel is an http(s) based file synchronization tool which can be used to 
mirror a central repository of files to multiple UNIX shell accounts.  It was
built to create a simple and automated mechanism for users to push out their
various shell environment config files to all the machines where they have
shell accounts.

In particular it's useful for pushing out a common .cshrc, .bashrc, or ssh
authorized_keys file to all your hosts.  Using netskel will allow you to make
centralized changes and updates to your unix enviornment without suffering
through the tedium of applying those changes to the multitude of hosts where
you have an account.
