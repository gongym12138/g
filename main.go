package main

import (
	"errors"
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

func CheckAndRemoveJava() {
	checkJava := "rpm -qa | grep java"
	findByPort := exec.Command("/bin/sh", "-c", checkJava)
	findByPortOut, _ := findByPort.Output()
	results := HandleData(strings.Trim(string(findByPortOut), "\n"))
	fmt.Println(results)
}

var SearchMap = map[string]string{
	"新建": "新建",
	"查询": "查询",
}

func main() {
	app := &cli.App{
		Name:     "g",
		Version:  "v1.0.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name: "gongym",
			},
		},
		Usage:     "轻松学习使用Linux命令",
		UsageText: "g/g.exe [global options] command [command options] [arguments...]",
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
					if context.Args().Len() == 0 {
						return errors.New("error: 未定义参数")
					}
					portStr := context.Args().First()
					var port uint
					if err := StrToUint(portStr, &port); err != nil {
						return err
					}
					if port < 0 || 65535 < port {
						return errors.New("error: 端口号不符合规则")
					}
					findByPort := exec.Command("/bin/sh", "-c", fmt.Sprintf("netstat -nlp | grep :%d", port))
					findByPortOut, _ := findByPort.Output()
					results := HandleData(strings.Trim(string(findByPortOut), "\n"))
					if len(results) == 0 {
						// 没有查找到对应端口号的内容
						return nil
					}
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
				Subcommands: []*cli.Command{
					{
						Name:  "jdk8",
						Usage: "安装JDK.V8",
						Action: func(c *cli.Context) error {
							CheckAndRemoveJava()
							return nil
						},
					},
					{
						Name:  "jdk11",
						Usage: "安装JDK.V11",
						Action: func(c *cli.Context) error {
							fmt.Println("delete subcommand")
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
