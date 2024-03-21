package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("testdata", "wal_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a Wal instance
	w := &Wal{
		file:     tmpfile,
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}

	// Test case 1: Save empty entities
	err = w.Save(nil)
	assert.NoError(t, err)

	// Test case 2: Save non-empty entities
	entities := [][]byte{
		[]byte("entity1"),
		[]byte("entity22"),
		[]byte("entity333"),
	}
	err = w.Save(entities)
	assert.NoError(t, err)

	// Verify the contents of the file
	fileContent, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	const crcLen = 4

	// Verify the length of each record in the file
	offset := 0
	for _, entity := range entities {
		recordLength := binary.LittleEndian.Uint64(fileContent[offset : offset+8])
		assert.Equal(t, uint64(len(entity)), recordLength-crcLen)
		offset += 8 + int(recordLength)
	}

	// Verify the CRC checksum of each record in the file
	offset = 0
	for _, entity := range entities {
		recordLength := binary.LittleEndian.Uint64(fileContent[offset : offset+8])
		recordData := fileContent[offset+8 : offset+8+int(recordLength)]
		r := &record{}
		r.UnMarshal(recordData)
		recordCRC := crc32.Checksum(r.data, w.crcTable)
		assert.Equal(t, recordCRC, r.crc)
		assert.Equal(t, entity, r.data)
		offset += 8 + int(recordLength)
	}
}

func TestReadRecord(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("testdata", "wal_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a Wal instance
	w := &Wal{
		file:     tmpfile,
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}

	// Write test data to the file
	testData := []byte("test data")

	err = w.Save([][]byte{testData})
	assert.NoError(t, err)

	w.file.Seek(0, 0)
	// Call the readRecord function
	record, err := w.readRecord(w.file)
	assert.NoError(t, err)
	assert.Equal(t, testData, record)
}

func TestReadAll(t *testing.T) {
	defer func() {
		files, _ := os.ReadDir("testdata")
		for _, f := range files {
			os.Remove(fmt.Sprintf("%s/%s", "testdata", f.Name()))
		}
	}()
	// Create a Wal instance
	w, err := CreateWal("testdata", 20)
	assert.NoError(t, err)

	// Write test data to the file
	entities := [][]byte{
		[]byte("entity1"),
		[]byte("entity22"),
		[]byte("entity333"),
	}
	err = w.Save(entities)
	assert.NoError(t, err)

	offsetMatrix := []int64{0, 19, 39, 60}
	expectMatrix := [][][]byte{
		{
			[]byte("entity1"),
			[]byte("entity22"),
			[]byte("entity333"),
		},
		{
			[]byte("entity22"),
			[]byte("entity333"),
		},
		{
			[]byte("entity333"),
		},
		{},
	}
	for i, offset := range offsetMatrix {
		// Call the ReadAll function
		readEntities, err := w.ReadAll(offset)
		assert.NoError(t, err)
		// Verify the returned entities
		assert.Equal(t, expectMatrix[i], readEntities)
	}
}

func TestMaybeRoll(t *testing.T) {
	// 创建一个 Wal 实例
	w, err := CreateWal("testdata", 6656)
	assert.NoError(t, err)
	defer func() {
		for _, f := range w.segmentFiles {
			f.Close()
			os.RemoveAll(f.Name())
		}
	}()

	entities := [][]byte{}
	for i := 0; i < 1024; i++ {
		c := make([]byte, 1)
		// 写入超过512字节的内容
		entities = append(entities, c)
	}

	err = w.Save(entities)
	assert.NoError(t, err)

	// 验证文件是否被截断并打开了新的文件
	assert.Equal(t, int64(13*1024), w.totalOffset)
	entity, err := os.ReadDir("testdata")
	assert.NoError(t, err)
	assert.Len(t, entity, 2)
}

func TestFileOffset(t *testing.T) {
	// 创建一个临时文件用于测试
	tmpfile, err := os.CreateTemp("testdata", "wal_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte("test data"))
	assert.NoError(t, err)
	off, err := tmpfile.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	t.Logf("current offset is %d", off)

	// 文件可以多次打开， 而且以os.O_RDWR打开的文件,offset为0
	f, err := os.OpenFile(tmpfile.Name(), os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	curoff, err := f.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	t.Logf("open finished, curr offset is %d", curoff)

	// 文件可以多次打开， 而且以os.O_RDONLy打开的文件,offset也为0
	rf, err := os.OpenFile(tmpfile.Name(), os.O_RDONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rfOffset, err := rf.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	t.Logf("open finished, curr offset is %d", rfOffset)
}

func TestSelectFiles(t *testing.T) {
	// Create a Wal instance
	w := &Wal{
		dir: "testdata",
	}

	// Test case 1: offset < 0
	names := []string{"testdata/segment_1.wal", "testdata/segment_5.wal", "testdata/segment_9.wal"}
	offset := int64(-1)
	result, _, err := w.selectFiles(names, offset)
	assert.NoError(t, err)
	assert.Equal(t, names, result)

	var offsetMatrix [][]int64 = [][]int64{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
	}
	var expectMatrix [][]string = [][]string{
		{"testdata/segment_1.wal", "testdata/segment_5.wal", "testdata/segment_9.wal"},
		{"testdata/segment_5.wal", "testdata/segment_9.wal"},
		{"testdata/segment_9.wal"},
	}

	// Test case 3: offset within range
	for i, item := range offsetMatrix {
		for _, offset := range item {
			result, _, err = w.selectFiles(names, offset)
			assert.NoError(t, err)
			assert.Equal(t, expectMatrix[i], result)
		}
	}

	names = []string{"testdata/segment_.wal", "testdata/segment_2.wal", "testdata/segment_3.wal"}
	offset = int64(1)
	_, _, err = w.selectFiles(names, offset)
	assert.Error(t, err)
}
