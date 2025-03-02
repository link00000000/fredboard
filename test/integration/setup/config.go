package setup

import (
	"errors"
	"os"
	"path"
	"testing"

	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/optional"
)

func SetupConfig(t *testing.T) config.Config {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal("failed to get working directory", err)
	}

	if err := config.Init(optional.Make(path.Join(cwd, "config.json"))); err != nil {
		t.Fatal("failed to initialize config", err)
	}

	if ok, errs := config.Validate(); !ok {
		t.Fatal("invalid config", errors.Join(errs...))
	}

	return config.Get()
}
