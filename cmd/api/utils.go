package main



import (
	"time"

)

func artificialSlowdown(seconds int) {
	time.Sleep(time.Second * time.Duration(seconds))
} 

