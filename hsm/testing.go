package hsm

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/net/context"

	"github.intel.com/hpdd/logging/debug"
	"github.intel.com/hpdd/lustre"
	"github.intel.com/hpdd/lustre/llapi"
)

type (
	// SignalChan is a channel that waiters can block on while
	// waiting for certain events to occur.
	SignalChan chan struct{}

	// TestSource implements hsm.ActionSource, but provides a
	// Lustre-independent way of generating hsm requests.
	TestSource struct {
		outgoing   chan ActionRequest
		nextAction chan ActionRequest
		rng        *rand.Rand
	}

	// TestRequest implements hsm.ActionRequest with additional
	// methods for controlling/inpecting request state.
	TestRequest struct {
		archive                uint
		action                 llapi.HsmAction
		testFid                *lustre.Fid
		handleProgressReceived chan *TestProgressUpdate
		handleEndReceived      SignalChan
	}

	testHandle struct {
		req              ActionRequest
		fid              *lustre.Fid
		progressReceived chan *TestProgressUpdate
		endReceived      SignalChan
	}

	// TestProgressUpdate contains information about progress updates
	// received by the test handle.
	TestProgressUpdate struct {
		Offset uint64
		Length uint64
		Total  uint64
	}
)

func (p *TestProgressUpdate) String() string {
	return fmt.Sprintf("Progress: %d->%d/%d", p.Offset, p.Length, p.Total)
}

// NewTestSource returns an ActionSource implementation suitable for testing
func NewTestSource() *TestSource {
	return &TestSource{
		nextAction: make(chan ActionRequest),
		outgoing:   make(chan ActionRequest),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddAction allows test code to inject arbitrary ActionRequests.
func (s *TestSource) AddAction(ar ActionRequest) {
	s.nextAction <- ar
}

// GenerateRandomAction generates a random action request
func (s *TestSource) GenerateRandomAction() {
	s.nextAction <- &TestRequest{}
}

// Actions returns a channel for callers to receive ActionRequests
func (s *TestSource) Actions() <-chan ActionRequest {
	return s.outgoing
}

func (s *TestSource) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			debug.Print("Shutting down test action generator")
			close(s.outgoing)
			return
		case next := <-s.nextAction:
			s.outgoing <- next
		}
	}
}

// Start starts the action generator
func (s *TestSource) Start(ctx context.Context) error {
	go s.run(ctx)

	// Bit of magic to let the test harness know that things are
	// started up.
	if signalFn, ok := ctx.Value("startSignal").(func()); ok {
		signalFn()
	}
	return nil
}

// NewTestRequest returns a new *TestRequest
func NewTestRequest(archive uint, action llapi.HsmAction, fid *lustre.Fid) *TestRequest {
	return &TestRequest{
		testFid:                fid,
		archive:                archive,
		action:                 action,
		handleProgressReceived: make(chan *TestProgressUpdate),
		handleEndReceived:      make(SignalChan),
	}
}

// Begin returns a new test handle
func (r *TestRequest) Begin(flags int, isError bool) (ActionHandle, error) {
	return &testHandle{
		req:              r,
		fid:              r.testFid,
		progressReceived: r.handleProgressReceived,
		endReceived:      r.handleEndReceived,
	}, nil
}

// FailImmediately immediately fails the request
func (r *TestRequest) FailImmediately(errval int) {
	return
}

// ArchiveID returns the backend archive number
func (r *TestRequest) ArchiveID() uint {
	return r.archive
}

func (r *TestRequest) String() string {
	return fmt.Sprintf("Test Request: %s", r.Action())
}

// Action returns the HSM action type
func (r *TestRequest) Action() llapi.HsmAction {
	return r.action
}

// Test-only methods for TestRequest follow

// ProgressUpdates returns a channel through which progress updates
// may be received.
func (r *TestRequest) ProgressUpdates() chan *TestProgressUpdate {
	return r.handleProgressReceived
}

// Finished returns a channel on which waiters may block until the
// request is finished.
func (r *TestRequest) Finished() SignalChan {
	return r.handleEndReceived
}

func (h *testHandle) Progress(offset, length, total uint64, flags int) error {
	h.progressReceived <- &TestProgressUpdate{
		Offset: offset,
		Length: length,
		Total:  total,
	}
	return nil
}

func (h *testHandle) End(offset, length uint64, flags int, errval int) error {
	close(h.progressReceived)
	close(h.endReceived)
	return nil
}

func (h *testHandle) Action() llapi.HsmAction {
	return h.req.Action()
}

func (h *testHandle) Fid() *lustre.Fid {
	return h.fid
}

func (h *testHandle) Cookie() uint64 {
	return 0
}

func (h *testHandle) DataFid() (*lustre.Fid, error) {
	return h.fid, nil
}

func (h *testHandle) Fd() (int, error) {
	return 0, nil
}

func (h *testHandle) Offset() uint64 {
	return 0
}

func (h *testHandle) ArchiveID() uint {
	return h.req.ArchiveID()
}

func (h *testHandle) Length() uint64 {
	return 0
}

func (h *testHandle) String() string {
	return ""
}

func (h *testHandle) Data() []byte {
	return nil
}