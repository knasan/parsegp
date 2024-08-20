// Pacakge parsegp provides functionality for parsing Guitar Pro files (.gp3, .gp4, gp5, gpx).
// Description: This file contains the implementation of the GP5Parser struct and its methods.

// Replica of C++ version by PhilPotter
// https://github.com/PhilPotter/gp_parser/blob/master/gp_parser.cpp

package parsegp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	QUARTER_TIME                     = 960
	QUARTER                          = 4
	TGVELOCITIES_MIN_VELOCITY        = 15
	TGVELOCITIES_VELOCITY_INCREMENT  = 16
	TGEFFECTBEND_MAX_POSITION_LENGTH = 12
	TGEFFECTBEND_SEMITONE_LENGTH     = 1
	GP_BEND_SEMITONE                 = 25
	GP_BEND_POSITION                 = 60
)

var VERSIONS = []string{
	"FICHIER GUITAR PRO v5.00",
	"FICHIER GUITAR PRO v5.10",
}

type Parser struct {
	FileBuffer         []byte
	BufferPosition     int
	VersionIndex       int
	Channels           []Channel
	MeasureHeaders     []MeasureHeader
	Measures           []Measure
	Tracks             []Track
	BitStream          *BitStream
	Lyric              Lyric
	TempoValue         int
	GlobalKeySignature int
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
	Comments           []string
	TabFile            *TabFile
	Version            string
}

// NewParser creates a new Parser instance from the given file path.
// It reads and parses the file's header information, including version, title, subtitle, artist, album,
// lyrics author, music author, copyright, tab, and instructions.
//
// Parameters:
// filePath (string): The path to the Guitar Pro tab file.
//
// Returns:
// *Parser: A pointer to a new Parser instance, or nil if an error occurs.
// error: An error if the file path is empty, if the file cannot be opened, or if an error occurs during parsing.
func NewParser(filePath string) (*Parser, error) {
	if filePath == "" {
		return nil, errors.New("null file path passed to constructor")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBuffer, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		FileBuffer:     fileBuffer,
		BufferPosition: 0,
	}

	// Read and check version
	version, err := parser.ReadVersion()
	if err != nil {
		return nil, err
	}
	parser.Version = version

	if !parser.IsSupportedVersion(parser.Version) {
		return nil, errors.New("unsupported version")
	}

	// Parse out major and minor version numbers
	majorAndMinorExp := regexp.MustCompile(`(\d+)\.(\d+)`)
	matches := majorAndMinorExp.FindStringSubmatch(version)
	if len(matches) < 3 {
		return nil, errors.New("failed to parse version numbers")
	}

	parser.Major, err = strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	parser.Minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	parser.Title, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Subtitle, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Artist, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Album, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.LyricsAuthor, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.MusicAuthor, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Copyright, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Tab, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}

	parser.Instructions, err = parser.readStringByteSizeOfInteger()
	if err != nil {
		return nil, err
	}
	// Assume additional initialization occurs here

	return parser, nil
}

// ReadInt reads the next 4 bytes from the file buffer as a 32-bit integer (int32).
// It returns the integer value and an error if there are not enough bytes to read.
// The function also updates the buffer position by 4 after reading.
func (p *Parser) readInt() (int32, error) {
	// Check if there are enough bytes to read
	if p.BufferPosition+4 > len(p.FileBuffer) {
		return 0, errors.New("not enough bytes to read int")
	}

	// Reading the 4 bytes and converting them to a 32-bit integer
	returnVal := int32(
		((uint32(p.FileBuffer[p.BufferPosition+3]) & 0xFF) << 24) |
			((uint32(p.FileBuffer[p.BufferPosition+2]) & 0xFF) << 16) |
			((uint32(p.FileBuffer[p.BufferPosition+1]) & 0xFF) << 8) |
			(uint32(p.FileBuffer[p.BufferPosition]) & 0xFF))

	// Increase buffer position by 4 after reading
	p.BufferPosition += 4

	return returnVal, nil
}

// readByte reads a single byte from the buffer and increments the position by one.
//
// The function checks if there are still bytes available in the buffer. If not, it returns an error.
// If there are bytes available, it reads the byte at the current position, increments the buffer position,
// and returns the byte value as an byte along with a nil error.
func (p *Parser) readByte() (byte, error) {
	// Check if there are still bytes in the buffer
	if p.BufferPosition >= len(p.FileBuffer) {
		return 0, errors.New("not enough bytes to read")
	}

	// Read the byte and increment the buffer position
	byteValue := p.FileBuffer[p.BufferPosition]
	p.BufferPosition++

	return byteValue, nil
}

// readUnsignedByte reads a single unsigned byte from the buffer and increments the position by one.
//
// The function checks if there are still bytes available in the buffer. If not, it returns an error.
// If there are bytes available, it reads the byte at the current position, increments the buffer position,
// and returns the byte value as an uint8 along with a nil error.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the byte.
//
// Returns:
//
//	uint8 - The byte value read from the buffer.
//	error - An error if there are not enough bytes to read.
func (p *Parser) readUnsignedByte() (byte, error) {
	// Check if there are still bytes in the buffer
	if p.BufferPosition >= len(p.FileBuffer) {
		return 0, errors.New("not enough bytes to read")
	}

	// Read the byte and increment the buffer position
	byteValue := p.FileBuffer[p.BufferPosition]
	p.BufferPosition++

	return byteValue, nil
}

// readString reads a string of specified size from the file buffer.
// It returns the string value and an error if there are not enough bytes to read.
// The function also updates the buffer position by the size of the string after reading.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the string.
//	size - The number of bytes to read as a string.
//
// Returns:
//
//	string - The string value read from the buffer.
//	error - An error if there are not enough bytes to read.
func (p *Parser) readString(size int) (string, error) {
	// Check if there are enough bytes in the buffer

	if p.BufferPosition+size > len(p.FileBuffer) {
		return "", errors.New("not enough bytes to read string")
	}

	// Read the bytes and create the string
	byteSlice := p.FileBuffer[p.BufferPosition : p.BufferPosition+size]
	p.BufferPosition += size

	return string(byteSlice), nil
}

// readByteString reads a string of specified size or length from the file buffer.
// If size is less than or equal to 0, it reads the specified length of bytes.
// It returns the string value and an error if there are not enough bytes to read.
// The function also updates the buffer position by the size of the string after reading.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the string.
//	size - The number of bytes to read as a string if it is greater than 0.
//	len - The length of the string to read if size is less than or equal to 0.
//
// Returns:
//
//	string - The string value read from the buffer.
//	error - An error if there are not enough bytes to read.
func (p *Parser) readByteString(size, len int) (string, error) {
	// Determine the number of bytes to read
	bytesToRead := size
	if bytesToRead <= 0 {
		bytesToRead = len
	}

	// Check if there are enough bytes in the buffer
	if p.BufferPosition+bytesToRead > binary.Size(p.FileBuffer) { // len(p.fileBuffer) {
		return "", errors.New("not enough bytes to read string")
	}

	// Read the bytes from the buffer
	bytes := p.FileBuffer[p.BufferPosition : p.BufferPosition+bytesToRead]
	p.BufferPosition += bytesToRead

	// Determine the actual length of the string to return
	actualLength := bytesToRead
	if len >= 0 && len <= bytesToRead {
		actualLength = len
	}

	// Return the string
	return string(bytes[:actualLength]), nil
}

// readStringByte reads a string from the file buffer using the specified size and the length read from the next byte.
// It calls readUnsignedByte to get the length of the string and then calls readByteString to read the string.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the string.
//	size - The number of bytes to read as a string if it is greater than 0.
//
// Returns:
//
//	string - The string value read from the buffer.
//	error - An error if there are not enough bytes to read.
func (p *Parser) readStringByte(size int) (string, error) {
	num, err := p.readUnsignedByte()
	if err != nil {
		return "", err
	}
	return p.readByteString(size, int(num))
}

// readStringByteSizeOfInteger reads a string from the file buffer using the size of the next byte as the string length.
// It first reads an unsigned byte to determine the length of the string. Then, it calls readStringByte to read the string.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the string.
//
// Returns:
//
//	string - The string value read from the buffer.
//	error - An error if there are not enough bytes to read.
func (p *Parser) readStringByteSizeOfInteger() (string, error) {
	num, err := p.readUnsignedByte()
	if err != nil {
		return "", err
	}
	return p.readStringByte(int(num) - 1)
}

// readStringInteger reads a string from the file buffer using the size read from the next 4 bytes as the string length.
// It first reads an integer from the buffer using the readInt method. Then, it calls readString to read the string.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the string.
//
// Returns:
//
//	string - The string value read from the buffer.
//	error - An error if there are not enough bytes to read or if the readInt method returns an error.
func (p *Parser) readStringInteger() (string, error) {
	num, err := p.readInt()
	if err != nil {
		return "", err
	}
	return p.readString(int(num))
}

// skip skips the specified number of bytes in the file buffer.
// It updates the buffer position by the given number of bytes.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to skip bytes.
//	n - The number of bytes to skip.
//
// Returns:
//
//	None
func (p *Parser) skip(n int) {
	p.BufferPosition += n
}

// ReadVersion reads a version string from the file buffer using the specified size.
// It calls the readStringByte method of the Parser struct to read the version string.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the version string.
//
// Returns:
//
//	string - The version string read from the buffer.
//	error - An error if there are not enough bytes to read or if the readStringByte method returns an error.
func (p *Parser) ReadVersion() (string, error) {
	return p.readStringByte(30)
}

// readLyrics reads and parses lyrics data from the file buffer.
// It reads the starting position of the lyrics, the lyrics text, and skips over 4 unknown fields.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the lyrics data.
//
// Returns:
//
//	Lyric - A struct containing the parsed lyrics data.
//		From: The starting position of the lyrics.
//		Lyric: The text of the lyrics.
//		If an error occurs during reading, the returned Lyric will have default values.
func (p *Parser) readLyrics() Lyric {
	lyric := Lyric{}
	num, err := p.readInt()
	if err != nil {
		return lyric
	}
	lyric.From = int(num)

	lyric.Lyric, err = p.readStringInteger()
	if err != nil {
		return lyric
	}

	for i := 0; i < 4; i++ {
		p.readInt()
		p.readStringInteger()
	}

	return lyric
}

// readPageSetup reads and processes page setup data from the file buffer.
// It skips over the page setup data based on the version index.
// If the version index is greater than 0, it skips over 49 bytes.
// Otherwise, it skips over 30 bytes.
// After skipping the initial bytes, it iterates over 11 times, skipping over 4 bytes and reading a string of length 0.
func (p *Parser) readPageSetup() {
	if p.VersionIndex > 0 {
		p.skip(49)
	} else {
		p.skip(30)
	}
	for i := 0; i < 11; i++ {
		p.skip(4)
		p.readStringByte(0)
	}
}

// readKeySignature reads the key signature from the file buffer and returns it.
// It reads a single byte from the buffer and adjusts the value based on the version index.
// If an error occurs during reading, it returns the original keySignature value minus 1.
//
// Parameters:
// p - A pointer to the Parser struct from which to read the key signature.
//
// Return:
// byte - The key signature value read from the file buffer.
// If an error occurs during reading, it returns the original keySignature value minus 1.
func (p *Parser) readKeySignature() byte {
	keySignature, err := p.readByte()
	if err != nil {
		return keySignature - 1
	}

	// if keySignature < 0 { // This fix addresses the staticcheck warning "SA4003: no value of type byte is less than 0".
	keySignature = 7 + keySignature // Fix: Add '+' instead of '-'
	// }

	return keySignature
}

// readChannels reads and parses the channel data from the file buffer.
// It iterates over 64 channels, reading the program, volume, balance, chorus, reverb, pan, phaser, tremolo,
// and bank information for each channel. It also sets the IsPercussionChannel flag for the 10th channel.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the channel data.
//
// Returns:
//
//	[]Channel - A slice of Channel structs containing the parsed channel data.
//		Each Channel struct contains the program, volume, balance, chorus, reverb, pan, phaser, tremolo,
//		bank, IsPercussionChannel flag, and name.
func (p *Parser) readChannels() []Channel {
	var channels []Channel
	for i := 0; i < 64; i++ {
		channel := Channel{}
		var err error
		if channel.Program, err = p.readInt(); err != nil {
			fmt.Println("Error reading channel program:", err)
		}

		if channel.Volume, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel volume:", err)
		}

		if channel.Balance, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel balance:", err)
		}

		if channel.Chorus, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel chorus:", err)
		}

		if channel.Reverb, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel reverb:", err)
		}

		if channel.Pan, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel pan:", err)
		}

		if channel.Phaser, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel phaser:", err)
		}

		if channel.Tremolo, err = p.readByte(); err != nil {
			fmt.Println("Error reading channel tremolo:", err)
		}

		if i == 9 {
			channel.Bank = "default percussion bank"
			channel.IsPercussionChannel = true
		} else {
			channel.Bank = "default bank"
		}

		if channel.Program < 0 {
			channel.Program = 0
		}

		channels = append(channels, channel)
		p.skip(2)
	}

	return channels
}

// readColor reads the next three bytes from the file buffer as unsigned integers representing the red, green, and blue
// components of a color. It then skips over the next byte.
//
// The function reads the red, green, and blue components using the readUnsignedByte method of the Parser struct.
// If any of these read operations fail, it prints an error message to the console.
// After reading the color components, it skips over the next byte using the skip method of the Parser struct.
//
// Parameters:
//
//	p - A pointer to the Parser struct from which to read the color components.
//
// Returns:
//
//	Color - A struct representing the color.
//		R: The red component of the color (0-255).
//		G: The green component of the color (0-255).
//		B: The blue component of the color (0-255).
func (p *Parser) readColor() Color {
	c := Color{}
	var err error
	if c.R, err = p.readUnsignedByte(); err != nil {
		fmt.Println("Error reading color red:", err)
	}

	if c.G, err = p.readUnsignedByte(); err != nil {
		fmt.Println("Error reading color green:", err)
	}

	if c.B, err = p.readUnsignedByte(); err != nil {
		fmt.Println("Error reading color blue:", err)
	}

	p.skip(1)

	return c
}

// readChannel reads and processes channel data from the file buffer.
// It maps GM channel numbers to the corresponding channel in the Parser's channels slice.
// If the GM channel number is valid, it creates temporary ChannelParam objects for GM channel 1 and GM channel 2.
// It then copies the corresponding channel from the Parser's channels slice to a temporary variable.
// If the copied channel's ID is 0, it assigns a new ID, sets the name to "TODO", appends the temporary ChannelParam objects,
// and adds the channel to the Parser's channels slice.
// Finally, it sets the track's ChannelID to the copied channel's ID.
func (p *Parser) readChannel(track *Track) {
	gmChannel1, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading gm channel 1:", err)
		return
	}
	gmChannel1 = gmChannel1 - 1

	gmChannel2, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading gm channel 2:", err)
		return
	}
	gmChannel2 = gmChannel2 - 1

	if gmChannel1 >= 0 && gmChannel1 < int32(len(p.Channels)) {
		// Temporäre ChannelParam Objekte
		gmChannel1Param := ChannelParam{
			Key:   "gm channel 1",
			Value: fmt.Sprintf("%d", gmChannel1),
		}

		gmChannel2Value := gmChannel2
		if gmChannel1 == 9 {
			gmChannel2Value = gmChannel1
		}

		gmChannel2Param := ChannelParam{
			Key:   "gm channel 2",
			Value: fmt.Sprintf("%d", gmChannel2Value),
		}

		// Kopiere Channel in eine temporäre Variable
		channel := p.Channels[gmChannel1]

		// TODO: channel auxiliary, JS code below:
		/*for i := 0; i < len(p.channels); i++ {
		    channelAux := p.channels[i]
		    for n := 0; n < len(channelAux.); n++ {

		    }
		}*/

		if channel.ID == 0 {
			channel.ID = int32(len(p.Channels) + 1)
			channel.Name = "TODO"
			channel.Parameters = append(channel.Parameters, gmChannel1Param, gmChannel2Param)
			p.Channels = append(p.Channels, channel)
		}
		track.ChannelID = channel.ID
	}
}

// readMeasure reads and processes a measure from the MIDI file.
// It iterates through the beats in the measure, calculates the start time for each beat,
// and updates the measure's beats, clef, and key signature.
func (p *Parser) readMeasure(measure *Measure, track *Track, tempo *Tempo, keySignature int8) {
	for voice := 0; voice < 2; voice++ {
		start := float64(measure.Start)

		beats, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading beats:", err)
			return
		}
		for k := 0; k < int(beats); k++ {
			start += p.readBeat(int32(start), measure, track, tempo, voice)
		}
	}

	var emptyBeats []*Beat
	for i := 0; i < len(measure.Beats); i++ {
		beatPtr := &measure.Beats[i]
		empty := true
		for v := 0; v < len(beatPtr.Voices); v++ {
			if len(beatPtr.Voices[v].Notes) != 0 {
				empty = false
				break
			}
		}
		if empty {
			emptyBeats = append(emptyBeats, beatPtr)
		}
	}

	for _, beatPtr := range emptyBeats {
		for i := 0; i < len(measure.Beats); i++ {
			if beatPtr == &measure.Beats[i] {
				measure.Beats = append(measure.Beats[:i], measure.Beats[i+1:]...)
				break
			}
		}
	}

	measure.Clef.Name = p.getClef(track)
	measure.KeySignature = keySignature
}

// getLength calculates the length of a measure based on its time signature and tempo.
// It uses the time signature's numerator and the denominator to convert it into a Duration,
// then multiplies it by the tempo to get the total length in seconds.
// The result is rounded to the nearest integer and returned as an int32.
func (p *Parser) getLength(header *MeasureHeader) int32 {
	return int32(math.Round(float64(header.TimeSignature.Numerator) *
		p.getTime(p.denominatorToDuration(header.TimeSignature.Denominator))))
}

// getBeat retrieves a beat from the given measure at the specified start time.
// If the beat already exists in the measure, it returns a pointer to the existing beat.
// If the beat does not exist, it creates a new beat, adds it to the measure, and returns a pointer to the new beat.
//
// Parameters:
// - measure: A pointer to the measure where the beat is located.
// - start: The start time of the beat.
//
// Return:
// - A pointer to the beat in the measure.
func (p *Parser) getBeat(measure *Measure, start int32) *Beat {
	for i := range measure.Beats {
		if measure.Beats[i].Start == start {
			return &measure.Beats[i]
		}
	}

	beat := Beat{}
	beat.Voices = make([]Voice, 2)
	beat.Start = start
	measure.Beats = append(measure.Beats, beat)

	return &measure.Beats[len(measure.Beats)-1]
}

// readMixChange reads the mix change data from the MIDI file.
// It updates the tempo value and skips the unnecessary data.
func (p *Parser) readMixChange(tempo *Tempo) {
	// Read the instrument byte and skip the next 16 bytes
	p.readByte()
	p.skip(16)

	// Read the volume, pan, chorus, reverb, phaser, and tremolo bytes
	volume, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change volume:", err)
		return
	}

	pan, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change pan:", err)
		return
	}

	chorus, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change chorus:", err)
		return
	}

	reverb, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change reverb:", err)
		return
	}

	phaser, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change phaser:", err)
		return
	}

	tremolo, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading mix change tremolo:", err)
		return
	}

	// Read the tempo name as a byte-sized string and skip it
	p.readStringByteSizeOfInteger()

	// Read the tempo value as an integer and update the tempo
	tempoValue, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading mix change tempo value:", err)
		return
	}

	if volume >= 0 {
		p.readByte()
	}
	if pan >= 0 {
		p.readByte()
	}
	if chorus >= 0 {
		p.readByte()
	}
	if reverb >= 0 {
		p.readByte()
	}
	if phaser >= 0 {
		p.readByte()
	}
	if tremolo >= 0 {
		p.readByte()
	}

	if tempoValue >= 0 {
		tempo.Value = tempoValue
		p.skip(1)
		if p.VersionIndex > 0 {
			p.skip(1)
		}
	}

	// Read the last byte and skip the next byte
	p.readByte()
	p.skip(1)

	// If the version index is greater than 0, read and skip the last two byte-sized strings
	if p.VersionIndex > 0 {
		p.readStringByteSizeOfInteger()
		p.readStringByteSizeOfInteger()
	}
}

// readBeatEffects reads the beat effects flags and applies them to the given beat and noteEffect.
func (p *Parser) readBeatEffects(beat *Beat, noteEffect *NoteEffect) {
	flags1, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading beat effects flags1:", err)
		return
	}

	flags2, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading beat effects flags2:", err)
		return
	}

	noteEffect.FadeIn = (flags1 & 0x10) != 0
	noteEffect.Vibrato = (flags1 & 0x02) != 0

	if (flags1 & 0x20) != 0 {
		effect, err := p.readUnsignedByte()
		if err != nil {
			fmt.Println("Error reading beat effects effect:", err)
			return
		}
		noteEffect.Tapping = effect == 1
		noteEffect.Slapping = effect == 2
		noteEffect.Pop = effect == 3
	}

	if (flags2 & 0x04) != 0 {
		p.readTremoloBar(noteEffect)
	}

	if (flags1 & 0x40) != 0 {
		strokeUp, err := p.readByte()
		if err != nil {
			fmt.Println("Error reading beat effects strokeUp:", err)
			return
		}

		strokeDown, err := p.readByte()
		if err != nil {
			fmt.Println("Error reading beat effects strokeDown:", err)
			return
		}

		// TODO: Implement the correct logic here
		if strokeUp > 0 {
			beat.Stroke.Direction = "stroke_up"
			beat.Stroke.Value = "stroke_down"
		} else if strokeDown > 0 {
			beat.Stroke.Direction = "stroke_down"
			beat.Stroke.Value = "stroke_down"
		}
	}

	if (flags2 & 0x02) != 0 {
		p.readByte()
	}
}

// readTremoloBar reads tremolo bar data from the MIDI file and populates the NoteEffect struct.
// It skips 5 bytes, reads the number of tremolo bar points, and then iterates over each point to populate
// the TremoloBar and TremoloPoint structs.
func (p *Parser) readTremoloBar(effect *NoteEffect) {
	p.skip(5)

	tremoloBar := TremoloBar{}
	numPoints, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading tremolo bar numPoints:", err)
		return
	}

	for i := 0; i < int(numPoints); i++ {
		position, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading tremolo bar position:", err)
			return
		}

		value, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading tremolo bar value:", err)
			return
		}

		p.readByte()

		point := TremoloPoint{}
		point.Position = int32(math.Round(
			float64(position) * 1.0 / 1.0)) // TODO: 'max position length' und 'bend position'

		point.Value = int32(math.Round(
			float64(value) / (1.0 * 0x2f))) // TODO: 'GP_BEND_SEMITONE'

		tremoloBar.Points = append(tremoloBar.Points, point)
	}

	if len(tremoloBar.Points) > 0 {
		effect.TremoloBar = tremoloBar
	}
}

// readText reads text from the MIDI file and populates the Text struct in the Beat struct.
//
// Parameters:
// beat: A pointer to the Beat struct where the text will be stored.
//
// Returns:
// This function does not return any value.
func (p *Parser) readText(beat *Beat) {
	text, err := p.readStringByteSizeOfInteger()
	if err != nil {
		fmt.Println("Error reading text:", err)
		return
	}
	beat.Text.Value = text
}

// readChord reads chord information from the MIDI file and populates the Chord struct in the Beat struct.
//
// Parameters:
// strings: An array of GuitarString representing the strings on which the chord is played.
// beat: A pointer to the Beat struct where the chord information will be stored.
//
// This function does not return any value.
func (p *Parser) readChord(strings []GuitarString, beat *Beat) {
	chord := Chord{
		Strings: &strings,
	}

	p.skip(17)

	chordName, err := p.readStringByte(21)
	if err != nil {
		fmt.Println("Error reading chord name:", err)
		return
	}

	chord.Name = chordName

	p.skip(4)

	chord.Frets = make([]int32, 6)
	chordFrets, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading chord fret 0:", err)
		return
	}
	chord.Frets = append(chord.Frets, chordFrets)

	for i := 0; i < 7; i++ {
		fret, err := p.readInt()
		if err != nil {
			fmt.Printf("Error reading chord fret %d: %v\n", i+1, err)
			return
		}
		if i < len(strings) {
			chord.Frets[i] = fret
		}
	}

	p.skip(32)

	if len(strings) > 0 {
		beat.Chord = chord
	}
}

// getTime calculates the time duration of a given musical duration.
//
// Parameters:
// duration: A Duration struct representing the musical duration.
//
// Return:
// The function returns a float64 representing the time duration of the given musical duration.
// The time duration is calculated based on the duration's value, dotted and double-dotted status,
// and the division of the duration.
func (p *Parser) getTime(duration Duration) float64 {
	time := QUARTER_TIME * 4.0 / float64(duration.Value)
	if duration.Dotted {
		time += time / 2
	} else if duration.DoubleDotted {
		time += (time / 4) * 3
	}

	return time * float64(duration.Division.Times) / float64(duration.Division.Enters)
}

// readDuration reads the duration of a note based on the given flags.
// It calculates the duration value, checks for dotted notes, and handles
// different types of note divisions.
//
// Parameters:
// - flags: A uint8 value representing the flags of the duration.
//
// Returns:
//   - float64: The calculated duration value. If an error occurs during reading,
//     it returns 0.0.
func (p *Parser) readDuration(flags uint8) float64 {
	duration := Duration{}
	b, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading duration flags:", err)
		return 0.0
	}
	duration.Value = math.Pow(2, float64(b+4)) / 4
	duration.Dotted = (flags & 0x01) != 0

	if (flags & 0x20) != 0 {
		divisionType, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading division type:", err)
			return 0.0
		}
		switch divisionType {
		case 3:
			duration.Division.Enters = 3
			duration.Division.Times = 2
		case 5:
			duration.Division.Enters = 5
			duration.Division.Times = 5
		case 6:
			duration.Division.Enters = 6
			duration.Division.Times = 4
		case 7:
			duration.Division.Enters = 7
			duration.Division.Times = 4
		case 9:
			duration.Division.Enters = 9
			duration.Division.Times = 8
		case 10:
			duration.Division.Enters = 10
			duration.Division.Times = 8
		case 11:
			duration.Division.Enters = 11
			duration.Division.Times = 8
		case 12:
			duration.Division.Enters = 12
			duration.Division.Times = 8
		case 13:
			duration.Division.Enters = 13
			duration.Division.Times = 8
		}
	}
	if duration.Division.Enters == 0 {
		duration.Division.Enters = 1
		duration.Division.Times = 1
	}

	return p.getTime(duration)
}

// readBeat reads a beat from the MIDI file and returns its duration.
// It also populates the beat with notes, chords, text, beat effects, and mix changes.
//
// Parameters:
// - start: The start position of the beat in the measure.
// - measure: The measure to which the beat belongs.
// - track: The track from which the beat is being read.
// - tempo: The tempo information for the track.
// - voiceIndex: The index of the voice within the beat.
//
// Return:
//   - float64: The duration of the beat. If the beat is empty (i.e., no notes are present),
//     the function returns 0.0.
func (p *Parser) readBeat(start int32, measure *Measure, track *Track, tempo *Tempo, voiceIndex int) float64 {
	flags, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading beat flags:", err)
		return 0.0
	}

	beat := p.getBeat(measure, start)
	voice := &beat.Voices[voiceIndex]

	if (flags & 0x40) != 0 {
		beatType, err := p.readUnsignedByte()
		if err != nil {
			fmt.Println("Error reading beat type:", err)
			return 0.0
		}

		voice.Empty = (beatType & 0x02) == 0
	}

	duration := p.readDuration(flags)
	effect := NoteEffect{}

	if (flags & 0x02) != 0 {
		p.readChord(track.GuitarStrings, beat)
	}
	if (flags & 0x04) != 0 {
		p.readText(beat)
	}
	if (flags & 0x08) != 0 {
		p.readBeatEffects(beat, &effect)
	}
	if (flags & 0x10) != 0 {
		p.readMixChange(tempo)
	}

	stringFlags, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading string flags:", err)
		return 0.0
	}

	for i := 6; i >= 0; i-- {
		if stringFlags&(1<<i) != 0 && (6-i) < len(track.GuitarStrings) {
			string := track.GuitarStrings[6-i]
			note := p.readNote(string, track, effect)
			voice.Notes = append(voice.Notes, note)
		}
	}
	voice.Duration.Value = duration

	p.skip(1)

	read, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading note flags:", err)
		return 0.0
	}

	if (read & 0x02) != 0 {
		p.skip(1)
	}

	if len(voice.Notes) != 0 {
		return duration
	}
	return 0.0
}

// readNote reads and processes a note from the MIDI file.
// It extracts relevant information such as string number, note effects, velocity, fret value, and tied note status.
// If the note is tied, it retrieves the fret value from the previous note on the same string.
// If the note is not tied, it sets the fret value to the value read from the MIDI file.
// If the fret value is invalid, it sets it to 0.
// The function also skips over any additional data related to the note.
func (p *Parser) readNote(guitarString GuitarString, track *Track, effect NoteEffect) Note {
	flags, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading note flags:", err)
		return Note{}
	}

	note := Note{
		String: guitarString.Number,
		Effect: effect,
	}
	note.Effect.AccentuatedNote = (flags & 0x40) != 0
	note.Effect.HeavyAccentuatedNote = (flags & 0x02) != 0
	note.Effect.GhostNote = (flags & 0x04) != 0

	if (flags & 0x20) != 0 {
		noteType, err := p.readUnsignedByte()
		if err != nil {
			fmt.Println("Error reading note type:", err)
			return Note{}
		}

		note.TiedNote = noteType == 0x02
		note.Effect.DeadNote = noteType == 0x03
	}

	if (flags & 0x10) != 0 {
		velocity, err := p.readByte()
		if err != nil {
			fmt.Println("Error reading velocity:", err)
			return Note{}
		}

		note.Velocity = TGVELOCITIES_MIN_VELOCITY +
			(TGVELOCITIES_VELOCITY_INCREMENT * int(velocity)) -
			TGVELOCITIES_VELOCITY_INCREMENT // TODO: Ensure constants are defined
	}

	if (flags & 0x20) != 0 {
		fret, err := p.readByte()
		if err != nil {
			fmt.Println("Error reading fret:", err)
			return Note{}
		}
		value := fret
		if note.TiedNote {
			value = p.getTiedNoteValue(guitarString.Number, track)
		}
		if value >= 0 && value < 100 {
			note.Value = value
		} else {
			note.Value = 0
		}
	}
	if (flags & 0x80) != 0 {
		p.skip(2)
	}
	if (flags & 0x01) != 0 {
		p.skip(8)
	}
	p.skip(1)
	if (flags & 0x08) != 0 {
		p.readNoteEffects(&note.Effect)
	}

	return note
}

// getTiedNoteValue retrieves the fret value of the last tied note on the same string in the previous measures.
// It iterates through the measures and beats in reverse order, looking for the last tied note on the specified guitar string.
// If a tied note is found, its fret value is returned. If no tied note is found, 0 is returned.
func (p *Parser) getTiedNoteValue(guitarString int32, track *Track) uint8 {
	measureCount := len(track.Measures)
	if measureCount > 0 {
		for m := measureCount - 1; m >= 0; m-- {
			measure := track.Measures[m]
			for b := len(measure.Beats) - 1; b >= 0; b-- {
				beat := measure.Beats[b]
				for v := 0; v < len(beat.Voices); v++ {
					voice := beat.Voices[v]
					if !voice.Empty {
						for n := len(voice.Notes) - 1; n >= 0; n-- {
							note := voice.Notes[n]
							if note.String == guitarString && note.TiedNote {
								return note.Value
							}
						}
					}
				}
			}
		}
	}

	return 0
}

// readNoteEffects reads and processes note effects from the MIDI file.
// It extracts relevant information such as bend, grace notes, tremolo picking, slide, artificial harmonics, trills, hammer, let ring, vibrato, palm mute, and staccato.
//
// Parameters:
// - noteEffect: A pointer to a NoteEffect struct where the extracted note effects will be stored.
//
// The function reads the note effect flags from the MIDI file and calls the appropriate functions to process each type of note effect.
// It also sets the corresponding fields in the NoteEffect struct based on the extracted information.
func (p *Parser) readNoteEffects(noteEffect *NoteEffect) {
	flags1, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading note effect flags 1:", err)
		return
	}

	flags2, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading note effect flags 2:", err)
		return
	}

	if (flags1 & 0x01) != 0 {
		p.readBend(noteEffect)
	}
	if (flags1 & 0x10) != 0 {
		p.readGrace(noteEffect)
	}
	if (flags2 & 0x04) != 0 {
		p.readTremoloPicking(noteEffect)
	}
	if (flags2 & 0x08) != 0 {
		noteEffect.Slide = true
		p.readByte() // Assume it's a placeholder for additional data related to slide
	}
	if (flags2 & 0x10) != 0 {
		p.readArtificialHarmonic(noteEffect)
	}
	if (flags2 & 0x20) != 0 {
		p.readTrill(noteEffect)
	}
	noteEffect.Hammer = (flags1 & 0x02) != 0
	noteEffect.LetRing = (flags1 & 0x08) != 0
	noteEffect.Vibrato = (flags2 & 0x40) != 0
	noteEffect.PalmMute = (flags2 & 0x02) != 0
	noteEffect.Staccato = (flags2 & 0x01) != 0
}

// readBend reads and processes bend information from the MIDI file.
// It extracts relevant information such as bend points and their positions and values.
// The function skips over unnecessary data and handles potential errors.
// If the bend points are successfully extracted, they are stored in a Bend struct and assigned to the NoteEffect.
func (p *Parser) readBend(effect *NoteEffect) {
	p.skip(5) // Skip over unnecessary data

	bend := Bend{}

	numPoints, err := p.readInt()
	if err != nil {
		fmt.Println("Error reading bend points count:", err)
		return
	}

	for i := 0; i < int(numPoints); i++ {
		bendPosition, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading bend point position:", err)
			return
		}

		bendValue, err := p.readInt()
		if err != nil {
			fmt.Println("Error reading bend point value:", err)
			return
		}
		p.readByte() // Skip over padding or unused field

		point := BendPoint{
			Position: int32(math.Round(float64(bendPosition) *
				TGEFFECTBEND_MAX_POSITION_LENGTH /
				float64(GP_BEND_POSITION))),
			Value: int32(math.Round(float64(bendValue) *
				TGEFFECTBEND_SEMITONE_LENGTH /
				float64(GP_BEND_SEMITONE))),
		}
		bend.Points = append(bend.Points, point)
	}

	if len(bend.Points) > 0 {
		effect.Bend = bend
	}
}

// readGrace reads and processes grace note information from the MIDI file.
// It extracts relevant information such as fret, dynamic, transition, duration, and flags.
// The function handles potential errors during the reading process.
// If the grace note information is successfully extracted, it is stored in a Grace struct and assigned to the NoteEffect.
func (p *Parser) readGrace(effect *NoteEffect) {
	fret, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading grace fret:", err)
		return
	}

	dynamic, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading grace dynamic:", err)
		return
	}

	transition, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading grace transition:", err)
		return
	}

	duration, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading grace duration:", err)
		return
	}

	flags, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading grace flags:", err)
		return
	}

	grace := Grace{
		Fret: fret,
		Dynamic: TGVELOCITIES_MIN_VELOCITY +
			(TGVELOCITIES_VELOCITY_INCREMENT * int(dynamic)) -
			TGVELOCITIES_VELOCITY_INCREMENT,
		Duration: duration,
		Dead:     (flags & 0x01) != 0,
		OnBeat:   (flags & 0x02) != 0,
	}

	switch transition {
	case 0:
		grace.Transition = "none"
	case 1:
		grace.Transition = "slide"
	case 2:
		grace.Transition = "bend"
	case 3:
		grace.Transition = "hammer"
	}

	effect.Grace = grace
}

// readTremoloPicking reads and processes tremolo picking information from the MIDI file.
// It extracts relevant information such as the duration of the tremolo picking.
//
// Parameters:
// - effect: A pointer to a NoteEffect struct where the extracted tremolo picking information will be stored.
//
// The function reads the tremolo picking value from the MIDI file and sets the corresponding duration value in the TremoloPicking struct.
// If an error occurs during the reading process, an error message is printed and the function returns without any further action.
func (p *Parser) readTremoloPicking(effect *NoteEffect) {
	value, err := p.readUnsignedByte()
	if err != nil {
		fmt.Println("Error reading tremolo picking value:", err)
		return
	}

	tp := TremoloPicking{}

	switch value {
	case 1:
		tp.Duration.Value = "eighth"
	case 2:
		tp.Duration.Value = "sixteenth"
	case 3:
		tp.Duration.Value = "thirty_second"
	default:
		return // No valid value, so no action is taken
	}

	effect.TremoloPicking = tp
}

// readArtificialHarmonic reads and processes artificial harmonic information from the MIDI file.
// It extracts relevant information such as the type of artificial harmonic.
//
// Parameters:
// - effect: A pointer to a NoteEffect struct where the extracted artificial harmonic information will be stored.
//
// The function reads the artificial harmonic type value from the MIDI file and sets the corresponding type value in the Harmonic struct.
// If an error occurs during the reading process, an error message is printed and the function returns without any further action.
func (p *Parser) readArtificialHarmonic(effect *NoteEffect) {
	typeVal, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading artificial harmonic type:", err)
		return
	}

	harmonic := Harmonic{}

	switch typeVal {
	case 1:
		harmonic.Type = "natural"
	case 2:
		p.skip(3)
		harmonic.Type = "artificial"
	case 3:
		p.skip(1)
		harmonic.Type = "tapped"
	case 4:
		harmonic.Type = "pinch"
	case 5:
		harmonic.Type = "semi"
	default:
		return // Bei unbekanntem Typ keine Änderung
	}

	effect.Harmonic = harmonic
}

// readTrill reads and processes trill information from the MIDI file.
//
// Parameters:
// - effect: A pointer to a NoteEffect struct where the extracted trill information will be stored.
//
// The function reads the trill fret and period values from the MIDI file.
// If an error occurs during the reading process, an error message is printed and the function returns without any further action.
// The trill information is then stored in a Trill struct and assigned to the NoteEffect.
// If the period value is not recognized, no changes are made to the trill information.
func (p *Parser) readTrill(effect *NoteEffect) {
	fret, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading trill fret:", err)
		return
	}

	period, err := p.readByte()
	if err != nil {
		fmt.Println("Error reading trill period:", err)
		return
	}

	trill := Trill{
		Fret: fret,
	}

	switch period {
	case 1:
		trill.Duration.Value = "sixteenth"
	case 2:
		trill.Duration.Value = "thirty_second"
	case 3:
		trill.Duration.Value = "sixty_fourth"
	default:
		return // Bei unbekanntem period keine Änderung
	}

	effect.Trill = trill
}

// isPercussionChannel checks if the given MIDI channel ID corresponds to a percussion channel.
//
// Parameters:
// - channelId: The MIDI channel ID to be checked.
//
// Return:
//   - bool: A boolean value indicating whether the given channel ID corresponds to a percussion channel.
//     Returns true if the channel ID is found in the list of percussion channels, otherwise returns false.
func (p *Parser) isPercussionChannel(channelId int32) bool {
	for _, channel := range p.Channels {
		if channel.ID == channelId {
			return channel.IsPercussionChannel
		}
	}
	return false
}

// getClef retrieves the clef type for the given track based on the guitar strings' values.
// If the track is not a percussion channel, it iterates through the guitar strings and checks their values.
// If any guitar string has a value less than or equal to 34, it returns "CLEF_BASS".
// Otherwise, it returns "CLEF_TREBLE".
func (p *Parser) getClef(track *Track) string {
	if !p.isPercussionChannel(track.ChannelID) {
		for _, gstr := range track.GuitarStrings {
			if gstr.Value <= 34 {
				return "CLEF_BASS"
			}
		}
	}

	return "CLEF_TREBLE"
}

// getTabFile returns a copy of the TabFile struct from the Parser.
// This function is used to retrieve the parsed MIDI data and prepare it for further processing.
//
// Parameters:
//
//	p: A pointer to the Parser struct containing the parsed MIDI data.
//
// Return:
//
//	TabFile: A copy of the TabFile struct containing the parsed MIDI data.
//	  The returned TabFile struct contains various fields such as title, artist, album, tempo value,
//	  global key signature, channels, measures, track count, measure headers, and tracks.
func (p *Parser) getTabFile() TabFile {
	return TabFile{
		Major:              p.TabFile.Major,
		Minor:              p.TabFile.Minor,
		Title:              p.TabFile.Title,
		Subtitle:           p.TabFile.Subtitle,
		Artist:             p.TabFile.Artist,
		Album:              p.TabFile.Album,
		LyricsAuthor:       p.TabFile.LyricsAuthor,
		MusicAuthor:        p.TabFile.MusicAuthor,
		Copyright:          p.TabFile.Copyright,
		Tab:                p.TabFile.Tab,
		Instructions:       p.TabFile.Instructions,
		Comments:           p.TabFile.Comments,
		Lyric:              p.TabFile.Lyric,
		TempoValue:         p.TabFile.TempoValue,
		GlobalKeySignature: p.TabFile.GlobalKeySignature,
		Channels:           p.TabFile.Channels,
		Measures:           p.TabFile.Measures,
		TrackCount:         p.TabFile.TrackCount,
		MeasureHeaders:     p.TabFile.MeasureHeaders,
		Tracks:             p.TabFile.Tracks,
	}
}

// numOfDigits calculates the number of digits in a given integer.
//
// Parameters:
// - num: The integer for which the number of digits needs to be calculated.
//
// Return:
// - int: The number of digits in the given integer.
func numOfDigits(num int32) int {
	digits := 0
	for order := 1; int(num)/int(order) != 0; order *= 10 {
		digits++
	}
	return digits
}

// denominatorToDuration converts a given denominator to a Duration struct.
//
// Parameters:
// - denominator: A Denominator struct containing the value and division of the denominator.
//
// Return:
//   - Duration: A Duration struct containing the value and division of the denominator.
//     The returned Duration struct has the same value and division as the input Denominator.
func (p *Parser) denominatorToDuration(denominator Denominator) Duration {
	return Duration{
		Value:    denominator.Value,
		Division: denominator.Division,
	}
}

// IsSupportedVersion checks if the given MIDI file version is supported by the parser.
//
// Parameters:
// - version: A string representing the MIDI file version to be checked.
//
// Return:
//   - bool: A boolean value indicating whether the given version is supported.
//     Returns true if the version is found in the list of supported versions, otherwise returns false.
func (p *Parser) IsSupportedVersion(version string) bool {
	for i, v := range VERSIONS {
		if strings.Compare(version, v) == 0 {
			p.VersionIndex = i
			return true
		}
	}
	return false
}
