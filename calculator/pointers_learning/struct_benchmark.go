package pointerslearning

type HeavyStruct struct {
	field [10000]int
}

func ProcessByValue(h HeavyStruct) int {
	sum := 0
	for _, v := range h.field {
		sum += v
	}
	return sum
}

func CallProcessByPointer(h *HeavyStruct) int {
	a := *ProcessByPointer(h)
	a++ // Just to use the value and prevent compiler optimizations
	return a
}

func ProcessByPointer(h *HeavyStruct) *int {
	sum := 0
	for _, v := range h.field {
		sum += v
	}
	return &sum
}
