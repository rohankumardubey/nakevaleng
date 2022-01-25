package record

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"time"
)

const (
	RECORD_STATUS_DEFAULT    = 0 << 0
	RECORD_TOMBSTONE_REMOVED = 1 << 0
)

// Atomic unit of information with all required context.
type Record struct {
	Crc       uint32 // Checksum of key and value ONLY!!!
	Timestamp int64  // Creation time as UNIX timestamp
	Status    uint8  // Status bits, see the documentation for more info.
	TypeInfo  uint8  // Type ID of 'Value'. Defaults to 0 (no type).
	KeySize   uint64 // Size of Key (in bytes)
	ValueSize uint64 // Size of Value (in bytes)
	Key       []byte //
	Value     []byte //
}

// New() creates a Record object with the key and value specified as byte slices.
//	key	::	Key for this Record
//	val	::	Value for this Record
func New(key, val []byte) Record {
	return Record{
		Crc:       crc32.ChecksumIEEE(append(key[:], val[:]...)),
		Timestamp: time.Now().Unix(),
		Status:    RECORD_STATUS_DEFAULT,
		TypeInfo:  0,
		KeySize:   uint64(len(key)),
		ValueSize: uint64(len(val)),
		Key:       key,
		Value:     val,
	}
}

// NewTyped() creates a Record object with the given key and value and assigns it a type.
// Note that nakevaleng does not provide any context for the types.
func NewTyped(key, val []byte, typeinfo uint8) Record {
	r := New(key, val)
	r.TypeInfo = typeinfo
	return r
}

// NewFromString() is like New() but with key and val specified as strings.
func NewFromString(key, val string) Record {
	return New([]byte(key), []byte(val))
}

// Clone() creates a new Record object with fields the copied from 'rec'.
// The timestamp is NOT copied!
func Clone(rec Record) Record {
	return Record{
		Crc:       rec.Crc,
		Timestamp: time.Now().Unix(),
		Status:    rec.Status,
		TypeInfo:  rec.TypeInfo,
		KeySize:   rec.KeySize,
		ValueSize: rec.ValueSize,
		Key:       rec.Key,
		Value:     rec.Value,
	}
}

// NewEmpty() creates an empty Record object.
func NewEmpty() Record {
	return New(make([]byte, 0), make([]byte, 0))
}

// IsDeleted() checks for the Tombstone bit in the record's Status field.
func (this Record) IsDeleted() bool {
	return (this.Status & RECORD_TOMBSTONE_REMOVED) != 0
}

// ToString() returns a string representation of the record suitable for reading and debugging.
// The Status and TypeInfo fields are printed in binary.
func (this Record) ToString() string {
	return fmt.Sprintf("Record(%d %d %08b %08b %d %d %v %v)",
		this.Crc,
		this.Timestamp,
		this.Status,
		this.TypeInfo,
		this.KeySize,
		this.ValueSize,
		this.Key,
		this.Value,
	)
}

func (this *Record) CalcCRC() {
	this.Crc = crc32.ChecksumIEEE(append(this.Key[:], this.Value[:]...))
}

// Seralize() appends the contents of the Record to a file in binary mode.
func (this Record) Serialize(fname string) {
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	this.serialize(writer)
}

// Deserialize() reads data from buffered reader and overwrites this record.
// The checksum is recalculated and compared with the one read from the file.
// The function will panic if they don't match.
func (this *Record) Deserialize(reader *bufio.Reader) {
	err := binary.Read(reader, binary.LittleEndian, &this.Crc)
	err = binary.Read(reader, binary.LittleEndian, &this.Timestamp)
	err = binary.Read(reader, binary.LittleEndian, &this.Status)
	err = binary.Read(reader, binary.LittleEndian, &this.TypeInfo)
	err = binary.Read(reader, binary.LittleEndian, &this.KeySize)
	err = binary.Read(reader, binary.LittleEndian, &this.ValueSize)

	this.Key = make([]byte, this.KeySize)
	this.Value = make([]byte, this.ValueSize)

	err = binary.Read(reader, binary.LittleEndian, &this.Key)
	err = binary.Read(reader, binary.LittleEndian, &this.Value)

	if err != nil {
		panic(err.Error())
	}

	// Checksum
	crc := crc32.ChecksumIEEE(append(this.Key[:], this.Value[:]...))

	if crc != this.Crc {
		fmt.Println("Bad Record checksum (got ", crc, ", expected ", this.Crc, ")")
		fmt.Println(this.ToString())
		panic("")
	}
}

// serialize() appends the contents of the Record using a buffered writer, in binary mode.
// Note that the writer does NOT get flushed. It's up to the caller to invoke writer.Flush().
func (this Record) serialize(writer *bufio.Writer) {
	err := binary.Write(writer, binary.LittleEndian, this.Crc)
	err = binary.Write(writer, binary.LittleEndian, this.Timestamp)
	err = binary.Write(writer, binary.LittleEndian, this.Status)
	err = binary.Write(writer, binary.LittleEndian, this.TypeInfo)
	err = binary.Write(writer, binary.LittleEndian, this.KeySize)
	err = binary.Write(writer, binary.LittleEndian, this.ValueSize)
	err = binary.Write(writer, binary.LittleEndian, this.Key)
	err = binary.Write(writer, binary.LittleEndian, this.Value)

	if err != nil {
		panic(err.Error())
	}
}
