package runner

import (
	"fmt"
	"os"
	"strconv"
	//"io/ioutil"
	//"strings"
	"io"
	"bufio"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger"
	"github.com/remeh/sizedwaitgroup"
	"ktbs.dev/ssb/pkg/ssb"
)

type Job struct {
	host string
	port int
	username string
	password string
}

// New execute bruteforces
func New(opt *Options) {
	opt.showInfo()

	defer opt.Close()
	jobs := make(chan Job)
	cur := opt.concurrent
	swg := sizedwaitgroup.New(cur)

	for i := 0; i < cur; i++ {
		swg.Add()
		go func() {
			defer swg.Done()
			for job := range jobs {
				for i := 0; i < opt.retries; i++ {
					if opt.run(job) {
						break
					}
				}
			}
		}()
	}

	for scanner := opt.listScanner(); scanner.Scan(); {
		password := scanner.Text()
		port := opt.port
		user := opt.user
		if opt.hosts != ""{
			//content,_ := ioutil.ReadFile(opt.hosts)
			//host := strings.Split(string(content), "\n")
			fi, err := os.Open(opt.hosts)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			defer fi.Close()

			br := bufio.NewReader(fi)
			for {
				h, _, c := br.ReadLine()
				if c == io.EOF {
					break
				}

				jobs <- Job{string(h), port, user, string(password)}
			}

		}else{
			jobs <- Job{opt.host, port, user, password}
		}
		
		
		
	}

	close(jobs)
	swg.Wait()
	gologger.Infof("Done!")
}

func (opt *Options) run(job Job) bool {
	//fmt.Println(job.host)
	//fmt.Println(job.password)
	cfg := ssb.New(job.username, job.password, opt.timeout)

	con, err := ssb.Connect(job.host, job.port, cfg)
	if err != nil {
		if opt.verbose {
			gologger.Errorf("Failed '%s': %s.", job.password, err.Error())
		}
	}

	if con {
		fmt.Printf("[%s] %s Connected with '%s'.\n", aurora.Green("VLD"), aurora.Magenta(job.host), aurora.Magenta(job.password))

		if opt.file != nil {
			fmt.Fprintf(opt.file, "%s\n", job.password)
		}

		vld = true
	}

	return vld
}

func (opt *Options) showInfo() {
	info := "________________________\n"
	info += "\n :: Username: " + opt.user
	info += "\n :: Hostname: " + opt.host
	info += "\n :: Port    : " + strconv.Itoa(opt.port)
	info += "\n :: Wordlist: " + opt.wordlist
	info += "\n :: Threads : " + strconv.Itoa(opt.concurrent)
	info += "\n :: Timeout : " + opt.timeout.String()
	info += "\n________________________\n\n"

	fmt.Fprint(os.Stderr, info)
}
