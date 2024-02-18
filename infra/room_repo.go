package infra

import (
	"sync"

	"github.com/Kurichi/carcharhinus-chat/domain"
	"github.com/pkg/errors"
)

type RoomRepository struct {
	lock  *sync.RWMutex
	users map[string]map[string]*domain.User
}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{}
}

func (r *RoomRepository) IsRoomExist(roomID string) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	_, ok := r.users[roomID]
	return ok
}

func (r *RoomRepository) CreateRoom(roomID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.users[roomID]; !ok {
		r.users[roomID] = map[string]*domain.User{}
	}
}

func (r *RoomRepository) DeleteRoom(roomID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.users, roomID)
}

func (r *RoomRepository) AddUser(roomID string, user *domain.User) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.users[roomID]; !ok {
		r.users[roomID] = map[string]*domain.User{}
		// return errors.WithStack(domain.ErrRoomNotFound)
	}
	r.users[roomID][user.ID] = user

	return nil
}

func (r *RoomRepository) RemoveUser(roomID, userID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.users[roomID]; !ok {
		return errors.WithStack(domain.ErrRoomNotFound)
	}
	delete(r.users[roomID], userID)

	return nil
}

func (r *RoomRepository) GetUsers(roomID string) ([]*domain.User, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	room, ok := r.users[roomID]
	if !ok {
		return nil, errors.WithStack(domain.ErrRoomNotFound)
	}

	users := make([]*domain.User, 0, len(room))
	for _, user := range room {
		users = append(users, user)
	}

	return users, nil
}
