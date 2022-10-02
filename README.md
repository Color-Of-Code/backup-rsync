# backup-rsync

Local backup using rsync as a data copier and ZFS datasets as a destination.

Ruby based scripts used for my own private purposes.

**Use at your own risk!**

## Goals

* Have the coverage of paths in the source devices checked for completeness
* Document the data copy jobs
* Easily run single jobs on the command line
* Provide checks for ZFS (snapshot management, size limits reached, ...)
