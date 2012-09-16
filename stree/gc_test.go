// Test script illustrating GC performance problem

package stree

import (
	"log"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestTreeGC(t *testing.T) {
	tree := NewTree()

	for i := 0; i < 500000; i++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
	}
	log.Print("Building Tree...")
	tree.BuildTree()
	log.Print("...done!")

        memstats := new(runtime.MemStats)
        for i := 0; i < 5; i++ {
                log.Print("GC()")
                runtime.GC()
                runtime.ReadMemStats(memstats)
                this_pause := time.Duration(memstats.PauseNs[(memstats.NumGC-1)%256])
                all_pause := time.Duration(memstats.PauseTotalNs)
                log.Printf("GC paused for %v -- total %v -- N %d", this_pause, all_pause, memstats.NumGC)
                log.Printf("alloc'd = %6d MB; (+footprint = %6d MB)",
                        memstats.HeapAlloc/1024/1024,
                        (memstats.Sys-memstats.HeapAlloc)/1024/1024)
        }
}
