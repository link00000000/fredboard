//go:build !debug

package include

import "fmt"

import "accidentallycoded.com/fredboard/v3/lib/gui/internal"

func Load(path string) error {
	fmt.Println("LOADED STATIC LIB")
	return nil
}

func Render() error {
	return internal.Render()
}
