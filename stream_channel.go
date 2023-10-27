package xm

type streamChannel struct {
	// Note-related data.
	inst       *instrument
	note       *patternNote
	period     float64
	sampleStep float64
	effect     effectKey
	keyOn      bool

	panning float64

	sampleOffset   float64
	volume         float64
	fadeoutVolume  float64
	computedVolume [2]float64

	// Arpeggio effect state.
	arpeggioRunning    bool
	arpeggioNoteOffset float64

	panningSlideValue   float64
	volumeSlideValue    float64
	portamentoUpValue   float64
	portamentoDownValue float64

	notePortamentoTargetPeriod float64
	notePortamentoValue        float64

	// Vibrato effect state.
	vibratoRunning      bool
	vibratoPeriodOffset float64
	vibratoDepth        float64
	vibratoStep         uint8
	vibratoSpeed        uint8

	// Ping-pong loop state.
	reverse bool

	volumeEnvelope  envelopeRunner
	panningEnvelope envelopeRunner

	// This ID is needed mostly for debugging,
	// therefore we put it to the object's tail.
	id int
}

type envelopeRunner struct {
	envelope

	value float64
	frame int
}

func (ch *streamChannel) Reset() {
	*ch = streamChannel{}
}

func (ch *streamChannel) resetEnvelopes() {
	ch.fadeoutVolume = 1
	ch.volumeEnvelope.value = 1
	ch.volumeEnvelope.frame = 0
	ch.panningEnvelope.value = 0.5
	ch.panningEnvelope.frame = 0
}

func (ch *streamChannel) assignNote(n *patternNote) {
	// Some sensible row note states:
	//
	//	[note] [instrument]
	//	no     no           keep playing the current note (if any)
	//	no     yes          "ghost instrument" (keeps the sample offset)
	//	yes    no           "ghost note" (keeps the volume)
	//	yes    yes          normal note play
	//
	// In practice, it's more complicated due to various effects
	// that may affect the logical consistency.

	ch.note = n
	ch.effect = n.effect
	noteKind := n.Kind()

	if noteKind == noteEmpty {
		return
	}

	hasNotePortamento := n.flags.Contains(noteHasNotePortamento)
	if !hasNotePortamento && noteKind == noteNormal {
		ch.inst = n.inst
		ch.volumeEnvelope.envelope = n.inst.volumeEnvelope
		ch.panningEnvelope.envelope = n.inst.panningEnvelope
	}

	ch.vibratoPeriodOffset = 0
	ch.keyOn = true
	ch.resetEnvelopes()

	if !hasNotePortamento && n.flags.Contains(noteValid) {
		ch.period = n.period
	}

	if !hasNotePortamento && noteKind != noteGhostInstrument {
		ch.sampleOffset = 0
		ch.reverse = false
	}

	if ch.inst != nil {
		if noteKind != noteGhost {
			ch.volume = ch.inst.volume
		}
		ch.panning = ch.inst.panning
	}
}
