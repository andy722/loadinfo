package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
	"log"
	"math"
	"runtime"
	"time"
)

type CpuLoad struct {
	CpuBusyPercent float64

	totalPrev, totalNow     float64
	nonIdlePrev, nonIdleNow float64
}

const rateInterval = 1 * time.Second

func (c *CpuLoad) Update() {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		log.Fatalf("%v", err)
	}

	stat, err := fs.Stat()
	if err != nil {
		log.Fatalf("%v", err)
	}

	total := 0 +
		stat.CPUTotal.User +
		stat.CPUTotal.Nice +
		stat.CPUTotal.System +
		stat.CPUTotal.Idle +
		stat.CPUTotal.Iowait +
		stat.CPUTotal.IRQ +
		stat.CPUTotal.SoftIRQ +
		stat.CPUTotal.Steal +
		stat.CPUTotal.Guest +
		stat.CPUTotal.GuestNice

	nonIdle := 0 +
		stat.CPUTotal.User +
		stat.CPUTotal.Nice +
		stat.CPUTotal.System +
		stat.CPUTotal.Iowait +
		stat.CPUTotal.IRQ +
		stat.CPUTotal.SoftIRQ +
		stat.CPUTotal.Steal +
		stat.CPUTotal.Guest +
		stat.CPUTotal.GuestNice

	// Like `irate` in Prometheus
	c.nonIdlePrev, c.nonIdleNow = c.nonIdleNow, nonIdle
	c.totalPrev, c.totalNow = c.totalNow, total
	nonIdleRate := (c.nonIdleNow - c.nonIdlePrev) / rateInterval.Seconds()
	totalRate := (c.totalNow - c.totalPrev) / rateInterval.Seconds()

	if totalRate != 0 {
		c.CpuBusyPercent = (nonIdleRate / totalRate) * float64(100)
	}
}

func main() {
	if runtime.GOOS != "linux" {
		log.Fatalf("Only works on Linux")
	}

	cpuLoad := &CpuLoad{}

	go (func() {
		for {
			cpuLoad.Update()
			// TODO: more precise scheduler accounting for re-calculation
			time.Sleep(rateInterval)
		}
	})()

	drawStuff(cpuLoad)
}

func drawStuff(cpuLoad *CpuLoad) {
	if err := ui.Init(); err != nil {
		log.Fatalf("%v", err)
	}
	defer ui.Close()

	gauge := widgets.NewGauge()
	gauge.Title = "CPU Busy"

	pie := widgets.NewPieChart()
	pie.Title = "CPU Busy"
	pie.AngleOffset = -.5 * math.Pi
	pie.LabelFormatter = func(i int, v float64) string {
		if i == 0 {
			return fmt.Sprintf("Busy: %.1f%%", v)

		} else {
			return fmt.Sprintf("Idle: %.1f%%", v)
		}
	}

	resize := func(w, h int) {
		gauge.SetRect(0, h-10, w, h)
		pie.SetRect(0, 0, w, h-11)
	}

	resize(ui.TerminalDimensions())

	go (func() {
		for {
			cpuBusyPercent := cpuLoad.CpuBusyPercent

			gauge.Percent = int(cpuBusyPercent)
			pie.Data = []float64{cpuBusyPercent, 100 - cpuBusyPercent}
			ui.Render(pie, gauge)

			time.Sleep(1 * time.Second)
		}
	})()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break

		} else if e.Type == ui.ResizeEvent {
			payload := e.Payload.(ui.Resize)
			resize(payload.Width, payload.Height)

			ui.Clear()
			ui.Render(pie, gauge)
		}
	}
}
