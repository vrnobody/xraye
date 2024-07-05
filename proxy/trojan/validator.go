package trojan

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
)

// Validator stores valid trojan users.
type Validator struct {
	// Considering email's usage here, map + sync.Mutex/RWMutex may have better performance.
	email         sync.Map
	users         sync.Map
	userInfoCache sync.Map
}

// Add a trojan user, Email must be empty or unique.
func (v *Validator) Add(u *protocol.MemoryUser) error {
	if u.Email != "" {
		_, loaded := v.email.LoadOrStore(strings.ToLower(u.Email), u)
		if loaded {
			return errors.New("User ", u.Email, " already exists.")
		}
	}
	v.users.Store(hexString(u.Account.(*MemoryAccount).Key), u)
	return nil
}

// Del a trojan user with a non-empty Email.
func (v *Validator) Del(e string) error {
	if e == "" {
		return errors.New("Email must not be empty.")
	}
	le := strings.ToLower(e)
	u, _ := v.email.Load(le)
	if u == nil {
		return errors.New("User ", e, " not found.")
	}
	mu := u.(*protocol.MemoryUser)
	v.userInfoCache.Delete(mu)
	v.email.Delete(le)
	v.users.Delete(hexString(mu.Account.(*MemoryAccount).Key))
	return nil
}

// Get a trojan user with hashed key, nil if user doesn't exist.
func (v *Validator) Get(hash string) *protocol.MemoryUser {
	u, _ := v.users.Load(hash)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}

// GetAll users info.
func (v *Validator) GetAll() ([]string, bool) {
	type info struct {
		Password string
		Email    string
		Level    uint32
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
				Password: ma.Password,
				Email:    mu.Email,
				Level:    mu.Level,
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
