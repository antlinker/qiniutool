package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/antlinker/store/qiniu"

	"gopkg.in/urfave/cli.v2"
)

type importCmd struct {
}

func (e importCmd) createCmd() *cli.Command {
	return &cli.Command{
		Name:    "import",
		Aliases: []string{"imp"},
		Usage:   "导入文件到七牛云",
		Flags:   e.createFlag(),
		Action:  e.createAction,
	}
}
func (e importCmd) createAction(c *cli.Context) error {
	pprof(c)
	b := c.String("bucket")
	if b == "" {
		log.Fatal("请输入容器名,使用 qiniutool help import 查看帮助.")
		return nil
	}
	p := c.String("path")
	if p == "" {
		log.Fatal("请输入本地路径,使用 qiniutool help import 查看帮助.")
		return nil
	}
	pf := c.String("prefix")
	tc := c.Int("taskcnt")
	fc := c.Int("failcnt")
	if !auth(c) {
		return nil
	}
	if tc <= 0 {
		tc = 1
	}
	if fc <= 0 {
		fc = 0
	}
	if checkCode(fmt.Sprintf("导入目录(%s)下的所有文件到七牛云容器(\033[1;32;40m%s\033[0m).", b, p)) {
		e.importBucket(b, pf, p, tc, fc)
	}
	return nil

}
func (e importCmd) createFlag() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "bucket",
			Aliases: []string{"b"},
			Value:   "",
			Usage:   "指定需要清空的容器",
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "指定需要上传的本地目录,上传文件包含目录及子目录下的所有文件",
		},
		&cli.StringFlag{
			Name:    "prefix",
			Aliases: []string{"pf"},
			Value:   "",
			Usage:   "追加上传的前缀",
		},
		&cli.IntFlag{
			Name:    "taskcnt",
			Aliases: []string{"tc"},
			Value:   1,
			Usage:   "同时上传的任务数量",
		},
		&cli.IntFlag{
			Name:    "failcnt",
			Aliases: []string{"fc"},
			Value:   1,
			Usage:   "失败重试次数",
		},
	}
	tmp := createAuth()
	flags = append(flags, tmp...)
	return flags
}

func (e importCmd) importBucket(bucket, prefix, path string, tc, fc int) {
	start := time.Now()
	s := qiniu.CreateStore(bucket, 3600)
	root, err := filepath.Abs(path)
	if err != nil {
		log.Printf("目录设置错误:%s", err)
		return
	}
	atask := createAsyncTask(tc, fc)

	l := len(root) + 1
	var success int64
	var faildcnt int64
	wf := filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		key := prefix + path[l:]
		atask.put(createAsyncHandler(fmt.Sprintf("上传文件%s=>%s开始....", path, bucket), func() error {
			//	log.Printf("上传文件%s=>%s开始", path, bucket)
			err = s.SaveFile(key, path)
			if err != nil {
				atomic.AddInt64(&faildcnt, -1)
				//	log.Printf("上传文件%s=>%s失败:%v", path, bucket, err)
				return err
			}
			atomic.AddInt64(&success, 1)
			//log.Printf("上传文件%s=>%s|%s成功\n", path, bucket, key)
			return nil
		}))
		return nil
	})
	go func() {
		filepath.Walk(root, wf)
		atask.stop()
	}()
	atask.start()
	end := time.Now()
	log.Printf("导入目录%s下的所有文件到(%s)完成.共导入文件:%d个,失败%d个", path, bucket, success, faildcnt)
	log.Printf("开始时间%v,结束时间%v,共用%v时", start, end, end.Sub(start))
}
