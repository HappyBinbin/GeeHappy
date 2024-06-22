package singleflight

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do("key", func() (interface{}, error) {
		return "bar", nil
	})

	if v != "bar" || err != nil {
		t.Errorf("Do v = %v, error = %v", v, err)
	}
}

func TestCurl(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			_, err := http.Get("http://localhost:9999/api?key=Tom")
			if err != nil {
				return
			}
			wg.Done()
		}()
	}
	// 等待所有 goroutine 完成
	wg.Wait()
}

func TestGroup_Do(t *testing.T) {
	var g Group

	// 用于模拟并发的 WaitGroup
	var wg sync.WaitGroup
	wg.Add(10)

	// 记录函数调用次数
	//var callCount int
	//mu := sync.Mutex{}

	// 模拟的函数，每次调用增加 callCount
	fn := func() (interface{}, error) {
		fmt.Println("call count =======")
		time.Sleep(2 * time.Millisecond) // 模拟长时间执行的操作
		return "result", nil
	}

	// 启动 10 个 goroutine 并发调用 g.Do
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			v, err := g.Do("key", fn)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if v != "result" {
				t.Errorf("unexpected value: %v", v)
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 检查函数是否只被调用了一次
	//if callCount != 1 {
	//	t.Errorf("expected callCount to be 1, but got %d", callCount)
	//}
}
