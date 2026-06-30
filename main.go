package main

import (
	"fmt"

	"github.com/gotracker/playback/format/xm"
	"github.com/gotracker/playback/index"
	"github.com/gotracker/playback/player/feature"
)

func main() {
	data, err := xm.XM.Load("./track.xm", []feature.Feature{})
	if err != nil {
		panic(err)
	}

	fmt.Println("GetPeriodType", data.GetPeriodType())
	fmt.Println("GetGlobalVolumeType", data.GetGlobalVolumeType())
	fmt.Println("GetChannelMixingVolumeType", data.GetChannelMixingVolumeType())
	fmt.Println("GetChannelVolumeType", data.GetChannelVolumeType())
	fmt.Println("GetChannelPanningType", data.GetChannelPanningType())

	bpm := data.GetInitialBPM()
	fmt.Println("GetInitialBPM", bpm)
	tempo := data.GetInitialTempo()
	fmt.Println("GetInitialTempo", tempo)
	fmt.Println("GetMixingVolumeGeneric", data.GetMixingVolumeGeneric())
	fmt.Println("GetTickDuration", data.GetTickDuration(bpm))
	orderList := data.GetOrderList()
	fmt.Println("GetOrderList", orderList)

	channels := data.GetNumChannels()
	fmt.Println("GetNumChannels", channels)
	fmt.Println("GetChannelSettings")
	for i := range index.Channel(channels) {
		fmt.Println("   ", data.GetChannelSettings(i))
	}

	instruments := data.NumInstruments()
	fmt.Println("NumInstruments", instruments)
	// fmt.Println("GetInstrument")
	// for i := range index.Instrument(instruments) {
	// 	fmt.Println("   ", data.GetInstrument(i))
	// }

	fmt.Println("GetName", data.GetName())
	fmt.Println("GetPattern")
	for _, pi := range orderList {
		p, err := data.GetPattern(pi)
		if err != nil {
			fmt.Println("   ", pi, ":", err)
			continue
		}

		fmt.Println("   ", p)
	}

	fmt.Println("GetPeriodCalculator", data.GetPeriodCalculator())
	fmt.Println("GetInitialOrder", data.GetInitialOrder())
	// fmt.Println("GetRowRenderStringer", data.GetRowRenderStringer())
	fmt.Println("GetSystem", data.GetSystem())
	fmt.Println("GetMachineSettings", data.GetMachineSettings())
	// fmt.Println("ForEachChannel", data.ForEachChannel())
	fmt.Println("IsOPL2Enabled", data.IsOPL2Enabled())
}
