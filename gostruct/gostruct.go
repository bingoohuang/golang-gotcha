package gostruct

import "reflect"

type Builder struct {
	fields []reflect.StructField
}

func New() *Builder {
	return &Builder{}
}

func (b *Builder) AddField(name string, type_ reflect.Type) *Builder {
	b.fields = append(b.fields, reflect.StructField{Name: name, Type: type_})
	return b
}

func (b *Builder) AddString(name string) *Builder { return b.AddField(name, reflect.TypeOf("")) }
func (b *Builder) AddBool(name string) *Builder   { return b.AddField(name, reflect.TypeOf(true)) }
func (b *Builder) AddInt64(name string) *Builder  { return b.AddField(name, reflect.TypeOf(int64(0))) }
func (b *Builder) AddFloat64(name string) *Builder {
	return b.AddField(name, reflect.TypeOf(float64(0)))
}

func (b *Builder) Build() Struct {
	strct := reflect.StructOf(b.fields)

	index := make(map[string]int)
	for i := 0; i < strct.NumField(); i++ {
		index[strct.Field(i).Name] = i
	}

	return Struct{strct: strct, index: index}
}

type Struct struct {
	strct reflect.Type
	index map[string]int
}

func (s *Struct) New() *Instance {
	instance := reflect.New(s.strct).Elem()
	return &Instance{internal: instance, index: s.index}
}

type Instance struct {
	internal reflect.Value
	index    map[string]int
}

func (i *Instance) Field(name string) reflect.Value       { return i.internal.Field(i.index[name]) }
func (i *Instance) SetString(name, value string)          { i.Field(name).SetString(value) }
func (i *Instance) SetBool(name string, value bool)       { i.Field(name).SetBool(value) }
func (i *Instance) SetInt64(name string, value int64)     { i.Field(name).SetInt(value) }
func (i *Instance) SetFloat64(name string, value float64) { i.Field(name).SetFloat(value) }
func (i *Instance) Interface() interface{}                { return i.internal.Interface() }
func (i *Instance) Addr() interface{}                     { return i.internal.Addr().Interface() }
