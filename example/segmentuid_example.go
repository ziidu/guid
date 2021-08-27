package main

import (
	"fmt"
	"time"

	"github.com/ziidu/guid"
	"github.com/ziidu/guid/dao"
)

func oneServerMultiG() {
	const connectURL = "root:admin@tcp(192.168.56.130:3306)/leaf_test?parseTime=true"
	dao := dao.NewDefaultDao(connectURL, "leaf_alloc")
	segmentUID := guid.NewSegmentUID(dao)

	for i := 0; i <= 1000; i++ {
		go func() {
			resp := segmentUID.Get("pay_order")
			fmt.Printf("%d-%d--\n", resp.Code, resp.Id)
		}()
	}
	time.Sleep(time.Second * 3)
}

func multiServer() {
	const connectURL = "root:admin@tcp(192.168.56.130:3306)/leaf_test?parseTime=true"
	dao := dao.NewDefaultDao(connectURL, "leaf_alloc")
	segmentUID1 := guid.NewSegmentUID(dao)
	segmentUID2 := guid.NewSegmentUID(dao)
	segmentUID3 := guid.NewSegmentUID(dao)

	for i := 0; i < 100; i++ {
		go func() {
			resp := segmentUID1.Get("pay_order")
			fmt.Printf("1: %d-%d--\n", resp.Code, resp.Id)
		}()
		go func() {
			resp := segmentUID2.Get("pay_order")
			fmt.Printf("2: %d-%d--\n", resp.Code, resp.Id)
		}()
		go func() {
			resp := segmentUID3.Get("pay_order")
			fmt.Printf("3: %d-%d--\n", resp.Code, resp.Id)
		}()
	}
	time.Sleep(time.Second * 10)
}
