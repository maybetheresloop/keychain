package data

type Entry struct {
	Timestamp int64
	FileID    uint64
	ValueSize int64
	ValuePos  int64
}

func NewEntry(fileID uint64, valueSize int64, valuePos int64, timestamp int64) *Entry {
	return &Entry{
		FileID:    fileID,
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}
}
