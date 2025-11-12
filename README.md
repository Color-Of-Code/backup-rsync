# backup-rsync

Local backup using rsync as a data copier and ZFS datasets as a destination.

Go tool used for my own private purposes.

**Use at your own risk!**

## Goals

* Have the coverage of paths in the source devices checked for completeness
* Document the data copy jobs
* Easily run single jobs on the command line
* Extensive logging of the operations performed
* Dry run for checking what would be performed
* Provide checks for ZFS (snapshot management, size limits reached, ...)
