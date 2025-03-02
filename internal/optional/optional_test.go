package optional_test

import (
	"encoding/json"
	"slices"
	"testing"

	"accidentallycoded.com/fredboard/v3/internal/optional"
)

func TestMakeFromValue(t *testing.T) {
	v := optional.Make(1)

	if !v.IsSet() {
		t.Fatal("value was not marked as set")
	}

	if v.Get() != 1 {
		t.Fatal("stored value is incorrect")
	}
}

func TestMakeNonNilFromPtr(t *testing.T) {
	vv := 1
	v := optional.MakePtr(&vv)

	if !v.IsSet() {
		t.Fatal("value was not marked as set")
	}

	if v.Get() != 1 {
		t.Fatal("stored value is incorrect")
	}
}

func TestOptionalMakeNilFromPtr(t *testing.T) {
	var vv *int = nil
	v := optional.MakePtr(vv)

	if v.IsSet() {
		t.Fatal("value was marked as set")
	}
}

func TestMakeNone(t *testing.T) {
	v := optional.None[int]()

	if v.IsSet() {
		t.Fatal("value was marked as set")
	}
}

func TestOverrideValue(t *testing.T) {
	v := optional.Make(1)
	v.Set(2)

	if !v.IsSet() {
		t.Fatal("value was not marked as set")
	}

	if v.Get() != 2 {
		t.Fatal("stored value is incorrect")
	}
}

func TestUnsetValue(t *testing.T) {
	v := optional.Make(1)
	v.Unset()

	if v.IsSet() {
		t.Fatal("value was marked as set")
	}
}

func TestPanicOnGetUnsetValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic but none occurred")
		}
	}()

	v := optional.None[int]()
	_ = v.Get()
}

func TestJsonMarshallString(t *testing.T) {
	v := optional.Make("testing")

	bytes, err := json.Marshal(v)

	if err != nil {
		t.Fatal("failed to marshall data", err)
	}

	if !slices.Equal(bytes, []byte(`"testing"`)) {
		t.Logf("bytes = %s", string(bytes))
		t.Fatal("marshalled value is incorrect")
	}
}

func TestJsonUnmarshallString(t *testing.T) {
}

func TestJsonMarshallInt(t *testing.T) {
	v := optional.Make(1)

	bytes, err := json.Marshal(v)

	if err != nil {
		t.Fatal("failed to marshall data", err)
	}

	if !slices.Equal(bytes, []byte(`1`)) {
		t.Logf("bytes = %s", string(bytes))
		t.Fatal("marshalled value is incorrect")
	}
}

func TestJsonUnmarshallInt(t *testing.T) {
}

func TestJsonMarshallFloat(t *testing.T) {
	v := optional.Make(1.23)

	bytes, err := json.Marshal(v)

	if err != nil {
		t.Fatal("failed to marshall data", err)
	}

	if !slices.Equal(bytes, []byte(`1.23`)) {
		t.Logf("bytes = %s", string(bytes))
		t.Fatal("marshalled value is incorrect")
	}
}

func TestJsonUnmarshallFloat(t *testing.T) {
}

func TestJsonMarshallBool(t *testing.T) {
	v := optional.Make(false)

	bytes, err := json.Marshal(v)

	if err != nil {
		t.Fatal("failed to marshall data", err)
	}

	if !slices.Equal(bytes, []byte(`false`)) {
		t.Logf("bytes = %s", string(bytes))
		t.Fatal("marshalled value is incorrect")
	}
}

func TestJsonUnmarshallBool(t *testing.T) {
}

func TestJsonMarshallNil(t *testing.T) {
	v := optional.None[int]()

	bytes, err := json.Marshal(v)

	if err != nil {
		t.Fatal("failed to marshall data", err)
	}

	if !slices.Equal(bytes, []byte(`null`)) {
		t.Logf("bytes = %s", string(bytes))
		t.Fatal("marshalled value is incorrect")
	}
}

func TestJsonUnmarshallNil(t *testing.T) {
}
