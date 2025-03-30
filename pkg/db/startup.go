package db

import (
	"os",
	"log",
	"bytes",
	"sync",
	"strings",
	"os/exec",
	"syscall",
	"strconv"

	"github.com/fuxxcss/redi2fuxx/pkg/fuxx"
)

const (

)

// export 
func StartUp(target,tool string) *Shm,error {

	var path,port string

	// Fuxx Target (redis, keydb, redis-stack)
	t,ok := Targets[target]
	if ok {

		// path, port
		path = t[Path]
		port = t[Port]
		
		redi := SingleRedi(port)
		alive := redi.CheckAlive()

		// need to startup
		if !alive {
			return startupCore(path,port,tool),nil

		// already startup
		}else {
			return nil,errors.New("Already StartUp.")
		}
		
	// target not support 
	}else {
		log.Fatalf("err: %v is not support\n",target)
	}

}

// static
func startupCore(path,port,tool string) *Shm{

	// cannot find path
	_,err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v",path,err)
	}

	// set ENV_DEBUG get map size
	var stdout bytes.Buffer
	os.Setenv(utils.Tools[tool][TOOLS_ENV_DEBUG],"1")

	debugProc := exec.Command(path)
	debugProc.Stdout = &stdout
	err = debugProc.Run()

	// cannot run path
	if err != nil {
		log.Fatalf("err: %v %v\n",path,err)
	}

	// loop stdout
	log.Println("[*] Loop Get Debug Size.")
	for {
		if strings.Contains(string(stdout),utils.Tools[tool][TOOLS_ENV_DEBUG_SIZE]){
			break
		}
	}
	debugProc.Process.Kill()

	// get debug size
	index := strings.Index(string(stdout),fuxx.AFL_DEBUG_SIZE)
	shmsize := ""
	for char := stdout[index] ; char != ',' {
		if char >= '0' && char <= '9' {
			shmsize += char
		}
	}
	
	// startup shm
	shm := SingleShm(shmsize)

	// clean up shm
	shm.CleanUp()

	// startup db
	// DB ENVs
	os.Setenv(utils.Tools[tool][TOOLS_ENV_DEBUG],"0")
	os.Setenv(utils.Tools[tool][TOOLS_ENV_MAX_SIZE],shm.ShmSize)
	os.Setenv(utils.Tools[tool][TOOLS_ENV_SHM_ID],shm.ShmID)
	// DB args
	args := []string {
		// port
		RediSep + " " + port,
		// daemon
		RediDeamon,
	}
	rediProc = exec.Command(path,args...)
	err := rediProc.Run()
	
	redi := SingleRedi(port)
	alive := redi.CheckAlive()

	// db failed
	if err != nil {
		log.Fatalf("err: db %v\n",err)
	}
	if !alive {
		log.Fatalln("err: redi failed.")
	}

	// db succeed
	redi.Proc = rediProc
	log.Printf("[*] DB %v StartUp.\n",path)
	
}

func ShutDown() {

	// kill redis
	redi := SingleRedi(nil)
	redi.Proc.Process.Kill()

	// close shm
	shm := SingleShm(nil)
	shm.Close()
}