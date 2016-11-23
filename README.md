# 使用说明


## 安装

```go
go get github.com/skyrunner2012/validation
```

## 例子

```go
package main

import (
    "validation"
    "fmt"
)

func main() {
    df := demoForm{Name:"name", Amount: 1, Amount2:100, Password:"a", Password2:"dd"}
    validForm(&df)
}

type demoForm struct {
    Name      string        `json:"name" valid:"Required;" vdesc:"名称不能为空" required:"true" description:"名称"`
    Amount    int           `json:"amount" valid:"Required;Range(1, 140)" vdesc:"金额不能为空" required:"true" description:"金额"`
    Amount2   int           `json:"amount" valid:"Required;Range(1, 140)" vdesc:"金额不能为空;无效金额，范围为1到140之间" required:"true" description:"金额"`
    Password  string      `json:"password" valid:"Required" vdesc:"密码不能为空" required:"true" description:"密码"`
    Password2 string      `json:"password" valid:"Required;Match(/^(test)?\\w*@;com$/)" vdesc:"密码不能为空;密码不符合密码规范" required:"true" description:"密码"`
    Memo      string        `json:"memo" valid:"Required" required:"true" description:"备注"`
}

func validForm(form interface{}) {
    valid := validation.Validation{}
    err := valid.Valid(form); if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Everything is ok")
    }
}
```

## 解释说明

* tag: `valid` 对应验证函数，最后有列出支持的验证函数，可以使用 `;` 号隔开，配置多个
* tag: `vdesc` 和valid标签配合使用，如果没有配置，则会使用系统默认值（目前默认值是中文版的），也可以使用 `;' 隔开，定义不同的错误描述
* 支持 `valid` 和 `vdesc` 一对一，也支持 `valid` 和 `vdesc` 多对一

## 支持的验证函数列表

```go
	Required
	Min(min int)
	Max(max int)
	Range(min, max int)
	MinSize(min int)
	MaxSize(max int)
	Length(length int)
	Alpha
	Numeric
	AlphaNumeric
	Match(pattern string)
	AlphaDash
	Email
	IP
	Base64
	Mobile
	Tel
	Phone
	ZipCode
```


## LICENSE

BSD License http://creativecommons.org/licenses/BSD/
