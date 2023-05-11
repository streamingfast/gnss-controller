package logger

import (
	"fmt"
	"time"

	"github.com/daedaleanai/ublox/ubx"
)

var noTime = time.Time{}

type Logger interface {
	Log(data *Data) error
}

type Data struct {
	Ttff       int64       `json:"ttff"`
	SystemTime time.Time   `json:"systemtime"`
	Timestamp  time.Time   `json:"timestamp"`
	Fix        string      `json:"fix"`
	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
	Altitude   float64     `json:"height"`
	Heading    float64     `json:"heading"`
	Speed      float64     `json:"speed"`
	Dop        *Dop        `json:"dop"`
	Satellites *Satellites `json:"satellites"`
	Sep        float64     `json:"sep"` // Estimated Spherical (3D) Position Error in meters. Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown.
	Eph        float64     `json:"eph"` // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
	loggers    []Logger
	startTime  time.Time
}
type Dop struct {
	GDop float64 `json:"gdop"`
	HDop float64 `json:"hdop"`
	PDop float64 `json:"pdop"`
	TDop float64 `json:"tdop"`
	VDop float64 `json:"vdop"`
	XDop float64 `json:"xdop"`
	YDop float64 `json:"ydop"`
}

type Satellites struct {
	Seen int `json:"seen"`
	Used int `json:"used"`
}

func NewLoggerData(loggers ...Logger) *Data {
	return &Data{
		SystemTime: noTime,
		Timestamp:  noTime,
		loggers:    loggers,
		Dop: &Dop{
			GDop: 99.99,
			HDop: 99.99,
			PDop: 99.99,
			TDop: 99.99,
			VDop: 99.99,
			XDop: 99.99,
			YDop: 99.99,
		},
		Satellites: &Satellites{},
	}
}

func (d *Data) Clone() Data {
	clone := Data{
		Ttff:       d.Ttff,
		SystemTime: d.SystemTime,
		Timestamp:  d.Timestamp,
		Fix:        d.Fix,
		Latitude:   d.Latitude,
		Longitude:  d.Longitude,
		Altitude:   d.Altitude,
		Heading:    d.Heading,
		Speed:      d.Speed,
		Dop: &Dop{
			GDop: d.Dop.GDop,
			HDop: d.Dop.HDop,
			PDop: d.Dop.PDop,
			TDop: d.Dop.TDop,
			VDop: d.Dop.VDop,
			XDop: d.Dop.XDop,
			YDop: d.Dop.YDop,
		},
		Satellites: &Satellites{
			Seen: d.Satellites.Seen,
			Used: d.Satellites.Used,
		},
		Sep:       d.Sep,
		Eph:       d.Eph,
		loggers:   d.loggers,
		startTime: d.startTime,
	}

	return clone
}

// GNSSfix Type: 0: no fix 1: dead reckoning only 2: 2D-fix 3: 3D-fix 4: GNSS + dead reckoning combined 5: time only fix
var fix = []string{"none", "dead reckoning only", "2D", "3D", "GNSS + dead reckoning combined", "time only fix"}

func (d *Data) HandleUbxMessage(msg interface{}) error {
	d.SystemTime = time.Now()

	switch m := msg.(type) {
	case *ubx.NavPvt:
		now := time.Date(int(m.Year_y), time.Month(int(m.Month_month)), int(m.Day_d), int(m.Hour_h), int(m.Min_min), int(m.Sec_s), int(m.Nano_ns), time.UTC)
		d.Timestamp = now
		d.Fix = fix[m.FixType]
		if d.Ttff == 0 && d.Fix == "3D" && d.Dop.HDop < 5.0 {
			fmt.Println("setting ttff to: ", now)
			d.Ttff = time.Since(d.startTime).Milliseconds()
		}
		d.Latitude = float64(m.Lat_dege7) * 1e-7
		d.Longitude = float64(m.Lon_dege7) * 1e-7

		if m.FixType == 3 {
			d.Altitude = float64(m.Height_mm) / 1000 //tv.Althae
		} else {
			d.Altitude = float64(m.HMSL_mm) / 1000 //tv.Althmsl
		}
		d.Eph = float64(m.HAcc_mm) / 1000

		d.Heading = float64(m.HeadMot_dege5) * 1e-5 //tv.HeadMot
		d.Speed = float64(m.GSpeed_mm_s) / 1000     //tv.Speed
	case *ubx.NavDop:
		d.Dop.GDop = float64(m.GDOP) * 0.01
		d.Dop.HDop = float64(m.HDOP) * 0.01
		d.Dop.PDop = float64(m.PDOP) * 0.01
		d.Dop.TDop = float64(m.TDOP) * 0.01
		d.Dop.VDop = float64(m.VDOP) * 0.01
		d.Dop.XDop = float64(m.EDOP) * 0.01
		d.Dop.YDop = float64(m.NDOP) * 0.01
		for _, logger := range d.loggers {
			c := d.Clone()
			if err := logger.Log(&c); err != nil {
				return err
			}
		}
	case *ubx.NavSat:
		d.Satellites.Seen = int(m.NumSvs)
		d.Satellites.Used = 0
		for _, sv := range m.Svs {
			if sv.Flags&ubx.NavSatSvUsed != 0x00 {
				d.Satellites.Used++
			}
		}
	}

	return nil
}

func (d *Data) SetStartTime(startTime time.Time) {
	d.startTime = startTime
}
