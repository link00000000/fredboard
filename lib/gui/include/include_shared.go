//go:build debug

package include

import (
	"errors"
	"fmt"
	"plugin"
)

var ErrUnexpectedSymbolInterface = errors.New("unexpected symbol interface")

var (
	renderImpl func() error
)

func Load(path string) error {
	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	symRender, err := plug.Lookup("Render")
	if err != nil {
		return fmt.Errorf("failed to lookup symbol Render: %w", err)
	}

	render, ok := symRender.(func() error)
	if !ok {
		return fmt.Errorf("failed to resolve interface for symbol Render: %w", ErrUnexpectedSymbolInterface)
	}

	fmt.Println("LOADED SHARED LIB")

	renderImpl = render
	return nil
}

func Render() error {
	if renderImpl == nil {
		panic("no implementation for \"Render() error\"")
	}

	return renderImpl()
}
