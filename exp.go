package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"gopkg.in/urfave/cli.v2"

	"qiniupkg.com/api.v7/kodo"
)

type export struct {
	bucket kodo.Bucket
}

func (e *export) createCmd() *cli.Command {
	return &cli.Command{
		Name:    "export",
		Aliases: []string{"exp"},
		Usage:   "备份到处容器文件",
		Flags:   e.createFlag(),
		Action:  e.createAction,
	}
}
func (e *export) createAction(c *cli.Context) error {

	pprof(c)
	b := c.String("bucket")
	if b == "" {
		log.Fatal("请输入容器名,使用 qiniutool help empty 查看帮助.")
		return nil
	}
	p := c.String("path")
	if p == "" {
		log.Fatal("请输入备份路径.")
		return nil
	}
	host := c.String("host")
	if host == "" {
		log.Fatal("请输入容器绑定域名.")
		return nil
	}
	tc := c.Int("taskcnt")
	fc := c.Int("failcnt")
	if tc <= 0 {
		tc = 1
	}
	if fc <= 0 {
		fc = 0
	}
	if !auth(c) {
		return nil
	}

	if checkCode(fmt.Sprintf("导出容器(\033[1;32;40m%s\033[0m)内文件,操作时间可能很长.", b)) {
		e.exportBucket(b, host, p, tc, fc)
	}
	return nil

}
func (e *export) createFlag() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "bucket",
			Aliases: []string{"b"},
			Value:   "",
			Usage:   "指定需要导出的容器",
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   "",
			Usage:   "指定本地备份路径",
		},
		&cli.StringFlag{
			Name:    "host",
			Aliases: []string{"hh"},
			Value:   "",
			Usage:   "指定容器绑定域名",
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
func (e *export) exportURL(url string, file *os.File) error {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(file, res.Body)
	return err
}
func (e *export) getVisitURL(c *kodo.Client, host, key string) string {
	baseURL := kodo.MakeBaseUrl(host, key)
	policy := kodo.GetPolicy{}
	//生成一个client对象
	//调用MakePrivateUrl方法返回url
	return c.MakePrivateUrl(baseURL, &policy)
}
func (e *export) download(url, path, key string) int {

	filename := path + "/" + key
	stat, err := e.bucket.Stat(nil, key)

	if err != nil {
		return -1
	}
	fstat, err := os.Stat(filename)
	if err != nil {
		err1 := os.Remove(filename)
		if err1 != nil {
			log.Printf("删除文件%s失败", filename)
			return -1
		}
	} else {
		if same(stat, fstat) {
			return 1
		}
	}

	dir := filepath.Dir(filename)
	os.MkdirAll(dir, os.ModePerm)

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("创建文件%s失败", filename)
		return -1
	}
	err = e.exportURL(url, file)
	if err != nil {
		log.Printf("下载%s失败", url)
		return -1
	}
	return 0
}
func (e *export) exportBucket(bucket string, host string, path string, tc, fc int) {
	start := time.Now()
	cli := kodo.New(0, nil)
	e.bucket = cli.Bucket(bucket)
	atask := createAsyncTask(tc, fc)
	var success int64
	var faildcnt int64
	var skipcnt int64
	h := walkhandler(func(cli *kodo.Client, b kodo.Bucket, item kodo.ListItem) (stop bool, err error) {

		atask.put(createAsyncHandler(fmt.Sprintf("下载%s|%s 开始....", bucket, item.Key), func() error {
			url := e.getVisitURL(cli, host, item.Key)
			switch e.download(url, path, item.Key) {
			default:
				atomic.AddInt64(&faildcnt, -1)
			case 0:
				atomic.AddInt64(&success, 1)
			case 1:
				atomic.AddInt64(&skipcnt, 1)
			}
			return nil
		}))
		return
	})
	go func() {
		walkBucket(bucket, h)
		atask.stop()
	}()

	atask.start()
	log.Printf("导出容器%s下的所有文件到目录(%s)完成.导出文件:%d个,失败%d个,跳过%d个", bucket, path, success, faildcnt, skipcnt)
	end := time.Now()
	log.Printf("开始时间%v,结束时间%v,共用%v时", start, end, end.Sub(start))
}
