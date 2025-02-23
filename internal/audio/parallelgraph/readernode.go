package parallelgraph

import (
	"errors"
	"io"

	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var _ Node = (*ReaderNode)(nil)

type ReaderNode struct {
	logger *logging.Logger
	r      io.Reader
	ins    []<-chan byte
	outs   []chan<- byte
	errs   chan error
	stop   chan FlushPolicy

	OnEOF *events.EventEmitter[struct{}]
}

func (node *ReaderNode) Start() error {
	if len(node.ins) != 0 {
		return newInvalidConnectionConfigErr(0, 0, len(node.ins))
	}

	if len(node.outs) != 1 {
		return newInvalidConnectionConfigErr(0, 1, len(node.outs))
	}

	go node.process()

	return nil
}

func (node *ReaderNode) Stop(flush FlushPolicy) error {
	defer func() {
		close(node.errs)
		node.logger.Debug("closed Errors() channel")
	}()

	node.stop <- flush

	return nil
}

func (node *ReaderNode) Errors() <-chan error {
	return node.errs
}

func (node *ReaderNode) addInput(in <-chan byte) {
	node.ins = append(node.ins, in)
}

func (node *ReaderNode) addOutput(out chan<- byte) {
	node.outs = append(node.outs, out)
}

func (node *ReaderNode) process() {
	data, errs := makeChanFromReader(node.r)

	for {
		select {
		case flush, ok := <-node.stop:
			if !ok {
				flush = FlushPolicy_NoFlush
			}

			node.logger.Debug("received done signal")
			_ = flush // TODO: handle flush policy
			return
		case err, ok := <-errs:
			if !ok {
				break
			}

			node.logger.Debug("received error from reader", "error", err)
			node.errs <- err
		case b, ok := <-data:
			if !ok {
				node.logger.Debug("reader channel closed")
				node.OnEOF.Broadcast(struct{}{})
				break // do not report EOF errors
			}

			node.logger.Debug("received data from reader", "data", []byte{b})

			select {
			case node.outs[0] <- b:
				node.logger.Debug("wrote bytes to next node", "data", []byte{b})
			default:
			}
		}
	}
}

func NewReaderNode(logger *logging.Logger, r io.Reader) *ReaderNode {
	return &ReaderNode{
		logger: logger,
		r:      r,
		ins:    make([]<-chan byte, 0),
		outs:   make([]chan<- byte, 0),
		errs:   make(chan error, 1),
		stop:   make(chan FlushPolicy),

		OnEOF: events.NewEventEmitter[struct{}](),
	}
}

func makeChanFromReader(r io.Reader) (data <-chan byte, errs <-chan error) {
	d := make(chan byte)
	e := make(chan error)

	go func() {
		defer close(d)
		defer close(e)

		buf := make([]byte, 0x1000)
		for {
			n, err := r.Read(buf)
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				e <- err
				break
			}

			if n > 0 {
				for _, b := range buf[:n] {
					d <- b
				}
			}
		}
	}()

	return d, e
}
