package usage

import (
	"time"
	"testing"
)

//=============================================================
//
//=============================================================

func TestDaysInYearMonth(t *testing.T) {
	for year := 2000; year < 2100; year++ {
		for month := time.January; month <= time.December; month++ {
			days := daysInMonth(year, month)
			
			tt := time.Date(year, month, days, 0, 0, 0, 0, time.UTC)
			if days > 31 || days < 28 || days != tt.Day() {
				t.Errorf("Wrong cal result of daysInMonth(%d, %d)=%d", year, month, days)
			}
		}
	} 
}