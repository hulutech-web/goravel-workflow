package factories

import "github.com/brianvoe/gofakeit/v6"

type UserFactory struct {
}

// Definition Define the model's default state.
func (f *UserFactory) Definition() map[string]any {
	return map[string]any{
		"nickName":  gofakeit.Name(),
		"AvatarUrl": gofakeit.ImageURL(100, 100),
		"Mobile":    gofakeit.Phone(),
		"Openid":    gofakeit.UUID(),
		"Unionid":   gofakeit.UUID(),
		"Address":   []string{},
	}
}
