package gofaster

import (
	"context"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

// 限流中间件
func Limit(limit, cap int) MiddlewareFunc {
	limiter := rate.NewLimiter(rate.Limit(limit), cap)
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			//实现限流
			con, cancelFunc := context.WithTimeout(context.Background(), time.Duration(1)*time.Second)
			defer cancelFunc()
			err := limiter.WaitN(con, 1)
			if err != nil {
				ctx.String(http.StatusForbidden, "限流了")
				return
			}
			next(ctx)

		}
	}
}
