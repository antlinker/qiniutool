package main

import (
	"fmt"
	"log"

	"gopkg.in/urfave/cli.v2"

	"qiniupkg.com/api.v7/kodo"
)

type emptyCmd struct {
}

func (e emptyCmd) createCmd() *cli.Command {
	return &cli.Command{
		Name:    "empty",
		Aliases: []string{"empty"},
		Usage:   "清空七牛云容器内文件",
		Flags:   e.createEmpty(),
		Action:  e.createAction,
	}
}
func (e emptyCmd) createAction(c *cli.Context) error {
	pprof(c)
	b := c.String("bucket")
	if b == "" {
		log.Fatal("请输入容器名,使用 qiniutool help empty 查看帮助.")
		return nil
	}
	bb := c.String("bakbucket")
	if bb == "" {
		log.Fatal("请输入备份容器名,使用 qiniutool help empty 查看帮助.")
		return nil
	}
	if !auth(c) {
		return nil
	}
	if checkCode(fmt.Sprintf("清空七牛云容器(\033[1;32;40m%s\033[0m)内文件操作危险.", b)) {
		e.empty(b, bb)
	}
	return nil

}
func (e emptyCmd) createEmpty() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "bucket",
			Aliases: []string{"b"},
			Value:   "",
			Usage:   "指定需要清空的容器",
		},
		&cli.StringFlag{
			Name:    "bakbucket",
			Aliases: []string{"bb"},
			Value:   "",
			Usage:   "指定需要备份的容器,清除前会自动将原来容器的内容备份到新的容器,在清空,请不要使用已存在的容器名",
		},
	}
	tmp := createAuth()
	flags = append(flags, tmp...)
	return flags
}

func (e emptyCmd) empty(bucket string, bakbucket string) {
	h := walkhandler(func(cli *kodo.Client, b kodo.Bucket, item kodo.ListItem) (stop bool, err error) {

		log.Printf("移动%s|%s=>%s|%s 开始\n", bucket, item.Key, bakbucket, item.Key)
		err = b.MoveEx(nil, item.Key, bakbucket, item.Key)
		if err != nil {
			log.Printf("移动%s|%s=>%s|%s失败:%v\n", bucket, item.Key, bakbucket, item.Key, err)
		} else {
			log.Printf("移动%s|%s=>%s|%s成功.", bucket, item.Key, bakbucket, item.Key)
		}

		return
	})
	walkBucket(bucket, h)

}
