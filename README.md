## CLI命令行实用程序开发基础

#### 概述

CLI（Command Line Interface）实用程序是Linux下应用开发的基础。正确的编写命令行程序让应用与操作系统融为一体，通过shell或script使得应用获得最大的灵活性与开发效率。Linux提供了cat、ls、copy等命令与操作系统交互；go语言提供一组实用程序完成从编码、编译、库管理、产品发布全过程支持；容器服务如docker、k8s提供了大量实用程序支撑云服务的开发、部署、监控、访问等管理任务；git、npm等都是大家比较熟悉的工具。尽管操作系统与应用系统服务可视化、图形化，但在开发领域，CLI在编程、调试、运维、管理中提供了图形化程序不可替代的灵活性与效率。

***

#### 目标

参阅[Selpg命令行程序设计逻辑](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html)，实现selpg命令行程序，满足选择文档页码进行打印的要求。

#### 参数规则

如果程序可以根据其输入或用户的首选参数有不同的行为，则应将它编写为接受名为 *选项*的命令行参数，这些参数允许用户指定什么行为将用于这个调用。

作为选项的命令行参数由前缀“-”（连字符）标识。另一类参数是那些不是选项的参数，也就是说，它们并不真正更改程序的行为，而更象是数据名称。

Linux 实用程序语法图看起来如下：

```
$ command mandatory_opts [ optional_opts ] [ other_args ]
```

其中：

·	`command` 是命令本身的名称

·	`mandatory_opts` 是为使命令正常工作必须出现的选项列表

·	`optional_opts` 是可指定也可不指定的选项列表，这由用户来选择；但是，其中一些参数可能是互斥的，如同 selpg 的“-f”和“-l”选项的情况（详情见下文）

·	`other_args` 是命令要处理的其它参数的列表；这可以是任何东西，而不仅仅是文件名

下面介绍selpg的**具体参数规则**。

***

·	“-sNumber”和“-eNumber”，eg：

```
$ selpg -s10 -e20...
```

`selpg` 要求用户输入的前两个参数是需要打印文档的起始页和终止页，即

startPage <= page <= endPage    的页码会被打印。

如上例中，指定打印第10页到第20页。`selpg`会对所给进行合理性检查；换句话说，它会检查两个数字是否为有效的正整数以及结束页是否不小于起始页，否则输出错误信息。

***

·	“-lNumber”和“-f”可选选项：

selpg 可以处理两种输入文本：

1. 文本的每一页有固定的行数。这是缺省类型，因此不必给出选项进行说明，如有需要，则用“-f”选项进行说明。比如以下命令：

```
$ selpg -s10 -e20 -l60...
```

表示需要打印的文档的每一页固定为60行。

2. 文本的每一页没有固定的函数，通过换页符'\f'来定界。该格式与“每页行数固定”格式相比的好处在于，当每页的行数有很大不同而且文件有很多页时，该格式可以节省磁盘空间。在含有文本的行后面，类型 2 的页只需要一个字符 ― 换页 ― 就可以表示该页的结束。打印机会识别换页符并自动根据在新的页开始新行所需的行数移动打印头。比如以下命令：

```
$ selpg -s10 -e20 -f...
```

该命令告诉 selpg 在输入中寻找换页符，并将其作为页定界符处理。

注：如果既没有指定”-lNumber“也没有指定”-f“，则默认以”-l72“作为参数。之所以用72作为缺省值，是因为打印机一般都以72作为页长度。

***

·	”-dDestination“可选选项：

selpg 还允许用户使用“-dDestination”选项将选定的页直接发送至打印机。这里，“Destination”应该是 lp 命令“-d”选项。如果当前有打印机连接至该目的地并且是启用的，则打印机应打印该输出。比如以下命令：

```
$ selpg -s10 -e20 -dDestination -source
```

该命令将source文件选定的页作为打印作业发送至Destination 打印目的地。由于实验中我们没有实体打印机，直接将选定的内容打印出来。

***

#### 代码实现

1.基础知识

使用os，flag包，最简单处理参数：

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    for i, a := range os.Args[1:] {
        fmt.Printf("Argument %d is %s\n", i+1, a)
    }

}
```

使用flag包。flag包提供了一系列解析命令行参数的功能接口，主要步骤有：定义flag参数；调用flag.Parse()解析命令行参数到定义的flag；调用Parse解析，就可以直接使用flag本身。示例：

```go
package main

import (
    "flag" 
    "fmt"
)

func main() {
    var port int
    flag.IntVar(&port, "p", 8000, "specify port to use.  defaults to 8000.")
    flag.Parse()

    fmt.Printf("port = %d\n", port)
    fmt.Printf("other args: %+v\n", flag.Args())
}
```

2.定义保存参数的结构体。

selpg所需参数有必须的开始页码-s以及结束页码-e，可选的输入文件名、自定页长-l、遇换页符换页-f和输出地址。其中自定页长和遇换页符换页两个选项是互斥的，不能同时使用。

```go
type selpgArgs struct {
    startPage   int
    endPage     int
    inputFile  string
    destFile   string
    pageLength int
    pageType    bool //-f or not
}
```

用pageType来表示分页方法，pageType为true代表输入了参数”-f“，采用分页符作为页码界线；否则表示输入了参数”-l“或没有输入参数。

3.使用pflag包提供的方法，给出命令行规范:

```go
func GetArgs(args *selpgArgs) {
    pflag.IntVarP(&(args.startPage), "startPage", "s", -1, "the start page of a text")
    pflag.IntVarP(&(args.endPage), "endPage", "e", -1, "the end page of a text")
    pflag.IntVarP(&(args.pageLength), "pageLength", "l", 72, "the page length of a text")
    pflag.StringVarP(&(args.destFile), "destFile", "d", "", "the destination file")
    pflag.BoolVarP(&(args.pageType), "pageType", "f", false, "the page type: paging by \\f")
    pflag.Parse()

    argLeft := pflag.Args()
    if len(argLeft) > 0 {
        args.inputFile = string(argLeft[0])
    } else {
        args.inputFile = ""
    }
}
```

4.检查参数错误：

```go
func CheckArgs(args *selpgArgs) {

    if (args.startPage == -1) || (args.endPage == -1) {
        fmt.Printf("Usage of selpg:\n")
        fmt.Printf("-d, --destFile string   the destination file (default -1)\n")
        fmt.Printf("-e, --endPage int       the end page of a text\n")
        fmt.Printf("-l, --pageLength int    the page length of a text (default 72)\n")
        fmt.Printf("-f, --pageType          the page type: paging by \\f\n")
        fmt.Printf("-s, --startPage int     the start page of a text (default -1)\n")
        os.Exit(0)
    } else if (args.startPage <= 0) || (args.endPage <= 0) {
        fmt.Fprintf(os.Stderr, "\n[Error]The startPage and endPage can't be negative!\n")
        os.Exit(0)
    } else if args.startPage > args.endPage {
        fmt.Fprintf(os.Stderr, "\n[Error]The startPage can't be bigger than the endPage!\n")
        os.Exit(0)
    } else if (args.pageType == true) && (args.pageLength != 72) {
        fmt.Fprintf(os.Stderr, "\n[Error]The arguments -l and -f are conflicting!\n")
        os.Exit(0)
    } else if args.pageLength <= 0 {
        fmt.Fprintf(os.Stderr, "\n[Error]The pageLength can't be less than 1 !\n")
        os.Exit(0)
    } else {
        pageType := "page length."
        if args.pageType == true {
            pageType = "The end sign \f."
            fmt.Printf("--------------------------------------------------\n")
            fmt.Printf("startPage: %d\nendPage: %d\ninputFile: %s\npageLength: according to end sign\npageType: %s\ndestFile: %s\n", args.startPage, args.endPage, args.inputFile, pageType, args.destFile)
            
        }else{
            fmt.Printf("--------------------------------------------------\n")
            fmt.Printf("startPage: %d\nendPage: %d\ninputFile: %s\npageLength: %d\npageType: %s\ndestFile: %s\n", args.startPage, args.endPage, args.inputFile, args.pageLength, pageType, args.destFile)
        }
    }

}
```

以达到仅仅输入”selpg“，就输出标准命令格式的目的；同时打印其他类型的参数错误或缺失信息。检查的顺序是：先检查了开始页args.startPage和结束页args.endPage是否被赋值，然后检查开始页args.startPage和结束页args.endPage是否为正数，接下来检查开始页args.startPage是否大于结束页args.endPage，然后检查自定页长-l和遇换页符换页-f是否同时出现，最后判断当自定页长-l出现时args.pageLength是否小于1。

5.执行命令

先检查输入，如果没有给定文件名，从标准输入中获取；给顶了文件名，则检查文件是否存在；然后打开文件，正常打开后，判断是否有-d参数。

```go
func DoSelpg(args *selpgArgs) {
    var fin *os.File
    if args.inputFile == "" {
        fin = os.Stdin
    } else {
        CheckFileAccess(args.inputFile)
        var err error
        fin, err = os.Open(args.inputFile)
        CheckError(err, "File input")
    }

    if len(args.destFile) == 0 {
        Output2Des(os.Stdout, fin, args.startPage, args.endPage, args.pageLength, args.pageType)
    } else {
        Output2Des(os.Stdout, fin, args.startPage, args.endPage, args.pageLength, args.pageType)
    }
}
```

***

#### 程序测试

新建一个test.txt作为测试文件(涵盖在本项目中)。为了方便地看出文本内容属于哪一页，我创建的txt文件每一行的内容为当前的行数。例如，第一行为1，第二行为2，一共145行(比72*2多一行，以此凑成三页)。

1.`selpg `

![1570197583158](https://github.com/farthjun/Go-CLI-/blob/master/img/1570197583158.png?raw=true)

2.`selpg -s1 -e1 test.txt`

![1570197764940](https://github.com/farthjun/Go-CLI-/blob/master/img/1570197764940.png?raw=true)

...

![1570197874136](https://github.com/farthjun/Go-CLI-/blob/master/img/1570197825605.png?raw=true)

可以看到刚好输出72行(一页)。

3.`selpg -s1 -e2 test.txt >outfile.txt`

![1570198189472](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198189472.png?raw=true)

![1570198218520](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198218520.png?raw=true)

正好144行(2页)。

4.`selpg -s1 -e4 test.txt 2>error_file.txt`

![1570198341479](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198341479.png?raw=true)

5.新建测试文件test_f.txt（涵盖在本项目中），在50行、100行之前插入'\f'。测试以下指令：

`selpg -s1 -e2 -f test_f.txt`

每页行数并不固定，显示在命令行中。

![1570198524592](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198524592.png?raw=true)

![1570198570142](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198570142.png?raw=true)

可以看到正好输出两页(100行之前)。

6.`selpg -s1 -e1 -dDestination test.txt`

![1570198748590](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198748590.png?raw=true)

![1570198768184](https://github.com/farthjun/Go-CLI-/blob/master/img/1570198768184.png?raw=true)

正好72行(1页)。
