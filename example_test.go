package thainum_test

import (
	"fmt"

	thainum "github.com/ultramcu/go-thainum"
)

func ExampleSpell() {
	fmt.Println(thainum.Spell(21))
	fmt.Println(thainum.Spell(1000000))
	// Output:
	// ยี่สิบเอ็ด
	// หนึ่งล้าน
}

func ExampleBaht() {
	fmt.Println(thainum.Baht(100))        // 100 baht (whole-baht unit)
	fmt.Println(thainum.BahtSatang(2121)) // 2121 satang = 21.21 baht
	fmt.Println(thainum.BahtSatang(25))   // 25 satang
	// Output:
	// หนึ่งร้อยบาทถ้วน
	// ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์
	// ยี่สิบห้าสตางค์
}

func ExampleToThaiDigits() {
	fmt.Println(thainum.ToThaiDigits("2566"))
	// Output: ๒๕๖๖
}

func ExampleFormatTHB() {
	fmt.Println(thainum.FormatTHB(2121))
	// Output: ฿21.21
}

func ExampleParseInt() {
	n, _ := thainum.ParseInt("ยี่สิบเอ็ด")
	fmt.Println(n)
	// Output: 21
}

func ExampleParseBaht() {
	satang, _ := thainum.ParseBaht("ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์")
	fmt.Println(satang)
	// Output: 2121
}
