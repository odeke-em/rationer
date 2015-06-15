package rationer

var Sentinel = "sentinel"

type Job func() interface{}

type Rationer struct {
	done          chan bool
	finished      bool
	results       chan interface{}
	tokenCapacity uint64
}

func NewRationer(capacity uint64) *Rationer {
	results := make(chan interface{})
	done := make(chan bool, 1)

	r := Rationer{
		tokenCapacity: capacity,
		done:          done,
		results:       results,
		finished:      false,
	}

	return &r
}

func (r *Rationer) Run() chan interface{} {
	loader := make(chan interface{})

	go func() {
		doneChan := make(chan bool)
		doneCount := uint64(0)

		tokens := make(chan bool, r.tokenCapacity)

		// Load them up
		for i := uint64(0); i < r.tokenCapacity; i += 1 {
			tokens <- true
		}

		for {
			head, hasContent := <-loader

			if !hasContent || head == Sentinel { // Cancellation requested or chan closed
				break
			}

			job, ok := head.(Job)
			if !ok {
				continue
			}

			<-tokens
			doneCount += 1

			go func(results chan interface{}) {
				results <- job()

				tokens <- true
				doneChan <- true

			}(r.results)
		}

		for i := uint64(0); i < doneCount; i += 1 {
			<-doneChan
		}

		r.done <- true
	}()

	return loader
}

func (r *Rationer) Wait() {
	r.finished = true
	<-r.done
}

func (r *Rationer) Results() chan interface{} {
	return r.results
}
