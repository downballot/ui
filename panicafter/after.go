package panicafter

var globalCount int

func After(count int) {
	if globalCount == 0 {
		globalCount = count
	} else {
		globalCount--
		if globalCount == 0 {
			panic("panicafter: globalCount is 0")
		}
	}
}
