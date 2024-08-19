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
	Numerator   int `json:"numerator"`
	Denominator struct {
		Value    int `json:"value"`
		Division struct {
			Enters int `json:"enters"`
			Times  int `json:"times"`
		} `json:"division"`
	} `json:"denominator"`
}

// Channel represents a channel in a Guitar Pro file.
type Channel struct {
	Program int   `json:"program"`
	Color   Color `json:"color"`
}

// Track represents a track in a Guitar Pro file.
type Track struct {
	Number   int       `json:"number"`
	Name     string    `json:"name"`
	Strings  []string  `json:"strings"`
	Measures []Measure `json:"measures"`
}

// Measure represents a measure in a Guitar Pro file.
type Measure struct {
	Header MeasureHeader `json:"header"`
	Start  int           `json:"start"`
	Beats  []Beat        `json:"beats"`
}

// Beat represents a beat in a Guitar Pro measure.
type Beat struct {
	Start int `json:"start"`
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
