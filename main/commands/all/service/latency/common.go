package latency

func logIf(logf func(...interface{}), uid string, args ...interface{}) {
	var args2 []interface{}
	if len(uid) > 0 {
		args2 = append(args2, "uid: ", uid, ", ")
	}
	for _, arg := range args {
		args2 = append(args2, arg)
	}
	logf(args2...)
}
