function communicateWithMe(||channel) {
	var m = channel()
    print("You sent me: ", m)
    outbox("I sent you this")
}

software {
	var p = ||
	fork communicateWithMe(p)
	p("Hello thread")
	print("Reply: ", inbox())
}
