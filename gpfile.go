package parsegp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SupportedFormats returns a slice of strings representing the supported Guitar Pro file formats.
// The supported formats are ".gp3", ".gp4", and ".gp5".
//
// This function does not take any parameters and returns a slice of strings.
//
// Example usage:
//
//	formats := parsegp.SupportedFormats()
//	fmt.Println(formats) // Output: [".gp3" ".gp4" ".gp5"]
func SupportedFormats() []string {
	return []string{".gp3", ".gp4", ".gp5"} //, ".gpx"}
}

// NewGPFile creates a new GPFile instance for the specified file path.
// It checks if the file exists, is not empty, and has a supported Guitar Pro format (.gp3, .gp4, .gp5, .gpx).
// If the file is valid, it opens the file, sets the FullPath property, and returns the GuitarProFileInfo instance.
// If the file is not a supported format, it returns a notGPFile error.
//
// Parameters:
// p (string): The file path of the Guitar Pro file.
//
// Returns:
// gp (*GPFile): A pointer to the GPFile instance for the specified file path.
// err (error): An error if any issues occur during the file validation or opening process.
func NewGPFile(p string) (gp *GPFile, err error) {
	gp = &GPFile{}
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	ext := filepath.Ext(file.Name())
	if ext != ".gp3" && ext != ".gp4" && ext != ".gp5" && ext != ".gpx" {
		return nil, &notGPFile{msg: "no supported file format"}
	}

	gp.FullPath = p

	return gp, nil
}

// LoadHeader reads the header of the Guitar Pro file (gp3, gp4, gp5, gpx)
// It first checks if the file exists and is not empty. If the file is valid,
// it opens the file and seeks to the beginning. Then, it determines the type of
// Guitar Pro file (gp3, gp4, gp5, gpx) by reading the header.
//
// If the file is a Guitar Pro file (gp3, gp4, gp5), it calls the appropriate
// function to read and store the file's information such as title, artist,
// version, and full path.
//
// If the file is a Guitar Pro XML file (gpx), it calls the loadGPXFile function
// to parse and store the XML data.
//
// The function returns an error if any issues occur during the file reading,
// header detection, or information extraction process.
func (gp *GPFile) LoadHeader() error {
	// return gp.loadFileHeader()
	if fi, err := os.Stat(gp.FullPath); err != nil || fi.Size() == 0 {
		if err == nil {
			return errors.New("file is empty or does not exist")
		}
		return err
	}

	f, err := os.Open(gp.FullPath)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", gp.FullPath, err)
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)

	fo, err := gpSeek(f)
	if err != nil {
		return err
	}

	headerlen, head, err := headerLen(fo)
	if err != nil {
		return err
	}

	if headerlen == 0 {
		return &notGPFile{"Invalid Guitar Pro file"}
	}

	switch headerlen {
	case 4:
		err = gp.loadGPXFile()
		return err
	default:
		return gp.uncompressedGpInfo(fo, head)
	}
}

// readLongString reads a long string from the given reader.
// The long string is represented as a sequence of bytes, where the first byte indicates the length of the string.
// If the first byte is zero, the length is read as a 32-bit integer.
// The function reads the specified number of bytes from the reader and returns the string representation.
//
// Parameters:
// fo (io.Reader): The reader from which to read the long string.
//
// Returns:
// string: The string representation of the long string read from the reader.
// error: An error if any issues occur during the reading process.
func (gp *GPFile) readLongString(fo io.Reader) (string, error) {
	var size int32
	if err := binary.Read(fo, binary.LittleEndian, &size); err != nil {
		return "", err
	}

	s := make([]byte, 1)
	if _, err := io.ReadFull(fo, s); err != nil {
		return "", err
	}
	if size == 0 {
		size = int32(s[0])
	}

	stringBytes := make([]byte, size-1)
	if _, err := io.ReadFull(fo, stringBytes); err != nil {
		return "", err
	}

	return string(stringBytes), nil
}

// uncompressedGpInfo reads and stores the information from an uncompressed Guitar Pro file (gp3, gp4, gp5).
// It reads the version, seeks to the appropriate position in the file, and then reads the title, artist,
// subtitle, album, lyricist, musician, copyright, transcriber, and notice (if applicable) from the file.
//
// Parameters:
// fo (io.ReadSeeker): The reader and seeker for the Guitar Pro file.
// head ([]byte): The header of the Guitar Pro file.
//
// Returns:
// error: An error if any issues occur during the reading or seeking process.
func (gp *GPFile) uncompressedGpInfo(fo io.ReadSeeker, head []byte) error {
	version := make([]byte, 4)
	if _, err := io.ReadFull(fo, version); err != nil {
		return err
	}
	gp.Version = string(version)
	hsize := head[0]
	_, err := fo.Seek(int64(hsize)+6, io.SeekStart)
	if err != nil {
		return err
	}

	switch gp.Version {
	case "1T\x03\x04", "1.04", "1.02", "1.03":
		_, err := fo.Seek(3, io.SeekCurrent)
		if err != nil {
			return err
		}
	/*
	   case "2.21":
	       _, err := fo.Seek(1, io.SeekCurrent)
	       if err != nil {
	           return err
	       }
	*/
	default:
		_, err := fo.Seek(1, io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	gp.Title, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Artist, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Subtitle, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Album, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.LyricsAuthor, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.MusicAuthor, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Copyright, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Tab, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	// Instructions (Version 5 gp5, tested) gp4 break for this
	switch gp.Version {
	case "v5.0", "v5.1":
		gp.Instructions, err = gp.readLongString(fo)
		if err != nil {
			return err
		}
	}

	// gp.FullPath set from NewGPFile
	return nil
}

// headerLen determines the length of the header of the Guitar Pro file.
// It reads the first 4 bytes of the file to check for a compressed format (BCFZ).
// If the file is compressed, it returns 4 as the header length.
// If the file is not compressed, it resets the file pointer and reads the next 19 bytes.
// It checks for the presence of specific strings in the header to identify GP3, GP4, and GP5 formats.
// If a match is found, it seeks to the appropriate position in the file and returns the header length.
// If no match is found, it returns 0 as the header length.
//
// Parameters:
// fo (io.ReadSeeker): The reader and seeker for the Guitar Pro file.
//
// Returns:
// int: The length of the header.
// []byte: The header bytes read from the file.
// error: An error if any issues occur during the reading or seeking process.
func headerLen(fo io.ReadSeeker) (int, []byte, error) {
	head := make([]byte, 4)

	if _, err := io.ReadFull(fo, head); err != nil {
		return 0, nil, err
	}

	if bytes.HasPrefix(head, []byte("BCFZ")) {
		return 4, head, nil
	}

	// reset the file pointer
	fo, err := gpSeek(fo)
	if err != nil {
		return 0, nil, err
	}

	// check for GP3, GP4, GP5
	head = make([]byte, 19)
	if _, err := io.ReadFull(fo, head); err != nil {
		return 0, nil, err
	}

	if bytes.HasPrefix(head[1:], []byte("FICHIER GUITAR PRO")) {
		_, err := fo.Seek(1, io.SeekCurrent)
		if err != nil {
			return 0, nil, err
		}
		return 19, head, nil
	}

	if bytes.HasPrefix(head[1:], []byte("FICHIER GUITARE PRO")) {
		_, err := fo.Seek(2, io.SeekCurrent)
		if err != nil {
			return 0, nil, err
		}
		return 20, head, nil
	}

	return 0, head, nil
}

// gpSeek is a helper function to seek to the beginning of the file.
// It takes an io.ReadSeeker as input and returns a new io.ReadSeeker positioned at the beginning of the file.
// If any error occurs during the seeking process, it returns an error.
//
// Parameters:
// fo (io.ReadSeeker): The reader and seeker for the file.
//
// Returns:
// io.ReadSeeker: A new io.ReadSeeker positioned at the beginning of the file.
// error: An error if any issues occur during the seeking process.
func gpSeek(fo io.ReadSeeker) (io.ReadSeeker, error) {
	_, err := fo.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return fo, nil
}
