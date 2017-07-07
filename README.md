# stash_users
Command for managing stash users

How to use:

```
usage: stash_users [--help|-h] [-v] [-s] --log LOGPATH
                [ [--cuser CROWD_USER] [--cpass CROWD_PASS] [--buser BITBUCKET_USER] [--bpass BITBUCKET_PASS] | [--cred CRED_FILE] ]
  -bhost string
                The Bitbucket host url.
                        Example: https://host.com
  -bpass string
                The password for Bitbucket
                        If missing, and no credentials file is provided, it will be requested at runtime
  -buser string
                The Bitbucket user
                        If missing, and no credentials file is provided, it will be requested at runtime
  -chost string
                The Crowd host url.
                        Example: https://host.com
  -cpass string
                The password for Crowd
                        If missing, and no credentials file is provided, it will be requested at runtime
  -cred string
                A file containing the credentials for Bitbucket and Crowd
                        Every line must have the following structure (2 columns): {Bitbucket | Crowd} user:password
                        If there are any parameters missing, they will be requested at runtime
  -cuser string
                The Crowd user
                        If missing, and no credentials file is provided, it will be requested at runtime
  -help
                Show this help message and exit
  -log string
                The log path where the loggers will write to.
                        Example: /var/etc/log/stash
  -s            Execute in simulation mode (default true)
  -v            Execute in verbose mode

```
