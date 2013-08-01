package metapi

import (
	"encoding/json"
	"fmt"
	"github.com/bonnefoa/gobot/message"
	bmath "github.com/bonnefoa/gobot/utils/math"
	"log"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
)

type PiQuery struct {
	Num     int64
	Channel string
}

type PiState struct {
	Iteration   int
	Numerator   string
	Denominator string
}

const kNum1 = 13591409
const kNum2 = 545140134
const kDenom = 640320

var piNumerator = big.NewRat(1, 1).SetFloat64(math.Sqrt(10005))
var piFactor = big.NewRat(1, 1).Quo(piNumerator, big.NewRat(4270934400, 1))

func _factorial(n int, current []int) []int {
	if n == 0 {
		return current
	}
	current = append(current, n)
	return _factorial(n-1, current)
}

func ratPow(rat *big.Rat, n int) *big.Rat {
	res := big.NewRat(1, 1)
	for i := 0; i < n; i++ {
		res.Mul(res, rat)
	}
	return res
}

func factorial(n int) []int {
	current := make([]int, 0)
	return _factorial(n, current)
}

func concatInt(slc1, slc2 []int) []int {
	dest := append(slc1, slc2...)
	return dest
}

func computeMul(lst []int) *big.Rat {
	res := big.NewRat(1, 1)
	for _, v := range lst {
		res.Mul(res, big.NewRat(int64(v), 1))
	}
	return res
}

func reduceSlices(numerator, denominator []int) ([]int, []int) {
	sortedNum := sort.IntSlice(numerator)
	sortedNum.Sort()
	sortedDenom := sort.IntSlice(denominator)
	sortedDenom.Sort()
	j := 0
	resNum := make([]int, 0)
	resDenom := make([]int, 0)
	for i := 0; i < len(sortedNum); {
		if j > len(sortedDenom)-1 {
			resNum = concatInt(resNum, sortedNum[i:])
			break
		}
		if sortedNum[i] == sortedDenom[j] {
			j++
			i++
			continue
		}
		if sortedNum[i] > sortedDenom[j] {
			resDenom = append(resDenom, sortedDenom[j])
			j++
			continue
		}
		resNum = append(resNum, sortedNum[i])
		i++
	}
	if j < len(sortedDenom) {
		resDenom = concatInt(resDenom, sortedDenom[j:])
	}
	return resNum, resDenom
}

func snFactorialPart(n int) *big.Rat {
	n64 := int(n)
	numeratorList := factorial(6 * n64)
	denominatorList := factorial(3 * n64)
	denominatorList2 := factorial(n)
	for i := range denominatorList2 {
		num := denominatorList2[i]
		denominatorList2[i] = num * num * num
	}
	denominatorList = concatInt(denominatorList, denominatorList2)
	num, denom := reduceSlices(numeratorList, denominatorList)
	ratNum := computeMul(num)
	ratDenom := computeMul(denom)
	ratNum.Quo(ratNum, ratDenom)
	return ratNum
}

func snQuoPart(n int) *big.Rat {
	num := big.NewRat(kNum1+kNum2*int64(n), 1)
	denom := ratPow(big.NewRat(kDenom, 1), 3*n)
	num.Quo(num, denom)
	return num
}

func sn(n int) *big.Rat {
	factRat := snFactorialPart(n)
	quoRat := snQuoPart(n)
	factRat.Mul(factRat, quoRat)
	if n%2 != 0 {
		factRat.Neg(factRat)
	}
	return factRat
}

func getPiFromSum(sum *big.Rat) *big.Rat {
	pi := big.NewRat(1, 1).Set(piFactor)
	pi.Mul(pi, sum)
	pi.Inv(pi)
	return pi
}

func EstimatePiFromPrevious(iteration, previousIndex int, sum *big.Rat) (*big.Rat, *big.Rat) {
	for i := previousIndex; i < iteration; i++ {
		sum.Add(sum, sn(i))
	}
	pi := getPiFromSum(sum)
	return pi, sum
}

func EstimatePi(iteration int) (*big.Rat, *big.Rat) {
	return EstimatePiFromPrevious(iteration, 0, big.NewRat(0, 1))
}

func formatFoundResponse(pi string, num string, index int) string {
	low := bmath.MaxInt(index-10, 0)
	high := bmath.MinInt(index+len(num)+10, len(pi)-1)
	return fmt.Sprintf("Found %q at position %v, ...%s...", num, index, pi[low:high])
}

func savePiState(iteration int, sum *big.Rat) {
	file, _ := os.OpenFile("pi_cache", os.O_WRONLY|os.O_CREATE, 400)
	enc := json.NewEncoder(file)
	enc.Encode(PiState{iteration, sum.Num().String(), sum.Denom().String()})
	file.Close()
}

func loadPiState() PiState {
	file, err := os.Open("pi_cache")
	if err != nil {
		return PiState{0, "1", "1"}
	}
	dec := json.NewDecoder(file)
	var piState PiState
	dec.Decode(&piState)
	file.Close()
	return piState
}

func processUntilFound(numStr string, iteration int, piStr string, sum *big.Rat,
	responseChannel chan fmt.Stringer, ircChannel string) (int, string, *big.Rat) {
	for {
		index := strings.Index(piStr, numStr)
		if index > 0 {
			responseChannel <- message.MsgSend{ircChannel,
				formatFoundResponse(piStr, numStr, index)}
			return iteration, piStr, sum
		}
		responseChannel <- message.MsgSend{ircChannel,
			"Computing next 500 iteration of  pi"}
		newPi, newSum := EstimatePiFromPrevious(iteration+500, iteration, sum)
		sum = newSum
		piStr = newPi.FloatString(iteration * 14)
		log.Printf("Pi is %s\n", piStr)
		iteration += 500
		savePiState(iteration, sum)
	}
}

func SearchWorker(readChannel chan PiQuery, responseChannel chan fmt.Stringer) {
	piState := loadPiState()
	iteration := piState.Iteration
	sum, _ := big.NewRat(1, 1).SetString(piState.Numerator)
	den, _ := big.NewRat(1, 1).SetString(piState.Denominator)
	sum.Quo(sum, den)
	piStr := getPiFromSum(sum).FloatString(iteration * 14)
	for {
		piQuery := <-readChannel
		numStr := strconv.FormatInt(piQuery.Num, 10)
		iteration, piStr, sum = processUntilFound(numStr, iteration,
			piStr, sum, responseChannel, piQuery.Channel)
	}
}
