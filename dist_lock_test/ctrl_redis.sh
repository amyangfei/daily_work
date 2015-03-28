#!/bin/bash

redis_bin="/usr/local/bin/redis-server"

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
    echo "starting redis servers..."
    $redis_bin --port 6379 $* >/dev/null &
    $redis_bin --port 6380 $* >/dev/null &
    $redis_bin --port 6381 $* >/dev/null &

    $redis_bin --port 6479 --appendonly yes --appendfsync always &
    $redis_bin --port 6480 --appendonly yes --appendfsync always &
    $redis_bin --port 6481 --appendonly yes --appendfsync always &
}

stop() {
    echo "stopping redis servers..."
    pids=$(ps aux|grep '[r]edis-server'|awk '{print $2}')
    for pid in $pids
    do
        kill $pid
    done
}

status() {
    ps aux|grep '[r]edis-server'
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
