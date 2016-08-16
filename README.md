# <img alt="logo" src="/assets/images/thyme.png" height="40"> Thyme

Spice up your day-to-day productivity with some free Thyme, courtesy of the team at [Sourcegraph](https://sourcegraph.com). Automatically track
which applications you use and for how long.

- Simple CLI to track and analyze your application usage
- Detailed charts that let you profile how you spend your time
- Stores data locally, giving you full control and privacy
- [Open-source](https://sourcegraph.com/github.com/sourcegraph/thyme/-/def/GoPackage/github.com/sourcegraph/thyme/cmd/thyme/-/main.go/TrackCmd/Execute), [well-documented](https://godoc.org/github.com/sourcegraph/thyme), and easily extensible

Thyme is a work in progress, so please report bugs! Want to see how it works? [Dive into the source here.](https://sourcegraph.com/github.com/sourcegraph/thyme/-/def/GoPackage/github.com/sourcegraph/thyme/cmd/thyme/-/main.go/TrackCmd/Execute)

Want to share what you've learned about your Thyme? Join the discussion on [Twitter](https://twitter.com/intent/tweet?url=https%3A%2F%2Fgithub.com%2Fsourcegraph%2Fthyme&original_referer=https%3A%2F%2Fgithub.com).

## Features

### Simple CLI

1. Record which applications you use every 30 seconds:
   ```
   $ watch -n 30 thyme track -o thyme.json
   ```
   An alternative to using ```watch``` is:
   ```
   $ while true; do thyme track -o thyme.json; sleep 30s; done;
   ```

2. Create charts showing application usage over time. In a new window:
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

1. [Install Go.](https://golang.org/dl/) (If you have Homebrew on macOS, you can also run `brew install go`.)
2. Run
```
$ go get -u github.com/sourcegraph/thyme/cmd/thyme
$ thyme dep
```
3. Follow the instructions printed by `thyme dep`.
4. Verify `thyme` works with
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
