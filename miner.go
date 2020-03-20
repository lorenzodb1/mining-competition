/*
 * COMPETITION RULES:
 * START DATE: OCT 28, 2019 AT 9:00PM
 * END DATE: NOV 26, 2019 AT 9:00PM
 * PRIZE: 10% ON FINAL GRADE PER MINER_ID
 *
 * YOU NEED TO SELECT A SECURE private_id AND USE md5(private_id) AS MINER ID
 */

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
	"strconv"
	"time"
)

var lastCoin = make([]byte, 32)
var sharedCoinBlob = make([]byte, 89)
var reqBody  = make([]byte, 77)
var client1  = &http.Client{}
var client2  = &http.Client{}
var difficulty = 4
var isEven = false

func main() {
	fmt.Printf("Mining coins with %v goroutines...\n", runtime.NumCPU())
	minerId  := "8cb4fd57c36968a268d4a7179d69c706"
	copy(sharedCoinBlob, fmt.Sprintf("CPEN 442 Coin2019%s%s%s", [32]byte{}, [8]byte{}, minerId))
	copy(reqBody, fmt.Sprintf(`{"coin_blob":"%s","id_of_miner":"%s"}`, [12]byte{}, minerId))
	getLastCoin("http://cpen442coin.ece.ubc.ca/last_coin", 12, 44)
	getDifficulty()
	for i:= 0; i < runtime.NumCPU()-1; i++ {
		go generateCoin()
	}
	findLastCoin()
}

func findLastCoin() {
	flag := false
	for range time.Tick(100*time.Millisecond) {
		if flag {
			getLastCoin("http://cpen442coin.ece.ubc.ca/last_coin", 12, 44)
			getDifficulty()
		} else {
			getLastCoin("http://cpen442coin.ece.ubc.ca/get_last_mined_coin", 9, 41)
		}
		flag = !flag
	}
}

func getLastCoin(url string, start int, end int) {
	req, _ := http.NewRequest("POST", url, nil)
	res, err := client1.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		resBody, _ := ioutil.ReadAll(res.Body)
		tempCoinBlob := resBody[start:end]
		if bytes.Compare(lastCoin, tempCoinBlob) != 0  {
			copy(lastCoin, tempCoinBlob)
		}
	}
}
func getDifficulty() {
	req, _ := http.NewRequest("POST", "http://cpen442coin.ece.ubc.ca/difficulty", nil)
	res, err := client1.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		resBody, _ := ioutil.ReadAll(res.Body)
		tempDifficulty, err := strconv.Atoi(string(resBody[27:29]))
		if err != nil {
			return
		}
		isEven = (tempDifficulty % 2) == 0
		if isEven {
			difficulty = tempDifficulty / 2
		} else {
			difficulty = (tempDifficulty / 2) + 1
		}
	}
}

func generateCoin() {
	coinBlob := make([]byte, len(sharedCoinBlob))
	copy(coinBlob, sharedCoinBlob)
	randomBytes := make([]byte, 8)
	coin := [16]byte{}
	_, _ = rand.Read(randomBytes)
	randomInt := binary.LittleEndian.Uint64(randomBytes)
	for {
		copy(coinBlob[17:], lastCoin)
		copy(coinBlob[49:], randomBytes)
		coin = md5.Sum(coinBlob)
		if verifyCoin(coin) {
			claimCoin(randomBytes, coin)
		}
		randomInt = (randomInt+1)%math.MaxUint64
		binary.LittleEndian.PutUint64(randomBytes, randomInt)
	}
}

func verifyCoin(coin [16]byte) bool {
	if isEven {
		for i := 0; i < difficulty; i++ {
			if coin[i] != 0 {
				return false
			}
		}
	} else {
		if coin[difficulty] < 16 {
			return false
		}
		for i := 0; i < difficulty-1; i++ {
			if coin[i] != 0 {
				return false
			}
		}
	}
	return true
}

func claimCoin(foundBytes []byte, coin [16]byte) {
	copy(lastCoin, hex.EncodeToString(coin[:]))
	copy(reqBody[14:], base64.StdEncoding.EncodeToString(foundBytes))
	req, _ := http.NewRequest("POST", "http://cpen442coin.ece.ubc.ca/claim_coin", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := client2.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}
