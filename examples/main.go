package main

import (
	"./lib"
	"./src"
	"os"
	"log"
	"net/http"
)
import _ "net/http/pprof"

func main() {

	go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

	lib.SIM_init()
	lib.SIM_platform_init(os.Args[1])

	lib.SIM_function_register("load_balancer", src.LoadBalancer)
	lib.SIM_function_register("server_manager", src.ServerManager)
	lib.SIM_function_register("PCIeFabric_manager", src.PCIeFabricManager)
	lib.SIM_function_register("Cache_manager", src.CacheManager)
	lib.SIM_function_register("disk", src.Disk)
	lib.SIM_function_register("tracer", src.Tracer)

	lib.SIM_launch_application(os.Args[2])

	lib.SIM_run(20.)
}
