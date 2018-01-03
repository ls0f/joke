package controllers

import (
	"github.com/astaxie/beego"
	cluster2 "github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/redis"
	"strings"
	"sync"
	"time"
)

const (
	Ping = time.Second * 10
)

var (
	cluster *cluster2.Cluster
	lock    sync.Mutex
)

func keepAlive(cluster *cluster2.Cluster, interval time.Duration) {
	for {
		<-time.After(interval)
		// redis cluster heart beat.
		clients, err := cluster.GetEvery()
		if err != nil {
			beego.BeeLogger.Error(err.Error())
			// unable to do redis heart beat, log the error.
			continue
		}
		for _, c := range clients {
			if err := c.Cmd("PING").Err; err != nil {
				beego.BeeLogger.Error(err.Error())
				// unable to keep redis conn alive, log the error.
				c.Close()
				continue
			}
			cluster.Put(c)
		}
		// finished redis heart beat.
	}
}

func newCluster() {
	lock.Lock()
	defer lock.Unlock()
	if cluster != nil {
		return
	}
	var err error
	cluster, err = cluster2.NewWithOpts(cluster2.Opts{
		Addr: beego.AppConfig.String("redisaddr"),
		Dialer: func(network, addr string) (cli *redis.Client, err error) {
			conn, err := redis.DialTimeout(network, addr, 2*time.Second)
			if err != nil {
				return nil, err
			}
			password := beego.AppConfig.String("redispassword")
			if password != "" {
				if conn.Cmd("AUTH", password).Err != nil {
					beego.BeeLogger.Error(err.Error())
					return nil, err
				}
			}
			return conn, nil

		},
	})
	if err != nil {
		beego.BeeLogger.Error(err.Error())
		if cluster == nil {
			panic(err)
		}
	}

	go keepAlive(cluster, Ping)
}

type Host struct {
	Domain string `form:"domain"`
	IP     string `form:"ip"`
}

type DNSController struct {
	beego.Controller
}

// http basic auth
// init redis connect
func (c *DNSController) Prepare() {
	CheckAuth(c.Ctx)
	newCluster()
}

func (c *DNSController) Get() {
	var HostsRecord = make(map[string]string)
	bindkey := beego.AppConfig.String("bindkey")
	rsp := cluster.Cmd("HGETALL", bindkey)
	HostsRecord, err := rsp.Map()
	if err != nil {
		beego.BeeLogger.Error(rsp.Err.Error())
	}
	c.Data["Redis"] = beego.AppConfig.String("redisaddr")
	c.Data["Hosts"] = HostsRecord
	c.Data["Version"] = beego.AppConfig.String("version")
	c.Layout = "layout.html"
	c.TplName = "dns.html"
}

func (c *DNSController) Post() {
	h := new(Host)
	if err := c.ParseForm(h); err != nil {
		c.Ctx.Abort(400, "Invalid post data")
		return
	}
	if h.Domain == "" || h.IP == "" {
		c.Ctx.Abort(400, "Both domain and ip needed")
		return
	}
	bindkey := beego.AppConfig.String("bindkey")

	if err := cluster.Cmd("HSET", bindkey, strings.ToLower(h.Domain), []byte(h.IP)).Err; err != nil {
		c.Ctx.Abort(500, "Save hosts record failed")
		beego.BeeLogger.Error(err.Error())
		return
	}
	beego.BeeLogger.Info("Insert [%s:%s] into redis", strings.ToLower(h.Domain), h.IP)
	c.Layout = "layout.html"
	c.TplName = "dns.html"

}

type DNSDelController struct {
	beego.Controller
}

func (c *DNSDelController) Prepare() {
	CheckAuth(c.Ctx)
	newCluster()
}

func (c *DNSDelController) Post() {
	h := new(Host)
	if err := c.ParseForm(h); err != nil {
		c.Ctx.Abort(400, "Invalid post data")
		return
	}
	bindkey := beego.AppConfig.String("bindkey")
	if err := cluster.Cmd("HDEL", bindkey, h.Domain).Err; err != nil {
		c.Ctx.Abort(500, "Delete hosts record failed")
		beego.BeeLogger.Error(err.Error())
		return
	}
	beego.BeeLogger.Info("Delete [%s:%s] from redis", h.Domain, h.IP)
	c.Layout = "layout.html"
	c.TplName = "dns.html"

}
