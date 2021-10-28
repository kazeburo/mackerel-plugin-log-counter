# mackerel-plugin-log-counter

mackerel metric plugin for count logs

## Usage

```
Usage:
  mackerel-plugin-log-counter [OPTIONS]

Application Options:
  -v, --version     Show version
      --filter=     filter string used before check pattern.
  -p, --pattern=    Regexp pattern to search for.
  -k, --key-name=   Key name for pattern
      --prefix=     Metric key prefix
      --log-file=   Path to log file (default: /var/log/messages)
      --per-second  calcurate per-seconds count. default per minute count

Help Options:
  -h, --help        Show this help message
```

```
./mackerel-plugin-log-counter --prefix something-log \
 --filter something-happen --per-second \
 --key-name new --pattern 'New xxx' \
 --key-name removed --pattern 'Removed yyy'
something-log.new  0.087719        1635396141
something-log.removed      0.087719        1635396141
```

## Install

Please download release page or `mkr plugin install kazeburo/mackerel-plugin-log-counter`.