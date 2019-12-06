package data

// An item stored in a Keychain database. An item has the following representation
// when serialized to disk:
//
// +-----------+----------+-----+------------+-------+
// | Timestamp | Key Size | Key | Value Size | Value |
// +-----------+----------+-----+------------+-------+
//
// The item's timestamp is used to determine the most recent version of a item.
type Item struct {
	Timestamp int64
	KeySize   int64
	Key       []byte
	ValueSize int64
	Value     []byte
}

func NewItem(key []byte, value []byte) *Item {
	keySize := -1
	if key != nil {
		keySize = len(key)
	}

	valueSize := -1
	if value != nil {
		valueSize = len(value)
	}

	return &Item{
		KeySize:   int64(keySize),
		ValueSize: int64(valueSize),
		Key:       key,
		Value:     value,
	}
}

func NewItemDeleteMarker(key []byte) *Item {
	return &Item{
		KeySize:   int64(len(key)),
		Key:       key,
		ValueSize: -1,
		Value:     nil,
	}
}
