package metapi

import "math"
import "math/big"

var kNum1 = big.NewRat(545140134, 1);
var kNum2 = big.NewRat(13591409, 1);
var piNumerator = big.NewRat(1, 1).SetFloat64(53360 * math.Sqrt(640320));
var kDenominator = big.NewRat(262537412640768000, 1);

func _factorial(n int64, current *big.Rat) *big.Rat {
        if n == 0 {
                return current
        }
        return _factorial(n - 1, current.Mul(current, big.NewRat(n, 1)))
}

func ratPow(rat *big.Rat, n int) *big.Rat {
        res := big.NewRat(1, 1)
        for i := 0; i < n ; i++ {
                res.Mul(res, rat)
        }
        return res
}

func factorial(n int64) *big.Rat {
        return _factorial(n, big.NewRat(1, 1))
}

func sn(n int) *big.Rat {
        n64 := int64(n)
        numerator := factorial(6 * n64)

        kNumeratorRight := big.NewRat(n64, 1)
        kNumeratorRight.Mul(kNum1, kNumeratorRight)
        kNumeratorRight.Add(kNumeratorRight, kNum2)
        numerator.Mul(numerator, kNumeratorRight)

        denominator := ratPow(factorial(n64), 3)
        denominator.Mul(denominator, factorial(3 * n64))
        denominator.Mul(denominator, ratPow(kDenominator, n))

        numerator.Quo(numerator, denominator)
        if n % 2 != 0 {
                numerator.Neg(numerator)
        }
        return numerator
}

func estimatePi(iteration int) *big.Rat {
        pi := big.NewRat(1, 1)
        pi.Mul(pi, piNumerator)
        sum := big.NewRat(1, 1)
        for i := 0; i < iteration; i++ {
                sum.Add(sum, sn(i))
        }
        pi.Quo(pi, sum)
        return pi
}
