# codis-rebalancer
```
A tool to auto rebalance slots between groups of a codis cluster.
But now only works with groups with same memory.
And only for codis 3.0.
```

##Usage
```
Usage of ./codis-rebalancer:
  -d string
        Dashboard Addr. (default "127.0.0.1:18080")
  -f    Do the action. Default only show the actions.
  -i int
        Migrate interval. (default 10)
```

##Todo
```
1.  auto rebalance base on the MAXMEMORY of codis-server settings.
```
