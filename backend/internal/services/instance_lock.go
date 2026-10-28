package services

import (
	"sync"

	"github.com/google/uuid"
)

// instanceWriteLocks serializes pg_hba.conf writes per instance across every
// call site (rule create/reorder/clone, temp-access grant/revoke, and the
// scheduler's auto-expire sweep). These are triggered from independent
// goroutines with no other coordination, and each one truncates and
// rewrites the same remote file over its own SSH session — without a lock,
// two overlapping writes to the same instance can interleave or leave the
// file truncated if one of them is interrupted mid-write.
var instanceWriteLocks sync.Map // map[uuid.UUID]*sync.Mutex

// lockInstanceWrite blocks until any other in-flight write for this instance
// finishes, then returns an unlock func to release it.
func lockInstanceWrite(instanceID uuid.UUID) func() {
	v, _ := instanceWriteLocks.LoadOrStore(instanceID, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}
