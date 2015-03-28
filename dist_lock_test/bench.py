#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sys
from os import mkdir, chdir
from os.path import join
from subprocess import check_output, PIPE, Popen
from getopt import getopt

output_path = lambda s="": join("output", s)


# All test_client runs and their cli args.
runs = {
    "nolock": ["./build/nolock"],
    "redlock": ["./build/redlock"],
    "redlock_withsync": [
        "./build/redlock",
        "-r='tcp://127.0.0.1:6479;tcp://127.0.0.1:6480;tcp://127.0.0.1:6481'"],
}


# Consistent graph colours defined for each of the runs.
colours = {
    "nolock": "red",
    "redlock": "green",
    "redlock_withsync": "blue",
    # "blue", "violet", "orange"
}


# Groups of runs mapped to each graph.
plots = {
    "rt_under_different_concurrent": ["nolock", "redlock", "redlock_withsync"],
}

xlabels = {
    "rt_under_different_concurrent": "并发客户端数",
}


def run_clients(args):
    results = []
    iter_counts = 3
    # iter_counts = 11
    total_reqs = 2 ** (iter_counts + 1)
    num_clients = [2**i for i in xrange(iter_counts)]
    req_per_cli = [total_reqs / num_clients[i] for i in xrange(iter_counts)]

    for idx in range(iter_counts):
        bar = ("#" * (idx + 1)).ljust(iter_counts)
        sys.stdout.write("\r[%s] %s/%s " % (bar, idx + 1, iter_counts))
        sys.stdout.flush()

        run_args = args[:]
        run_args.extend(["-c=%d" % num_clients[idx]])
        run_args.extend(["-n=%s" % req_per_cli[idx]])

        print run_args
        out = check_output(run_args, stderr=PIPE)

        results.append(out.split(" ")[2].strip())

    sys.stdout.write("\n")
    return results


def prepare():
    # Store all results in an output directory.
    try:
        mkdir(output_path())
    except OSError as e:
        print e


def bench():
    prepare()

    for name, args in runs.iteritems():
        with open(output_path(name + ".dat"), "w") as f:
            f.write("\n".join(run_clients(args)))


def draw():
    # change working dir to output
    chdir(output_path())

    plot_basic = """set terminal png enhanced font 'Georgia,12' size 960,600
                    set output "%(name)s.png"
                    set grid y
                    set xlabel "%(xlabel)s"
                    set ylabel "总时间（单位毫秒）"
                    set decimal locale
                    set format y "%%'g"
                    set xrange [1:%(clients)s]
                    plot %(lines)s"""

    line = '"%s.dat" using ($0+1):1 with lines title "%s" lw 2 lt rgb "%s"'
    for name, names in plots.items():
        #name = output_path(name)
        with open(names[0] + ".dat", "r") as f:
            clients = len(f.read().split())
        with open(name + ".p", "w") as f:
            lines = ", ".join([line % (l, l.replace("_", " "), colours[l])
                               for l in names])
            f.write(plot_basic %
                    {"name": name, "lines": lines,
                        "clients": clients, "xlabel": xlabels[name]})
        Popen(["gnuplot", name + ".p"], stderr=sys.stdout)


def show_help():
    print '''usage: bench.py [-h] [--draw-only] [--run-only]

optional arguments:
  -h, --help            show this help message and exit
  --draw-only           generate graphs only without running benchmark
  --run-only            run benchmark only

'''


if __name__ == '__main__':
    do_bench = True
    do_draw = True
    shortopts = 'h'
    longopts = ['draw-only', 'run-only']
    optlist, args = getopt(sys.argv[1:], shortopts, longopts)
    for key, val in optlist:
        if key == '--draw-only':
            do_bench = False
        elif key == '--run-only':
            do_draw = False
        elif key == '-h':
            show_help()
            sys.exit(0)
    if do_bench:
        bench()
    if do_draw:
        draw()
