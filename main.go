package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

var (
	currentProcess *exec.Cmd
	cancelCurrent  context.CancelFunc
	processLock    sync.Mutex

	currentLive *URL

	GlobalQueue *Queue
	limiter     = rate.NewLimiter(1, 5)
)

type URL struct {
	Url       string `json:"url"`
	Reachable bool   `json:"reachable"`
}

type Queue struct {
	q  []*URL
	mu sync.Mutex
}

func Init() {
	GlobalQueue = NewQueue()
}

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewURL(url string) *URL {
	return &URL{
		Url:       url,
		Reachable: isReachable(url),
	}
}

func NewQueue() *Queue {
	return &Queue{
		q: make([]*URL, 0),
	}
}

func isReachable(url string) bool {
	client := &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(url)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return res.StatusCode >= 200 && res.StatusCode <= 299
}

func (q *Queue) Enque(url *URL) {
	q.q = append(q.q, url)
}

func (q *Queue) Dequeue() *URL {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.q) == 0 {
		fmt.Print("Queue is empty")
		return nil
	}

	first := q.q[0]
	q.q = q.q[1:]
	return first
}

func startStreamlink(ctx context.Context, url *URL) {
	processLock.Lock()
	defer processLock.Unlock()

	if currentProcess != nil {
		cancelCurrent()
		currentProcess.Process.Kill()
		currentProcess = nil
	}

	cmd := exec.Command("streamlink", "--player-external-http", "--player-external-http-port", "8801", url.Url, "best")
	currentProcess = cmd

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	cancelCurrent = cancel

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting Streamlink:", err)
		return
	}

	currentLive = url

	go func() {
		select {
		case <-ctx.Done():
			if err := cmd.Process.Kill(); err != nil {
				fmt.Println("Failed to kill Streamlink process:", err)
			}
			fmt.Println("Streamlink stopped")
			currentLive = nil
		case <-func() chan struct{} {
			done := make(chan struct{})
			go func() {
				cmd.Wait()
				close(done)
			}()
			return done
		}():
			fmt.Println("Streamlink exited naturally")
			currentLive = nil
		}

		processLock.Lock()
		currentProcess = nil
		processLock.Unlock()
	}()
}
func GoNext(w http.ResponseWriter, r *http.Request) {
	ProcessItem()
}

func stopCurrentStreamlink() {
	processLock.Lock()
	defer processLock.Unlock()

	if currentProcess != nil {
		if cancelCurrent != nil {
			cancelCurrent()
		}
		currentProcess.Process.Kill()
		currentProcess = nil
		fmt.Println("Current Streamlink process stopped")
		currentLive = nil
	}
}

func StopStream(w http.ResponseWriter, r *http.Request) {
	stopCurrentStreamlink()
}

func GetLiveVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if currentLive != nil {
		json.NewEncoder(w).Encode(currentLive.Url)
	} else {
		json.NewEncoder(w).Encode("None")
	}
}

func ProcessItem() {
	stopCurrentStreamlink()

	url := GlobalQueue.Dequeue()
	if url != nil {
		go func(url *URL) {
			if url.Reachable {
				ctx, cancel := context.WithCancel(context.Background())
				cancelCurrent = cancel
				startStreamlink(ctx, url)
			}
		}(url)
	}
}

func (q *Queue) PrintQueue(w http.ResponseWriter) {
	for i := range q.q {
		fmt.Fprintf(w, "URL: %s\nReachable: %t\n", q.q[i].Url, q.q[i].Reachable)
	}
}

// DEBUG
func PrintElems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GlobalQueue.q)
}

func main() {
	Init()
	Router()
}

func Router() {
	http.Handle("/streamurl", rateLimit(http.HandlerFunc(ServeMain)))
	http.Handle("/addstreamurl", rateLimit(http.HandlerFunc(AddQueue)))
	http.Handle("/print", rateLimit(http.HandlerFunc(PrintElems)))
	http.Handle("/gonext", rateLimit(http.HandlerFunc(GoNext)))
	http.Handle("/stop", rateLimit(http.HandlerFunc(StopStream)))
	http.Handle("/getCurrent", rateLimit(http.HandlerFunc(GetLiveVideo)))

	fmt.Println("Running Server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Print("Error: ", err)
	}
}

func AddQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error: wrong endpoint. This one is for posting streamURL only", 422)
	}

	var url *URL

	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil {
		http.Error(w, "Wrong Format", 400)
		return
	}

	url.Reachable = isReachable(url.Url)
	if url.Reachable == false {
		http.Error(w, "Domain not reachable", 404)
		return
	}
	GlobalQueue.Enque(url)
}

func ServeMain(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Error: wrong endpoint. This one is for streamURL-Page only", 422)
	}
	http.ServeFile(w, r, "index.html")
}
