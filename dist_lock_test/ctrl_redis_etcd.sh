#!/bin/bash

redis_bin="/usr/local/bin/redis-server"
etcd_bin="/usr/local/bin/etcd"
cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

sep="----------------------------------------------------"
red='\e[0;31m'
NC='\e[0m' # No Color

show_help() {
    echo "Usage:"
    echo ${sep}

    echo -e "${red}start${NC} multi redis master"
    echo example: "$0 start"
    echo ""

    echo -e "${red}stop${NC} all redis master"
    echo example: "$0 stop"
    echo ""

    echo -e "${red}restart${NC} all redis master"
    echo example: "$0 restart"
    echo ""

    echo ${sep}
    exit 1
}

start() {
    mkdir -p $cur/build/redis_6379 $cur/build/redis_6380 $cur/build/redis_6381
    mkdir -p $cur/build/redis_6382 $cur/build/redis_6383 $cur/build/redis_6384
    mkdir -p $cur/build/redis_6385
    mkdir -p $cur/build/redis_6479 $cur/build/redis_6480 $cur/build/redis_6481

    echo "starting redis servers..."
    $redis_bin --port 6379 --dir $cur/build/redis_6379 $* > /dev/null &
    $redis_bin --port 6380 --dir $cur/build/redis_6380 $* > /dev/null &
    $redis_bin --port 6381 --dir $cur/build/redis_6381 $* > /dev/null &
    $redis_bin --port 6382 --dir $cur/build/redis_6382 $* > /dev/null &
    $redis_bin --port 6383 --dir $cur/build/redis_6383 $* > /dev/null &
    $redis_bin --port 6384 --dir $cur/build/redis_6384 $* > /dev/null &
    $redis_bin --port 6385 --dir $cur/build/redis_6385 $* > /dev/null &

    $redis_bin --port 6479 --dir $cur/build/redis_6479 --appendonly yes --appendfsync always > /dev/null &
    $redis_bin --port 6480 --dir $cur/build/redis_6480 --appendonly yes --appendfsync always > /dev/null &
    $redis_bin --port 6481 --dir $cur/build/redis_6481 --appendonly yes --appendfsync always > /dev/null &

    $etcd_bin > /dev/null &
}

stop() {
    echo "stopping redis servers..."
    pids=$(ps aux|grep '[r]edis-server'|awk '{print $2}')
    for pid in $pids
    do
        kill $pid
    done

    echo "stopping etcd server..."
    pids=$(ps aux|grep [${etcd_bin:0:1}]${etcd_bin:1}|awk '{print $2}')
    for pid in $pids
    do
        kill $pid
    done
}

status() {
    ps aux|grep '[r]edis-server'
    ps aux|grep [${etcd_bin:0:1}]${etcd_bin:1}
}

case $1 in
    start)
        shift 1
        start $*
        ;;
    stop)
        stop $*
        ;;
    restart)
        shift 1
        stop $*
        sleep 1
        start $*
        ;;
    status)
        status $*
        ;;
    *)
        show_help $*
        ;;
esac
