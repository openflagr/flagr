These tests are for determining the physical memory from the data returned by
/proc/meminfo on Linux hosts. The total physical memory of the linux system is 
reported as part of the enviornment values. The key used by the Python agent
is 'Total Physical Memory (MB)'. 

The names of all test files should be of the form `meminfo_nnnnMB.txt`. The
value `nnnn` in the filename is the physical memory of that system in MB.
