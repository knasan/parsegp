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

// SupportedFormats returns the supported formats
func SupportedFormats() []string {
	return []string{".gp3", ".gp4", ".gp5"} //, ".gpx"}
}

// NewGPFile New checks if the file is a supported format
func NewGPFile(p string) (gp *GuitarProFileInfo, err error) {
	gp = &GuitarProFileInfo{}
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
func (gp *GuitarProFileInfo) LoadHeader() error {
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

func (gp *GuitarProFileInfo) readLongString(fo io.Reader) (string, error) {
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

// uncompressedGpInfo holds the information about the Guitar Pro file
// such as the title, artist, version, and the full path to the file
// only for uncompressed Guitar Pro files (gp3, gp4, gp5)
func (gp *GuitarProFileInfo) uncompressedGpInfo(fo io.ReadSeeker, head []byte) error {
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

	gp.LyricBy, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.MusicBy, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Copyright, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	gp.Transcriber, err = gp.readLongString(fo)
	if err != nil {
		return err
	}

	// Notice (Vesion 5 gp5, testet) gp4 break for this
	switch gp.Version {
	case "v5.0", "v5.1":
		gp.Notice, err = gp.readLongString(fo)
		if err != nil {
			return err
		}
	}

	// gp.FullPath set from NewGPFile
	return nil
}

// headerLen returns the length of the header of the Guitar Pro file
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

// gpSeek is a helper function to seek to the beginning of the file
func gpSeek(fo io.ReadSeeker) (io.ReadSeeker, error) {
	_, err := fo.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return fo, nil
}
