package thyme

import (
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"
)

type AggTime struct {
	Charts []*BarChart
}

type BarChart struct {
	ID     string
	YLabel string
	XLabel string
	Title  string
	Series map[string]int
}

type Bar struct {
	Label string
	Count int
}

func NewBarChart(id, x, y, title string) *BarChart {
	return &BarChart{ID: id, XLabel: x, YLabel: y, Title: title, Series: make(map[string]int)}
}

func (c *BarChart) Plus(label string, n int) {
	c.Series[label] += n
}

type sortBars struct {
	bars []Bar
}

func (s sortBars) Len() int           { return len(s.bars) }
func (s sortBars) Less(a, b int) bool { return s.bars[a].Count > s.bars[b].Count }
func (s sortBars) Swap(a, b int)      { s.bars[a], s.bars[b] = s.bars[b], s.bars[a] }

func (c *BarChart) OrderedBars() []Bar {
	var bars []Bar
	for l, c := range c.Series {
		bars = append(bars, Bar{Label: l, Count: c})
	}
	s := sortBars{bars}
	sort.Sort(s)
	return s.bars
}

func NewAggTime(stream *Stream, labelFunc func(*Window) string) *AggTime {
	active := NewBarChart("Active", "App", "Samples", "Time (multiplied by number of windows) application was active")
	visible := NewBarChart("Visible", "App", "Samples", "Time (multiplied by number of windows) application was visible")
	all := NewBarChart("All", "App", "Samples", "Time (multiplied by number of windows) application was open")
	for _, snap := range stream.Snapshots {
		windows := make(map[int64]*Window)
		for _, win := range snap.Windows {
			windows[win.ID] = win
		}

		active.Plus(labelFunc(windows[snap.Active]), 1)
		for _, v := range snap.Visible {
			visible.Plus(labelFunc(windows[v]), 1)
		}
		for _, win := range snap.Windows {
			all.Plus(labelFunc(win), 1)
		}
	}
	return &AggTime{Charts: []*BarChart{active, visible, all}}
}

type Timeline struct {
	Start time.Time
	End   time.Time
	Rows  map[string][]*Range
}

type Range struct {
	Label string
	Start time.Time
	End   time.Time
}

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
			win := windows[snap.Active]
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
		}

		for _, prevRange := range lastVisible {
			prevRange.End = snap.Time
		}
		nextVisible := make(map[string]*Range)
		for _, v := range snap.Visible {
			win := windows[v]
			winLabel := labelFunc(win)
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

func timeToJS(t time.Time) string {
	return fmt.Sprintf(`new Date(%d, %d, %d, %d, %d, %d)`, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

type statsPage struct {
	Fine   *Timeline
	Coarse *Timeline
	Agg    *AggTime
}

var statsTmpl = template.Must(template.New("").Funcs(map[string]interface{}{
	"timeToJS": timeToJS,
}).Parse(`<html>
  <head>
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
				"{{.Label}}",
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.Visible}}
			[
				"Visible",
				"{{.Label}}",
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.All}}
			[
				"All",
				"{{.Label}}",
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
		['{{.Label}}', {{.Count}}],
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
        bars: 'horizontal'
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
				"{{.Label}}",
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.Visible}}
			[
				"Visible",
				"{{.Label}}",
				{{timeToJS .Start}},
				{{timeToJS .End}},
			],
		{{end}}
		{{range .Rows.All}}
			[
				"All",
				"{{.Label}}",
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
		This is a coarse-grained chart of all the applications you over the course of the day. Every bar represents an application.
	</div>
    <div id="timeline_coarse" style="min-height: 500px;"></div>
	<hr>

	<div class="description">
		This is a fine-grained chart of all the applications you over the course of the day. Every bar represents a distinct window.
	</div>
    <div id="timeline_fine" style="min-height: 500px;"></div>
	<hr>

	{{range $chart := .Agg.Charts}}
	<div id="bar_chart_{{$chart.ID}}"></div>
	<hr>
	{{end}}

  </body>
</html>`))

func appID(w *Window) string {
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

func List(stream *Stream) {
	fmt.Printf("%s", stream.Print())
}
