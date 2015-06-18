package mailmanv2

// An endpoint function
type endpoint func(wr *WorkRequest, w *Worker)

// A structure  to represent a device
type Phone struct {
	name string
	platform string
	reg_id string
}
