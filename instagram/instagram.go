package instagram

import (
	"context"
	"errors"
	"net/url"

	"github.com/Davincible/goinsta/v3"
)

type AuthError struct{ error }

func RestoreSession(path string) (*Instagram, error) {
	i, err := goinsta.Import(path)
	if errors.Is(err, goinsta.ErrLoggedOut) {
		return nil, &AuthError{err}
	}

	return &Instagram{i}, err
}

func Login(user, pass, exportPath string) (*Instagram, error) {
	i := goinsta.New(user, pass)
	err := i.Login()
	if err == nil {
		i.Export(exportPath)

		return &Instagram{i}, nil
	}

	var authErr *goinsta.Error400
	if errors.As(err, &authErr) {
		return nil, AuthError{errors.New(authErr.Message)}
	}

	return nil, err
}

type Instagram struct{ *goinsta.Instagram }

func (i Instagram) Ping() error {
	err := i.Account.Sync()
	var errN *goinsta.ErrorN
	if !errors.As(err, &errN) {
		return err
	}
	if errN.Message == "login_required" {
		return AuthError{errN}
	}

	return nil
}

func (i Instagram) Followers(_ context.Context) (map[int64]User, error) {
	return getAll(i.Account.Followers(""))
}

func (i Instagram) Following(_ context.Context) (map[int64]User, error) {
	return getAll(i.Account.Following("", goinsta.LatestOrder))
}

func getAll(i *goinsta.Users) (map[int64]User, error) {
	if err := i.Error(); err != nil {
		return nil, err
	}

	users := make(map[int64]User)

	for i.Next() {
		for _, u := range i.Users {
			users[u.ID] = User{
				ID:         u.ID,
				Username:   u.Username,
				Fullname:   u.FullName,
				ProfilePic: url.QueryEscape(u.ProfilePicURL),
			}
		}
	}

	return users, nil
}
