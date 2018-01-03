package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/ls0f/joke/controllers"
	"os"
)

/*
Joke: The web console of godns
*/

var (
	GitTag  = "dev"
	Version = "0.1.1"
	Build   = "2017-01-01"
)

func main() {
	version := flag.Bool("v", false, "version")
	conf := flag.String("c", "conf/app.conf", "config")
	static := flag.String("s", "static", "static dir")
	flag.Parse()
	if *version {
		fmt.Fprintf(os.Stdout, "GitTag: %s\n", GitTag)
		fmt.Fprintf(os.Stdout, "Version: %s\n", Version)
		fmt.Fprintf(os.Stdout, "Build: %s\n", Build)
		return
	}
	beego.LoadAppConfig("ini", *conf)
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/dns", &controllers.DNSController{})
	beego.Router("/dns/del", &controllers.DNSDelController{})
	beego.SetStaticPath("/static", *static)
	beego.AppConfig.Set("version", fmt.Sprintf("%s build at %s", Version, Build))
	beego.Run()
}


