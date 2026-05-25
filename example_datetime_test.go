package thainum_test

import (
	"fmt"
	"time"

	thainum "github.com/ultramcu/go-thainum"
)

func ExampleFormatDate() {
	d := time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC)
	fmt.Println(thainum.FormatDate(d))
	fmt.Println(thainum.FormatDateFull(d))
	// Output:
	// 5 มิถุนายน 2567
	// วันพุธที่ 5 มิถุนายน พ.ศ. 2567
}

func ExampleParseDate() {
	d, _ := thainum.ParseDate("5 มิถุนายน 2567")
	fmt.Println(d.Format("2006-01-02"))
	// Output: 2024-06-05
}

func ExampleFormatTime() {
	t := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	fmt.Println(thainum.FormatTime(t))  // formal
	fmt.Println(thainum.FormatClock(t)) // colloquial
	// Output:
	// สิบสี่นาฬิกาสามสิบนาที
	// บ่ายสองโมงครึ่ง
}

func ExampleFormatDuration() {
	fmt.Println(thainum.FormatDuration(90 * time.Minute))
	// Output: หนึ่งชั่วโมงสามสิบนาที
}

func ExampleOrdinal() {
	fmt.Println(thainum.Ordinal(21))
	fmt.Println(thainum.Fraction(3, 4))
	fmt.Println(thainum.Year(2566))
	// Output:
	// ที่ยี่สิบเอ็ด
	// เศษสามส่วนสี่
	// พุทธศักราชสองพันห้าร้อยหกสิบหก
}
