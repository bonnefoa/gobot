package utilstring

import (
        "math/rand"
        "strings"
        "fmt"
        "unicode"
)

var rotateLetter = [...]rune {'ɐ', 'q', 'ɔ', 'p', 'ǝ', 'ɟ', 'ƃ', 'ɥ', 'ı', 'ɾ', 'ʞ', 'l', 'ɯ', 'u', 'o', 'd', 'b', 'ɹ', 's', 'ʇ', 'n', 'ʌ', 'ʍ', 'x', 'ʎ', 'z'}
var rotateDigit = [...]rune {'0', 'Ɩ', '2', 'Ɛ', '4', '5', '9', '7', '8', '6'}

func RotateString(s string) string {
        return strings.Map(RotateRune, s)
}

func RotateRune(r rune) rune {
        if unicode.IsLetter(r) {
                if r > 'z' || r < 'a' {
                        return r
                }
                index := int(r) - int('a')
                return rotateLetter[index]
        } else if unicode.IsDigit(r) {
                index := int(r) - int('0')
                return rotateDigit[index]
        }
        return r
}

func KeepLettersAndSpace(r rune) rune {
        if unicode.IsLetter(r) { return r }
        if unicode.IsSpace(r) { return r }
        return -1
}

func StringContains(el string, sl []string) bool {
        for _, m := range sl {
                if strings.Contains(el, m) {return true}
        }
        return false
}

func TriggerIn(words []string, msg string) bool {
        for _, w := range words {
                if strings.Contains(w, " ") {
                        if strings.Contains(msg, w) {
                                return true
                        }
                }
        }
        return SliceIn(words, strings.Split(msg, " "))
}

func SliceIn(el []string, sl []string) bool {
        for _, e := range el {
                for _, m := range sl {
                        if e == m { return true }
                }
        }
        return false
}

func SliceContains(el string, sl []string) bool {
        for _, m := range sl {
                if el == m { return true }
        }
        return false
}

func RandomString(sl []string) string {
        n := rand.Intn(len(sl))
        return sl[n]
}

func ShuffleString(str string) []string {
        splitted := strings.Split(str, "")
        for i := range splitted {
                j := rand.Intn(i+1)
                splitted[i], splitted[j] = splitted[j], splitted[i]
        }
        return splitted
}

func ColorStringSlice(lst []string) string {
        res := make([]string, len(lst))
        for i := 0; i < len(lst); i++ {
                res[i] = ColorString(lst[i])
        }
        return strings.Join(res, "")
}

func GetHearts(number int) string {
        hearts := strings.Repeat("❤♥", rand.Intn(number) + 1)
        shuffledHearts := ShuffleString(hearts)
        strHearts := ColorStringSlice(shuffledHearts)
        return strHearts
}

func Truncate(msg string, max int) string {
        if len(msg) < max { max = len(msg) }
        return msg[:max]
}

func ConcatStrings(lst1, lst2 []string) []string {
        dest := make([]string, len(lst1) + len(lst2))
        copy(dest, lst1)
        copy(dest[:len(lst1)], lst2)
        return dest
}

func ColorString(str string) string {
        res := fmt.Sprintf("%d%s", rand.Intn(15), str)
        return res
}

func GetRandUtf8() string {
        tableNum := rand.Intn(len(unicode.GraphicRanges) - 1)
        table := *unicode.GraphicRanges[tableNum]
        rangeNum := rand.Intn(len(table.R32) - 1)
        var hi, lo uint32
        if rand.Intn(1) == 0 {
                theRange := table.R32[rangeNum]
                hi, lo = theRange.Hi, theRange.Lo
        } else {
                theRange := table.R16[rangeNum]
                hi, lo = uint32(theRange.Hi), uint32(theRange.Lo)
        }
        diff := hi - lo
        randDiff := uint32(rand.Int31n(int32(diff)))
        theRune := lo + randDiff
        return string(theRune)
}
