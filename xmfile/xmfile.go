package xmfile

import (
	"fmt"
	"io"
)

// Module is a parsed XM file contents.
// This is a raw module format that is not optimized for anything.
type Module struct {
	Name string

	TrackerName string

	// Major and minor version numbers.
	// Version[0] is a major version.
	// Version[1] is a minor version.
	Version [2]byte

	SongLength      int
	RestartPosition int

	NumChannels    int
	NumPatterns    int
	NumInstruments int

	// 0 - Amiga
	// 1 - Linear
	Flags uint16

	DefaultTempo int
	DefaultBPM   int

	PatternOrder []uint8

	Patterns []Pattern

	// This pattern is generated only once and then used for every empty pattern in Patterns.
	EmptyPattern Pattern

	Instruments []Instrument
}

type Pattern struct {
	// Whether this pattern is an auto-generated empty pattern.
	IsEmpty bool

	Rows []PatternRow
}

type PatternRow struct {
	Notes []PatternNote
}

type PatternNote struct {
	Note            uint8
	Instrument      uint8
	Volume          uint8
	EffectType      uint8
	EffectParameter uint8
}

type Instrument struct {
	Name string

	KeymapAssignments []byte
	EnvelopeVolume    []EnvelopePoint
	EnvelopePanning   []EnvelopePoint

	VolumeSustainPoint    uint8
	VolumeLoopStartPoint  uint8
	VolumeLoopEndPoint    uint8
	PanningSustainPoint   uint8
	PanningLoopStartPoint uint8
	PanningLoopEndPoint   uint8

	VolumeFlags  EnvelopeFlags
	PanningFlags EnvelopeFlags

	VibratoType  uint8
	VibratoSweep uint8
	VibratoDepth uint8
	VibratoRate  uint8

	VolumeFadeout int

	Samples []InstrumentSample
}

type EnvelopePoint struct {
	X uint16
	Y uint16
}

type InstrumentSample struct {
	Name         string
	Length       int
	LoopStart    int
	LoopLength   int
	Volume       int
	Finetune     int
	TypeFlags    uint8
	Panning      uint8
	RelativeNote int
	Format       SampleFormat
	Data         []uint8
}

type SampleLoopType int

const (
	SampleLoopNone SampleLoopType = iota
	SampleLoopForward
	SampleLoopPingPong
	SampleLoopUnknown
)

func (s *InstrumentSample) LoopType() SampleLoopType {
	bits := s.TypeFlags & 0b11
	return SampleLoopType(bits)
}

func (s *InstrumentSample) Is16bits() bool {
	return (s.TypeFlags & (1 << 4)) != 0
}

type EnvelopeFlags uint8

func (f EnvelopeFlags) IsOn() bool {
	return f&(1<<0) != 0
}

func (f EnvelopeFlags) SustainEnabled() bool {
	return f&(1<<1) != 0
}

func (f EnvelopeFlags) LoopEnabled() bool {
	return f&(1<<2) != 0
}

type SampleFormat int

const (
	SampleFormatDeltaPacked SampleFormat = iota
	SampleFormatADPCM
)

// Parse reads XM file data and decodes it into a module.
//
// A non-nil error is usually a *ParseError object.
func Parse(r io.Reader) (*Module, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}
	p := &parser{data: data}
	return p.Parse()
}
