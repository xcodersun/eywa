package utils

import (
	"encoding/json"
	"encoding/xml"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
	"time"
)

type OneLevel struct {
	OneInt      int           `assign:"one_int;;-"`
	OneInt64    int64         `assign:"one_int64;;-"`
	OneBool     bool          `assign:"one_bool;;-"`
	OneString   string        `assign:"one_string;;-"`
	OneFloat64  float64       `assign:"one_float64;;-"`
	OneDuration time.Duration `assign:"one_duration;duration;-"`

	OneIntPtr      *int           `assign:"one_int_ptr;;-"`
	OneInt64Ptr    *int64         `assign:"one_int64_ptr;;-"`
	OneBoolPtr     *bool          `assign:"one_bool_ptr;;-"`
	OneStringPtr   *string        `assign:"one_string_ptr;;-"`
	OneFloat64Ptr  *float64       `assign:"one_float64_ptr;;-"`
	OneDurationPtr *time.Duration `assign:"one_duration_ptr;duration;-"`
}

type TopLevel struct {
	TopInt      int           `assign:"top_int;;-"`
	TopInt64    int64         `assign:"top_int64;;-"`
	TopBool     bool          `assign:"top_bool;;-"`
	TopString   string        `assign:"top_string;;-"`
	TopFloat64  float64       `assign:"top_float64;;-"`
	TopDuration time.Duration `assign:"top_duration;duration;-"`

	TopIntPtr      *int           `assign:"top_int_ptr;;-"`
	TopInt64Ptr    *int64         `assign:"top_int64_ptr;;-"`
	TopBoolPtr     *bool          `assign:"top_bool_ptr;;-"`
	TopStringPtr   *string        `assign:"top_string_ptr;;-"`
	TopFloat64Ptr  *float64       `assign:"top_float64_ptr;;-"`
	TopDurationPtr *time.Duration `assign:"top_duration_ptr;duration;-"`

	Nested NestedLevel `assign:"nested;;-"`

	NestedPtr *NestedLevel `assign:"nested_ptr;;-"`
}

type NestedLevel struct {
	NestedInt      int           `assign:"nested_int;;-"`
	NestedInt64    int64         `assign:"nested_int64;;-"`
	NestedBool     bool          `assign:"nested_bool;;-"`
	NestedString   string        `assign:"nested_string;;-"`
	NestedFloat64  float64       `assign:"nested_float64;;-"`
	NestedDuration time.Duration `assign:"nested_duration;duration;-"`

	NestedIntPtr      *int           `assign:"nested_int_ptr;;-"`
	NestedInt64Ptr    *int64         `assign:"nested_int64_ptr;;-"`
	NestedBoolPtr     *bool          `assign:"nested_bool_ptr;;-"`
	NestedStringPtr   *string        `assign:"nested_string_ptr;;-"`
	NestedFloat64Ptr  *float64       `assign:"nested_float64_ptr;;-"`
	NestedDurationPtr *time.Duration `assign:"nested_duration_ptr;duration;-"`
}

type PartialTopLevel struct {
	SeenField   int `assign:"seen"`
	IgnoreField string

	SeenStruct   *PartialNestedLevel `assign:"nested_seen"`
	IngoreStruct PartialNestedLevel
}

type PartialNestedLevel struct {
	IgnoreField bool
	SeenField   time.Duration `assign:"seen;duration"`
}

type NilNested struct {
	Dummy int `assign:"dummy"`
}

type NilTopLevel struct {
	Nested *NilNested `assign:"nested"`
	Int    *int       `assign:"int"`
}

type StructAssign struct {
	D    JSONDuration  `assign:"d;jd"`
	DPtr *JSONDuration `assign:"dp;jd"`
}

func TestAssigns(t *testing.T) {
	Convey("successfully assigns values to top level fields/ptrs when they are initialized", t, func() {
		oneInt := 11
		oneInt64 := int64(22)
		oneBool := false
		oneString := "xx"
		oneFloat64 := 0.1122
		oneDuration := 11 * time.Second

		tl := OneLevel{
			OneInt:      1,
			OneInt64:    2,
			OneBool:     false,
			OneString:   "x",
			OneFloat64:  0.12,
			OneDuration: 1 * time.Second,

			OneIntPtr:      &oneInt,
			OneInt64Ptr:    &oneInt64,
			OneBoolPtr:     &oneBool,
			OneFloat64Ptr:  &oneFloat64,
			OneStringPtr:   &oneString,
			OneDurationPtr: &oneDuration,
		}

		js := []byte(`
			{
				"one_int": 5,
				"one_int64": 6,
				"one_bool": true,
				"one_string": "m",
				"one_float64": 0.56,
				"one_duration": "5s",

				"one_int_ptr": 55,
				"one_int64_ptr": 66,
				"one_bool_ptr": true,
				"one_string_ptr": "mm",
				"one_float64_ptr": 0.5656,
				"one_duration_ptr": "55s"
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = ForceAssign(&tl, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		_oneInt := 55
		_oneInt64 := int64(66)
		_oneBool := true
		_oneString := "mm"
		_oneFloat64 := 0.5656
		_oneDuration := 55 * time.Second

		exp := OneLevel{
			OneInt:      5,
			OneInt64:    6,
			OneBool:     true,
			OneString:   "m",
			OneFloat64:  0.56,
			OneDuration: 5 * time.Second,

			OneIntPtr:      &_oneInt,
			OneInt64Ptr:    &_oneInt64,
			OneBoolPtr:     &_oneBool,
			OneFloat64Ptr:  &_oneFloat64,
			OneStringPtr:   &_oneString,
			OneDurationPtr: &_oneDuration,
		}
		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)
	})

	Convey("successfully assigns values to top level fields/ptrs when they are not initialized", t, func() {
		tl := OneLevel{}

		js := []byte(`
			{
				"one_int": 5,
				"one_int64": 6,
				"one_bool": true,
				"one_string": "m",
				"one_float64": 0.56,
				"one_duration": "5s",

				"one_int_ptr": 55,
				"one_int64_ptr": 66,
				"one_bool_ptr": true,
				"one_string_ptr": "mm",
				"one_float64_ptr": 0.5656,
				"one_duration_ptr": "55s"
			}
	 `)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = ForceAssign(&tl, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		_oneInt := 55
		_oneInt64 := int64(66)
		_oneBool := true
		_oneString := "mm"
		_oneFloat64 := 0.5656
		_oneDuration := 55 * time.Second

		exp := OneLevel{
			OneInt:      5,
			OneInt64:    6,
			OneBool:     true,
			OneString:   "m",
			OneFloat64:  0.56,
			OneDuration: 5 * time.Second,

			OneIntPtr:      &_oneInt,
			OneInt64Ptr:    &_oneInt64,
			OneBoolPtr:     &_oneBool,
			OneFloat64Ptr:  &_oneFloat64,
			OneStringPtr:   &_oneString,
			OneDurationPtr: &_oneDuration,
		}
		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)
	})

	Convey("successfully assigns values to nested level fields/ptrs when they are initialized", t, func() {
		nestedInt := 11
		nestedInt64 := int64(22)
		nestedBool := false
		nestedString := "xx"
		nestedFloat64 := 0.1122
		nestedDuration := 11 * time.Second

		nl := NestedLevel{
			NestedInt:      1,
			NestedInt64:    2,
			NestedBool:     false,
			NestedString:   "x",
			NestedFloat64:  0.12,
			NestedDuration: 1 * time.Second,

			NestedIntPtr:      &nestedInt,
			NestedInt64Ptr:    &nestedInt64,
			NestedBoolPtr:     &nestedBool,
			NestedFloat64Ptr:  &nestedFloat64,
			NestedStringPtr:   &nestedString,
			NestedDurationPtr: &nestedDuration,
		}

		nestedIntPtr := 1111
		nestedInt64Ptr := int64(2222)
		nestedBoolPtr := false
		nestedStringPtr := "xxxx"
		nestedFloat64Ptr := 0.11112222
		nestedDurationPtr := 1111 * time.Second

		nlP := &NestedLevel{
			NestedInt:      111,
			NestedInt64:    222,
			NestedBool:     false,
			NestedString:   "xxx",
			NestedFloat64:  0.111222,
			NestedDuration: 111 * time.Second,

			NestedIntPtr:      &nestedIntPtr,
			NestedInt64Ptr:    &nestedInt64Ptr,
			NestedBoolPtr:     &nestedBoolPtr,
			NestedFloat64Ptr:  &nestedFloat64Ptr,
			NestedStringPtr:   &nestedStringPtr,
			NestedDurationPtr: &nestedDurationPtr,
		}

		topInt := -1
		topInt64 := int64(-1)
		topBool := false
		topString := "-x"
		topFloat64 := -0.1122
		topDuration := 10 * time.Second

		tl := TopLevel{
			TopInt:      -2,
			TopInt64:    -44,
			TopBool:     false,
			TopString:   "-y",
			TopFloat64:  -0.12,
			TopDuration: 12 * time.Second,

			TopIntPtr:      &topInt,
			TopInt64Ptr:    &topInt64,
			TopBoolPtr:     &topBool,
			TopFloat64Ptr:  &topFloat64,
			TopStringPtr:   &topString,
			TopDurationPtr: &topDuration,

			Nested:    nl,
			NestedPtr: nlP,
		}

		js := []byte(`
			{
			  "top_int": 5,
			  "top_int64": 6,
			  "top_bool": true,
			  "top_string": "m",
			  "top_float64": 0.56,
			  "top_duration": "5s",
			  "top_int_ptr": 55,
			  "top_int64_ptr": 66,
			  "top_bool_ptr": true,
			  "top_string_ptr": "mm",
			  "top_float64_ptr": 0.5656,
			  "top_duration_ptr": "55s",
			  "nested": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  },
			  "nested_ptr": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  }
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = ForceAssign(&tl, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		_nestedInt := 55
		_nestedInt64 := int64(66)
		_nestedBool := true
		_nestedString := "mm"
		_nestedFloat64 := 0.5656
		_nestedDuration := 55 * time.Second

		_nl := NestedLevel{
			NestedInt:      5,
			NestedInt64:    6,
			NestedBool:     true,
			NestedString:   "m",
			NestedFloat64:  0.56,
			NestedDuration: 5 * time.Second,

			NestedIntPtr:      &_nestedInt,
			NestedInt64Ptr:    &_nestedInt64,
			NestedBoolPtr:     &_nestedBool,
			NestedFloat64Ptr:  &_nestedFloat64,
			NestedStringPtr:   &_nestedString,
			NestedDurationPtr: &_nestedDuration,
		}

		_nestedIntPtr := 55
		_nestedInt64Ptr := int64(66)
		_nestedBoolPtr := true
		_nestedStringPtr := "mm"
		_nestedFloat64Ptr := 0.5656
		_nestedDurationPtr := 55 * time.Second

		_nlP := &NestedLevel{
			NestedInt:      5,
			NestedInt64:    6,
			NestedBool:     true,
			NestedString:   "m",
			NestedFloat64:  0.56,
			NestedDuration: 5 * time.Second,

			NestedIntPtr:      &_nestedIntPtr,
			NestedInt64Ptr:    &_nestedInt64Ptr,
			NestedBoolPtr:     &_nestedBoolPtr,
			NestedFloat64Ptr:  &_nestedFloat64Ptr,
			NestedStringPtr:   &_nestedStringPtr,
			NestedDurationPtr: &_nestedDurationPtr,
		}

		_topInt := 55
		_topInt64 := int64(66)
		_topBool := true
		_topString := "mm"
		_topFloat64 := 0.5656
		_topDuration := 55 * time.Second

		exp := TopLevel{
			TopInt:      5,
			TopInt64:    6,
			TopBool:     true,
			TopString:   "m",
			TopFloat64:  0.56,
			TopDuration: 5 * time.Second,

			TopIntPtr:      &_topInt,
			TopInt64Ptr:    &_topInt64,
			TopBoolPtr:     &_topBool,
			TopFloat64Ptr:  &_topFloat64,
			TopStringPtr:   &_topString,
			TopDurationPtr: &_topDuration,

			Nested:    _nl,
			NestedPtr: _nlP,
		}

		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)
	})

	Convey("successfully assigns values to nested level fields/ptrs when they are not initialized", t, func() {
		tl := TopLevel{}

		js := []byte(`
			{
			  "top_int": 5,
			  "top_int64": 6,
			  "top_bool": true,
			  "top_string": "m",
			  "top_float64": 0.56,
			  "top_duration": "5s",
			  "top_int_ptr": 55,
			  "top_int64_ptr": 66,
			  "top_bool_ptr": true,
			  "top_string_ptr": "mm",
			  "top_float64_ptr": 0.5656,
			  "top_duration_ptr": "55s",
			  "nested": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  },
			  "nested_ptr": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  }
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = ForceAssign(&tl, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		_nestedInt := 55
		_nestedInt64 := int64(66)
		_nestedBool := true
		_nestedString := "mm"
		_nestedFloat64 := 0.5656
		_nestedDuration := 55 * time.Second

		_nl := NestedLevel{
			NestedInt:      5,
			NestedInt64:    6,
			NestedBool:     true,
			NestedString:   "m",
			NestedFloat64:  0.56,
			NestedDuration: 5 * time.Second,

			NestedIntPtr:      &_nestedInt,
			NestedInt64Ptr:    &_nestedInt64,
			NestedBoolPtr:     &_nestedBool,
			NestedFloat64Ptr:  &_nestedFloat64,
			NestedStringPtr:   &_nestedString,
			NestedDurationPtr: &_nestedDuration,
		}

		_nestedIntPtr := 55
		_nestedInt64Ptr := int64(66)
		_nestedBoolPtr := true
		_nestedStringPtr := "mm"
		_nestedFloat64Ptr := 0.5656
		_nestedDurationPtr := 55 * time.Second

		_nlP := &NestedLevel{
			NestedInt:      5,
			NestedInt64:    6,
			NestedBool:     true,
			NestedString:   "m",
			NestedFloat64:  0.56,
			NestedDuration: 5 * time.Second,

			NestedIntPtr:      &_nestedIntPtr,
			NestedInt64Ptr:    &_nestedInt64Ptr,
			NestedBoolPtr:     &_nestedBoolPtr,
			NestedFloat64Ptr:  &_nestedFloat64Ptr,
			NestedStringPtr:   &_nestedStringPtr,
			NestedDurationPtr: &_nestedDurationPtr,
		}

		_topInt := 55
		_topInt64 := int64(66)
		_topBool := true
		_topString := "mm"
		_topFloat64 := 0.5656
		_topDuration := 55 * time.Second

		exp := TopLevel{
			TopInt:      5,
			TopInt64:    6,
			TopBool:     true,
			TopString:   "m",
			TopFloat64:  0.56,
			TopDuration: 5 * time.Second,

			TopIntPtr:      &_topInt,
			TopInt64Ptr:    &_topInt64,
			TopBoolPtr:     &_topBool,
			TopFloat64Ptr:  &_topFloat64,
			TopStringPtr:   &_topString,
			TopDurationPtr: &_topDuration,

			Nested:    _nl,
			NestedPtr: _nlP,
		}

		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)
	})

	Convey("successfully ignores untagged fields or structs", t, func() {
		p := PartialTopLevel{
			SeenField:   1,
			IgnoreField: "yes",

			IngoreStruct: PartialNestedLevel{
				IgnoreField: true,
				SeenField:   1 * time.Second,
			},
		}

		js := []byte(`
			{
				"seen": 6,
				"ignore": false,

				"nested_seen": {
					"seen": "5s",
					"ignore": "anything"
				},

				"nested_ignore": {
					"seen": "anything",
					"ignore": "anything"
				}
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = Assign(&p, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		exp := PartialTopLevel{
			SeenField:   6,
			IgnoreField: "yes",

			SeenStruct: &PartialNestedLevel{
				SeenField:   5 * time.Second,
				IgnoreField: false,
			},

			IngoreStruct: PartialNestedLevel{
				IgnoreField: true,
				SeenField:   1 * time.Second,
			},
		}
		So(reflect.DeepEqual(p, exp), ShouldBeTrue)
	})

	Convey("successfully set initialized fields to nil", t, func() {
		i := 1
		p := NilTopLevel{
			Int:    &i,
			Nested: &NilNested{Dummy: 1},
		}

		js := []byte(`
			{
				"nested": null,
				"int": null
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = Assign(&p, assignments, map[string]AssignReader{})
		So(err, ShouldBeNil)

		exp := NilTopLevel{}
		So(reflect.DeepEqual(p, exp), ShouldBeTrue)
	})

	Convey("skips assignments for - fields or structs", t, func() {
		nestedInt := 11
		nestedInt64 := int64(22)
		nestedBool := false
		nestedString := "xx"
		nestedFloat64 := 0.1122
		nestedDuration := 11 * time.Second

		nl := NestedLevel{
			NestedInt:      1,
			NestedInt64:    2,
			NestedBool:     false,
			NestedString:   "x",
			NestedFloat64:  0.12,
			NestedDuration: 1 * time.Second,

			NestedIntPtr:      &nestedInt,
			NestedInt64Ptr:    &nestedInt64,
			NestedBoolPtr:     &nestedBool,
			NestedFloat64Ptr:  &nestedFloat64,
			NestedStringPtr:   &nestedString,
			NestedDurationPtr: &nestedDuration,
		}

		nestedIntPtr := 1111
		nestedInt64Ptr := int64(2222)
		nestedBoolPtr := false
		nestedStringPtr := "xxxx"
		nestedFloat64Ptr := 0.11112222
		nestedDurationPtr := 1111 * time.Second

		nlP := &NestedLevel{
			NestedInt:      111,
			NestedInt64:    222,
			NestedBool:     false,
			NestedString:   "xxx",
			NestedFloat64:  0.111222,
			NestedDuration: 111 * time.Second,

			NestedIntPtr:      &nestedIntPtr,
			NestedInt64Ptr:    &nestedInt64Ptr,
			NestedBoolPtr:     &nestedBoolPtr,
			NestedFloat64Ptr:  &nestedFloat64Ptr,
			NestedStringPtr:   &nestedStringPtr,
			NestedDurationPtr: &nestedDurationPtr,
		}

		topInt := -1
		topInt64 := int64(-1)
		topBool := false
		topString := "-x"
		topFloat64 := -0.1122
		topDuration := 10 * time.Second

		tl := TopLevel{
			TopInt:      -2,
			TopInt64:    -44,
			TopBool:     false,
			TopString:   "-y",
			TopFloat64:  -0.12,
			TopDuration: 12 * time.Second,

			TopIntPtr:      &topInt,
			TopInt64Ptr:    &topInt64,
			TopBoolPtr:     &topBool,
			TopFloat64Ptr:  &topFloat64,
			TopStringPtr:   &topString,
			TopDurationPtr: &topDuration,

			Nested:    nl,
			NestedPtr: nlP,
		}

		js := []byte(`
			{
			  "top_int": 5,
			  "top_int64": 6,
			  "top_bool": true,
			  "top_string": "m",
			  "top_float64": 0.56,
			  "top_duration": "5s",
			  "top_int_ptr": 55,
			  "top_int64_ptr": 66,
			  "top_bool_ptr": true,
			  "top_string_ptr": "mm",
			  "top_float64_ptr": 0.5656,
			  "top_duration_ptr": "55s",
			  "nested": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  },
			  "nested_ptr": {
			    "nested_int": 5,
			    "nested_int64": 6,
			    "nested_bool": true,
			    "nested_string": "m",
			    "nested_float64": 0.56,
			    "nested_duration": "5s",
			    "nested_int_ptr": 55,
			    "nested_int64_ptr": 66,
			    "nested_bool_ptr": true,
			    "nested_string_ptr": "mm",
			    "nested_float64_ptr": 0.5656,
			    "nested_duration_ptr": "55s"
			  }
			}
		`)

		//make deep copy
		asBytes, err := xml.Marshal(tl)
		So(err, ShouldBeNil)
		exp := TopLevel{}
		err = xml.Unmarshal(asBytes, &exp)
		So(err, ShouldBeNil)
		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)

		assignments := map[string]interface{}{}
		err = json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = Assign(&tl, assignments, map[string]AssignReader{"duration": DurationAssignReader})
		So(err, ShouldBeNil)

		So(reflect.DeepEqual(tl, exp), ShouldBeTrue)
	})

	Convey("successfully assigns whole struct to a field", t, func() {
		s := StructAssign{}
		js := []byte(`
			{
				"d": "5s",
				"dp": "10s"
			}
		`)

		assignments := map[string]interface{}{}
		err := json.Unmarshal(js, &assignments)
		So(err, ShouldBeNil)
		err = Assign(&s, assignments, map[string]AssignReader{"jd": JSONDurationAssignReader})
		So(err, ShouldBeNil)

		exp := StructAssign{D: JSONDuration{Duration: 5 * time.Second}, DPtr: &JSONDuration{Duration: 10 * time.Second}}
		So(reflect.DeepEqual(s, exp), ShouldBeTrue)

	})
}
