package main

import (
	"flag"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/coreos/go-systemd/v22/journal"
	"github.com/coreos/go-systemd/v22/login1"
)

func diff(a, b uint64) uint64 {
	if a < b {
		return math.MaxUint64 - b + a
	}

	return a - b
}

func main() {
	threshold := flag.Uint64("threshold", 50, "network traffic threshold in KB/s")
	wait := flag.Uint64("wait", 30, "time in seconds to wait before uninhibit")
	flag.Parse()

	journal.Print(journal.PriDebug, "threshold: %d KB/s, wait: %d s", *threshold, *wait)

	prevStats, err := netstat()
	if err != nil {
		journal.Print(journal.PriCrit, "Failed to get network stats: %s", err.Error())
		return
	}

	counter := uint64(0)

	login, err := login1.New()
	if err != nil {
		journal.Print(journal.PriCrit, "Failed to initialize login1: %s", err.Error())
		return
	}
	defer login.Close()

	var inhibit *os.File
	defer func() {
		if inhibit != nil {
			inhibit.Close()
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	daemon.SdNotify(false, daemon.SdNotifyReady)
	defer daemon.SdNotify(false, daemon.SdNotifyStopping)

	daemon.SdNotify(false, "STATUS=Uninhibited")

	for {
		select {
		case <-ticker.C:
			stats, err := netstat()
			if err != nil {
				journal.Print(journal.PriCrit, "Failed to get network stats: %s", err.Error())
				return
			}

			rxBytes := uint64(0)
			txBytes := uint64(0)
			for name, stat := range stats {
				if prevStat, ok := prevStats[name]; ok {
					rxBytes += diff(stat.RxBytes, prevStat.RxBytes)
					txBytes += diff(stat.TxBytes, prevStat.TxBytes)
				}
			}

			if *threshold*1024 < rxBytes+txBytes {
				counter = *wait

				if inhibit == nil {
					inhibit, err = login.Inhibit("sleep:shutdown", "shinobu-go", "network activity", "block")
					if err != nil {
						journal.Print(journal.PriCrit, "Failed to inhibit: %s", err.Error())
						return
					}

					daemon.SdNotify(false, "STATUS=Inhibited")
				}
			} else if 0 < counter {
				counter--
			} else if inhibit != nil {
				err = inhibit.Close()
				if err != nil {
					journal.Print(journal.PriCrit, "Failed to uninhibit: %s", err.Error())
					return
				}

				inhibit = nil
				daemon.SdNotify(false, "STATUS=Uninhibited")
			}

			prevStats = stats

		case <-sig:
			return
		}
	}
}
