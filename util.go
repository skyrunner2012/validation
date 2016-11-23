// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
    "fmt"
    "reflect"
    "regexp"
    "strconv"
    "strings"
    "bytes"
)

const (
    // ValidTag struct tag
    ValidTag = "valid"
    // valid里面可以有多个，分号隔开，vdesc里面也可以有多个分号隔开，如果是多个需要一一对应匹配或者vdesc只提供一个，就是一对多也行
    ValidErrDescTag = "vdesc"
)

var (
    // key: function name
    // value: the number of parameters
    funcs = make(Funcs)

    // doesn't belong to validation functions
    unFuncs = map[string]bool{
        "Clear":     true,
        "HasErrors": true,
        "ErrorMap":  true,
        "Error":     true,
        "apply":     true,
        "Check":     true,
        "Valid":     true,
        "NoMatch":   true,
    }
)

func init() {
    v := &Validation{}
    t := reflect.TypeOf(v)
    for i := 0; i < t.NumMethod(); i++ {
        m := t.Method(i)
        if !unFuncs[m.Name] {
            funcs[m.Name] = m.Func
        }
    }
}

// CustomFunc is for custom validate function
type CustomFunc func(v *Validation, obj interface{}, key string)

// AddCustomFunc Add a custom function to validation
// The name can not be:
//   Clear
//   HasErrors
//   ErrorMap
//   Error
//   Check
//   Valid
//   NoMatch
// If the name is same with exists function, it will replace the origin valid function
func AddCustomFunc(name string, f CustomFunc) error {
    if unFuncs[name] {
        return fmt.Errorf("invalid function name: %s", name)
    }

    funcs[name] = reflect.ValueOf(f)
    return nil
}

// ValidFunc Valid function type
type ValidFunc struct {
    Name   string
    ErrMsg string
    Params []interface{}
}

// Funcs Validate function map
type Funcs map[string]reflect.Value

// Call validate values with named type string
func (f Funcs) Call(name string, params ...interface{}) (result []reflect.Value, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("%v", r)
        }
    }()
    if _, ok := f[name]; !ok {
        err = fmt.Errorf("%s does not exist", name)
        return
    }
    if len(params) != f[name].Type().NumIn() {
        err = fmt.Errorf("The number of params is not adapted")
        return
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }
    result = f[name].Call(in)
    return
}

func isStruct(t reflect.Type) bool {
    return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
    return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

// 增加对错误描述tag的处理
func getValidFuncs(f reflect.StructField) (vfs []ValidFunc, err error) {
    tag := strings.TrimSpace(f.Tag.Get(ValidTag))
    if len(tag) == 0 {
        return
    }
    errDescTag := strings.TrimSpace(f.Tag.Get(ValidErrDescTag))
    if vfs, tag, errDescTag, err = getRegFuncs(tag, errDescTag, f.Name); err != nil {
        return
    }
    fs := strings.Split(tag, ";")
    var fsDesc []string
    isFsArray := false
    if strings.Index(errDescTag, ";") >= 0 {
        fsDesc = strings.Split(errDescTag, ";")
        isFsArray = true
    }
    for idx, vfunc := range fs {
        var vf ValidFunc
        if len(vfunc) == 0 {
            continue
        }
        var errDesc string
        if !isFsArray {
            errDesc = errDescTag
        } else {
            if len(fsDesc) > idx {
                errDesc = fsDesc[idx]
            }
        }

        vf, err = parseFunc(vfunc, errDesc, f.Name)
        if err != nil {
            return
        }

        vfs = append(vfs, vf)
    }
    return
}

// Get Match function
// May be get NoMatch function in the future
func getRegFuncs(tag, errDescTag, key string) (vfs []ValidFunc, newTag string, newErrDescTag string, err error) {
    index := strings.Index(tag, "Match(/")
    if index == -1 {
        newTag = tag
        newErrDescTag = errDescTag
        return
    }
    end := strings.LastIndex(tag, "/)")
    if end < index {
        err = fmt.Errorf("invalid Match function")
        return
    }
    reg, err := regexp.Compile(tag[index + len("Match(/") : end])
    if err != nil {
        return
    }
    regDescTag := ""
    if len(errDescTag) > 0 {
        if strings.Index(errDescTag, ";") < 0 {
            regDescTag = errDescTag
        } else {
            errDescTags := strings.Split(errDescTag, ";")
            errDestTagBuffer := bytes.Buffer{}
            for idx, aTag := range strings.Split(tag, ";") {
                if strings.Index(strings.TrimSpace(aTag), "Match(/") >= 0 {
                    if len(errDescTags) > idx {
                        regDescTag = errDescTags[idx]
                    }
                } else {
                    if idx > 0 && len(errDescTags) > idx {
                        errDestTagBuffer.WriteString(";")
                    }
                    if len(errDescTags) > idx {
                        errDestTagBuffer.WriteString(errDescTags[idx])
                    }
                }
            }
            newErrDescTag = errDestTagBuffer.String()
        }
    }
    vfs = []ValidFunc{{"Match", regDescTag, []interface{}{reg, key + ".Match"}}}
    newTag = strings.TrimSpace(tag[:index]) + strings.TrimSpace(tag[end + len("/)"):])
    return
}

func parseFunc(vfunc, errDesc, key string) (v ValidFunc, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("%v", r)
        }
    }()

    vfunc = strings.TrimSpace(vfunc)
    start := strings.Index(vfunc, "(")
    var num int

    // doesn't need parameter valid function
    if start == -1 {
        if num, err = numIn(vfunc); err != nil {
            return
        }
        if num != 0 {
            err = fmt.Errorf("%s require %d parameters", vfunc, num)
            return
        }
        v = ValidFunc{vfunc, errDesc, []interface{}{key + "." + vfunc}}
        return
    }

    end := strings.Index(vfunc, ")")
    if end == -1 {
        err = fmt.Errorf("invalid valid function")
        return
    }

    name := strings.TrimSpace(vfunc[:start])
    if num, err = numIn(name); err != nil {
        return
    }

    params := strings.Split(vfunc[start + 1:end], ",")
    // the num of param must be equal
    if num != len(params) {
        err = fmt.Errorf("%s require %d parameters", name, num)
        return
    }

    tParams, err := trim(name, key + "." + name, params)
    if err != nil {
        return
    }
    v = ValidFunc{name, errDesc, tParams}
    return
}

func numIn(name string) (num int, err error) {
    fn, ok := funcs[name]
    if !ok {
        err = fmt.Errorf("doesn't exsits %s valid function", name)
        return
    }
    // sub *Validation obj and key and errDesc
    num = fn.Type().NumIn() - 4
    return
}

func trim(name, key string, s []string) (ts []interface{}, err error) {
    ts = make([]interface{}, len(s), len(s) + 1)
    fn, ok := funcs[name]
    if !ok {
        err = fmt.Errorf("doesn't exsits %s valid function", name)
        return
    }
    for i := 0; i < len(s); i++ {
        var param interface{}
        // skip *Validation and obj params
        if param, err = parseParam(fn.Type().In(i + 2), strings.TrimSpace(s[i])); err != nil {
            return
        }
        ts[i] = param
    }
    ts = append(ts, key)
    return
}

// modify the parameters's type to adapt the function input parameters' type
func parseParam(t reflect.Type, s string) (i interface{}, err error) {
    switch t.Kind() {
    case reflect.Int:
        i, err = strconv.Atoi(s)
    case reflect.String:
        i = s
    case reflect.Ptr:
        if t.Elem().String() != "regexp.Regexp" {
            err = fmt.Errorf("does not support %s", t.Elem().String())
            return
        }
        i, err = regexp.Compile(s)
    default:
        err = fmt.Errorf("does not support %s", t.Kind().String())
    }
    return
}

func mergeParam(v *Validation, obj interface{}, errDesc string, params []interface{}) []interface{} {
    return append(append([]interface{}{v, obj}, params...), errDesc)
}
