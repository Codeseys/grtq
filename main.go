package main

import (
	"fmt"
	"strconv"
	"time"
	// "github.com/scottkiss/grtm"
)

// func testting(msg ...interface{}) {
// 	fmt.Println(msg[0].([]interface{})[0].(string))
// 	time.Sleep(2 * time.Second)
// }
func testting(args interface{}) {
	fmt.Println(args.(string))
	// fmt.Println(args[0].([]interface{})[1].(string))
	time.Sleep(time.Millisecond * 2000)
}

func main() {
	// defer ants.Release()
	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()
	// input := scanner.Text()
	// fmt.Println("Search Parameter: ", input)
	// runtimes := 1000
	// var wg sync.WaitGroup
	// syncCalculateSum := func() {
	// 	time.Sleep(10 * time.Second)
	// 	// fmt.Println("Hello World!")
	// 	wg.Done()
	// }
	// for i := 0; i < runtimes; i++ {
	// 	wg.Add(1)
	// 	_ = ants.Submit(syncCalculateSum)
	// 	// fmt.Println(mlep)
	// }
	// wg.Wait()
	// fmt.Printf("running goroutines: %d\n", ants.Running())
	// fmt.Printf("finish all tasks.\n")
	gm := NewDefaultGRManager()
	defer gm.pool.Release()
	for i := 0; i < 500; i++ {
		gm.newloopTask("Task "+strconv.Itoa(i), false, testting, "Task "+strconv.Itoa(i)+" running")
		// time.Sleep(time.Millisecond*200)
	}
	// gm.newloopTask("Task 1", testting, "Task 1 running")
	time.Sleep(time.Second * 10)
	// gm.StartGRT("Task 1")
	gm.StartAllGRT()
	// gm.newloopTask("Task 2", testting, "Task 2 running")
	// gm.newloopTask("Task 3", testting, "Task 3 running")
	time.Sleep(time.Second * 20)
	// gm.StopLoopGoroutine("Task 1")
	// gm.StopLoopGoroutine("Task 2")
	// gm.StopLoopGoroutine("Task 3")
	fmt.Println("Stopping")
	gm.StopAllGRT()
	time.Sleep(time.Second * 20)
	// gm.StartAllGRT()
	time.Sleep(time.Second * 20)

	gm.DeleteAllGRT()
	// for i := 0; i < 10; i++ {
	// 	gm.DeleteGRT("Task "+strconv.Itoa(i))
	// }
	// gm.wg.Wait()
	fmt.Println("All Stopped")
}
