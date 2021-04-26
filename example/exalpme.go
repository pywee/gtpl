package main

/**
 * Auther pywee
 * Email 702365381@qq.com
 * Date 2021/04/27
 */

import (
	"fmt"
	"gotpl/handle/templates"
)

// example struct
type Data struct {
	Id       int64
	WebTitle string
	Users    []*User
	Winner   *Demo
}

// example struct
type Demo struct {
	Str string
}

type Demo2 struct {
	Name string
}

// example struct
type User struct {
	Id       int
	UserName string
	MyList   []*UserCustom
}

type UserCustom struct {
	PhoneNumber string
}

func main() {

	// create data that you want to render to HTML
	// 组装要输出到 HTML 的数据
	data := &Data{
		Id:       1,
		WebTitle: "I am a title",
		Winner: &Demo{
			Str: "pywee",
		},
		Users: []*User{
			{
				Id:       1,
				UserName: "Jim",
				MyList: []*UserCustom{
					{PhoneNumber: "11"},
					{PhoneNumber: "22"},
				},
			},
			{
				Id:       2,
				UserName: "Lucy",
				MyList: []*UserCustom{
					{PhoneNumber: "33"},
					{PhoneNumber: "44"},
				},
			},
		},
	}

	// data := []*Data{
	// 	{
	// 		Id: 1,
	// 		Users: []*User{
	// 			{
	// 				Id:       2,
	// 				UserName: "title",
	// 				MyList: []*UserCustom{
	// 					{A: "1"},
	// 					{A: "2"},
	// 					{A: "3"},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// example 1

	// 实例化gtpl (instantiation gtpl)
	p := templates.NewParser()

	// 引入模板文件，将 data 数据匹配模板中的调用
	re, err := p.ParseFile("example/index.html", data)
	if err != nil {
		panic(err)
	}

	// print string
	fmt.Println(re.String())

	// example 2

	// http server example
	// http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {

	// 	// 实例化gtpl (instantiation gtpl)
	// 	p := templates.NewParser()

	// 	// 引入模板文件，将 data 数据匹配模板中的调用
	// 	re, err := p.ParseFile("example/index.html", data)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// print string
	// 	// fmt.Println(re.String())

	// 	// 输出到屏幕，它的作用相当于 template.Excute(w, "%s", data)
	// 	re.Fprint(w)
	// })

	// log.Fatal(http.ListenAndServe(":8080", nil))
}
