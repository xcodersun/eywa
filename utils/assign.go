package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var TagName = "assign"

func assign(d interface{}, s map[string]interface{}, rds map[string]AssignReader, withForce bool) error {
	if len(s) == 0 {
		return nil
	}

	dv := reflect.ValueOf(d)
	if dv.Kind() != reflect.Ptr || dv.Type().Elem().Kind() != reflect.Struct {
		return errors.New("expecting assignee to be a pointer to struct")
	}

	if dv.IsNil() {
		return errors.New("expecting non nil pointer")
	}

	dt := dv.Elem().Type()

	for key, value := range s {
		var fieldName string
		var tagName string
		var rdName string

		for i := 0; i < dt.NumField(); i++ {
			field := dt.Field(i)
			tag := field.Tag.Get(TagName)
			ts := strings.Split(tag, ";")
			tagName = ts[0]
			if tagName == key {
				if len(ts) > 1 {
					rdName = ts[1]
				}
				if withForce {
					fieldName = field.Name
				} else {
					if len(ts) <= 2 || ts[2] != "-" {
						fieldName = field.Name
					}
				}
				break
			}
		}

		if len(fieldName) == 0 {
			continue
		}

		field := dv.Elem().FieldByName(fieldName)
		if conV, ok := value.(map[string]interface{}); ok {
			if field.Kind() != reflect.Struct && (field.Kind() != reflect.Ptr || field.Type().Elem().Kind() != reflect.Struct) {
				return errors.New(fmt.Sprintf("expecting field: %s in %s to be a struct or a pointer to struct", fieldName, dt.Name()))
			}

			var err error
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					val := reflect.New(field.Type().Elem())
					field.Set(val)
				}
				err = assign(field.Interface(), conV, rds, withForce)
			} else {
				err = assign(field.Addr().Interface(), conV, rds, withForce)
			}

			if err != nil {
				return err
			}
		} else {
			if value == nil {
				val := reflect.Zero(field.Type())
				field.Set(val)
			} else {
				var rd AssignReader
				var found bool
				if rds != nil {
					rd, found = rds[rdName]
				}
				if !found {
					if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
						return errors.New(fmt.Sprintf("expecting field: %s in %s not to be a struct, neither a pointer to struct", fieldName, dt.Name()))
					}

					var tp string
					if field.Kind() == reflect.Ptr {
						tp = field.Type().Elem().Name()
					} else {
						tp = field.Type().Name()
					}
					switch tp {
					case "int":
						rd = IntAssignReader
					case "int64":
						rd = Int64AssignReader
					case "string":
						rd = StringAssignReader
					case "bool":
						rd = BoolAssignReader
					case "float64":
						rd = Float64AssignReader
					default:
						return errors.New(fmt.Sprintf("don't know how to assign field: %s of type %s in %s", fieldName, field.Type(), dt.Name()))
					}

				}

				v, err := rd(value, field.Kind() == reflect.Ptr)
				if err != nil {
					return err
				}

				if field.CanSet() {
					field.Set(v)
				} else {
					return errors.New(fmt.Sprintf("field: %s in %s cannot be assigned", fieldName, dt.Name()))
				}

			}
		}
	}

	return nil
}

func ForceAssign(d interface{}, s map[string]interface{}, rds map[string]AssignReader) error {
	return assign(d, s, rds, true)
}

func Assign(d interface{}, s map[string]interface{}, rds map[string]AssignReader) error {
	return assign(d, s, rds, false)
}

type AssignReader func(interface{}, bool) (reflect.Value, error)

var BoolAssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	if b, ok := v.(bool); !ok {
		return reflect.Value{}, errors.New("interface is not a bool")
	} else {
		if isPtr {
			return reflect.ValueOf(&b), nil
		} else {
			return reflect.ValueOf(b), nil
		}
	}
}

var IntAssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	if i, ok := v.(int); !ok {
		if f, ok := v.(float64); !ok {
			return reflect.Value{}, errors.New("interface is not an int")
		} else {
			if isPtr {
				pf := int(f)
				return reflect.ValueOf(&pf), nil
			} else {
				return reflect.ValueOf(int(f)), nil
			}
		}
	} else {
		if isPtr {
			return reflect.ValueOf(&i), nil
		} else {
			return reflect.ValueOf(i), nil
		}
	}
}

var Int64AssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	if i, ok := v.(int64); !ok {
		if f, ok := v.(float64); !ok {
			return reflect.Value{}, errors.New("interface is not an int64")
		} else {
			if isPtr {
				pf := int64(f)
				return reflect.ValueOf(&pf), nil
			} else {
				return reflect.ValueOf(int64(f)), nil
			}
		}
	} else {
		if isPtr {
			return reflect.ValueOf(&i), nil
		} else {
			return reflect.ValueOf(i), nil
		}
	}
}

var Float64AssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	if f, ok := v.(float64); !ok {
		return reflect.Value{}, errors.New("interface is not a float64")
	} else {
		if isPtr {
			return reflect.ValueOf(&f), nil
		} else {
			return reflect.ValueOf(f), nil
		}
	}
}

var StringAssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	if s, ok := v.(string); !ok {
		return reflect.Value{}, errors.New("interface is not a string")
	} else {
		if isPtr {
			return reflect.ValueOf(&s), nil
		} else {
			return reflect.ValueOf(s), nil
		}
	}
}

var DurationAssignReader = func(v interface{}, isPtr bool) (reflect.Value, error) {
	var d time.Duration
	var err error
	if f, ok := v.(float64); ok {
		d = time.Duration(int64(f))
	} else if i, ok := v.(int64); ok {
		d = time.Duration(i)
	} else if s, ok := v.(string); ok {
		d, err = time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, err
		}
	} else {
		return reflect.Value{}, errors.New("interface is not a duration")
	}

	if isPtr {
		return reflect.ValueOf(&d), nil
	} else {
		return reflect.ValueOf(d), nil
	}
}
