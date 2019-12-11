package main

import (
	"errors"
	"fmt"
	"github.com/bitfield/script"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type LinuxType int

const (
	_ LinuxType = iota
	Centos
	Ubuntu
	Unknown
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

func CheckLinuxType() LinuxType {
	checkType := exec.Command("/bin/sh", "-c", "cat /etc/issue")
	checkTypeOut, _ := checkType.Output()
	checkTypeOutFmt := strings.ToLower(strings.Trim(string(checkTypeOut), "\n"))
	if strings.Contains(checkTypeOutFmt, "kernel") {
		return Centos
	} else if strings.Contains(checkTypeOutFmt, "ubuntu") {
		return Ubuntu
	} else {
		return Unknown
	}
}

func CheckAndRemoveJava() {
	linuxType := CheckLinuxType()
	fmt.Println("检查是否需要移除OpenJDK")
	if linuxType == Centos {
		checkJava := exec.Command("/bin/sh", "-c", "rpm -qa | grep java")
		checkJavaOut, _ := checkJava.Output()
		pkgs := strings.Split(strings.Trim(string(checkJavaOut), "\n"), "\n")
		if len(pkgs) == 0 {
			return
		}
		for _, pkg := range pkgs {
			if strings.Contains(pkg, "openjdk") {
				fmt.Println("正在移除：" + pkg)
				checkJava := exec.Command("/bin/sh", "-c", "rpm -e --nodeps "+pkg)
				_, _ = checkJava.Output()
			}
		}
	}
	fmt.Println("检查完成")
}

func InstallJava(version int) error {
	downloadUrl := ""
	if version == 8 {
		downloadUrl = "https://github.com/gongym12138/oracle-java/releases/download/8u231/jdk-8u231-linux-x64.tar.gz"
		fileName := "java8.tar.gz"
		response, err := http.Get(downloadUrl)
		if err != nil {
			return err
		}
		if response.StatusCode == http.StatusOK {
			fmt.Println("正在下载JDK")
		}

	}
	return nil
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
						Action: func(context *cli.Context) error {
							CheckAndRemoveJava()
							_ = InstallJava(8)
							installPath := "/usr/local/java/"
							if context.Args().Len() > 0 {
								installPath = context.Args().First()
							}
							cmdPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
							if err != nil {
								return err
							}
							fmt.Println(installPath)
							fmt.Println(cmdPath)
							return nil
						},
					},
					{
						Name:  "jdk11",
						Usage: "安装JDK.V11",
						Action: func(c *cli.Context) error {
							_ = InstallJava(11)
							return nil
						},
					},
					{
						Name:  "mysql",
						Usage: "安装MySQL",
						Action: func(c *cli.Context) error {
							return nil
						},
					},
				},
			},
			{
				Name:  "show",
				Usage: "show 需要查看的信息",
				Action: func(c *cli.Context) error {

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
