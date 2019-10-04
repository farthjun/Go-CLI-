package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "os/exec"
    "github.com/spf13/pflag"
)

type selpgArgs struct {
    startPage   int
    endPage     int
    inputFile  string
    destFile   string
    pageLength int
    pageType    bool //-f or not
}

func main() {
    var args selpgArgs
    GetArgs(&args)
    CheckArgs(&args)
    DoSelpg(&args)
}

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

func CheckError(err error, object string) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "\n[Error]%s:", object)
        panic(err)
    }
}

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

func CheckFileAccess(filename string) {
    _, errFileExits := os.Stat(filename)
    if os.IsNotExist(errFileExits) {
        fmt.Fprintf(os.Stderr, "\n[Error]: input file \"%s\" does not exist\n", filename)
        os.Exit(0)
    }
}


func CMDExec(destFile string) io.WriteCloser {
    cmd := exec.Command("lp", "-d"+destFile)
    fout, err := cmd.StdinPipe()
    CheckError(err, "StdinPipe")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    errStart := cmd.Start()
    CheckError(errStart, "CMD Start")
    return fout
}

func Output2Des(fout interface{}, fin *os.File, pageStart int, pageEnd int, pageLength int, pageType bool) {

    lineCount := 0
    pageCount := 1
    buf := bufio.NewReader(fin)
    for true {

        var line string
        var err error
        if pageType {
            //If the command argument is -f
            line, err = buf.ReadString('\f')
            pageCount++
        } else {
            //If the command argument is -lnumber
            line, err = buf.ReadString('\n')
            lineCount++
            if lineCount > pageLength {
                pageCount++
                lineCount = 1
            }
        }

        if err == io.EOF {
            break
        }
        CheckError(err, "file read in")
        if (pageCount >= pageStart) && (pageCount <= pageEnd) {
            var outputErr error
            if stdOutput, ok := fout.(*os.File); ok {
                _, outputErr = fmt.Fprintf(stdOutput, "%s", line)
            } else if pipeOutput, ok := fout.(io.WriteCloser); ok {
                _, outputErr = pipeOutput.Write([]byte(line))
            } else {
                fmt.Fprintf(os.Stderr, "\n[Error]:fout type error. ")
                os.Exit(0)
            }
            CheckError(outputErr, "An error occurred when output the pages.")
        }
    }
    if pageCount < pageStart {
        fmt.Fprintf(os.Stderr, "\n[Error]: startPage (%d) greater than total pages (%d), no output written\n", pageStart, pageCount)
        os.Exit(0)
    } else if pageCount < pageEnd {
        fmt.Fprintf(os.Stderr, "\n[Error]: endPage (%d) greater than total pages (%d), less output than expected\n", pageEnd, pageCount)
        os.Exit(0)
    }
}
