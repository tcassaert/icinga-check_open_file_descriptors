# Icinga check for open file descriptors

This check provides functionality to check the open file descriptors per process running on the system.

## Usage

```
# check_open_files -h
  -critical float
        critical treshold (default 0.9)
  -process string
        Name of the process to watch
  -warning float
        warning treshold (default 0.8
```

You can define a single processname to watch. When doing that, this process will be watched and will always send its performance data.
If you do not add a processname, all processes will be checked and performance data will only be added when there's a warning or a critical value.
