package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"qiniupkg.com/api.v7/kodo"

	"git.oschina.net/antlinker/antmqtt/debug"

	"gopkg.in/urfave/cli.v2"
)

func createDefault() []cli.Flag {
	return []cli.Flag{

		&cli.StringFlag{
			Name:    "pprof",
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "启动pprof分析程序,默认不启动，可以设置:9090,指定一个监听ip端口通过http访问",
		},
	}
}

func createAuth() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "AK",
			Aliases: []string{"ak"},
			Value:   "",
			Usage:   "AK",
		},
		&cli.StringFlag{
			Name:    "SK",
			Aliases: []string{"sk"},
			Value:   "",
			Usage:   "sk",
		},
		&cli.StringFlag{
			Name:    "pprof",
			Aliases: []string{"pp"},
			Value:   "",
			Usage:   "启动pprof分析程序,默认不启动，可以设置:9090,指定一个监听ip端口通过http访问",
		},
	}
}

func checkCode(msg string) bool {
	reader := bufio.NewReader(os.Stdin)
	fcnt := 0
	for {
		code := CreateRandomString(6)
		fmt.Printf("%s.\n请输入验证码(\033[1;31;40m%s\033[0m)进行确认:", msg, code)
		line, _, err2 := reader.ReadLine()
		if err2 != nil || io.EOF == err2 {
			log.Println("请输入容器名,使用 qiniutool help empty 查看帮助.")
			return false
		}
		// if !isPrefix {
		// 	fmt.Printf("请重新输入.\n")
		// 	continue
		// }
		if string(line) != code {
			fcnt++
			fmt.Printf("验证码输入错误次数\033[1;31;40m%d\033[0m(3次错误后退出).\n", fcnt)
			if fcnt >= 3 {
				return false
			}
			continue
		}
		return true
	}
}

func auth(c *cli.Context) bool {
	ak := c.String("AK")
	sk := c.String("SK")
	if ak == "" {
		log.Fatal("请输入AK参数,使用 qiniutool help empty 查看帮助.")
		return false
	}
	if sk == "" {
		log.Fatal("请输入SK参数,使用 qiniutool help empty 查看帮助.")
		return false
	}
	fmt.Println(ak)
	fmt.Println(sk)
	kodo.SetMac(ak, sk)
	return true
}
func main() {

	app := &cli.App{
		Authors: []*cli.Author{
			{Name: "@antlinker.com"},
		},
		Name:    "qiniu管理工具",
		Version: "1.0",
		Usage:   "antmqtt 服务程序",
		Commands: []*cli.Command{
			emptyCmd{}.createCmd(),
			(&export{}).createCmd(),
			importCmd{}.createCmd(),
		},
	}
	go regSigin()
	err := app.Run(os.Args)
	if err != nil {
		log.Println("启动失败:", err)
	}
}
func regSigin() {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	<-sigc
	log.Println("收到了退出程序信号")
	os.Exit(0)

}
func pprof(c *cli.Context) {
	pprofaddr := c.String("pprof")
	if pprofaddr != "" {
		debug.StartHTTPPprof(pprofaddr)
	}
}
