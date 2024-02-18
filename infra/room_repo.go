package infra

import (
	"fmt"
	"sync"

	"github.com/Kurichi/carcharhinus-chat/domain"
	"github.com/pkg/errors"
)

type RoomRepository struct {
	lock  *sync.RWMutex
	users map[string]map[string]*domain.User
}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{
		lock:  &sync.RWMutex{},
		users: map[string]map[string]*domain.User{},
	}
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
	r.users[roomID][user.Name] = user

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

func (r *RoomRepository) PushMsg(roomID string, msg *domain.Comment) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	fmt.Println("PushMsg", roomID, msg)

	room, ok := r.users[roomID]
	if !ok {
		return errors.WithStack(domain.ErrRoomNotFound)
	}

	// users := make([]*domain.User, 0, len(room))
	for _, user := range room {
		fmt.Println("PushMsg", user.Name, msg)
		// users = append(users, user)
		user.Ch <- *msg
	}

	return nil
}
