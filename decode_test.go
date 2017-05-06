package d2b

import (
	"encoding/binary"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDecode(t *testing.T) {
	Convey("Test Decode", t, func() {
		Convey("Should decode int8", func() {
			var result int8
			err := Decode([]byte{255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, -1)
		})
		Convey("Should decode uint8", func() {
			var result uint8
			err := Decode([]byte{255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 255)
		})
		Convey("Should decode int16", func() {
			var result int16
			err := Decode([]byte{255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, -1)
		})
		Convey("Should decode uint16", func() {
			var result uint16
			err := Decode([]byte{255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 65535)
		})
		Convey("Should decode int32", func() {
			var result int32
			err := Decode([]byte{255, 255, 255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, -1)
		})
		Convey("Should decode uint32", func() {
			var result uint32
			err := Decode([]byte{255, 255, 255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 4294967295)
		})

		Convey("Should decode int64", func() {
			var result int64
			err := Decode([]byte{255, 255, 255, 255, 255, 255, 255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, -1)
		})
		Convey("Should decode uint64", func() {
			var result uint64
			err := Decode([]byte{255, 255, 255, 255, 255, 255, 255, 255}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, uint64(18446744073709551615))
		})

		Convey("Should create new value for nil pointers and decode there data", func() {
			var result *int32
			err := Decode([]byte{1, 2, 3, 4}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(*result, ShouldEqual, 67305985)
		})

		Convey("Should decode float32", func() {
			var result float32
			err := Decode([]byte{1, 2, 3, 4}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 1.5399896e-36)
		})
		Convey("Should decode float64", func() {
			var result float64
			err := Decode([]byte{1, 2, 3, 4, 5, 6, 7, 8}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 5.447603722011605e-270)
		})
		Convey("Should decode struct", func() {
			type Struct struct {
				A     *[]int32  `d2b:"length:2"`
				B     *[]uint32 `d2b:"length:2"`
				C     [2]int32
				D     int    `d2b:"-"`
				Test  string `d2b:"length:6"`
				Test1 string `d2b:"length:4"`
			}
			bVal := make([]uint32, 1)
			var result = Struct{B: &bVal}
			err := Decode([]byte{
				1, 2, 3, 4,
				1, 2, 3, 4,
				1, 2, 3, 4,
				1, 2, 3, 4,
				1, 2, 3, 4,
				1, 2, 3, 4,
				'H', 'e', 'l', 'l', 0, 0,
				'H', 'e', 'l', 'l',
			}, binary.LittleEndian, &result)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, Struct{
				A:     &[]int32{67305985, 67305985},
				B:     &[]uint32{513, 513},
				C:     [2]int32{67305985, 67305985},
				Test:  "Hell",
				Test1: "Hell",
			})
		})
		Convey("Should return error if trying to decode unsupported type", func() {
			var result int
			err := Decode([]byte{1, 2, 3, 4}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})
		Convey("Should return error if trying to decode to nil pointer", func() {
			var result *int
			err := Decode([]byte{1, 2, 3, 4}, binary.LittleEndian, result)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, nil)
		})
		Convey("Should return error if trying to decode to non-pointer type", func() {
			var result int
			err := Decode([]byte{1, 2, 3, 4}, binary.LittleEndian, result)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})
		Convey("Should return error if trying to decode struct with bad field type", func() {
			type Struct struct {
				A int
			}
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
		Convey("Should return error if trying to decode struct with bad tags", func() {
			type Struct struct {
				A string `d2b:"length:hello"`
			}
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
		Convey("Should return error if trying to decode struct string field without length", func() {
			type Struct struct{ A string }
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
		Convey("Should return error if trying to decode struct slice field without length", func() {
			type Struct struct{ A []int32 }
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
		Convey("Should return error if trying to decode struct slice field with bad elements", func() {
			type Struct struct {
				A []int `d2b:"length:2"`
			}
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
		Convey("Should return error if trying to decode struct slice field with bad elements with already filled slice", func() {
			type Struct struct {
				A []int `d2b:"length:2"`
			}
			var result = Struct{A: []int{1, 2}}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{A: []int{1, 2}})
		})
		Convey("Should return error if trying to decode struct array field with bad elements", func() {
			type Struct struct {
				A [2]int
			}
			var result = Struct{}
			err := Decode([]byte{
				1, 2, 3, 4,
			}, binary.LittleEndian, &result)
			So(err, ShouldNotBeNil)
			So(result, ShouldResemble, Struct{})
		})
	})
}
