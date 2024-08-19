package parsegp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

/* GPX Dont work - experimental */

// BitStream represents a stream of bits that can be read bit by bit.
type BitStream struct {
	reader io.Reader
	buffer byte
	count  uint8
}

// NewBitStream creates a new BitStream reader.
func newBitStream(reader io.Reader) *BitStream {
	return &BitStream{
		reader: reader,
		buffer: 0,
		count:  0,
	}
}

// ReadBit reads a single bit from the stream.
func (bs *BitStream) readBit() (uint8, error) {
	if bs.count == 0 {
		if err := binary.Read(bs.reader, binary.BigEndian, &bs.buffer); err != nil {
			return 0, err
		}
		bs.count = 8
	}

	bs.count--
	return (bs.buffer >> bs.count) & 1, nil
}

// ReadBits reads n bits from the stream and returns them as an integer.
func (bs *BitStream) readBits(n uint8) (uint32, error) {
	var value uint32
	for i := uint8(0); i < n; i++ {
		bit, err := bs.readBit()
		if err != nil {
			return 0, err
		}
		value = (value << 1) | uint32(bit)
	}
	return value, nil
}

// DecompressBCFZ decompresses a BCFZ file.
func decompressBCFZ(file *os.File) ([]byte, error) {
	// Read the expected decompressed length (32-bit little-endian integer)
	var expectedLength uint32
	if err := binary.Read(file, binary.LittleEndian, &expectedLength); err != nil {
		return nil, err
	}
	// fmt.Println("expectedLength", expectedLength)

	bitStream := newBitStream(file)
	var decompressed bytes.Buffer

	for {
		// Read the next bit to determine the chunk type
		chunkType, err := bitStream.readBit()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if chunkType == 0 {
			// Uncompressed chunk
			length, err := bitStream.readBits(2) // Read 2 bits for length
			if err != nil {
				return nil, err
			}
			buf := make([]byte, length)
			if _, err := io.ReadFull(file, buf); err != nil {
				return nil, err
			}
			decompressed.Write(buf)
		} else {
			// Compressed chunk
			wordSize, err := bitStream.readBits(4) // Read 4 bits for word size
			if err != nil {
				return nil, err
			}
			offset, err := bitStream.readBits(uint8(wordSize)) // Read wordSize bits for offset
			if err != nil {
				return nil, err
			}
			length, err := bitStream.readBits(uint8(wordSize)) // Read wordSize bits for length
			if err != nil {
				return nil, err
			}
			start := decompressed.Len() - int(offset)
			if start < 0 || start+int(length) > decompressed.Len() {
				fmt.Println("start:", start, "length:", length)
				return nil, fmt.Errorf("invalid offset/length in compressed chunk")
			}
			buf := decompressed.Bytes()[start : start+int(length)]
			fmt.Println("buf:", string(buf))
			decompressed.Write(buf)
		}

		// Stop if we reach the expected decompressed length or the file ends
		if uint32(decompressed.Len()) >= expectedLength {
			break
		}
	}

	// Return the decompressed data
	return decompressed.Bytes(), nil
}

// loadGPXFile loads a GPX file.
func (gp *GPFile) loadGPXFile() error {
	fi, err := os.Stat(gp.FullPath)
	if err != nil || fi.Size() == 0 {
		return fmt.Errorf("file is empty or does not exist")
	}

	// Read the BCFZ data
	f, err := os.Open(gp.FullPath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(f)

	data, err := decompressBCFZ(f)
	if err != nil {
		return err
	}
	fmt.Println("Decompressed data:", data)

	/*

		data := make([]byte, fi.Size())
		zeros := 0
		for {
			data = data[:cap(data)]
			fmt.Println(string(data))
			n, err := f.Read(data)
			if err != nil {
				if err == io.EOF {
					fmt.Println("EOF")
					break
				}
				fmt.Println("err:", err)
				return err
			}
			data = data[:n]
			for _, b := range data {
				if b == 0 {
					zeros++
				}
			}
		}
	*/
	return nil
}
