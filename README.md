# Icinga check for open file descriptors

This check provides functionality to check the open file descriptors per process running on the system.

## Usage

```
# check_open_files -h
  -critical float
        critical treshold (default 0.9)
  -warning float
        warning treshold (default 0.8)
```

The tresholds are a percentage of the maximum open files per process.
