// Package parsegp provides functionality for parsing Guitar Pro files (.gp3, .gp4, .gp5, .gpx).
package parsegp

// GPFile represents a Guitar Pro file structure.
type GPFile struct {
	FullPath       string          `json:"-"`
	Version        string          `json:"version"`
	Title          string          `json:"title"`
	Subtitle       string          `json:"subtitle"`
	Artist         string          `json:"artist"`
	Album          string          `json:"album"`
	LyricsAuthor   string          `json:"lyricsAuthor"`
	MusicAuthor    string          `json:"musicAuthor"`
	Copyright      string          `json:"copyright"`
	Tab            string          `json:"tab"`
	Instructions   string          `json:"instructions"`
	Comments       []string        `json:"comments"`
	Lyric          Lyric           `json:"lyric"`
	TempoValue     int             `json:"tempoValue"`
	KeySignature   int             `json:"keySignature"`
	Channels       []Channel       `json:"channels"`
	Measures       int             `json:"measures"`
	TrackCount     int             `json:"trackCount"`
	MeasureHeaders []MeasureHeader `json:"measureHeaders"`
	Tracks         []Track         `json:"tracks"`
}

// Color represents a color in RGB format.
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// Lyric represents a lyric in a Guitar Pro file.
type Lyric struct {
	From  int    `json:"from"`
	Lyric string `json:"lyric"`
}

// MeasureHeader represents a measure header in a Guitar Pro file.
type MeasureHeader struct {
	Number        int           `json:"number"`
	Start         int           `json:"start"`
	Tempo         int           `json:"tempo"`
	RepeatOpen    bool          `json:"repeatOpen"`
	TimeSignature TimeSignature `json:"timeSignature"`
}

// TimeSignature represents a time signature in a Guitar Pro file.
type TimeSignature struct {
	Numerator   int         `json:"numerator"`
	Denominator Denominator `json:"denominator"`
	Division    Division    `json:"division"`
}

// Channel represents a channel in a Guitar Pro file.
type Channel struct {
	Program             int32          `json:"program"`
	Volume              byte           `json:"volume"`
	Pan                 byte           `json:"pan"`
	Chorus              byte           `json:"chorus"`
	Reverb              byte           `json:"reverb"`
	Phaser              byte           `json:"phaser"`
	Tremolo             byte           `json:"tremolo"`
	Balance             byte           `json:"balance"`
	Color               Color          `json:"color"`
	Bank                string         `json:"bank"`
	IsPercussionChannel bool           `json:"isPercussionChannel"`
	ID                  int32          `json:"id"`
	Name                string         `json:"name"`
	Parameters          []ChannelParam `json:"param"`
}

// Track represents a track in a Guitar Pro file.
type Track struct {
	ChannelID     int32          `json:"channelID"`
	Number        int            `json:"number"`
	Name          string         `json:"name"`
	GuitarStrings []GuitarString `json:"guitarStrings"`
	Measures      []Measure      `json:"measures"`
	Channel       Channel        `json:"channel"`
}

// Measure represents a measure in a Guitar Pro file.
type Measure struct {
	Header       MeasureHeader `json:"header"`
	Start        int           `json:"start"`
	Beats        []Beat        `json:"beats"`
	TempoValue   int           `json:"tempoValue"`
	KeySignature int8          `json:"keySignature"`
	Clef         Clef          `json:"clef"`
	TrackID      int           `json:"-"`
	ID           int           `json:"-"`
	Voices       []Voice       `json:"-"`
}

type Clef struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Octave int    `json:"octave"`
	Name   string `json:"name"`
}

// Beat represents a beat in a Guitar Pro measure.
type Beat struct {
	Start  int32   `json:"start"`
	Voices []Voice `json:"voices"`
	Stroke Stroke  `json:"stroke"`
	Pitch  Pitch   `json:"pitch"`
	Effect Effect  `json:"effect"`
	Text   Text    `json:"text"`
	Chord  Chord   `json:"chord"`
}

type Text struct {
	Start  int32      `json:"start"`
	Text   string     `json:"text"`
	Effect TextEffect `json:"effect"`
	Pitch  Pitch      `json:"pitch"`
	Value  string     `json:"value"`
}

type TextEffect struct {
	Type   string         `json:"type"`
	Params []ChannelParam `json:"params"`
	Pitch  Pitch          `json:"pitch"`
}

type Stroke struct {
	Direction string `json:"direction"`
	Value     string `json:"value"`
}

type Voice struct {
	Start    int      `json:"start"`
	Notes    []Note   `json:"note"`
	Empty    bool     `json:"empty"`
	Duration Duration `json:"duration"`
}

type Note struct {
	String   int32      `json:"string"`
	Effect   NoteEffect `json:"effect"`
	TiedNote bool       `json:"tiedNote"`
	Velocity int        `json:"velocity"`
	Value    uint8      `json:"value"`
}

type Effect struct {
	Type   string         `json:"type"`
	Params []ChannelParam `json:"params"`
}

type NoteEffect struct {
	TremoloBar           TremoloBar
	FadeIn               bool
	Vibrato              bool
	Tapping              bool
	Slapping             bool
	Pop                  bool
	AccentuatedNote      bool
	HeavyAccentuatedNote bool
	GhostNote            bool
	DeadNote             bool
	Slide                bool
	Hammer               bool
	LetRing              bool
	PalmMute             bool
	Staccato             bool
	Bend                 Bend
	Grace                Grace
	TremoloPicking       TremoloPicking
	Harmonic             Harmonic
	Trill                Trill
}

type Trill struct {
	Fret     byte
	Duration struct {
		Value string
	}
}

type Bend struct {
	Points []BendPoint `json:"points"`
}

type Harmonic struct {
	Type string `json:"type"`
}

type BendPoint struct {
	Position int32 `json:"position"`
	Value    int32 `json:"value"`
}

type TremoloBar struct {
	Points []TremoloPoint `json:"points"`
}

type TremoloPoint struct {
	Position int32 `json:"position"`
	Value    int32 `json:"value"`
}

type Pitch struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ChannelParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Tempo struct {
	BPM             int
	TimeSig         TimeSignature
	TimeSigBeats    []int    `json:"timeSigBeats"`
	TempoValue      int      `json:"tempoValue"`
	TempoValueStr   string   `json:"tempoValueStr"`
	TimeSigBeatsStr []string `json:"timeSigBeatsStr"`
	TempoStr        string   `json:"tempoStr"`
	Value           int32    `json:"value"`
}

type GuitarString struct {
	Number int32 `json:"number"`
	Value  int32 `json:"value"`
}

type Chord struct {
	Name    string          `json:"name"`
	Strings *[]GuitarString `json:"strings"`
	Frets   []int32         `json:"fret"`
}

type Duration struct {
	Value        float64
	Dotted       bool
	DoubleDotted bool
	Division     Division
}

type Division struct {
	Times  int
	Enters int
}

// Grace represents the grace note effect in the parser.
type Grace struct {
	Fret       uint8  `json:"fret"`
	Dynamic    int    `json:"dynamic"`
	Duration   uint8  `json:"duration"`
	Dead       bool   `json:"dead"`
	OnBeat     bool   `json:"onBeat"`
	Transition string `json:"transition"`
}

type TremoloPicking struct {
	Duration struct {
		Value string
	}
}

type Denominator struct {
	Value    float64
	Division Division
}

type TabFile struct {
	Major              int
	Minor              int
	Title              string
	Subtitle           string
	Artist             string
	Album              string
	LyricsAuthor       string
	MusicAuthor        string
	Copyright          string
	Tab                string
	Instructions       string
	Comments           string
	Lyric              Lyric
	TempoValue         int
	GlobalKeySignature int
	Channels           []Channel
	Measures           []Measure
	TrackCount         int
	MeasureHeaders     []MeasureHeader
	Tracks             []Track
}

// notGPFile represents an error indicating the file is not a Guitar Pro file.
// It implements the error interface, allowing it to be used with functions that expect an error return value.
type notGPFile struct {
	msg string // The error message.
}

// Error returns the error message associated with the notGPFile instance.
// This method is required to satisfy the error interface.
func (e *notGPFile) Error() string {
	return e.msg
}
