package dtb

import (
	"encoding/binary"
	"reflect"
	"testing"
)

type TestStruct1 struct {
	Abcd  string `bytes_length:"10"`
	Abcde string `bytes_length:"10" bytes_ignore:"true"`
}
type TestStruct2 struct {
	Abcd  string `bytes_length:"10"`
	Abcde string `bytes_length:"10"`
	B     byte
}
type CustomFnStruct struct {
	Ip string `bytes_fn:"EncodeIP,DecodeIP"`
}

func (t CustomFnStruct) EncodeIP() ([]byte, error) {
	return []byte{192, 168, 0, 1}, nil
}

func (t CustomFnStruct) DecodeIP(data []byte) (int, error) {
	t.Ip = "192.168.0.1"
	return 4, nil
}

func TestIgnore(t *testing.T) {
	data := TestStruct1{}
	newData := TestStruct1{}
	data.Abcd = "1234"
	data.Abcde = "456"

	// Test copy with non-empty ignored field (shouldn't equal to original)
	bytes, err := ConvertDataToBytes(data, binary.LittleEndian)
	if err != nil {
		t.Error(err)
	}
	err = ConvertBytesToData(bytes, binary.LittleEndian, &newData)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(data, newData) {
		t.Error("New struct shouldn't ne equal original one")
	}
	// Test copy with empty ignored field (should equal to original)
	data.Abcde = ""
	bytes, err = ConvertDataToBytes(data, binary.LittleEndian)
	if err != nil {
		t.Error(err)
	}
	err = ConvertBytesToData(bytes, binary.LittleEndian, &newData)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(data, newData) {
		t.Error("New struct doesn't equal original one")
	}
}
func TestCopyData(t *testing.T) {
	data := [2]TestStruct2{}
	newData := [2]TestStruct2{}

	data[0].Abcd = "1234"
	data[0].Abcde = "456"
	data[0].B = 15

	data[1].Abcd = "789"
	data[1].Abcde = "10111"
	data[1].B = 30

	bytes, err := ConvertDataToBytes(data, binary.LittleEndian)
	if err != nil {
		t.Error(err)
	}
	err = ConvertBytesToData(bytes, binary.LittleEndian, &newData)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(data, newData) {
		t.Error("New struct doesn't equal original one")
	}
}
func TestCustomFunction(t *testing.T) {
	data := CustomFnStruct{Ip: "192.168.0.1"}
	b, err := ConvertDataToBytes(&data, binary.LittleEndian)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(b, []byte{192, 168, 0, 1}) {
		t.Error("Bad custom function parsing")
		return
	}
	data1 := CustomFnStruct{}
	err = ConvertBytesToData(b, binary.LittleEndian, &data1)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(data1, data) {
		t.Error("Bad custom function parsing")
	}

}
