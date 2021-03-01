package instagram

import (
	"context"
	"errors"

	"github.com/kr/pretty"

	"github.com/ahmdrz/goinsta/v2"
)

type AuthError struct{ error }

func RestoreSession(path string) (*Instagram, error) {
	i, err := goinsta.Import(path)

	return &Instagram{i}, err
}

func Login(user, pass, exportPath string) (*Instagram, error) {
	i := goinsta.New(user, pass)
	err := i.Login()
	if err == nil {
		i.Export(exportPath)

		return &Instagram{i}, nil
	}

	authErr := goinsta.Error400{}
	if errors.As(err, &authErr) {
		return nil, AuthError{errors.New(authErr.Message)}
	}

	return nil, err
}

type Instagram struct{ *goinsta.Instagram }

func (i Instagram) Ping() error {
	err := i.Instagram.Account.Sync()
	errN := goinsta.ErrorN{}
	if !errors.As(err, &errN) {
		return err
	}
	if errN.Message == "login_required" {
		return AuthError{errN}
	}

	return nil
}

func (i Instagram) Followers(_ context.Context) (map[int64]User, error) {
	return getAll(i.Instagram.Account.Followers())
}

func (i Instagram) Following(_ context.Context) (map[int64]User, error) {
	return getAll(i.Instagram.Account.Following())
}

func getAll(i *goinsta.Users) (map[int64]User, error) {
	if err := i.Error(); err != nil {
		pretty.Println(err)
		return nil, err
	}

	users := make(map[int64]User)
	for i.Next() {
		for _, u := range i.Users {
			users[u.ID] = User{
				ID:         u.ID,
				Username:   u.Username,
				Fullname:   u.FullName,
				ProfilePic: u.ProfilePicURL,
			}
		}
	}

	return users, nil
}
