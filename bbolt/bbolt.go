package bbolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fahmifan/covidbot"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type UserServiceCfg struct {
	DB *bolt.DB
}

type UserService struct {
	*UserServiceCfg
}

func NewUserService(cfg *UserServiceCfg) *UserService {
	us := &UserService{cfg}
	us.init()
	return us
}

func (u *UserService) bucket() []byte {
	return []byte("user_service")
}

func (u *UserService) init() {
	u.DB.Update(func(t *bolt.Tx) error {
		_, err := u.createBucket(t, u.bucket())
		return err
	})
}

func (u *UserService) createBucket(t *bolt.Tx, bucket []byte) (*bolt.Bucket, error) {
	b, err := t.CreateBucketIfNotExists(bucket)
	if err != nil {
		logrus.Error(err)
	}
	return b, err
}

func (u *UserService) findByUserName(ctx context.Context, userName string) (*covidbot.User, error) {
	var user *covidbot.User
	err := u.DB.View(func(t *bolt.Tx) error {
		b := t.Bucket(u.bucket())
		v := b.Get([]byte(userName))
		if v == nil {
			return covidbot.ErrNotFound
		}

		us := &covidbot.User{}
		err := json.Unmarshal(v, us)
		if err != nil {
			return fmt.Errorf("unable to unmarshal: %w", err)
		}
		user = us
		return nil
	})
	return user, err
}

// Create ..
func (u *UserService) Create(ctx context.Context, user *covidbot.User) error {
	oldUser, err := u.findByUserName(ctx, user.Username)
	if err != nil && err != covidbot.ErrNotFound {
		return fmt.Errorf("unable to find user: %w", err)
	}

	// skip recreate
	if oldUser != nil {
		user.ID = oldUser.ID
		return nil
	}

	err = u.DB.Update(func(t *bolt.Tx) error {
		b := t.Bucket(u.bucket())
		id, err := gonanoid.New()
		if err != nil {
			return fmt.Errorf("unable to generate id: %w", err)
		}

		user.ID = id
		err = b.Put([]byte(user.Username), covidbot.Dump(user))
		if err != nil {
			return fmt.Errorf("unable to put user: %w", err)
		}

		return nil
	})

	return err
}

// ForEach ..
func (u *UserService) ForEach(ctx context.Context, cb func(user covidbot.User) error) (err error) {
	return u.DB.View(func(t *bolt.Tx) error {
		b := t.Bucket(u.bucket())
		return b.ForEach(func(k, v []byte) error {
			user := covidbot.User{}
			err := json.Unmarshal(v, &user)
			if err != nil {
				return fmt.Errorf("unable to unmarshal: %w", err)
			}
			return cb(user)
		})
	})
}
