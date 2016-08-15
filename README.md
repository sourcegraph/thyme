# <img alt="logo" src="/assets/images/thyme.png" height="40"> Thyme

Spice up your day-to-day productivity with Thyme. Automatically track
which applications you use and for how long.

- Simple CLI to track and analyze your application usage
- Detailed charts that let you profile how you spend your time
- Stores data locally, giving you full control and privacy
- [Open-source](https://sourcegraph.com/github.com/sourcegraph/thyme/-/def/GoPackage/github.com/sourcegraph/thyme/cmd/thyme/-/main.go/TrackCmd/Execute), [well-documented](https://godoc.org/github.com/sourcegraph/thyme), and easily extensible

Thyme is a work in progress, so please report bugs! Want to see how it works? [Dive into the source here.](https://sourcegraph.com/github.com/sourcegraph/thyme/-/def/GoPackage/github.com/sourcegraph/thyme/cmd/thyme/-/main.go/TrackCmd/Execute)

## Features

### Simple CLI

1. Record which applications you use every 30 seconds:
   ```
   $ watch -n 30 thyme track -o thyme.json
   ```

2. Create charts showing application usage over time.
   ```
   $ thyme show -i thyme.json -w stats > thyme.html
   ```

3. Open `thyme.html` in your browser of choice to see the charts
   below.

### Application usage timeline

![Application usage timeline](/assets/images/app_coarse.png)

### Detailed application window timeline

![Application usage timeline](/assets/images/app_fine.png)

### Aggregate time usage by app

![Application usage timeline](/assets/images/agg.png)


## Install

Install from source:

```
$ go get github.com/sourcegraph/thyme/cmd/thyme
$ thyme dep
```

Thyme may depend on a few OS-specific command-line tools. `thyme dep` displays instructions for installing them. Verify
`thyme` works with

```
$ thyme track
```

This should display JSON describing which applications are currently active, visible, and present on your system.

**Note:** Thyme currently supports Linux (using X-Windows) and macOS (via the AppleScript "System Events" API). Pull
requests are welcome for Windows!

## Use cases

Thyme was designed for developers who want to investigate their
application usage to make decisions that boost their day-to-day
productivity.

It can also be for other purposes such as:

- Tracking billable hours and constructing timesheets
- Studying application usage behavior in a given population

## How is Thyme different from other time trackers?

There are many time tracking products and services on the market.
Thyme differs from available offerings in the following ways:

- Thyme does not intend to be a fully featured time management product
  or service. Thyme adopts the Unix philosophy of a command-line tool
  that does one thing well and plays nicely with other command-line
  tools.

- Thyme does not require you to manually signal when you start or stop
  an activity. It automatically records which applications you use.

- Thyme is open source and free of charge.

- Thyme does not send data over the network. It stores the data it
  collects on local disk. It's up to you whether you want to share it
  or not.

## Attribution

The [Thyme logo](https://thenounproject.com/term/thyme/356887/)
<img alt="logo" src="/assets/images/thyme.png" height="40"> by
[Anthony Bossard](https://thenounproject.com/le101edaltonien/) is
licensed under
[Creative Commons 3.0](https://creativecommons.org/licenses/by/3.0/us/).
