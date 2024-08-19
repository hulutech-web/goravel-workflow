package feature

import (
	"github.com/goravel/framework/facades"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_, err := getGormDBConnection()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := facades.Route().Run(); err != nil {
			panic(err)
		}
	}()

	exit := m.Run()
	os.Exit(exit)
}
