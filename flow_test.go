package flow

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFlow_SetConcurrencyLevel(t *testing.T) {
	Convey("Given flow instance", t, func() {
		f := New()

		Convey("When setting concurrency level", func() {
			f.SetConcurrencyLevel(5)

			Convey("Then concurrency level should be set value", func() {
				So(f.concurrencyLevel, ShouldEqual, 5)

			})
		})
	})
}
func TestFlow_Serial(t *testing.T) {
	Convey("Given flow instance and funcs", t, func() {
		f := New()

		var counter int

		baseF := func() error {
			counter++
			return nil
		}

		funcs := []Func{
			WrapFunc(baseF),
			WrapFunc(baseF),
			WrapFunc(baseF),
		}

		ctx := context.Background()

		Convey("When executes sequentially", func() {
			err := f.Serial(funcs...)(ctx)

			Convey("Then all funcs should be executed", func() {
				So(err, ShouldBeNil)
				So(counter, ShouldEqual, len(funcs))

			})
		})
	})

	Convey("Given flow instance and errored func", t, func() {
		f := New()

		var counter int

		baseF := func() error {
			counter++
			return nil
		}

		funcs := []Func{
			WrapFunc(baseF),
			func(ctx context.Context) error {
				counter++
				return errors.New("This is error")
			},
			WrapFunc(baseF),
		}

		ctx := context.Background()

		Convey("When executing sequentially", func() {
			err := f.Serial(funcs...)(ctx)

			Convey("Then error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "This is error")

			})
		})
	})
}
func TestFlow_Parallel(t *testing.T) {
	Convey("Given flow instance and funcs", t, func() {
		f := New()

		res := make(chan int, 3)

		baseF := func() error {
			res <- 1
			return nil
		}

		funcs := []Func{
			WrapFunc(baseF),
			WrapFunc(baseF),
			WrapFunc(baseF),
		}

		ctx := context.Background()

		Convey("When executing parallelly", func() {
			err := f.Parallel(funcs...)(ctx)

			var counter int
			for i := 0; i < len(funcs); i++ {
				select {
				case <-res:
					counter++
				default:
				}
			}

			Convey("Then all funcs should be executed", func() {
				So(err, ShouldBeNil)
				So(counter, ShouldEqual, len(funcs))

			})
		})
	})

	Convey("Given flow instance and errored func", t, func() {
		f := New()

		res := make(chan int, 3)

		baseF := func() error {
			res <- 1
			return nil
		}

		funcs := []Func{
			WrapFunc(baseF),
			func(ctx context.Context) error {
				return errors.New("This is error")
			},
			WrapFunc(baseF),
		}

		ctx := context.Background()

		Convey("When executing parallelly", func() {
			err := f.Parallel(funcs...)(ctx)

			Convey("Then error should be returned", func() {
				So(err, ShouldNotBeNil)
				close(res)
				var counter int
				for i := range res {
					counter += i
				}
				So(err.Error(), ShouldEqual, "This is error")

			})
		})
	})
}
