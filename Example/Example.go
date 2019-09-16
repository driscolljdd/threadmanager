package main

import (
	"fmt"
	"github.com/driscolljdd/threads"
	"strconv"
	"time"
)

	func main() {

		// Our thread group - an extension of an errorgroup, which is an extension of a waitgroup
		threads := threads.GetManager()

		// Spin up some threads
		for i := 0; i < 5; i++ {

			// The thread ID is a string so can be a useful reference to the thread. Makes it easier to stop them later
			myThread := threads.Add("thread " + strconv.Itoa(i))

			// Having done the work and created the thread, run the actual back end process in the underlying error group
			threads.Run("thread " + strconv.Itoa(i), func() error {

				return RandomTask(myThread)
			})
		}

		// For demonstration purposes, wait three seconds...
		time.Sleep(time.Second * 3)

		// And then demonstrate that we can stop any thread we like at any time with a simple command
		threads.StopThread("thread 2")

		// Send a sample of a message comprised of a struct containing multiple types of information
		myMsg := Msg{ Message: "Hello", SomeUsefulInteger: 4 }
		threads.Send("thread 3", myMsg)

		// The main thread waits for all other threads to complete before ending itself, but a unix kill command will also stop all the threads in our group
		threads.Wait()
	}



	// This represents a worker process you will want to invoke in a very controlled, separate thread
	func RandomTask(myThread threads.Thread) error {

		for {

			// Instead of a select statement to check the context of our thread, it is wrapped up in an easy single function
			if(!myThread.Alive()) {

				// Make clear (again for demonstration purposes) that our thread is quitting
				fmt.Println("Quitting - ", myThread.ID)
				return nil
			}

			if raw := myThread.Read(); raw != nil {

				// Convert the raw message into the struct we have chosen and display the message body
				fmt.Println(myThread.ID, " Got a message: ", raw.(Msg).Message)
			}

			// This represents a repeating workload. It could be either some repetitive task as seen here, or a straight list of tasks. Check myThread.Alive() regularly to respond to requests to stop.
			fmt.Println(myThread.ID, " - Working")

			time.Sleep(time.Second * time.Duration(1))
		}
	}



	// This is a stand in for whatever complex struct you want to use for rich, detailed inter thread communication
	type Msg struct {

		Message string
		SomeUsefulInteger int
	}