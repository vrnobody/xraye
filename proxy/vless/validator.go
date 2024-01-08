package vless

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/uuid"
)

// Validator stores valid VLESS users.
type Validator struct {
	// Considering email's usage here, map + sync.Mutex/RWMutex may have better performance.
	email sync.Map
	users sync.Map

	userInfoCache sync.Map
}

// Add a VLESS user, Email must be empty or unique.
func (v *Validator) Add(u *protocol.MemoryUser) error {
	if u.Email != "" {
		_, loaded := v.email.LoadOrStore(strings.ToLower(u.Email), u)
		if loaded {
			return newError("User ", u.Email, " already exists.")
		}
	}
	v.users.Store(u.Account.(*MemoryAccount).ID.UUID(), u)
	return nil
}

// Del a VLESS user with a non-empty Email.
func (v *Validator) Del(e string) error {
	if e == "" {
		return newError("Email must not be empty.")
	}
	le := strings.ToLower(e)
	u, _ := v.email.Load(le)
	if u == nil {
		return newError("User ", e, " not found.")
	}
	mu := u.(*protocol.MemoryUser)
	v.userInfoCache.Delete(mu)
	v.email.Delete(le)
	v.users.Delete(mu.Account.(*MemoryAccount).ID.UUID())
	return nil
}

// GetAll users info.
func (v *Validator) GetAll() ([]string, bool) {
	type info struct {
		Id    string
		Flow  string
		Email string
		Level uint32
	}

	users := make([]string, 0)
	v.users.Range(func(_, value interface{}) bool {
		mu, ok := value.(*protocol.MemoryUser)
		if !ok {
			return true
		}
		if o, ok := v.userInfoCache.Load(mu); ok {
			if o == nil {
				return true
			}
			if user, ok := o.(string); ok {
				users = append(users, user)
				return true
			}
		}
		if ma, ok := mu.Account.(*MemoryAccount); ok {
			info := &info{
				Id:    ma.ID.String(),
				Flow:  ma.Flow,
				Email: mu.Email,
				Level: mu.Level,
			}
			if j, err := json.MarshalIndent(info, "", "  "); err == nil {
				user := string(j)
				users = append(users, user)
				v.userInfoCache.Store(mu, user)
				return true
			}
		}
		v.userInfoCache.Store(mu, nil)
		return true
	})
	return users, true
}

// Get a VLESS user with UUID, nil if user doesn't exist.
func (v *Validator) Get(id uuid.UUID) *protocol.MemoryUser {
	u, _ := v.users.Load(id)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}
