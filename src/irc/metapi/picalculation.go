package main

import "math"
import "math/big"
import "sort"

const kNum1  = 13591409;
const kNum2  = 545140134;
const kDenom = 640320;

var piNumerator = big.NewRat(1, 1).SetFloat64(math.Sqrt(10005));
var piFactor = big.NewRat(1, 1).Quo(piNumerator, big.NewRat(4270934400, 1))

func _factorial(n int, current []int) []int {
        if n == 0 {
                return current
        }
        current = append(current, n)
        return _factorial(n - 1, current)
}

func ratPow(rat *big.Rat, n int) *big.Rat {
        res := big.NewRat(1, 1)
        for i := 0; i < n ; i++ {
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
                if j > len(sortedDenom) - 1 {
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
        num := big.NewRat(kNum1 + kNum2 * int64(n), 1)
        denom := ratPow(big.NewRat(kDenom, 1), 3 * n)
        num.Quo(num, denom)
        return num
}

func sn(n int) *big.Rat {
        factRat := snFactorialPart(n)
        quoRat := snQuoPart(n)
        factRat.Mul(factRat, quoRat)
        if n % 2 != 0 {
                factRat.Neg(factRat)
        }
        return factRat
}

func EstimatePi(iteration int) *big.Rat {
        pi := big.NewRat(0, 1).Set(piFactor)
        sum := big.NewRat(1, 1)
        for i := 0; i < iteration; i++ {
                sum.Add(sum, sn(i))
        }
        pi.Mul(pi, sum)
        pi.Inv(pi)
        return pi
}
