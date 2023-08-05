package main

import (
	"fmt"
	"gofaster"
	"gofaster/fspool"
	"gofaster/log"
	"gofaster/token"
	"net/http"
	"sync"
	"time"
)

type Test struct {
	id int
}

func main() {
	engine := gofaster.Default()
	g := engine.Group("user")
	//jwtHandler := token.JwtHandler{Key: []byte("123456")}
	g.Use(gofaster.Logging, gofaster.Recovery)
	//g.Any("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "any hello")
	//
	//})
	//g.Get("/hello", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "get any hello")
	//})
	g.Use(func(next gofaster.HandlerFunc) gofaster.HandlerFunc {
		return func(ctx *gofaster.Context) {
			//fmt.Println("pre handler")
			next(ctx)
			//fmt.Println("post handler")
		}
	})
	var t *Test
	g.Get("/hello/user", func(ctx *gofaster.Context) {
		logger := ctx.E.Logger
		//err := fserror.Default()
		//err.Result(func(fsError *fserror.FsError) {
		//	logger.Error(fsError.Error())
		//})
		//err.Put(errors.New("hello"))

		logger.SetPath("./logger")
		t.id = 2
		logger.LogFileSize = 1 << 10 //1k
		//logger.Formatter = logger.JsonFormatter{}
		logger.Info("12")
		logger.Error("123")
		logger.WithFields(log.Fields{"name": "ccfuncy", "id": 2}).Debug("321")
		ctx.Template("login.html", 12)
	})
	g.Get("/get", func(ctx *gofaster.Context) {
		fmt.Println("run")
		fmt.Fprint(ctx.W, "get hello")
	})
	g.Post("/post", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "post hello")
	})
	//g.Get("/use/:id", func(ctx *gofaster.Context) {
	//	fmt.Fprint(ctx.W, "/use/:id get hello")
	//})
	g.Get("/use/*/get:id", func(ctx *gofaster.Context) {
		fmt.Fprint(ctx.W, "/use/*/get:id get hello")
	})

	g.Get("/html", func(ctx *gofaster.Context) {
		ctx.HTML(http.StatusOK, "<h1>Hello</h1>")
	})
	g.Get("/json", func(ctx *gofaster.Context) {
		type User struct {
			Id   int64
			Name string
		}
		model := User{
			Id:   12,
			Name: "12",
		}
		ctx.JSON(http.StatusCreated, &model)
	})
	g.Get("/htmlTemplate", func(ctx *gofaster.Context) {
		ctx.Template("login.html", "")
		//print(err.Error())
	})
	engine.LoadTemplate("tpl/*.html")
	g.Get("/template", func(ctx *gofaster.Context) {
		ctx.Template("index.html", "")
		//print(err.Error())
	})
	pool, _ := fspool.NewPool(1000)
	g.Get("/pool", func(ctx *gofaster.Context) {
		var wg sync.WaitGroup
		now := time.Now()
		wg.Add(5)
		pool.Submit(func() {
			fmt.Println("111111")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		pool.Submit(func() {
			fmt.Println("222222")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		pool.Submit(func() {
			fmt.Println("3333333")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		pool.Submit(func() {
			fmt.Println("4444444")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		pool.Submit(func() {
			fmt.Println("555555")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		wg.Wait()
		fmt.Println(time.Now().Sub(now).Truncate(time.Second))
		ctx.JSON(http.StatusOK, "success")
	})
	g.Get("/login", func(ctx *gofaster.Context) {
		jwt := token.JwtHandler{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = time.Minute * 10
		jwt.Authenticator = func(ctx *gofaster.Context) (map[string]any, error) {
			m := make(map[string]any)
			m["user"] = 1
			return m, nil
		}
		jwtResponse, err := jwt.LoginHandler(ctx)
		if err != nil {
			ctx.E.Logger.Error(err)
			ctx.JSON(http.StatusOK, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, jwtResponse)
	})
	g.Get("/refresh", func(ctx *gofaster.Context) {
		jwt := token.JwtHandler{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = time.Minute * 10
		jwt.RefreshTimeOut = time.Minute * 20
		jwt.RefreshKey = "blog_refresh_token"
		ctx.Set(jwt.RefreshKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODk2NjQ5NjAsImlhdCI6MTY4OTY2NDM2MCwidXNlciI6MX0.1n4zDi2b8ocAVtFcdVitdEOqFFGGOvCZE1rrRegrClk")
		jwtResponse, err := jwt.RefreshHandler(ctx)
		if err != nil {
			ctx.E.Logger.Error(err)
			ctx.JSON(http.StatusOK, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, jwtResponse)
	})
	engine.Run(":9001")
}
