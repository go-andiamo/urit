package urit

import "errors"

type PathVar struct {
	Name          string
	NamedPosition int
	Position      int
	Value         string
}

// PathVars is the interface used to pass path vars into a template and returned from a template after extracting
//
// Use either Positional or Named to create a new PathVars
type PathVars interface {
	GetPositional(position int) (string, bool)
	GetNamed(name string, position int) (string, bool)
	GetNamedFirst(name string) (string, bool)
	GetNamedLast(name string) (string, bool)
	Get(idents ...interface{}) (string, bool)
	GetAll() []PathVar
	Len() int
	Clear()
	// VarsType returns the path vars type (Positions or Names)
	VarsType() PathVarsType
	AddNamedValue(name string, val string) error
	AddPositionalValue(val string) error
}

type pathVars struct {
	named    map[string][]PathVar
	all      []PathVar
	varsType PathVarsType
}

func newPathVars(varsType PathVarsType) PathVars {
	return &pathVars{
		named:    map[string][]PathVar{},
		all:      make([]PathVar, 0),
		varsType: varsType,
	}
}

func (pvs *pathVars) GetPositional(position int) (string, bool) {
	if position < 0 && (len(pvs.all)+position) >= 0 {
		return pvs.all[len(pvs.all)+position].Value, true
	} else if position >= 0 && position < len(pvs.all) {
		return pvs.all[position].Value, true
	}
	return "", false
}

func (pvs *pathVars) GetNamed(name string, position int) (string, bool) {
	if vs, ok := pvs.named[name]; ok {
		if position < 0 && (len(vs)+position) >= 0 {
			return vs[len(vs)+position].Value, true
		} else if position >= 0 && position < len(vs) {
			return vs[position].Value, true
		}
	}
	return "", false
}

func (pvs *pathVars) GetNamedFirst(name string) (string, bool) {
	if vs, ok := pvs.named[name]; ok && len(vs) > 0 {
		return vs[0].Value, true
	}
	return "", false
}

func (pvs *pathVars) GetNamedLast(name string) (string, bool) {
	if vs, ok := pvs.named[name]; ok && len(vs) > 0 {
		return vs[len(vs)-1].Value, true
	}
	return "", false
}

func (pvs *pathVars) Get(idents ...interface{}) (string, bool) {
	firstInt, isFirstInt := idents[0].(int)
	firstStr, isFirstStr := idents[0].(string)
	if len(idents) < 1 || len(idents) > 2 ||
		(isFirstInt && len(idents) > 1) {
		return "", false
	}
	if isFirstInt {
		return pvs.GetPositional(firstInt)
	} else if isFirstStr {
		if len(idents) > 1 {
			if secondInt, ok := idents[1].(int); ok {
				return pvs.GetNamed(firstStr, secondInt)
			}
		} else {
			return pvs.GetNamedFirst(firstStr)
		}
	}
	return "", false
}

func (pvs *pathVars) GetAll() []PathVar {
	return pvs.all
}

func (pvs *pathVars) Len() int {
	return len(pvs.all)
}

func (pvs *pathVars) Clear() {
	pvs.named = map[string][]PathVar{}
	pvs.all = make([]PathVar, 0)
}

// VarsType returns the path vars type (Positions or Names)
func (pvs *pathVars) VarsType() PathVarsType {
	return pvs.varsType
}

func (pvs *pathVars) AddNamedValue(name string, val string) error {
	if pvs.varsType != Names {
		return errors.New("cannot add named var to non-names vars")
	}
	np := len(pvs.named[name])
	v := PathVar{
		Name:          name,
		NamedPosition: np,
		Position:      len(pvs.all),
		Value:         val,
	}
	pvs.named[name] = append(pvs.named[name], v)
	pvs.all = append(pvs.all, v)
	return nil
}

func (pvs *pathVars) AddPositionalValue(val string) error {
	if pvs.varsType != Positions {
		return errors.New("cannot add positional var to non-positionals vars")
	}
	pvs.all = append(pvs.all, PathVar{
		Position: len(pvs.all),
		Value:    val,
	})
	return nil
}

// Positional creates a positional PathVars from the values supplied
func Positional(values ...string) PathVars {
	result := newPathVars(Positions)
	for _, val := range values {
		_ = result.AddPositionalValue(val)
	}
	return result
}

// Named creates a named PathVars from the name and value pairs supplied
//
// Note: If there is not a value for each name - this function panics.
// So ensure that the number of varargs passed is an even number
func Named(namesAndValues ...string) PathVars {
	if len(namesAndValues)%2 != 0 {
		panic("must be a value for each name")
	}
	result := newPathVars(Names)
	for i := 0; i < len(namesAndValues); i += 2 {
		_ = result.AddNamedValue(namesAndValues[i], namesAndValues[i+1])
	}
	return result
}
