# mackerel-plugin-log-counter

mackerel metric plugin for count lines in log

## Options

```
Usage:
  mackerel-plugin-log-counter [OPTIONS]

Application Options:
  -v, --version          Show version
      --filter=          filter string used before check pattern.
      --ignore=          ignore string used before check pattern.
  -p, --pattern=         Regexp pattern to search for.
  -k, --key-name=        Key name for pattern. if key has '|uniq' suffix, this plugin count unique matches.
      --prefix=          Metric key prefix
      --log-file=        Path to log file (default: /var/log/messages)
      --log-archive-dir= Path to log archive directory
      --per-second       calculate per-second count. default per minute count
      --verbose          display infomational logs

Help Options:
  -h, --help             Show this help message
```

## Basic usage

```
./mackerel-plugin-log-counter 
 --prefix something-log --filter something-log  \
 --key-name new --pattern 'New xxx' \
 --key-name removed --pattern 'Removed yyy'
something-log.new       5.26314        1635396141
something-log.removed   5.26314        1635396141
```

## Uniqness count usage

If log was these.
```
2025-10-22 09:41:01 error: fileA something happen
2025-10-22 09:41:01 error: fileA something happen
2025-10-22 09:41:01 error: fileB something happen
2025-10-22 09:41:01 error: fileB something happen
2025-10-22 09:41:01 error: fileC something happen
2025-10-22 09:41:01 error: fileD something happen
```

Specify key-name with '|uniq' suffix. This plugin counts unique match.

```
./mackerel-plugin-log-counter --prefix something-happen \
 --filter something-happen --per-second \
 --key-name countall --pattern 'error: file[A-Z]' \
 --key-name errorfile|uniq --pattern 'error: file[A-Z]'
something-happen.countall    6        1635396141
something-happen.errorfile   4        1635396141
```

## Install

Please download release page or `mkr plugin install kazeburo/mackerel-plugin-log-counter`.