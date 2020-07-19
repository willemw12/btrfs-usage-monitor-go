Btrfs usage monitor
===================

A simple Btrfs usage monitor.

There is also a version written in [Rust](https://github.com/willemw12/btrfs-usage-monitor).


Feature
-------

- Print a warning if Btrfs data usage drops below the free limit percentage.


Installation
------------

Run, for example:

    $ git clone https://github.com/willemw12/btrfs-usage-monitor-go
    $ cd btrfs-usage-monitor-go
    $ GOBIN=$HOME/bin/ go install


Usage
-----

    # btrfs-usage-monitor /mnt/btrfs 10
    WARNING /mnt/btrfs free: 752.58GiB (min: 681.47GiB), 9% (limit: 10%)


License
-------

GPL-3.0 or later


Link
----

[GitHub](https://github.com/willemw12/btrfs-usage-monitor-go)

