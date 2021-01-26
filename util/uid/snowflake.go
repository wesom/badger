package uid

import (
	"fmt"

	"github.com/sony/sonyflake"
)

// ID wrapper a unique id generator implement with snowflake
type ID struct {
	sf *sonyflake.Sonyflake
}

// New return a ID instance
func New() *ID {
	var st sonyflake.Settings
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
	return &ID{
		sf: sf,
	}
}

// Next return a new unique id
func (s *ID) Next() uint64 {
	id, err := s.sf.NextID()
	if err != nil {
		panic(fmt.Sprintf("id generate err: %v", err))
	}
	return id
}
