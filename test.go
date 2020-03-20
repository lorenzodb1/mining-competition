package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"time"
)

var testLastCoin = make([]byte, 32)
var testCoinBlob = make([]byte, 89)
var testReqBody  = make([]byte, 77)
var testClient = &http.Client{}
const NumCycles = 1000000000
var NumCPU = runtime.NumCPU()

func main() {
	fmt.Printf("Testing with %v routines...\n", NumCPU)
	wait := make(chan bool, 1)

	for i:= 0; i < NumCPU; i++ {
		go TestGenerate(wait)
	}

	//go TestFindLast(wait)

	//for range time.Tick(6*time.Second) {
	//	TestClaim([]byte{}, [16]byte{})
	//}

	<- wait
}

func TestFindLast(wait chan bool) {
	var url string
	var start int
	var end int
	var tempCoinBlob []byte
	var resBody []byte
	flag := false
	for range time.Tick(1600*time.Millisecond) {
		if flag {
			url = "http://cpen442coin.ece.ubc.ca/last_coin"
			start = 12
			end = 44
		} else {
			url = "http://cpen442coin.ece.ubc.ca/get_last_mined_coin"
			start = 9
			end = 41
		}
		req, _ := http.NewRequest("POST", url, nil)
		res, err := testClient.Do(req)
		if err != nil {
			return
		}
		if res.StatusCode == 200 {
			resBody, _ = ioutil.ReadAll(res.Body)
			tempCoinBlob = resBody[start:end]
			if bytes.Compare(testLastCoin, tempCoinBlob) != 0  {
				copy(testCoinBlob[17:], tempCoinBlob)
				copy(testLastCoin, tempCoinBlob)
			}
		}
		fmt.Printf("%v: %s\n", res.StatusCode, url)
		res.Body.Close()
		flag = !flag
	}
	wait <- true
}

func TestGenerate(wait chan bool) {
	coinBlob := make([]byte, len(testCoinBlob))
	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes)
	randomInt := binary.LittleEndian.Uint64(randomBytes)
	for i := uint64(0); i < NumCycles; i++ {
		copy(coinBlob, testCoinBlob)
		copy(coinBlob[49:], randomBytes)
		coin := md5.Sum(coinBlob)
		if TestVerify(coin) {
			TestClaim(randomBytes, coin)
		}
		randomInt = (randomInt+1)%math.MaxUint64
		binary.LittleEndian.PutUint64(randomBytes, randomInt)
	}
	wait <- true
}

func TestVerify(coin [16]byte) bool {
	for i := 0; i < 4; i++ {
		if coin[i] != 0 {
			return false
		}
	}
	fmt.Printf("found: %v\n", time.Now().Nanosecond())
	return true
}

func TestClaim(foundBytes []byte, coin [16]byte) {
	copy(testReqBody[14:], base64.StdEncoding.EncodeToString(foundBytes))
	req, _ := http.NewRequest("POST", "http://cpen442coin.ece.ubc.ca/claim_coin", bytes.NewBuffer(testReqBody))
	fmt.Println(req.UserAgent())
	req.Header.Set("Content-Type", "application/json")
	res, err := testClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	hexCoin := hex.EncodeToString(coin[:])
	if res.StatusCode == 200 {
		copy(testCoinBlob[17:], hexCoin)
		copy(testLastCoin, hexCoin)
		fmt.Printf("CLAIMED: %s\n", hexCoin)
	} else  if res.StatusCode == 409 {
		fmt.Printf("CONFLICT - %s: %s\n", res.Status, hexCoin)
	} else {
		fmt.Printf("REJECTED - %s: %s\n", res.Status, hexCoin)
	}
}
