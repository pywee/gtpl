# gtpl

<br />

##### 使用必读

gtpl is a HTML template engine for golang

gtpl 是一个 go 语言模板引擎，它能以极快的速度进行模板语法分析。相比 go 语言官方库 html/template，gtpl 的语法有着简练、灵活、易用的特点。

gtpl 最终的目的就是完全替代 go 语言官方过于复杂的 html/template 渲染包，让模板调用变得更加灵活，更加适合理解，从而在很大程度上节约开发者的时间。

<hr />

##### 与 php 模板引擎不同之处

gtpl 与 那些基于 php 的模板引擎完全不同。

php 模板引擎直接将标签语法翻译为 php 代码，保存为 php 文件之后再重新从 php 引入，然后就可以直接运行了。而 gtpl 相当于一个超轻量级的编程语言，它实现了自己的语义分析，因此，从处理方式上有着本质区别。每一个标签的处理都需要使用递归逻辑，并且需要在递归过程中保持上下文。笔者为此付出了大量的工作。

目前，它仍然存在一些不足，比如对于 if 语句的解析存在一些未知问题需要解决，我会在接下来的版本中将会重点关注并修复它们。

gtpl 目前已经可以使用简单的 if 条件表达式，如算术运算、逻辑运算，它对运算符的优先级处理是与 go 语言一样的：
```html
{!if id==1+2-(10*5/(11))&&((1+11*(5-10))>=1)}
    123
{;elseif id!=pid}
    {!if title=="hello"}
        ...
    {/if}
{;else}
    789
{/if}
```

##### 我将在下一个版本着重处理这个问题。

<br />
<hr />

以下是 gtpl 常用的标签调用示例 (Example):

##### example a.go

```go

import(
    "github.com/pywee/gtpl/handle/templates"
    "log"
    "net/http"
)

// data
// example struct
type Data struct {
    Id      int64
    WebTitle string
    Users   []*User
    Winner  *Demo
}

// example struct
type Demo struct {
	Str string
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
    // 组装要渲染到 HTML 的数据
    data := &Data{
        Id:    1,
        WebTitle: "WebTitle",
        Winner: &Demo{
            Str: "pywee",
        },
        Users: []*User{
            {
                Id:       1,
                UserName: "Jim",
                MyList: []*UserCustom{
                    {PhoneNumber: "1"},
                    {PhoneNumber: "2"},
                },
            },
            {
                Id:       2,
                UserName: "Lucy",
                MyList: []*UserCustom{
                    {PhoneNumber: "1"},
                    {PhoneNumber: "2"},
                },
            },
        },
    }

    // http server example
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {

		// 实例化gtpl (instantiation gtpl)
		p := templates.NewParser()

		// 引入模板文件，将 data 数据匹配模板中的调用
		re, err := p.ParseFile("example/index.html", data)
		if err != nil {
			panic(err)
		}

		// print string
		// fmt.Println(re.String())

		// 输出到屏幕，它的作用相当于 template.Excute(w, "%s", data)
		re.Fprint(w)
	})
}

```

<hr />

针对以上代码，你可以用这种方式在html文件中获取数据(For the above code, you can get data in HTML file in this way)。

#### example a.html:

```
<html>
....

// 你可以在这里直接访问data结构体下的字段 
// (You can call the fields under the structure in this way)

// 访问结构体中的 Data.Id, Data.Title  在 html 中调用时请注意使用小写，用下划线替代结构体中的驼峰字段
// to call Id and WebTitle of struct Data
{:id}
{:web_title}

// 访问 Data 结构体中的 Other.Demo.Str
// to call Other.Demo.Str of struct Data
{:winner.str} 

// 在访问字段的同时，执行内置函数
{:trim(replace(title, "i am", "xx", -1), "xx")}

// 对于列表处理，例如循环 data 下的 users 数组，用如下方式
{:id}
{:winner.str}  // 访问 Data 下的 Winner结构体，再访问该结构体下的 Str 字段
{!list:users}
    {!if id==(1*1+1-1-1)}
        man: {:user_name}
    {;elseif 1==21}
        women: {:user_name}
    {;else}
        {!list:my_list}
            {:phone_number}
        {/list}
    {/if}
{/list}
</html>
```

针对以上标签代码，```{!list:users}``` 也可以明确写成 ```{!list:data.users}```，但底层处理方法略有不同。
标签中所声明的大多数调用，底层都会像剥大蒜一遍又一遍递归查找、匹配。因此，你应该尽可能使用 ```{!list:users}``` ，即使这样也是以递归的方式处理语法相关的工作，但相比之下，会减少递归的次数。
(You can also use ```{!list:data.users}``` to ergodic slice users, but it is not good for performance.)


<hr />

*此外，必须注意，出于效率考虑，目前 gtpl 并不支持返回结构体的 tag 来获得结构体指定字段的数据，所以，当你在html中调用结构体中的字段时，需转为下划线 "_"，而结构体中的字段仍然使用驼峰写法。
如调用结构体中的 UserName，在 html 文件中需要使用 {:user_name} 来调用。

gtpl 可支持部分 go 语言内置函数的调用：

```html
len()
tolower()
toupper()
replace()
trim()
ltrim()
rtrim()
trimspace()

// 示例 example
{!list:data}
    {:trim(title, "hello")}
{/list}
```

<hr >

另外，如果传入 ParseFile 函数的数所是一个数组结构，那么你必需在html中首先循环它，如：

##### example
```go
    d2 := []*Data{ // slice
		{
			Id: 1,
			Users: []*User{
				{
					Id:       2,
					UserName: "title",
					MyList: []*UserCustom{
						{PhoneNumber: "1"},
						{PhoneNumber: "2"},
						{PhoneNumber: "3"},
					},
				},
			},
		},
	}

	p := templates.NewParser()
	re, err := p.ParseFile("example/index.html", d2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(re.String())
```

```html
// 正确的调用方式 (right)
{:id}
{!list:data}
    {!list:users}
        {!list:my_list}
            {:a}
        {/list}
    {/list}
{/list}

// 正确的调用方式 (right)
{!list:users}
    {!if id==(1*1+1-1-1)}
        man: {:user_name}
    {;elseif id+1==21}
        women: {:user_name}
    {;else}
        {!list:my_list}
            {:phone_number}
        {/list}
    {/if}
{/list}

// 错误的调用方式 (wrong)
{!list:users}
    {!list:my_list}
        {:a}
    {/list}
{/list}
```

将来会支持更多的函数或自定义函数调用：

<br />
<br />