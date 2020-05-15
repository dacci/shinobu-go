package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type netStat struct {
	RxBytes      uint64
	RxPackets    uint64
	RxErrors     uint64
	RxDropped    uint64
	RxFIFO       uint64
	RxFrame      uint64
	RxCompressed uint64
	RxMulticast  uint64
	TxBytes      uint64
	TxPackets    uint64
	TxErrors     uint64
	TxDropped    uint64
	TxFIFO       uint64
	TxCollisions uint64
	TxCarrier    uint64
	TxCompressed uint64
}

func parseUint(s string) uint64 {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return u
}

func netstat() (map[string]netStat, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	r.ReadString('\n')
	r.ReadString('\n')

	stats := make(map[string]netStat)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Fields(line)

		stat := netStat{
			RxBytes:      parseUint(fields[1]),
			RxPackets:    parseUint(fields[2]),
			RxErrors:     parseUint(fields[3]),
			RxDropped:    parseUint(fields[4]),
			RxFIFO:       parseUint(fields[5]),
			RxFrame:      parseUint(fields[6]),
			RxCompressed: parseUint(fields[7]),
			RxMulticast:  parseUint(fields[8]),
			TxBytes:      parseUint(fields[9]),
			TxPackets:    parseUint(fields[10]),
			TxErrors:     parseUint(fields[11]),
			TxDropped:    parseUint(fields[12]),
			TxFIFO:       parseUint(fields[13]),
			TxCollisions: parseUint(fields[14]),
			TxCarrier:    parseUint(fields[15]),
			TxCompressed: parseUint(fields[16]),
		}

		stats[strings.TrimSuffix(fields[0], ":")] = stat
	}

	return stats, nil
}
