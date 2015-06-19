package mailmanv2

// An endpoint function
type endpoint func(wr *WorkRequest, w *Worker)
