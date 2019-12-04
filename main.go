package main

import (
	"fmt"
	"github.com/bitfield/script"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func StrToUint(strNumber string, value interface{}) (err error) {
	var number interface{}
	number, err = strconv.ParseUint(strNumber, 10, 64)
	switch v := number.(type) {
	case uint64:
		switch d := value.(type) {
		case *uint64:
			*d = v
		case *uint:
			*d = uint(v)
		case *uint16:
			*d = uint16(v)
		case *uint32:
			*d = uint32(v)
		case *uint8:
			*d = uint8(v)
		}
	}
	return
}

func HandleData(str string) []string {
	strs := []rune(str)
	var results []string
	temp := ""
	for i := 0; i < len(strs); i++ {
		char := strs[i]
		if !unicode.IsSpace(char) {
			temp += string(char)
		} else {
			if len(temp) > 0 {
				results = append(results, temp)
			}
			temp = ""
		}
	}
	if len(temp) > 0 {
		results = append(results, temp)
	}
	return results
}

var SearchMap = map[string]string{
	"新建": "新建",
	"查询": "查询",
}

func main() {
	app := &cli.App{
		Name:     "o",
		Version:  "v1.0.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name: "gongym",
			},
		},
		Usage:     "轻松学习使用Linux命令",
		UsageText: "o/o.exe [global options] command [command options] [arguments...]",
		Commands: []*cli.Command{
			{
				Name: "ls",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Value:   ".",
						Usage:   "路径",
					},
				},
				Usage: "ls --path|p=路径",
				Action: func(c *cli.Context) error {
					path := c.String("path")
					_, err := script.ListFiles(path).RejectRegexp(regexp.MustCompile(`[^/]*/\\.[^/]*$`)).Stdout()
					return err
				},
			},
			{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "find|f 端口号",
				Action: func(context *cli.Context) error {
					if context.Args().Len() > 0 {
						portStr := context.Args().First()
						var port uint
						if err := StrToUint(portStr, &port); err != nil {
							return err
						}
						if 0 < port && port <= 65535 {
							findByPort := exec.Command("/bin/sh", "-context", fmt.Sprintf("netstat -nlp | grep :%d", port))
							findByPortOut, _ := findByPort.Output()
							results := HandleData(strings.Trim(string(findByPortOut), "\n"))
							// 格式化命令执行结果
							pid := ""
							fmt.Println("连接协议：", results[0])
							fmt.Println("接收队列：", results[1])
							fmt.Println("接收队列：", results[2])
							fmt.Println("本地地址：", results[3])
							fmt.Println("外部地址：", results[4])
							if len(results) == 6 {
								fmt.Println("进程信息：", results[5])
								pid = strings.Split(results[5], "/")[0]
							}
							if len(results) == 7 {
								fmt.Println("连接状态：", results[5])
								fmt.Println("进程信息：", results[6])
								pid = strings.Split(results[6], "/")[0]
							}
							// 获取PID，查找文件位置
							if len(pid) > 0 {
								findByPid := exec.Command("ls", "-l", fmt.Sprintf("/proc/%s/cwd", pid))
								findByPidOut, err := findByPid.Output()
								if err != nil {
									return err
								}
								fmt.Println("文件位置：", strings.Trim(string(findByPidOut), "\n"))
							}
						}
					}
					return nil
				},
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "search|s 关键词",
				Action: func(c *cli.Context) error {
					fmt.Println(SearchMap[c.Args().First()])
					return nil
				},
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "install|i 要安装的软件",
				Action: func(c *cli.Context) error {
					fmt.Println(SearchMap[c.Args().First()])
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
