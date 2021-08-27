package main

import (
	"fmt"
	"sync"

	"github.com/ziidu/guid"
)

// snowflake example
func oneServerMultiG_Snowflake() {
	var lock sync.Mutex
	var wait sync.WaitGroup
	snowflake := guid.NewSnowflakeUID(guid.IpWorkIdHolder)
	results := make(map[int64]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			resp := snowflake.Get("")
			lock.Lock()
			defer lock.Unlock()
			results[resp.Id] = struct{}{}
		}()
	}

	wait.Wait()
	fmt.Println(len(results) == 1000)
}
