# codis-rebalancer
A tool to auto rebalance slots between groups of a codis cluster.
But now only works with groups with same memory.

##Usage
```
Usage of ./codis-rebalancer:
  -d 127.0.0.1:18080
        Dashboard Addr: 127.0.0.1:18080
  -i 10
        Migrate interval. (default 10)
```

##Todo
```
1.  auto rebalance base on the MAXMEMORY of codis-server settings.
```
