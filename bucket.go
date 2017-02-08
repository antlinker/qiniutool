package main

import (
	"io"
	"log"
	"os"

	"qiniupkg.com/api.v7/kodo"
)

const (
	badToken = "bad token"
)

type walkhandler func(cli *kodo.Client, b kodo.Bucket, item kodo.ListItem) (stop bool, err error)

func walkBucket(bucket string, handler walkhandler) {
	cli := kodo.New(0, nil)

	b := cli.Bucket(bucket)
	var (
		entries        []kodo.ListItem
		commonPrefixes []string
		markerOut      string
		err            error
	)
	//	host := b.BucketInfo.IoHost
	log.Printf("bucket:%#v\n", b.BucketInfo)
	for err == nil || (err != io.EOF && err.Error() != badToken) {
		//	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
		entries, commonPrefixes, markerOut, err = b.List(nil, "", "", markerOut, 10)
		log.Printf("%v |%v | %#v", commonPrefixes, markerOut, err)
		for _, en := range entries {
			stop, e := handler(cli, b, en)
			if stop {
				return
			}
			if e != nil && err == nil {
				err = e
			}

		}
		//cancel()
	}

	log.Println(err)
}

func same(qstat kodo.Entry, fstat os.FileInfo) bool {

	return qstat.Fsize == fstat.Size()
}
