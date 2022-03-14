package KVStore

import "time"

//StoreRequest struct sent to the store guardian. Includes a response channel as well as a done channel, which can be closed to cancel the request
type StoreRequest struct {
	command         string
	data            StoreData
	responseChannel chan StoreResponse
	doneChannel     chan struct{}
}

// SendData having this as a method allows it to run in its own goroutine and takes care of the shutdown case which doesn't have a channel
func (s StoreRequest) SendData(data StoreResponse) bool {
	if s.responseChannel != nil {
		defer close(s.responseChannel)
		select {
		case s.responseChannel <- data:
			return true
		case <-s.doneChannel: //requester is no longer waiting for the response
			return false
		}
	} else {
		return false
	}
}

//StoreResponse format of data expected as a response to a store request
type StoreResponse struct {
	json  []byte
	value string
	err   error
}

//StoreData format of data expected as in a store request
type StoreData struct {
	key   string
	user  string
	value string
}

func MakeRequest(request StoreRequest) StoreResponse {
	request.responseChannel = make(chan StoreResponse)
	request.doneChannel = make(chan struct{})
	defer close(request.doneChannel)
	select {
	case <-ShutdownChannel:
		return StoreResponse{err: ErrShutdown}
	case StoreChannel <- request:
		return <-request.responseChannel //Don't want to select on shutdown here because data return happens in its own go routine, may still be data waiting after guardian has shut down
	}
}

//MonitorRoutine is a struct that wraps around an instance of a monitor routine, storing the heartbeat, done channel and kill channel
//this struct means that each time the monitor routine is started it gets a new killChan, so we can be sure that we are killing one and only one instance
//the old way without using this struct meant that a monitor instance that did not receive the kill signal before a new instance was started would not be killed since the two instances had the same killchan
type MonitorRoutine struct {
	HeartBeat chan struct{}
	DoneChan  chan struct{}
	killChan  chan struct{}
}

func (m *MonitorRoutine) Kill() {
	close(m.killChan)
}

func NewMonitorRoutine(heartRate time.Duration) *MonitorRoutine {
	out := &MonitorRoutine{
		killChan: make(chan struct{}),
	}
	out.HeartBeat, out.DoneChan = monitor(heartRate, out.killChan)
	return out
}

func monitor(heartBeat time.Duration, killChan chan struct{}) (heartBeatChan chan struct{}, doneChan chan struct{}) {
	pulse := time.Tick(heartBeat)
	heartBeatChan = make(chan struct{})
	doneChan = make(chan struct{})
	go func() {
		defer close(doneChan)
		var response StoreResponse
		var storeRequest StoreRequest
		var storeOpen bool
	monitorLoop:
		for {
			select {
			case <-pulse: //send a heartbeat to let the steward know we're still alive
				heartBeatChan <- struct{}{}
				continue monitorLoop
			case <-killChan: //monitor has been killed by the steward
				break monitorLoop
			case storeRequest, storeOpen = <-StoreChannel:
				if !storeOpen { //store channel has been closed, cannot continue to monitor
					break monitorLoop
				}
			}

			select {
			case <-storeRequest.doneChannel: //check if this request is still needed
				continue monitorLoop
			default:
			}

			switch storeRequest.command {
			case LookupString:
				value, err := directLookupValue(storeRequest.data.key, storeRequest.data.user)
				response = StoreResponse{
					value: value,
					err:   err,
				}
			case PutString:
				err := directPutValue(storeRequest.data.key, storeRequest.data.user, storeRequest.data.value)
				response = StoreResponse{
					err: err,
				}
			case DeleteString:
				err := directDelete(storeRequest.data.key, storeRequest.data.user)
				response = StoreResponse{
					err: err,
				}
			case ListString:
				var json []byte
				var err error
				if storeRequest.data.key == "" {
					json, err = directListStore()
				} else {
					json, err = directListKey(storeRequest.data.key)
				}
				response = StoreResponse{
					json: json,
					err:  err,
				}
			case ShutdownString:
				break monitorLoop //this should cause the store guardian to complete and exit
			default:
				response = StoreResponse{
					err: ErrBadRequest,
				}
			}
			go storeRequest.SendData(response) //this is run on a new goroutine to prevent the store guardian getting stock waiting to send a response
		}
	}()
	return heartBeatChan, doneChan
}

//ListenForStoreRequests is the main monitoring routine. Waits for requests to be made and then calls the direct methods
func ListenForStoreRequests(timeout time.Duration) (doneChan chan struct{}) {
	doneChan = make(chan struct{})
	var monitorInstance *MonitorRoutine
	go func() {
		defer close(doneChan)
		monitorInstance = NewMonitorRoutine(timeout / 5) //set the monitors heartrate at 1/5 the timeout so as not to have uneccessary restarts
	monitorLoop:
		for {
			select {
			case <-monitorInstance.HeartBeat: //everything is OK, monitor is still doing its thing
				continue monitorLoop

			case <-doneWithTimeout(monitorInstance.DoneChan, timeout): //either monitor has crashed or hasn't checked in. Either way we need to check in on it
				close(monitorInstance.killChan) //monitor is about to be restarted so let's make sure the old instance is definitely dead
				select {
				case <-ShutdownChannel: //a shutdown has been initialised, so everything is fine, proceed to kill the steward
					break monitorLoop
				default:
				}

				request, storeOpen := <-StoreChannel //need to see if crash has caused the store channel to close
				if storeOpen {
					StoreChannel <- request //channel is still open, but we need to put the request we stole back in
				} else { //store channel has unexpectedly closed, reopen it (note, some requests will have been lost)
					StoreChannel = make(chan StoreRequest, BufferSize)
				}
				monitorInstance = NewMonitorRoutine(timeout / 5) //this restarts the monitor
			}
		}
	}()
	return doneChan
}
