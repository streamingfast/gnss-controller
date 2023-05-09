package handlers

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/ubx"
)

type MgaAnoLoader struct {
	anoPerSatellite map[uint8]int
	ackChannel      chan *ubx.MgaAckData0
}

func NewAnoLoader() *MgaAnoLoader {
	return &MgaAnoLoader{
		anoPerSatellite: map[uint8]int{},
		ackChannel:      make(chan *ubx.MgaAckData0),
	}
}

func (l *MgaAnoLoader) LoadAnoFile(file string, loadAll bool, now time.Time, stream io.Writer) error {
	fmt.Println("loading mga offline file:", file)
	mgaOfflineFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("opening mga offline file: %w", err)
	}

	mgaOfflineDecoder := ublox.NewDecoder(mgaOfflineFile)
	sentCount := 0
	ackCount := 0
	lastDay := 0
	go func() {
		for {
			select {
			case <-l.ackChannel:
				fmt.Print("!")
				ackCount++
			}
		}
	}()

	for {
		msg, err := mgaOfflineDecoder.Decode()
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				fmt.Println("reach mga EOF")
				break
			}
			return fmt.Errorf("decoding mga offline file: %w", err)
		}
		ano := msg.(*ubx.MgaAno)
		anoDate := time.Date(int(ano.Year)+2000, time.Month(ano.Month), int(ano.Day), 0, 0, 0, 0, time.UTC)
		if loadAll || (anoDate.Year() == now.Year() && anoDate.Month() == now.Month() && anoDate.Day() == now.Day()) { //todo: get system date
			if lastDay != now.Day() {
			}
			lastDay = anoDate.Day()
			fmt.Print(".")
			encoded, err := ubx.Encode(msg.(ubx.Message))
			if err != nil {
				return fmt.Errorf("encoding ano message: %w", err)
			}
			_, err = stream.Write(encoded)
			if err != nil {
				return fmt.Errorf("writing to stream: %w", err)
			}
			fmt.Print(".")
			sentCount++
			time.Sleep(10 * time.Millisecond)
		}
	}
	time.Sleep(2 * time.Second)
	fmt.Println("sent", sentCount, "messages and received", ackCount, "acks")

	return nil
}

func (l *MgaAnoLoader) HandleUbxMessage(message interface{}) error {
	ack := message.(*ubx.MgaAckData0)
	l.ackChannel <- ack
	return nil
}