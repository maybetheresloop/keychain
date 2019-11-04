package proto

type Item struct {
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
		KeySize: int64(keySize),
		ValueSize: int64(valueSize),
		Key: key,
		Value: value,
	}
}