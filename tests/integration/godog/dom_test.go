package suite

//
//import (
//	"context"
//	"flag"
//	"fmt"
//	"os"
//	"regexp"
//	"testing"
//
//	"github.com/cucumber/godog"
//	"github.com/cucumber/godog/colors"
//)
//
//var opts = godog.Options{
//	Output:      colors.Colored(os.Stdout),
//	Concurrency: 4,
//}
//
//var stepStore map[string]interface{}
//
//func init() {
//	godog.BindFlags("godog.", flag.CommandLine, &opts)
//
//	stepStore = make(map[string]interface{})
//	stepStore[`^there are (\d+) godogs$`] = thereAreGodogs
//	stepStore[`^I eat (\d+)$`] = iEat
//	stepStore[`^there should be (\d+) remaining$`] = thereShouldBeRemaining
//	stepStore[`^there should be none remaining$`] = thereShouldBeNoneRemaining
//}
//
//func TestFeatures(t *testing.T) {
//	o := opts
//	o.TestingT = t
//
//	status := godog.TestSuite{
//		Name:                 "godogs",
//		Options:              &o,
//		TestSuiteInitializer: InitializeTestSuite,
//		ScenarioInitializer:  InitializeScenario,
//	}.Run()
//
//	if status == 2 {
//		t.SkipNow()
//	}
//
//	if status != 0 {
//		t.Fatalf("zero status code expected, %d received", status)
//	}
//}
//
//type godogsCtxKey struct{}
//
//func godogsToContext(ctx context.Context, g Godogs) context.Context {
//	return context.WithValue(ctx, godogsCtxKey{}, &g)
//}
//
//func godogsFromContext(ctx context.Context) *Godogs {
//	g, _ := ctx.Value(godogsCtxKey{}).(*Godogs)
//
//	return g
//}
//
//// Concurrent execution of scenarios may lead to race conditions on shared resources.
//// Use context to maintain data separation and avoid data races.
//
//// Step definition can optionally receive context as a first argument.
//
//func thereAreGodogsByDom(ctx context.Context, available int) {
//	godogsFromContext(ctx).Add(available)
//}
//
//func thereAreGodogs(ctx context.Context, available int) {
//	godogsFromContext(ctx).Add(available)
//}
//
//// Step definition can return error, context, context and error, or nothing.
//
//func iEat(ctx context.Context, num int) error {
//	return godogsFromContext(ctx).Eat(num)
//}
//
//func thereShouldBeRemaining(ctx context.Context, remaining int) error {
//	available := godogsFromContext(ctx).Available()
//	if available != remaining {
//		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, available)
//	}
//
//	return nil
//}
//
//func thereShouldBeNoneRemaining(ctx context.Context) error {
//	return thereShouldBeRemaining(ctx, 0)
//}
//
//func InitializeTestSuite(ctx *godog.TestSuiteContext) {
//	ctx.BeforeSuite(func() { fmt.Println("Get the party started!") })
//}
//
//func InitializeScenario(ctx *godog.ScenarioContext) {
//	// better way to do this???
//	ctx.Step(`^there are (\d+) godogs$`, thereAreGodogsByDom)
//
//	ctx.Before(func(ctx2 context.Context, sc *godog.Scenario) (context.Context, error) {
//		for _, step := range sc.Steps {
//			for stepRegEx, fn := range stepStore {
//				matched, err := regexp.MatchString(stepRegEx, step.Text)
//				if err != nil {
//					panic(err)
//				}
//				if matched {
//					ctx.Step(stepRegEx, fn)
//				}
//			}
//		}
//
//		// Add initial godogs to context.
//		return godogsToContext(ctx2, 0), nil
//	})
//}
