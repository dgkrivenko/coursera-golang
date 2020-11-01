package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// SingleHashAsync - count left part of single hash
func SingleHashAsync(value string, mu *sync.Mutex) chan string {
	result := make(chan string, 1)
	go func(value string) {
		mu.Lock()
		r := DataSignerMd5(value)
		mu.Unlock()
		result <- DataSignerCrc32(r)
	}(value)
	return result
}

// SingleHash - count single hash of value
func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for data := range in {
		value, _ := data.(int)
		str := strconv.Itoa(value)
		wg.Add(1)
		// each hash is calculated in a separate goroutine
		go func(value string) {
			// calculate left part and right part of hash in different goroutine
			rightPartChan := SingleHashAsync(value, mu)
			leftPar := DataSignerCrc32(value)
			rightPart := <-rightPartChan
			out <- leftPar + "~" + rightPart
			wg.Done()
		}(str)
	}
	wg.Wait()
}

// MultiHashAsync - count multi hash of value
func MultiHashAsync(value string) string {
	var resultBuff = make([]string, 6, 6)
	var result string
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i <= 5; i++ {
		wg.Add(1)
		// Count each part of multi-hash in different goroutines and save it to resultBuff
		go func(idx int, value string) {
			res := DataSignerCrc32(strconv.Itoa(idx) + value)
			mu.Lock()
			resultBuff[idx] = res
			mu.Unlock()
			wg.Done()
		}(i, value)
	}
	wg.Wait()
	for _, val := range resultBuff {
		result += val
	}
	return result
}

// MultiHash - run goroutine for count hash for each value
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		value := data.(string)
		wg.Add(1)
		go func(value string) {
			out <- MultiHashAsync(value)
			wg.Done()
		}(value)
	}
	wg.Wait()
}

// CombineResults - count combine results hash
func CombineResults(in, out chan interface{}) {
	var buff []string
	var result string
	// save data while chanel is open
	for data := range in {
		str, _ := data.(string)
		buff = append(buff, str)
	}
	sort.Strings(buff)
	if len(buff) > 0 {
		result = buff[0]
	}
	for i := 1; i < len(buff); i++ {
		result += "_" + buff[i]
	}
	out <- result
}

// ExecutePipeline - run all jobs
func ExecutePipeline(hashSignJobs ...job) {
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	wg := &sync.WaitGroup{}
	for _, j := range hashSignJobs {
		wg.Add(1)
		go func (in, out chan interface{}, worker job) {
			worker(in, out)
			close(out)
			wg.Done()
		}(in, out, j)
		in = out
		out = make(chan interface{}, 1)
	}
	wg.Wait()
}

func main() {
	var testResult string
	var expectedResult = "1173136728138862632818075107442090076184424490584241521304_1696913515191343735512658979631549563179965036907783101867_27225454331033649287118297354036464389062965355426795162684_29568666068035183841425683795340791879727309630931025356555_3994492081516972096677631278379039212655368881548151736_4958044192186797981418233587017209679042592862002427381542_4958044192186797981418233587017209679042592862002427381542"
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
				return
			}
			testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	end := time.Since(start)

	expectedTime := 3 * time.Second

	if expectedResult != testResult {
		fmt.Printf("results not match\nGot: %v\nExpected: %v\n", testResult, expectedResult)
	} else {
		fmt.Println("Result value ok")
	}
	if end > expectedTime {
		fmt.Printf("execition too long\nGot: %s\nExpected: <%s\n", end, time.Second*3)
	} else {
		fmt.Println("Result time ok")
	}
}
