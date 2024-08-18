package parsegp

type GuitarProFileInfo struct {
	FullPath string

	Version string

	Title       string
	Artist      string
	Subtitle    string
	Album       string
	LyricBy     string
	MusicBy     string
	Copyright   string
	Transcriber string
	Notice      string

	/*
		Beats       []Beat
		Voice       []Voice
		Notes       []Note
		Chords      []Chord
		Rythm       []Rythm
		Automations []Automation
		Markers     []Marker
		Info        []Info
		Measure     []Measure
		Bar         []Bar
		Track       []Track
	*/
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
