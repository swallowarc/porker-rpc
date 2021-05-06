package room

import (
	"fmt"
	"time"

	"github.com/swallowarc/porker-rpc/internal/commons/random"
)

const (
	idPattern       = "1234567890"
	idKeyPrefix     = "porker_room_id"
	memberKeyPrefix = "porker_room_member"
	streamKeyPrefix = "porker_room_stream"
)

const (
	TimeoutDuration = 15 * time.Minute
)

type (
	ID string
)

func NewID() ID {
	return ID(random.RandString6ByParam(5, idPattern))
}

func (id ID) IDKey() string {
	return fmt.Sprintf("%s:%s", idKeyPrefix, id)
}

func (id ID) MemberKey() string {
	return fmt.Sprintf("%s:%s", memberKeyPrefix, id)
}

func (id ID) StreamKey() string {
	return fmt.Sprintf("%s:%s", streamKeyPrefix, id)
}

func (id ID) String() string {
	return string(id)
}
