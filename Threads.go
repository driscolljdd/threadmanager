package threads

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"golang.org/x/sync/errgroup"
)

	// Constructor - get a thread manager with this function
	func GetManager() threadGroup {

		threadGroup := threadGroup{}

		threadContext := threadContext{}
		threadContext.Setup()

		threadGroup.group, threadGroup.context = errgroup.WithContext(threadContext.Context)
		threadGroup.cancel = threadContext.Cancel
		threadGroup.threads = make(map[string]Thread)

		return threadGroup
	}



	// A holder for the group context and cancel properties - a struct because some setup is required
	type threadContext struct {

		Cancel context.CancelFunc
		Context context.Context
	}

	// Thread context setup - take the time here to capture unix kill signals. Makes our threads, all of them, have a done() context on kill signals
	func (tc *threadContext) Setup() {

		// Set up the context and cancel functions
		tc.Context, tc.Cancel = context.WithCancel(context.Background())

		// Spawn a background thread that will watch for Unix signals (sigint, sigterm)
		go func() {

			// Make a channel to receive unix signals
			signalChannel := make(chan os.Signal, 1)

			// Register for notifications of being asked to quit by unix
			signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

			// Wait for a quit instruction from outside
			<- signalChannel

			// Call the cancel function for our thread context
			tc.Cancel()
		}()

		// Send back the thread which has now completed setup
		return
	}



	// Our thread group - an extension of an errorgroup which is in turn an extension of a waitgroup
	type threadGroup struct {

		group *errgroup.Group
		context context.Context
		cancel context.CancelFunc
		lock sync.RWMutex
		threads map[string]Thread
	}

	// Add a new thread to this group
	func (tg *threadGroup) Add(id string) Thread {

		myThread := Thread{ ID: id, groupContext: tg.context }
		myThread.context, myThread.cancel = context.WithCancel(myThread.groupContext)
		myThread.channel = make(chan interface{}, 1)

		// Add this to our group
		tg.lock.Lock()

			tg.threads[myThread.ID] = myThread

		tg.lock.Unlock()
		return myThread
	}

	func (tg *threadGroup) Send(threadID string, msg interface{}) {

		var thread Thread
		if thread = tg.getThread(threadID); thread.ID != threadID {

			return
		}

		thread.channel <- msg
	}

	// Run a method inside a thread
	func (tg *threadGroup) Run(threadID string, f func() error) {

		// Check we haven't run something under the alias of this thread before - essentially make sure that for every call to this function, there is a prior call to tg.Add()
		var thread Thread
		if thread = tg.getThread(threadID); thread.ID != threadID {

			return
		}

		// Check if we have 'used up' this thread already
		if(thread.isRunning) {

			return
		}

		// Run the thread inside our parent thread group
		thread.isRunning = true
		tg.group.Go(f)

		// Save this back
		tg.lock.Lock()

			tg.threads[threadID] = thread

		tg.lock.Unlock()
	}

	// Wait for all threads in the group to be finished
	func (tg *threadGroup) Wait() error {

		return tg.group.Wait()
	}

	// Stop a particular thread
	func (tg *threadGroup) StopThread(threadID string) {

		var thread Thread
		if thread = tg.getThread(threadID); thread.ID != threadID {

			return
		}

		// Kill the thread
		thread.cancel()
	}

	// Stop all threads in the group
	func (tg *threadGroup) Stop() {

		tg.cancel()
	}

	func (tg *threadGroup) getThread(threadID string) Thread {

		tg.lock.RLock()

			thread, exists := tg.threads[threadID]

		tg.lock.RUnlock()

		if(!exists) {

			return Thread{}
		}

		return thread
	}



	// An individual thread - should respond both to the group context and it's own individual context
	type Thread struct {

		ID string
		groupContext context.Context
		context context.Context
		cancel context.CancelFunc
		isRunning bool
		channel chan interface{}
	}

	func (t *Thread) Alive() bool {

		select {

			case <- t.context.Done():

				return false

			case <- t.groupContext.Done():

				return false

			default:

				return true
		}
	}

	func (t *Thread) Read() interface{} {

		select {

			case msg := <- t.channel:

				return msg

			default:

				return nil
		}
	}