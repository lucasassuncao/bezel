package main

import (
	"math"
	"strings"
	"time"

	"github.com/lucasassuncao/bezel/animation"
	"github.com/lucasassuncao/bezel/component/chart"
	"github.com/lucasassuncao/bezel/component/progress"
	"github.com/lucasassuncao/bezel/theme"
)

const (
	chartWidth = 40
	grow       = 1200 * time.Millisecond
	sparkKeep  = chartWidth
)

type (
	growMsg  struct{} // g: the bars grow and the line is drawn again
	pushMsg  struct{} // p: a new value on the sparkline
	gaugeMsg struct{} // n on the data tab: the gauge moves
	warmMsg  struct{} // w: the heatmap warms and cools
)

var (
	barValues = []float64{72, 45, 88}
	barNames  = []string{"cpu", "disk", "net"}
	readouts  = []float64{18, 64, 92, 37}
	// heats are the two states the heatmap turns between.
	heats = [][][]float64{
		{{1, 3, 5, 7, 9, 7, 5, 3, 1, 0}, {3, 5, 7, 9, 10, 9, 7, 5, 3, 1}, {1, 3, 5, 7, 9, 7, 5, 3, 1, 0}},
		{{9, 7, 5, 3, 1, 0, 2, 4, 6, 8}, {7, 5, 3, 1, 0, 1, 3, 5, 7, 9}, {9, 7, 5, 3, 1, 0, 2, 4, 6, 8}},
	}
)

// waves is a smooth run of n values between 10 and 90, for the drawn line.
func waves(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = 50 + 40*math.Sin(float64(i)/4)
	}
	return out
}

// datas is the seventh tab: charts whose numbers move instead of jumping.
type datas struct {
	bars    animation.Series
	area    animation.Series
	spark   animation.Series
	fed     int
	gauge   animation.Spring
	reading int
	heat    animation.Heat
	warm    int
}

func newDatas() datas {
	spark := make([]float64, sparkKeep)
	for i := range spark {
		spark[i] = 50 + 30*math.Sin(float64(i)/3)
	}
	return datas{
		bars:  animation.NewGrow(barValues, 0),
		area:  animation.NewDraw(waves(chartWidth), 0),
		spark: animation.NewGrow(spark, 0),
		gauge: animation.NewSpring(readouts[0]).WithDamping(0.5),
		heat:  animation.NewHeat(heats[0], 900*time.Millisecond),
	}
}

// replay grows the bars and draws the line again.
func (d datas) replay() datas {
	d.bars = animation.NewGrow(barValues, grow).WithEase(animation.Out(animation.Back))
	d.area = animation.NewDraw(waves(chartWidth), grow)
	return d
}

// feed pushes the next value onto the sparkline.
func (d datas) feed() datas {
	d.fed++
	d.spark = d.spark.WithDuration(300 * time.Millisecond)
	d.spark, _ = d.spark.Push(50+45*math.Sin(float64(d.fed)*1.7), sparkKeep)
	return d
}

func (d datas) read() datas {
	d.reading++
	d.gauge, _ = d.gauge.To(readouts[d.reading%len(readouts)])
	return d
}

func (d datas) warmUp() datas {
	d.warm++
	d.heat, _ = d.heat.Set(heats[d.warm%len(heats)])
	return d
}

func (d datas) advance(at time.Time) datas {
	d.bars, _ = d.bars.Advance(at)
	d.area, _ = d.area.Advance(at)
	d.spark, _ = d.spark.Advance(at)
	d.gauge, _ = d.gauge.Advance(at)
	d.heat, _ = d.heat.Advance(at)
	return d
}

func (d datas) moving() bool {
	return d.bars.Active() || d.area.Active() || d.spark.Active() || d.gauge.Active() || d.heat.Active()
}

func (d datas) rows(th theme.Resolved, label func(kind, variant string) string) string {
	var bars []chart.Bar
	for i, v := range d.bars.Values() {
		bars = append(bars, chart.Bar{Label: barNames[i], Value: v, Text: " "})
	}
	var rows []string
	add := func(kind, variant, block string) {
		for i, line := range strings.Split(block, "\n") {
			if i > 0 {
				kind, variant = "", ""
			}
			rows = append(rows, label(kind, variant)+line)
		}
	}
	add("bars", "g grow", chart.Bars(bars, 100, chartWidth, th))
	add("spark", "p pushes", chart.Spark(d.spark.Values(), 100, chartWidth, th))
	add("area", "g draws", chart.Area(d.area.Values(), 100, chartWidth, 3, th))
	add("gauge", "n springs", progress.Gauge(d.gauge.Pos, 100, chartWidth, th))
	add("heat", "w warms", chart.Heat(d.heat.Values(), 10, th))
	return strings.Join(rows, "\n")
}
