package thyme

import (
	"fmt"
	"os"
	"text/template"
	"time"
)

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
	stats := NewTimeline(stream, func(w *Window) string { return w.Name })
	if err := statsTmpl.Execute(os.Stdout, &stats); err != nil {
		return err
	}
	return nil
}

func timeToJS(t time.Time) string {
	return fmt.Sprintf(`new Date(%d, %d, %d, %d, %d, %d)`, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

var statsTmpl = template.Must(template.New("").Funcs(map[string]interface{}{
	"timeToJS": timeToJS,
}).Parse(`<html>
  <head>
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
      google.charts.load('current', {'packages':['timeline']});
      google.charts.setOnLoadCallback(drawChart);
      function drawChart() {
        var container = document.getElementById('timeline');
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
  </head>
  <body>
    <div id="timeline" style="height: 100%;"></div>
  </body>
</html>`))

func List(stream *Stream) {
	fmt.Printf("%s", stream.Print())
}
