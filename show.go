package thyme

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"
)

const maxNumberOfBars = 30

// Stats renders an HTML page with charts using stream as its data
// source. Currently, it renders the following charts:
// 1. A timeline of applications active, visible, and open
// 2. A timeline of windows active, visible, and open
// 3. A barchart of applications most often active, visible, and open
func Stats(stream *Stream) error {
	tlFine := NewTimeline(stream, func(w *Window) string { return w.Name })
	tlCoarse := NewTimeline(stream, appID)
	agg := NewAggTime(stream, appID)

	if err := statsTmpl.Execute(os.Stdout, &statsPage{
		Fine:   tlFine,
		Coarse: tlCoarse,
		Agg:    agg,
	}); err != nil {
		return err
	}
	return nil
}

// AggTime is the list of bar charts that convey aggregate application time usage.
type AggTime struct {
	Charts []*BarChart
}

// NewAggTime returns a new AggTime created from a Stream.
func NewAggTime(stream *Stream, labelFunc func(*Window) string) *AggTime {
	n := strconv.Itoa(maxNumberOfBars)
	active := NewBarChart("Active", "App", "Samples", "Top "+n+" active applications by time (multiplied by window count)")
	visible := NewBarChart("Visible", "App", "Samples", "Top "+n+" visible applications by time (multiplied by window count)")
	all := NewBarChart("All", "App", "Samples", "Top "+n+" open applications by time (multiplied by window count)")
	for _, snap := range stream.Snapshots {
		windows := make(map[int64]*Window)
		for _, win := range snap.Windows {
			windows[win.ID] = win
		}

		if win := windows[snap.Active]; win != nil {
			active.Plus(labelFunc(windows[snap.Active]), 1)
		}
		for _, v := range snap.Visible {
			visible.Plus(labelFunc(windows[v]), 1)
		}
		for _, win := range snap.Windows {
			all.Plus(labelFunc(win), 1)
		}
	}
	return &AggTime{Charts: []*BarChart{active, visible, all}}
}

// BarChart is a representation of a bar chart.
type BarChart struct {
	ID     string
	YLabel string
	XLabel string
	Title  string
	Series map[string]int
}

// Bar represents a single bar in a bar chart.
type Bar struct {
	Label string
	Count int
}

// NewBarChart returns a new BarChart with the specified ID, x- and
// y-axis label, and title.
func NewBarChart(id, x, y, title string) *BarChart {
	return &BarChart{ID: id, XLabel: x, YLabel: y, Title: title, Series: make(map[string]int)}
}

// Plus adds n to the count associated with the label.
func (c *BarChart) Plus(label string, n int) {
	c.Series[label] += n
}

// OrderedBars returns a list of the top $maxNumberOfBars bars in the bar chart ordered by
// decreasing count.
func (c *BarChart) OrderedBars() []Bar {
	var bars []Bar
	for l, c := range c.Series {
		bars = append(bars, Bar{Label: l, Count: c})
	}
	s := sortBars{bars}
	sort.Sort(s)
	numberOfBars := maxNumberOfBars
	if numberOfBars > len(s.bars) {
		numberOfBars = len(s.bars)
	}
	return s.bars[:numberOfBars]
}

type sortBars struct {
	bars []Bar
}

func (s sortBars) Len() int           { return len(s.bars) }
func (s sortBars) Less(a, b int) bool { return s.bars[a].Count > s.bars[b].Count }
func (s sortBars) Swap(a, b int)      { s.bars[a], s.bars[b] = s.bars[b], s.bars[a] }

// Timeline represents a timeline of application usage.
// Start is the start time of the timeline.
// End is the end time of the timeline.
// Rows is a map where the keys are tags and the values are lists of
// time ranges. Each row is a distinct sub-timeline.
type Timeline struct {
	Start time.Time
	End   time.Time
	Rows  map[string][]*Range
}

// Range represents a labeled range of time.
type Range struct {
	Label string
	Start time.Time
	End   time.Time
}

// NewTimeline returns a new Timeline created from the specified
// Stream. labelFunc is used to determine the ID string to be used for
// a given Window. If you're tracking events by app, this ID should
// reflect the identity of the window's application. If you're
// tracking events by window name, the ID should be the window name.
func NewTimeline(stream *Stream, labelFunc func(*Window) string) *Timeline {
	if len(stream.Snapshots) == 0 {
		return nil
	}
	var active, visible, other []*Range
	var lastActive *Range
	var lastVisible, lastOther = make(map[string]*Range), make(map[string]*Range)
	for _, snap := range stream.Snapshots {
		windows := make(map[int64]*Window)
		for _, win := range snap.Windows {
			windows[win.ID] = win
		}

		{
			if win := windows[snap.Active]; win != nil {
				winLabel := labelFunc(win)
				if lastActive != nil && lastActive.Label == winLabel {
					lastActive.End = snap.Time
				} else {
					if lastActive != nil {
						lastActive.End = snap.Time
					}
					newRange := &Range{Label: winLabel, Start: snap.Time, End: snap.Time}
					active = append(active, newRange)
					lastActive = newRange
				}
			} else {
				lastActive = nil
			}
		}

		for _, prevRange := range lastVisible {
			prevRange.End = snap.Time
		}
		nextVisible := make(map[string]*Range)
		for _, v := range snap.Visible {
			var winLabel string
			if win := windows[v]; win != nil {
				winLabel = labelFunc(win)
			}
			if existRng, exists := lastVisible[winLabel]; !exists {
				newRange := &Range{Label: winLabel, Start: snap.Time, End: snap.Time}
				nextVisible[winLabel] = newRange
				visible = append(visible, newRange)
			} else {
				nextVisible[winLabel] = existRng
			}
		}
		lastVisible = nextVisible

		for _, prevRange := range lastOther {
			prevRange.End = snap.Time
		}
		nextOther := make(map[string]*Range)
		for _, win := range snap.Windows {
			winLabel := labelFunc(win)
			if existRng, exists := lastOther[winLabel]; !exists {
				newRange := &Range{Label: winLabel, Start: snap.Time, End: snap.Time}
				nextOther[winLabel] = newRange
				other = append(other, newRange)
			} else {
				nextOther[winLabel] = existRng
			}
		}
		lastOther = nextOther
	}
	return &Timeline{
		Start: stream.Snapshots[0].Time,
		End:   stream.Snapshots[len(stream.Snapshots)-1].Time,
		Rows:  map[string][]*Range{"Active": active, "Visible": visible, "All": other},
	}
}

// timeToJS is a template helper function that converts a time.Time to
// code that creates a JavaScript Date object.
func timeToJS(t time.Time) string {
	return fmt.Sprintf(`new Date(%d, %d, %d, %d, %d, %d)`, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// statsPage is the data rendered in statsTmpl.
type statsPage struct {
	Fine   *Timeline
	Coarse *Timeline
	Agg    *AggTime
}

// statsTmpl is the HTML template for the page rendered by the `Stats`
// function.
var statsTmpl = template.Must(template.New("").Funcs(map[string]interface{}{
	"timeToJS": timeToJS,
}).Parse(`<html>
  <head>
	<meta charset="utf-8">
	<style>
		.description {
			font-family: Roboto;
			font-size: 16px;
			padding: 16px 0;
			color: rgb(117, 117, 117);
		}
	</style>

    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
      google.charts.load('current', {'packages':['corechart', 'bar', 'timeline']});
	</script>

	{{with .Coarse}}
    <script type="text/javascript">
      google.charts.setOnLoadCallback(drawChartCoarse);
      function drawChartCoarse() {
        var container = document.getElementById('timeline_coarse');
        var chart = new google.visualization.Timeline(container);
        var dataTable = new google.visualization.DataTable();

        dataTable.addColumn({ type: 'string', id: 'Status' });
		dataTable.addColumn({ type: 'string', id: 'Name' });
        dataTable.addColumn({ type: 'date', id: 'Start' });
        dataTable.addColumn({ type: 'date', id: 'End' });
        dataTable.addRows([
		{{range .Rows.Active}}
			[
				"Active",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.Visible}}
			[
				"Visible",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.All}}
			[
				"All",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		]);

		var options = {
			timeline: { showRowLabels: true },
		};
        chart.draw(dataTable, options);
      }
    </script>
	{{end}}

	{{range $chart := .Agg.Charts}}
	<script type="text/javascript">
	google.charts.setOnLoadCallback(drawBarChart{{$chart.ID}});
	function drawBarChart{{$chart.ID}}() {
      var data = google.visualization.arrayToDataTable([
        ['Application', 'Number of samples'],
		{{range $chart.OrderedBars}}
		[{{printf "%q" .Label}}, {{.Count}}],
		{{end}}
      ]);

      var options = {
        chart: {
          title: '{{$chart.Title}}'
        },
		legend: { position: "none" },
        hAxis: {
          title: '{{$chart.YLabel}}',
          minValue: 0,
        },
        vAxis: {
          title: '{{$chart.XLabel}}'
        },
        bars: 'horizontal',
        height: 600
      };
      var material = new google.charts.Bar(document.getElementById('bar_chart_{{$chart.ID}}'));
      material.draw(data, options);
    }
	</script>
	{{end}}

	{{with .Fine}}
    <script type="text/javascript">
      google.charts.setOnLoadCallback(drawChartFine);
      function drawChartFine() {
        var container = document.getElementById('timeline_fine');
        var chart = new google.visualization.Timeline(container);
        var dataTable = new google.visualization.DataTable();

        dataTable.addColumn({ type: 'string', id: 'Status' });
		dataTable.addColumn({ type: 'string', id: 'Name' });
        dataTable.addColumn({ type: 'date', id: 'Start' });
        dataTable.addColumn({ type: 'date', id: 'End' });
        dataTable.addRows([
		{{range .Rows.Active}}
			[
				"Active",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.Visible}}
			[
				"Visible",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.All}}
			[
				"All",
				{{printf "%q" .Label}},
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		]);

		var options = {
			timeline: { showRowLabels: true },
		};
        chart.draw(dataTable, options);
      }
    </script>
	{{end}}


  </head>
  <body>

	<div class="description">
		This is a coarse-grained timeline of all the applications you use over the course of the day. Every bar represents an application.
	</div>
    <div id="timeline_coarse" style="min-height: 500px;"></div>
	<hr>

	<div class="description">
		This is a fine-grained timeline of all the applications you use over the course of the day. Every bar represents a distinct window.
	</div>
    <div id="timeline_fine" style="min-height: 500px;"></div>
	<hr>

	{{range $chart := .Agg.Charts}}
	<div id="bar_chart_{{$chart.ID}}"></div>
	<hr>
	{{end}}

  </body>
</html>`))

// appID returns a string that identifies the application of the
// window, w. It does so in best effort fashion. If the application
// can't be determined, it returns the the name of the window.
func appID(w *Window) string {
	if w == nil {
		return "(nil)"
	}
	if w.Info().App != "" {
		return w.Info().App
	}
	if w.Info().SubApp != "" {
		return fmt.Sprintf("%s :: %s", w.Info().App, w.Info().SubApp)
	}
	if w.Info().Title != "" {
		return w.Info().Title
	}
	return w.Name
}
