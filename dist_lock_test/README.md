
## Performance test between distributed lock using redis and nolock

redis can be set with aof and fsync=always support and no aof

### Dependency

* [go](https://golang.org/)
* [python](https://www.python.org/)
* [etcd](https://github.com/coreos/etcd), install to /usr/local/bin/etcd or modify ctrl_redis_etcd.sh
* [redis](http://redis.io), install to /usr/local/bin/redis-server or modify ctrl_redis_etcd.sh
* [gnuplot](http://www.gnuplot.info/)
* [chinese font wqy](http://wenq.org/wqy2/index.cgi), e.g. on ubuntu: sudo apt-get install ttf-wqy-microhei

### Build & Run benchmark

    make
    ./ctrl_redis_etcd.sh start
    python bench.py

